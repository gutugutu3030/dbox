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
	Name     string
	Template string
	Agent    string
	Path     string
	Clone    bool
	CPUs     int
	Memory   string
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
		if len(fields) >= 5 {
			lo := ListOutput{
				Sandbox:   fields[0],
				Agent:     fields[1],
				Status:    fields[2],
				Ports:     fields[3],
				Workspace: strings.Join(fields[4:], " "),
			}
			result = append(result, lo)
		}
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

// Start は停止中のサンドボックスを起動する
func (r *Runner) Start(name string) error {
	_, err := r.sbxExec("start", name)
	return err
}

// Stop はサンドボックスを停止する
func (r *Runner) Stop(name string) error {
	_, err := r.sbxExec("stop", name)
	return err
}

// Exec はサンドボックス内でコマンドを実行する
func (r *Runner) Exec(name string, command string) (string, error) {
	return r.sbxExec("exec", name, command)
}

// TemplateList はテンプレート一覧を取得する
func (r *Runner) TemplateList() (string, error) {
	return r.sbxExec("template", "ls")
}

// TemplateSave はサンドボックスをテンプレートとして保存する
func (r *Runner) TemplateSave(tag string) (string, error) {
	return r.sbxExec("template", "save", tag)
}

// HasTemplate は指定されたテンプレートが存在するか確認する
func (r *Runner) HasTemplate(templateName string) (bool, error) {
	out, err := r.TemplateList()
	if err != nil {
		return false, err
	}
	return strings.Contains(out, templateName), nil
}
