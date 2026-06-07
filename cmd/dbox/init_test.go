package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gutugutu3030/sbx-template/internal/config"
)

// createFile はテスト用のファイルを作成する
func createFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// createDir はテスト用のディレクトリを作成する
func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// TestGenerateSandboxName はサンドボックス名の生成を確認する
func TestGenerateSandboxName(t *testing.T) {
	tests := []struct {
		name  string
		agent string
		dir   string
		want  string
	}{
		{
			name:  "opencode + my-project",
			agent: "opencode",
			dir:   "/home/user/my-project",
			want:  "dbox-opencode-my-project",
		},
		{
			name:  "codex + カレントディレクトリ名",
			agent: "codex",
			dir:   "/workspace/sbx-template",
			want:  "dbox-codex-sbx-template",
		},
		{
			name:  "claude + ディレクトリ名にドット",
			agent: "claude",
			dir:   "/tmp/my.project",
			want:  "dbox-claude-my.project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSandboxName(tt.agent, tt.dir)
			if got != tt.want {
				t.Errorf("generateSandboxName(%q, %q) = %q, want %q", tt.agent, tt.dir, got, tt.want)
			}
		})
	}
}

// TestPublishPorts はポート公開関数がモックを正しく呼ぶことを確認する
func TestPublishPorts(t *testing.T) {
	mock := &MockSandboxOperator{}

	ports := []string{"8080", "3000:8080"}
	if err := publishPorts(mock, "test-sandbox", ports); err != nil {
		t.Fatalf("publishPorts() エラー: %v", err)
	}

	if mock.CallCount("PortPublish") != 2 {
		t.Errorf("PortPublish の呼び出し回数 = %d, want 2", mock.CallCount("PortPublish"))
	}
}

// TestPublishPorts_Empty は空のポートリストで何も起きないことを確認する
func TestPublishPorts_Empty(t *testing.T) {
	mock := &MockSandboxOperator{}

	if err := publishPorts(mock, "test-sandbox", nil); err != nil {
		t.Fatalf("publishPorts(nil) エラー: %v", err)
	}

	if mock.CallCount("PortPublish") != 0 {
		t.Errorf("PortPublish の呼び出し回数 = %d, want 0", mock.CallCount("PortPublish"))
	}
}

// TestApplyNetworkPolicies_WithDomains はドメインがある場合に PolicyAllowNetwork が呼ばれることを確認する
func TestApplyNetworkPolicies_WithDomains(t *testing.T) {
	mock := &MockSandboxOperator{}
	domains := []string{"opencode.ai:443", "example.com"}
	if err := applyNetworkPolicies(mock, "test-sandbox", domains); err != nil {
		t.Fatalf("applyNetworkPolicies() エラー: %v", err)
	}

	if !mock.AssertCalled("PolicyAllowNetwork") {
		t.Error("PolicyAllowNetwork が呼ばれていません")
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("呼び出し回数 = %d, want 1", len(mock.Calls))
	}
	gotDomains, ok := mock.Calls[0].Args[1].([]string)
	if !ok {
		t.Fatal("PolicyAllowNetwork の第2引数が []string ではありません")
	}
	if len(gotDomains) != 2 || gotDomains[0] != "opencode.ai:443" || gotDomains[1] != "example.com" {
		t.Errorf("PolicyAllowNetwork の引数 = %v, want [opencode.ai:443 example.com]", gotDomains)
	}
}

// TestApplyNetworkPolicies_EmptyDomains は空ドメインで何も起きないことを確認する
func TestApplyNetworkPolicies_EmptyDomains(t *testing.T) {
	mock := &MockSandboxOperator{}
	if err := applyNetworkPolicies(mock, "test-sandbox", nil); err != nil {
		t.Fatalf("applyNetworkPolicies(nil) エラー: %v", err)
	}

	if mock.AssertCalled("PolicyAllowNetwork") {
		t.Error("空ドメインで PolicyAllowNetwork が呼ばれています")
	}
}

// TestApplyNetworkPolicies_EmptySlice は空スライスでも何も起きないことを確認する
func TestApplyNetworkPolicies_EmptySlice(t *testing.T) {
	mock := &MockSandboxOperator{}
	if err := applyNetworkPolicies(mock, "test-sandbox", []string{}); err != nil {
		t.Fatalf("applyNetworkPolicies([]) エラー: %v", err)
	}

	if mock.AssertCalled("PolicyAllowNetwork") {
		t.Error("空スライスで PolicyAllowNetwork が呼ばれています")
	}
}

// TestApplyNetworkPoliciesFromFile は設定ファイルからドメインを読み込み適用することを確認する
func TestApplyNetworkPoliciesFromFile(t *testing.T) {
	dir := tempDir(t)
	cfg := &config.ProjectConfig{
		Agent: "opencode",
		Network: config.NetworkConfig{
			AllowedDomains: []string{"example.com"},
		},
	}
	if err := config.SaveProjectConfig(dir, cfg); err != nil {
		t.Fatalf("SaveProjectConfig() エラー: %v", err)
	}

	mock := &MockSandboxOperator{}
	if err := applyNetworkPoliciesFromFile(mock, "test-sandbox", dir); err != nil {
		t.Fatalf("applyNetworkPoliciesFromFile() エラー: %v", err)
	}

	if !mock.AssertCalled("PolicyAllowNetwork") {
		t.Error("PolicyAllowNetwork が呼ばれていません")
	}
}

// TestApplyNetworkPoliciesFromFile_ConfigNotFound は設定ファイルがない場合にエラーになることを確認する
func TestApplyNetworkPoliciesFromFile_ConfigNotFound(t *testing.T) {
	dir := tempDir(t)
	mock := &MockSandboxOperator{}
	if err := applyNetworkPoliciesFromFile(mock, "test-sandbox", dir); err == nil {
		t.Error("applyNetworkPoliciesFromFile() はエラーを返すべきですが nil でした")
	}
}

// TestPublishPorts_CallsPortPublish は各ポートが個別に PortPublish されることを確認する
func TestPublishPorts_CallsPortPublish(t *testing.T) {
	mock := &MockSandboxOperator{}

	ports := []string{"3000", "9090", "8080:80"}
	if err := publishPorts(mock, "test-sandbox", ports); err != nil {
		t.Fatalf("publishPorts() エラー: %v", err)
	}

	if mock.CallCount("PortPublish") != 3 {
		t.Errorf("PortPublish の呼び出し回数 = %d, want 3", mock.CallCount("PortPublish"))
	}

	expectedPorts := []string{"3000", "9090", "8080:80"}
	for i, expected := range expectedPorts {
		portSpec, ok := mock.Calls[i].Args[1].(string)
		if !ok {
			t.Fatalf("PortPublish[%d] の第2引数が string ではありません", i)
		}
		if portSpec != expected {
			t.Errorf("PortPublish[%d] = %q, want %q", i, portSpec, expected)
		}
	}
}

// TestResolveLanguages_Auto は自動検出が正しく機能することを確認する
func TestResolveLanguages_Auto(t *testing.T) {
	dir := tempDir(t)
	if err := createFile(filepath.Join(dir, "package.json"), "{}"); err != nil {
		t.Fatalf("package.json 作成エラー: %v", err)
	}

	langs, err := resolveLanguages(dir, "auto")
	if err != nil {
		t.Fatalf("resolveLanguages(dir, auto) エラー: %v", err)
	}
	if len(langs) == 0 {
		t.Fatal("言語が検出されませんでした")
	}
	found := false
	for _, l := range langs {
		if l == "node" {
			found = true
		}
	}
	if !found {
		t.Errorf("resolveLanguages(dir, auto) = %v, node を含むべきです", langs)
	}
}

// TestResolveLanguages_Explicit は明示的な言語指定が正しくパースされることを確認する
func TestResolveLanguages_Explicit(t *testing.T) {
	dir := tempDir(t)

	langs, err := resolveLanguages(dir, "node,go")
	if err != nil {
		t.Fatalf("resolveLanguages(dir, node,go) エラー: %v", err)
	}
	if len(langs) != 2 {
		t.Fatalf("len(langs) = %d, want 2", len(langs))
	}
	if string(langs[0]) != "node" || string(langs[1]) != "go" {
		t.Errorf("resolveLanguages(dir, node,go) = %v, want [node go]", langs)
	}
}

// TestResolveLanguages_Invalid は無効な言語指定でエラーになることを確認する
func TestResolveLanguages_Invalid(t *testing.T) {
	dir := tempDir(t)

	_, err := resolveLanguages(dir, "")
	if err == nil {
		t.Error("resolveLanguages(dir, \"\") はエラーを返すべきです")
	}
}

// TestResolveAgent_Explicit は明示的なエージェント指定を確認する
func TestResolveAgent_Explicit(t *testing.T) {
	agent, err := resolveAgent("codex")
	if err != nil {
		t.Fatalf("resolveAgent(codex) エラー: %v", err)
	}
	if agent != "codex" {
		t.Errorf("resolveAgent(codex) = %q, want %q", agent, "codex")
	}
}

// TestResolveAgent_Default は空文字で既定値が使われることを確認する
func TestResolveAgent_Default(t *testing.T) {
	origHome := t.TempDir()
	dir := filepath.Join(origHome, ".config", "dbox")
	if err := createDir(dir); err != nil {
		t.Fatalf("config dir 作成エラー: %v", err)
	}
	t.Setenv("HOME", origHome)

	agent, err := resolveAgent("")
	if err != nil {
		t.Fatalf("resolveAgent(\"\") エラー: %v", err)
	}
	if agent != "opencode" {
		t.Errorf("resolveAgent(\"\") = %q, want %q", agent, "opencode")
	}
}

// TestPublishPorts_Error は1つめのポートでエラーが発生した場合に即座にエラーを返すことを確認する
func TestPublishPorts_Error(t *testing.T) {
	mock := &MockSandboxOperator{
		PortPublishFunc: func(sandboxName, portSpec string) error {
			if portSpec == "3000" {
				return assertAnError("port publish failed")
			}
			return nil
		},
	}

	err := publishPorts(mock, "test-sandbox", []string{"3000", "8080"})
	if err == nil {
		t.Fatal("publishPorts() はエラーを返すべきです")
	}
}

// assertAnError は常にエラーを返すヘルパー
func assertAnError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
