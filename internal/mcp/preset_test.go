package mcp

import (
	"os"
	"testing"

	"github.com/gutugutu3030/sbx-template/internal/config"
)

// TestPresets はプリセット定義が正しく設定されていることを確認する
func TestPresets(t *testing.T) {
	if len(Presets) == 0 {
		t.Fatal("Presets が空です")
	}

	// 基本3種が含まれていること
	names := make(map[string]bool)
	for _, p := range Presets {
		if p.Name == "" {
			t.Error("プリセット名が空です")
		}
		if p.Package == "" {
			t.Errorf("プリセット %s のパッケージが空です", p.Name)
		}
		if names[p.Name] {
			t.Errorf("プリセット名 %s が重複しています", p.Name)
		}
		names[p.Name] = true
	}
}

// TestPresetByName は名前によるプリセット検索を確認する
func TestPresetByName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "GitNexus", want: true},
		{name: "Context7", want: true},
		{name: "Fetch", want: true},
		{name: "Unknown", want: false},
		{name: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PresetByName(tt.name)
			if tt.want && got == nil {
				t.Errorf("PresetByName(%q) = nil, want non-nil", tt.name)
			}
			if !tt.want && got != nil {
				t.Errorf("PresetByName(%q) = %v, want nil", tt.name, got)
			}
		})
	}
}

// TestPresetToMCPServer は Preset から MCPServer への変換を確認する
func TestPresetToMCPServer(t *testing.T) {
	preset := PresetByName("GitNexus")
	if preset == nil {
		t.Fatal("GitNexus プリセットが見つかりません")
	}

	server := preset.ToMCPServer()
	if server.Name != "GitNexus" {
		t.Errorf("Name = %q, want %q", server.Name, "GitNexus")
	}
	if server.Package != "@anthropic/gitnexus" {
		t.Errorf("Package = %q, want %q", server.Package, "@anthropic/gitnexus")
	}
	if len(server.Domains) == 0 {
		t.Error("Domains が空です")
	}
}

// TestWriteOpenCodeConfig は opencode 設定の書き込みを確認する
func TestWriteOpenCodeConfig(t *testing.T) {
	dir := t.TempDir()

	mcpCfg := GenerateOpenCodeConfig([]config.MCPServer{
		{Name: "test", Package: "@scope/test-pkg"},
	})

	if err := WriteOpenCodeConfig(dir, mcpCfg); err != nil {
		t.Fatalf("WriteOpenCodeConfig() エラー: %v", err)
	}

	// ファイルが作成されたことを確認
	path := OpenCodeConfigPath(dir)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("設定ファイルが作成されていません: %v", err)
	}
}
