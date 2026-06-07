package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner は sbx コマンドの実行を抽象化する
type Runner struct {
	// DryRun が true の場合、実際のコマンドは実行せずログのみ出力する
	DryRun bool
}

// NewRunner は Runner を作成する
func NewRunner(dryRun bool) *Runner {
	return &Runner{DryRun: dryRun}
}

// sbxExec は sbx コマンドを実行する
func (r *Runner) sbxExec(args ...string) (string, error) {
	cmd := exec.Command("sbx", args...)
	cmd.Stderr = os.Stderr

	if r.DryRun {
		fmt.Printf("[dry-run] sbx %s\n", strings.Join(args, " "))
		return "", nil
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sbx %s の実行に失敗: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CreateParams は sbx create のパラメータ
type CreateParams struct {
	Name         string
	Template     string
	Agent        string
	Path         string
	Clone        bool
	CPUs         int
	Memory       string
	PublishPorts []string // 作成後に公開するポート [[HOST_IP:]HOST_PORT:]SANDBOX_PORT[/PROTOCOL]
}

// Create はサンドボックスを作成する
func (r *Runner) Create(params CreateParams) (string, error) {
	args := []string{"create"}

	if params.Name != "" {
		args = append(args, "--name", params.Name)
	}
	if params.Template != "" {
		args = append(args, "--template", params.Template)
	}
	if params.Clone {
		args = append(args, "--clone")
	}
	if params.CPUs > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%d", params.CPUs))
	}
	if params.Memory != "" {
		args = append(args, "--memory", params.Memory)
	}

	args = append(args, params.Agent, params.Path)
	return r.sbxExec(args...)
}

// PortPublish はサンドボックスのポートを公開する。
// portSpec の形式: [[HOST_IP:]HOST_PORT:]SANDBOX_PORT[/PROTOCOL]
func (r *Runner) PortPublish(sandboxName, portSpec string) error {
	_, err := r.sbxExec("ports", sandboxName, "--publish", portSpec)
	return err
}

// CopyToSandbox はホストのファイル/ディレクトリをサンドボックス内にコピーする。
// followSymlinks が true の場合、シンボリックリンクを実体としてコピーする
func (r *Runner) CopyToSandbox(src, sandboxDst, sandboxName string, followSymlinks bool) error {
	args := []string{"cp"}
	if followSymlinks {
		args = append(args, "-L")
	}
	args = append(args, src, fmt.Sprintf("%s:%s", sandboxName, sandboxDst))
	_, err := r.sbxExec(args...)
	return err
}

// ListOutput は sbx ls の出力の1行を表す
type ListOutput struct {
	Name      string
	Sandbox   string
	Agent     string
	Status    string
	Ports     string
	Workspace string
}

// List はサンドボックス一覧を取得する
func (r *Runner) List() ([]ListOutput, error) {
	out, err := r.sbxExec("ls")
	if err != nil {
		return nil, err
	}

	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	var result []ListOutput

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SANDBOX") {
			continue
		}

		fields := strings.Fields(line)
		// 最低4フィールド必要: name, agent, status, workspace
		if len(fields) < 4 {
			continue
		}

		lo := ListOutput{
			Sandbox:   fields[0],
			Agent:     fields[1],
			Status:    fields[2],
			Ports:     "",
			Workspace: fields[len(fields)-1],
		}
		// PORTS カラムがある場合 (5フィールド以上)
		if len(fields) >= 5 {
			lo.Ports = strings.Join(fields[3:len(fields)-1], " ")
		}
		result = append(result, lo)
	}
	return result, nil
}

// FindByName は名前でサンドボックスを検索する
func (r *Runner) FindByName(name string) (*ListOutput, error) {
	sandboxes, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, sb := range sandboxes {
		if sb.Sandbox == name {
			return &sb, nil
		}
	}
	return nil, nil
}

// Run はサンドボックスにアタッチする
func (r *Runner) Run(name string) error {
	args := []string{"run", name}
	cmd := exec.Command("sbx", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if r.DryRun {
		fmt.Printf("[dry-run] sbx %s\n", strings.Join(args, " "))
		return nil
	}

	return cmd.Run()
}

// RunCommand はサンドボックス内で対話的にコマンドを実行する。
// サンドボックスが停止中の場合は自動で起動される
func (r *Runner) RunCommand(name string, command ...string) error {
	args := []string{"exec", "-i", "-t", name}
	args = append(args, command...)
	cmd := exec.Command("sbx", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if r.DryRun {
		fmt.Printf("[dry-run] sbx %s\n", strings.Join(args, " "))
		return nil
	}

	return cmd.Run()
}

// Start は停止中のサンドボックスを起動する
func (r *Runner) Start(name string) error {
	_, err := r.sbxExec("run", name)
	return err
}

// Stop はサンドボックスを停止する
func (r *Runner) Stop(name string) error {
	_, err := r.sbxExec("stop", name)
	return err
}

// Remove はサンドボックスを強制削除する
func (r *Runner) Remove(name string) error {
	_, err := r.sbxExec("rm", "--force", name)
	return err
}

// Exec はサンドボックス内でコマンドを実行する
func (r *Runner) Exec(name string, command string) (string, error) {
	args := []string{"exec", name}
	args = append(args, strings.Fields(command)...)
	return r.sbxExec(args...)
}

// TemplateList はテンプレート一覧を取得する
func (r *Runner) TemplateList() (string, error) {
	return r.sbxExec("template", "ls")
}

// HasTemplate は指定されたテンプレート名が sbx に存在するか確認する
func (r *Runner) HasTemplate(name string) (bool, error) {
	out, err := r.TemplateList()
	if err != nil {
		return false, err
	}
	// "dbox-go:latest" という名前を "dbox-go" でもマッチさせる
	// sbx template ls の出力: "dbox-go  latest  IMAGE_ID  ..."
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "REPOSITORY") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			repo := fields[0]
			tag := fields[1]
			fullName := repo + ":" + tag
			if repo == name || fullName == name {
				return true, nil
			}
		}
	}
	return false, nil
}

// TemplateRemove は指定したテンプレートを sbx から削除する
func (r *Runner) TemplateRemove(tag string) error {
	_, err := r.sbxExec("template", "rm", tag)
	return err
}

// TemplateSave はDockerイメージをエクスポートし、sbxテンプレートとしてロードする
func (r *Runner) TemplateSave(tag string) error {
	if r.DryRun {
		fmt.Printf("[dry-run] docker save %s | sbx template load\n", tag)
		return nil
	}

	tmpFile, err := os.CreateTemp("", "dbox-template-*.tar")
	if err != nil {
		return fmt.Errorf("一時ファイル作成に失敗: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// docker save -o でエクスポート
	saveArgs := []string{"save", "-o", tmpPath, tag}
	saveCmd := exec.Command("docker", saveArgs...)
	saveCmd.Stderr = os.Stderr
	if err := saveCmd.Run(); err != nil {
		return fmt.Errorf("Dockerイメージ %s のエクスポートに失敗: %w", tag, err)
	}

	// sbx template load でインポート
	_, err = r.sbxExec("template", "load", tmpPath)
	if err != nil {
		return fmt.Errorf("sbxテンプレートのロードに失敗: %w", err)
	}

	fmt.Printf("テンプレート %s を sbx にロードしました\n", tag)
	return nil
}
