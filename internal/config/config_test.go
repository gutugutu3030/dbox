package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// tempDir はテスト用の一時ディレクトリを作成する
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dbox-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// TestDefaultGlobalConfig は既定値が正しく設定されていることを確認する
func TestDefaultGlobalConfig(t *testing.T) {
	cfg := DefaultGlobalConfig()
	if cfg.DefaultAgent != "opencode" {
		t.Errorf("DefaultAgent = %q, want %q", cfg.DefaultAgent, "opencode")
	}
	if cfg.Template.Registry != "docker/sandbox-templates" {
		t.Errorf("Template.Registry = %q, want %q", cfg.Template.Registry, "docker/sandbox-templates")
	}
}

// TestDefaultProjectConfig はプロジェクト設定の既定値を確認する
func TestDefaultProjectConfig(t *testing.T) {
	cfg := DefaultProjectConfig()
	if cfg.Version != 2 {
		t.Errorf("Version = %d, want %d", cfg.Version, 2)
	}
	if cfg.Agent != "opencode" {
		t.Errorf("Agent = %q, want %q", cfg.Agent, "opencode")
	}
	if len(cfg.Langs) != 1 || cfg.Langs[0] != "base" {
		t.Errorf("Langs = %v, want [base]", cfg.Langs)
	}
	if cfg.Clone != false {
		t.Errorf("Clone = %v, want %v", cfg.Clone, false)
	}
	if cfg.Resources.CPUs != 0 {
		t.Errorf("Resources.CPUs = %d, want %d", cfg.Resources.CPUs, 0)
	}
	if cfg.Resources.Memory != "" {
		t.Errorf("Resources.Memory = %q, want empty", cfg.Resources.Memory)
	}
}

// TestLoadSaveGlobalConfig はグローバル設定の保存と読み込みをテストする
func TestLoadSaveGlobalConfig(t *testing.T) {
	// HOME を一時ディレクトリに変更
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	tmpHome := tempDir(t)
	os.Setenv("HOME", tmpHome)

	// 設定を保存
	cfg := &GlobalConfig{
		DefaultAgent: "codex",
		Template: TemplateConfig{
			Registry: "my-registry",
		},
	}

	if err := SaveGlobalConfig(cfg); err != nil {
		t.Fatalf("SaveGlobalConfig() エラー: %v", err)
	}

	// 保存されたファイルを確認
	path, _ := GlobalConfigPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("設定ファイルが作成されていません: %v", err)
	}

	// 読み込み
	loaded, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() エラー: %v", err)
	}

	if loaded.DefaultAgent != "codex" {
		t.Errorf("DefaultAgent = %q, want %q", loaded.DefaultAgent, "codex")
	}
	if loaded.Template.Registry != "my-registry" {
		t.Errorf("Template.Registry = %q, want %q", loaded.Template.Registry, "my-registry")
	}
}

// TestLoadGlobalConfig_NotExist は設定ファイルがない場合に既定値が返ることを確認する
func TestLoadGlobalConfig_NotExist(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	tmpHome := tempDir(t)
	os.Setenv("HOME", tmpHome)

	cfg, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() エラー: %v", err)
	}
	if cfg.DefaultAgent != "opencode" {
		t.Errorf("DefaultAgent = %q, want %q", cfg.DefaultAgent, "opencode")
	}
}

// TestSaveProjectConfig はプロジェクト設定の保存をテストする
func TestSaveProjectConfig(t *testing.T) {
	dir := tempDir(t)

	cfg := &ProjectConfig{
		Version:     2,
		Agent:       "opencode",
		Langs:       []string{"node"},
		Template:    "dbox-node",
		SandboxName: "dbox-opencode-test-project",
		Clone:       true,
		Resources: ResourceConfig{
			CPUs:   2,
			Memory: "4g",
		},
	}

	if err := SaveProjectConfig(dir, cfg); err != nil {
		t.Fatalf("SaveProjectConfig() エラー: %v", err)
	}

	// ファイルの存在確認
	path := filepath.Join(dir, ".dbox.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf(".dbox.yaml が作成されていません: %v", err)
	}

	// YAML の内容を検証
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(".dbox.yaml 読み込みエラー: %v", err)
	}

	var loaded ProjectConfig
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("YAML パースエラー: %v", err)
	}

	if loaded.Agent != "opencode" {
		t.Errorf("Agent = %q, want %q", loaded.Agent, "opencode")
	}
	if len(loaded.Langs) != 1 || loaded.Langs[0] != "node" {
		t.Errorf("Langs = %v, want [node]", loaded.Langs)
	}
	if loaded.SandboxName != "dbox-opencode-test-project" {
		t.Errorf("SandboxName = %q, want %q", loaded.SandboxName, "dbox-opencode-test-project")
	}
}

// TestLoadProjectConfig はプロジェクト設定の読み込みをテストする
func TestLoadProjectConfig(t *testing.T) {
	dir := tempDir(t)

	// 設定ファイルを作成
	cfg := DefaultProjectConfig()
	cfg.Agent = "claude"
	cfg.Langs = []string{"python"}
	if err := SaveProjectConfig(dir, cfg); err != nil {
		t.Fatalf("SaveProjectConfig() エラー: %v", err)
	}

	loaded, err := LoadProjectConfig(dir)
	if err != nil {
		t.Fatalf("LoadProjectConfig() エラー: %v", err)
	}

	if loaded.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", loaded.Agent, "claude")
	}
	if len(loaded.Langs) != 1 || loaded.Langs[0] != "python" {
		t.Errorf("Langs = %v, want [python]", loaded.Langs)
	}
}

// TestLoadProjectConfig_NotExist は設定ファイルがない場合にエラーが返ることを確認する
func TestLoadProjectConfig_NotExist(t *testing.T) {
	dir := tempDir(t)

	_, err := LoadProjectConfig(dir)
	if err == nil {
		t.Error("LoadProjectConfig() はエラーを返すべきですが nil でした")
	}
}

// TestGlobalConfigDir は設定ディレクトリパスが正しいことを確認する
func TestGlobalConfigDir(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	tmpHome := tempDir(t)
	os.Setenv("HOME", tmpHome)

	dir, err := GlobalConfigDir()
	if err != nil {
		t.Fatalf("GlobalConfigDir() エラー: %v", err)
	}

	expected := filepath.Join(tmpHome, ".config", "dbox")
	if dir != expected {
		t.Errorf("GlobalConfigDir() = %q, want %q", dir, expected)
	}
}

// TestFindProjectConfig は .dbox.yaml の検索をテストする
func TestFindProjectConfig(t *testing.T) {
	dir := tempDir(t)

	// 設定ファイルがない場合
	_, err := FindProjectConfig(dir)
	if err == nil {
		t.Error("FindProjectConfig() はエラーを返すべきですが nil でした")
	}

	// 設定ファイルを作成
	cfg := DefaultProjectConfig()
	SaveProjectConfig(dir, cfg)

	// 設定ファイルが見つかること
	path, err := FindProjectConfig(dir)
	if err != nil {
		t.Fatalf("FindProjectConfig() エラー: %v", err)
	}
	if filepath.Base(path) != ".dbox.yaml" {
		t.Errorf("ファイル名 = %q, want %q", filepath.Base(path), ".dbox.yaml")
	}
}

// TestMergeDomains はユーザー指定ドメインとエージェント既定値のマージを確認する
func TestMergeDomains(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected []string
	}{
		{
			name: "エージェント既定値のみ",
			cfg: &ProjectConfig{
				Agent:   "opencode",
				Network: NetworkConfig{},
			},
			expected: []string{"opencode.ai:443"},
		},
		{
			name: "ユーザー指定のみ",
			cfg: &ProjectConfig{
				Agent: "unknown",
				Network: NetworkConfig{
					AllowedDomains: []string{"example.com"},
				},
			},
			expected: []string{"example.com"},
		},
		{
			name: "両方あり（重複なし）",
			cfg: &ProjectConfig{
				Agent: "opencode",
				Network: NetworkConfig{
					AllowedDomains: []string{"example.com", "api.example.com"},
				},
			},
			expected: []string{"opencode.ai:443", "example.com", "api.example.com"},
		},
		{
			name: "重複あり（エージェント既定と同じドメインが指定された場合）",
			cfg: &ProjectConfig{
				Agent: "opencode",
				Network: NetworkConfig{
					AllowedDomains: []string{"opencode.ai:443"},
				},
			},
			expected: []string{"opencode.ai:443"},
		},
		{
			name: "全許可（**）",
			cfg: &ProjectConfig{
				Agent: "opencode",
				Network: NetworkConfig{
					AllowedDomains: []string{"**"},
				},
			},
			expected: []string{"opencode.ai:443", "**"},
		},
		{
			name: "エージェント既定値なし・ユーザー指定なし",
			cfg: &ProjectConfig{
				Agent:   "unknown",
				Network: NetworkConfig{},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeDomains(tt.cfg)
			if len(got) != len(tt.expected) {
				t.Fatalf("MergeDomains() = %v (len=%d), want %v (len=%d)", got, len(got), tt.expected, len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("MergeDomains()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestEnsureGlobalConfigDir は設定ディレクトリが作成されることを確認する
func TestEnsureGlobalConfigDir(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	tmpHome := tempDir(t)
	os.Setenv("HOME", tmpHome)

	if err := EnsureGlobalConfigDir(); err != nil {
		t.Fatalf("EnsureGlobalConfigDir() エラー: %v", err)
	}

	dir, _ := GlobalConfigDir()
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("設定ディレクトリが作成されていません: %v", err)
	}
}
