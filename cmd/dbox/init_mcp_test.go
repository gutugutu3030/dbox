package main

import "testing"

// TestHasLang は言語の存在確認をテストする
func TestHasLang(t *testing.T) {
	tests := []struct {
		name  string
		langs []string
		lang  string
		want  bool
	}{
		{name: "含まれる", langs: []string{"go", "node"}, lang: "go", want: true},
		{name: "含まれない", langs: []string{"go", "node"}, lang: "python", want: false},
		{name: "空スライス", langs: []string{}, lang: "go", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasLang(tt.langs, tt.lang)
			if got != tt.want {
				t.Errorf("hasLang(%v, %q) = %v, want %v", tt.langs, tt.lang, got, tt.want)
			}
		})
	}
}

// TestParseMCPFlags_NoFlags はフラグなしで空リストが返ることを確認する
func TestParseMCPFlags_NoFlags(t *testing.T) {
	initMCP = nil
	initNoMCP = false

	servers := parseMCPFlags()
	if servers != nil {
		t.Errorf("parseMCPFlags() = %v, want nil", servers)
	}
}

// TestParseMCPFlags_Preset はプリセット名を解釈することを確認する
func TestParseMCPFlags_Preset(t *testing.T) {
	initMCP = []string{"gitnexus"}
	initNoMCP = false

	servers := parseMCPFlags()
	if len(servers) != 1 {
		t.Fatalf("parseMCPFlags() の長さ = %d, want 1", len(servers))
	}
	if servers[0].Name != "GitNexus" {
		t.Errorf("Name = %q, want %q", servers[0].Name, "GitNexus")
	}
	if servers[0].Package != "@anthropic/gitnexus" {
		t.Errorf("Package = %q, want %q", servers[0].Package, "@anthropic/gitnexus")
	}
}

// TestParseMCPFlags_Custom はカスタムパッケージ名を解釈することを確認する
func TestParseMCPFlags_Custom(t *testing.T) {
	initMCP = []string{"@myorg/my-mcp"}
	initNoMCP = false

	servers := parseMCPFlags()
	if len(servers) != 1 {
		t.Fatalf("parseMCPFlags() の長さ = %d, want 1", len(servers))
	}
	if servers[0].Name != "@myorg/my-mcp" {
		t.Errorf("Name = %q, want %q", servers[0].Name, "@myorg/my-mcp")
	}
	if servers[0].Package != "@myorg/my-mcp" {
		t.Errorf("Package = %q, want %q", servers[0].Package, "@myorg/my-mcp")
	}
}

// TestParseMCPFlags_NoMCP は --no-mcp フラグで空リストが返ることを確認する
func TestParseMCPFlags_NoMCP(t *testing.T) {
	initMCP = []string{"gitnexus"}
	initNoMCP = true

	servers := parseMCPFlags()
	if servers != nil {
		t.Errorf("parseMCPFlags() with --no-mcp = %v, want nil", servers)
	}
}
