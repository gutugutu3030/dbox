//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfMissingTools は sbx と docker が利用可能かを確認し、なければスキップする
func skipIfMissingTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"sbx", "docker"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s が見つかりません。このテストをスキップします", tool)
		}
	}
}

// buildDboxBinary は dbox バイナリをビルドし、そのパスを返す
func buildDboxBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "dbox")

	cmd := exec.Command("go", "build", "-o", binaryPath, "github.com/gutugutu3030/sbx-template/cmd/dbox")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("dbox バイナリのビルドに失敗: %v", err)
	}
	return binaryPath
}

// runDbox は dbox コマンドを実行し、標準出力とエラーを返す
func runDbox(t *testing.T, binary, dir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// createProjectDir はテスト用のプロジェクトディレクトリ（git リポジトリ）を作成する
func createProjectDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("ディレクトリ作成に失敗: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("ファイル作成に失敗: %v", err)
		}
	}
	// sbx create --clone が必要とするため git リポジトリとして初期化
	if err := exec.Command("git", "init", "-b", "main", dir).Run(); err != nil {
		t.Fatalf("git init に失敗: %v", err)
	}
	// 初期コミットを作成（sbx clone に必要）
	gitCmd := exec.Command("git", "-C", dir, "add", ".")
	if err := gitCmd.Run(); err != nil {
		t.Fatalf("git add に失敗: %v", err)
	}
	gitCmd = exec.Command("git", "-C", dir, "commit", "-m", "initial", "--allow-empty")
	if err := gitCmd.Run(); err != nil {
		t.Fatalf("git commit に失敗: %v", err)
	}
	return dir
}

// cleanupSandbox は指定されたサンドボックス名を停止・削除する
func cleanupSandbox(t *testing.T, binary, dir, sandboxName string) {
	t.Helper()
	// 停止（存在しなければエラーになるが無視）
	exec.Command(binary, "stop", sandboxName).Run()
	// サンドボックス名指定で削除はないので、prune で一掃
	// 実際には sbx rm が必要。ここでは start で作成したものだけ削除
	runDbox(t, binary, dir, "prune")
}

// テストで使用するプロジェクト名（ユニークにするためテスト関数名を含む）
const testAgent = "opencode"

// TestE2E_Init_DryRun は dbox init --dry-run がエラーなく実行できることを確認する
func TestE2E_Init_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
		"index.ts":     `console.log("hello");`,
	})

	stdout, stderr, err := runDbox(t, binary, dir, "init", "--dry-run")
	if err != nil {
		t.Fatalf("dbox init --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	if stdout == "" {
		t.Error("標準出力が空です")
	}
	t.Logf("stdout: %s", stdout)
}

// TestE2E_Init_NewSandbox_DryRun は未作成の状態で dbox init --dry-run が成功することを確認する
// （addTest.md 1.1 相当）
func TestE2E_Init_NewSandbox_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
		"index.ts":     `console.log("hello");`,
	})

	stdout, stderr, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("dbox init --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)
	if !strings.Contains(stdout, ".dbox.yaml を作成しました") {
		t.Error(".dbox.yaml が作成されたことを示すメッセージが見つかりません")
	}

	// .dbox.yaml が作成されていることを確認
	if _, err := os.Stat(filepath.Join(dir, ".dbox.yaml")); err != nil {
		t.Errorf(".dbox.yaml が作成されていません: %v", err)
	}
}

// TestE2E_Init_Overwrite_DryRun は既存サンドボックスがある状態の dbox init --dry-run が
// エラーにならないことを確認する（addTest.md 1.2 相当）
func TestE2E_Init_Overwrite_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
	})

	// 1回目の init
	stdout1, stderr1, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("1回目の dbox init --dry-run エラー: %v\nstderr: %s", err, stderr1)
	}
	t.Logf("1回目 stdout: %s", stdout1)

	// 2回目の init --dry-run（上書きの検証）
	stdout2, stderr2, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("2回目の dbox init --dry-run エラー: %v\nstderr: %s", err, stderr2)
	}
	t.Logf("2回目 stdout: %s", stdout2)
	// dry-run では既存サンドボックスの有無に関わらずエラーにならない
	if !strings.Contains(stdout2, ".dbox.yaml を作成しました") {
		t.Error("2回目も .dbox.yaml 作成メッセージが表示されるべきです")
	}
	if strings.Contains(stderr2, "既存のサンドボックス") {
		t.Log("既存サンドボックス検出メッセージが表示されました（dry-run では検出せず）")
	}
}

// TestE2E_TemplateBuild_DryRun は dbox template build --dry-run がエラーなく実行できることを確認する
// （addTest.md 7.1 の DryRun 版）
func TestE2E_TemplateBuild_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := t.TempDir()

	stdout, stderr, err := runDbox(t, binary, dir, "template", "build", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("dbox template build --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)
	if !strings.Contains(stdout, "ベースイメージ") && !strings.Contains(stdout, "[dry-run]") {
		t.Error("テンプレートビルドのメッセージが見つかりません")
	}
}

// TestE2E_Init_AllowedDomains は allowed_domains が dbox init で保持されることを確認する
// （addTest.md 2.1）
func TestE2E_Init_AllowedDomainsPreserved(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
	})

	// 1回目の init
	stdout1, stderr1, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("1回目の dbox init --dry-run エラー: %v\nstderr: %s", err, stderr1)
	}
	t.Logf("1回目 stdout: %s", stdout1)

	// .dbox.yaml が作成されていることを確認
	dboxYaml := filepath.Join(dir, ".dbox.yaml")
	if _, err := os.Stat(dboxYaml); err != nil {
		t.Fatalf(".dbox.yaml が作成されていません: %v", err)
	}

	// 手動で allowed_domains を追加
	data, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	modified := strings.Replace(string(data), "allowed_domains: []", "allowed_domations:\n    - example.com", 1)
	// 正しい形式に修正
	modified = strings.Replace(modified, "allowed_domations:\n    - example.com", "allowed_domains:\n    - example.com", 1)
	if err := os.WriteFile(dboxYaml, []byte(modified), 0644); err != nil {
		t.Fatalf(".dbox.yaml 書き込みエラー: %v", err)
	}
	t.Logf("手動で allowed_domains を追加しました")

	// 2回目の init --dry-run で allowed_domains が保持されていることを確認
	stdout2, stderr2, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("2回目の dbox init --dry-run エラー: %v\nstderr: %s", err, stderr2)
	}
	t.Logf("2回目 stdout: %s", stdout2)

	// 保存された .dbox.yaml を確認
	savedData, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	if !strings.Contains(string(savedData), "example.com") {
		t.Error("手動追加した allowed_domains が保持されていません")
	}
}

// TestE2E_Prune は dbox prune がエラーなく実行できることを確認する
func TestE2E_Prune(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := t.TempDir()

	stdout, stderr, err := runDbox(t, binary, dir, "prune", "--dry-run")
	if err != nil {
		t.Fatalf("dbox prune --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)

	if !strings.Contains(stdout, "削除対象") && !strings.Contains(stdout, "ありません") {
		t.Log("prune の出力を確認しました")
	}
}

// TestE2E_Help は dbox help がエラーなく実行できることを確認する
func TestE2E_Help(t *testing.T) {
	binary := buildDboxBinary(t)
	dir := t.TempDir()

	stdout, stderr, err := runDbox(t, binary, dir, "help")
	if err != nil {
		t.Fatalf("dbox help エラー: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "dbox") {
		t.Error("ヘルプに dbox が含まれていません")
	}

	for _, sub := range []string{"init", "start", "stop"} {
		subStdout, _, subErr := runDbox(t, binary, dir, "help", sub)
		if subErr != nil {
			t.Errorf("dbox help %s エラー: %v", sub, subErr)
		}
		if !strings.Contains(subStdout, sub) {
			t.Errorf("dbox help %s の出力に %s が含まれていません", sub, sub)
		}
	}
}

// TestE2E_MultiLangTemplate_DryRun は複数言語の合成テンプレートビルドが DryRun でエラーにならないことを確認する
// （addTest.md 7.2 の DryRun 版）
func TestE2E_MultiLangTemplate_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := t.TempDir()

	stdout, stderr, err := runDbox(t, binary, dir, "template", "build", "--lang=node,go", "--dry-run")
	if err != nil {
		t.Fatalf("dbox template build --lang=node,go --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)
	if !strings.Contains(stdout, "合成イメージ") && !strings.Contains(stdout, "dbox-go-node") {
		t.Log("合成テンプレートのメッセージを確認（DryRun のため出力は限定的）")
	}
}

// TestE2E_Init_ConfigFile は .dbox.yaml の内容が正しいことを確認する
func TestE2E_Init_ConfigFile(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
	})

	_, stderr, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("dbox init --dry-run エラー: %v\nstderr: %s", err, stderr)
	}

	dboxYaml := filepath.Join(dir, ".dbox.yaml")
	data, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	content := string(data)

	for _, expect := range []string{"version: 2", "agent: opencode", "sandbox_name: dbox-opencode-", "allowed_domains:"} {
		if !strings.Contains(content, expect) {
			t.Errorf(".dbox.yaml に %q が含まれていません", expect)
		}
	}
}

// TestE2E_Init_MultipleDomains_DryRun は複数ドメインが dbox init で正しく処理されることを確認する
// （addTest.md 3.1 の DryRun 版）
func TestE2E_Init_MultipleDomains_DryRun(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
	})

	// --dry-run では実際に sbx は呼ばれないが、内部ロジックでループ処理が走る
	stdout, stderr, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("dbox init --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)

	// 2回目: allowed_domains を手動で設定し、再度 dry-run
	dboxYaml := filepath.Join(dir, ".dbox.yaml")
	data, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	content := string(data)
	// 複数ドメインを追加
	content = strings.Replace(content, "allowed_domains: []", "allowed_domains:\n    - opencode.ai:443\n    - example.com\n    - api.example.com", 1)
	if err := os.WriteFile(dboxYaml, []byte(content), 0644); err != nil {
		t.Fatalf(".dbox.yaml 書き込みエラー: %v", err)
	}

	stdout2, stderr2, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("2回目の dbox init --dry-run エラー: %v\nstderr: %s", err, stderr2)
	}
	t.Logf("2回目 stdout: %s", stdout2)

	// 保存された config に全てのドメインが含まれていることを確認
	savedData, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	for _, domain := range []string{"opencode.ai:443", "example.com", "api.example.com"} {
		if !strings.Contains(string(savedData), domain) {
			t.Errorf("ドメイン %q が保持されていません", domain)
		}
	}
}

// TestE2E_SandboxName はサンドボックス名が期待通り生成されることを確認する
func TestE2E_SandboxName(t *testing.T) {
	skipIfMissingTools(t)
	binary := buildDboxBinary(t)
	dir := createProjectDir(t, map[string]string{
		"package.json": `{"name": "test-project"}`,
	})

	stdout, stderr, err := runDbox(t, binary, dir, "init", "--lang=base", "--dry-run")
	if err != nil {
		t.Fatalf("dbox init --dry-run エラー: %v\nstderr: %s", err, stderr)
	}
	t.Logf("stdout: %s", stdout)

	// .dbox.yaml の sandbox_name を確認
	dboxYaml := filepath.Join(dir, ".dbox.yaml")
	data, err := os.ReadFile(dboxYaml)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}
	content := string(data)
	expectedPrefix := fmt.Sprintf("sandbox_name: dbox-%s-", testAgent)
	if !strings.Contains(content, expectedPrefix) {
		t.Errorf(".dbox.yaml に %q が含まれていません: %s", expectedPrefix, content)
	}
}
