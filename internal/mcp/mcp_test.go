package mcp

import (
	"testing"

	"github.com/gutugutu3030/sbx-template/internal/config"
)

// TestInstallCommand_Empty は空のサーバーリストで空文字列が返ることを確認する
func TestInstallCommand_Empty(t *testing.T) {
	got := InstallCommand(nil)
	if got != "" {
		t.Errorf("InstallCommand(nil) = %q, want empty", got)
	}

	got = InstallCommand([]config.MCPServer{})
	if got != "" {
		t.Errorf("InstallCommand([]) = %q, want empty", got)
	}
}

// TestInstallCommand_Single は単一パッケージのインストールコマンドを確認する
func TestInstallCommand_Single(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "test", Package: "@scope/test-pkg"},
	}
	got := InstallCommand(servers)
	want := "npm install -g @scope/test-pkg"
	if got != want {
		t.Errorf("InstallCommand() = %q, want %q", got, want)
	}
}

// TestInstallCommand_Multi は複数パッケージのインストールコマンドを確認する
func TestInstallCommand_Multi(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "a", Package: "@scope/a"},
		{Name: "b", Package: "@scope/b"},
	}
	got := InstallCommand(servers)
	want := "npm install -g @scope/a @scope/b"
	if got != want {
		t.Errorf("InstallCommand() = %q, want %q", got, want)
	}
}

// TestInstallCommand_NoPackage はパッケージが空の場合に空文字列を返すことを確認する
func TestInstallCommand_NoPackage(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "test", Package: ""},
	}
	got := InstallCommand(servers)
	if got != "" {
		t.Errorf("InstallCommand() with empty package = %q, want empty", got)
	}
}

// TestGenerateOpenCodeConfig_Empty は空のサーバーリストで空マップを返すことを確認する
func TestGenerateOpenCodeConfig_Empty(t *testing.T) {
	got := GenerateOpenCodeConfig(nil)
	if len(got) != 0 {
		t.Errorf("GenerateOpenCodeConfig(nil) = %v, want empty", got)
	}
}

// TestGenerateOpenCodeConfig_Single は単一サーバーの設定を確認する
func TestGenerateOpenCodeConfig_Single(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "gitnexus", Package: "@anthropic/gitnexus"},
	}
	got := GenerateOpenCodeConfig(servers)

	if len(got) != 1 {
		t.Fatalf("GenerateOpenCodeConfig() の長さ = %d, want 1", len(got))
	}

	srv, ok := got["gitnexus"]
	if !ok {
		t.Fatal("GenerateOpenCodeConfig() に gitnexus が含まれていません")
	}
	if srv.Type != "local" {
		t.Errorf("Type = %q, want %q", srv.Type, "local")
	}
	if srv.Command != "npx" {
		t.Errorf("Command = %q, want %q", srv.Command, "npx")
	}
	if len(srv.Args) != 2 || srv.Args[0] != "-y" || srv.Args[1] != "@anthropic/gitnexus" {
		t.Errorf("Args = %v, want [-y @anthropic/gitnexus]", srv.Args)
	}
}

// TestGenerateOpenCodeConfig_Multi は複数サーバーの設定を確認する
func TestGenerateOpenCodeConfig_Multi(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "a", Package: "@scope/a"},
		{Name: "b", Package: "@scope/b"},
	}
	got := GenerateOpenCodeConfig(servers)
	if len(got) != 2 {
		t.Fatalf("GenerateOpenCodeConfig() の長さ = %d, want 2", len(got))
	}
	if _, ok := got["a"]; !ok {
		t.Error("a が含まれていません")
	}
	if _, ok := got["b"]; !ok {
		t.Error("b が含まれていません")
	}
}

// TestGenerateOpenCodeConfig_SkipInvalid は空の名前やパッケージをスキップすることを確認する
func TestGenerateOpenCodeConfig_SkipInvalid(t *testing.T) {
	servers := []config.MCPServer{
		{Name: "", Package: "@scope/a"},
		{Name: "valid", Package: ""},
		{Name: "ok", Package: "@scope/ok"},
	}
	got := GenerateOpenCodeConfig(servers)
	if len(got) != 1 {
		t.Fatalf("GenerateOpenCodeConfig() の長さ = %d, want 1", len(got))
	}
	if _, ok := got["ok"]; !ok {
		t.Error("ok が含まれていません")
	}
}

// TestMergeMCPDomains は MCP ドメインと既存ドメインのマージを確認する
func TestMergeMCPDomains(t *testing.T) {
	tests := []struct {
		name     string
		servers  []config.MCPServer
		existing []string
		want     []string
	}{
		{
			name:     "サーバーのみ",
			servers:  []config.MCPServer{{Name: "t", Package: "@t/t", Domains: []string{"api.t.com:443"}}},
			existing: nil,
			want:     []string{"api.t.com:443"},
		},
		{
			name:     "既存のみ",
			servers:  nil,
			existing: []string{"existing.com:443"},
			want:     []string{"existing.com:443"},
		},
		{
			name:     "両方マージ",
			servers:  []config.MCPServer{{Name: "t", Package: "@t/t", Domains: []string{"api.t.com:443"}}},
			existing: []string{"existing.com:443"},
			want:     []string{"existing.com:443", "api.t.com:443"},
		},
		{
			name:     "重複除去",
			servers:  []config.MCPServer{{Name: "t", Package: "@t/t", Domains: []string{"same.com:443"}}},
			existing: []string{"same.com:443"},
			want:     []string{"same.com:443"},
		},
		{
			name: "複数サーバーのドメイン",
			servers: []config.MCPServer{
				{Name: "a", Package: "@a/a", Domains: []string{"a.com:443"}},
				{Name: "b", Package: "@b/b", Domains: []string{"b.com:443"}},
			},
			existing: nil,
			want:     []string{"a.com:443", "b.com:443"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeMCPDomains(tt.servers, tt.existing)
			if len(got) != len(tt.want) {
				t.Fatalf("MergeMCPDomains() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("MergeMCPDomains()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestHasNodeJS は Node.js ランタイムの必要性を判定する
func TestHasNodeJS(t *testing.T) {
	tests := []struct {
		name    string
		servers []config.MCPServer
		want    bool
	}{
		{
			name:    "空リスト",
			servers: nil,
			want:    false,
		},
		{
			name:    "パッケージあり",
			servers: []config.MCPServer{{Name: "t", Package: "@t/t"}},
			want:    true,
		},
		{
			name:    "パッケージなし",
			servers: []config.MCPServer{{Name: "t", Package: ""}},
			want:    false,
		},
		{
			name: "空文字列のパッケージも含む",
			servers: []config.MCPServer{
				{Name: "a", Package: ""},
				{Name: "b", Package: "@b/b"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasNodeJS(tt.servers)
			if got != tt.want {
				t.Errorf("HasNodeJS(%v) = %v, want %v", tt.servers, got, tt.want)
			}
		})
	}
}
