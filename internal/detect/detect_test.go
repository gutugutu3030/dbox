package detect

import (
	"os"
	"path/filepath"
	"testing"
)

// tempDirWithFiles はテスト用の一時ディレクトリにファイルを作成する
func tempDirWithFiles(t *testing.T, files []string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dbox-detect-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("ディレクトリ作成に失敗: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("ファイル作成に失敗: %v", err)
		}
	}
	return dir
}

// TestDetectByConfig_Node は package.json から Node を検出できることを確認する
func TestDetectByConfig_Node(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"package.json", "index.js"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	langs := DetectByConfig(infos)
	if len(langs) != 1 || langs[0] != LanguageNode {
		t.Errorf("DetectByConfig() = %v, want [node]", langs)
	}
}

// TestDetectByConfig_Python は requirements.txt から Python を検出できることを確認する
func TestDetectByConfig_Python(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"requirements.txt", "main.py"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	langs := DetectByConfig(infos)
	if len(langs) != 1 || langs[0] != LanguagePython {
		t.Errorf("DetectByConfig() = %v, want [python]", langs)
	}
}

// TestDetectByConfig_Go は go.mod から Go を検出できることを確認する
func TestDetectByConfig_Go(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"go.mod", "main.go"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	langs := DetectByConfig(infos)
	if len(langs) != 1 || langs[0] != LanguageGo {
		t.Errorf("DetectByConfig() = %v, want [go]", langs)
	}
}

// TestDetectByConfig_Multi は package.json と go.mod の両方から node と go を検出できることを確認する
func TestDetectByConfig_Multi(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"package.json", "go.mod", "main.go"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	langs := DetectByConfig(infos)
	if len(langs) != 2 {
		t.Fatalf("DetectByConfig() = %v, want 2 languages", langs)
	}

	hasNode, hasGo := false, false
	for _, l := range langs {
		if l == LanguageNode {
			hasNode = true
		}
		if l == LanguageGo {
			hasGo = true
		}
	}
	if !hasNode || !hasGo {
		t.Errorf("DetectByConfig() に node または go が含まれていません: %v", langs)
	}
}

// TestDetectByConfig_Empty は該当ファイルがない場合に空スライスを返すことを確認する
func TestDetectByConfig_Empty(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"README.md", "LICENSE"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	langs := DetectByConfig(infos)
	if len(langs) != 0 {
		t.Errorf("DetectByConfig() = %v, want empty", langs)
	}
}

// TestDetectByExtension_Node は .ts ファイルから Node を検出できることを確認する
func TestDetectByExtension_Node(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"src/index.ts",
		"src/app.tsx",
		"src/utils.js",
		"README.md",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageNode {
		t.Errorf("DetectByExtension() = %v, want [node]", langs)
	}
}

// TestDetectByExtension_Python は .py ファイルから Python を検出できることを確認する
func TestDetectByExtension_Python(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.py",
		"utils.py",
		"tests/test_main.py",
		"README.md",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguagePython {
		t.Errorf("DetectByExtension() = %v, want [python]", langs)
	}
}

// TestDetectByExtension_Go は .go ファイルから Go を検出できることを確認する
func TestDetectByExtension_Go(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.go",
		"internal/handler.go",
		"go.mod",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageGo {
		t.Errorf("DetectByExtension() = %v, want [go]", langs)
	}
}

// TestDetectByExtension_Rust は .rs ファイルから Rust を検出できることを確認する
func TestDetectByExtension_Rust(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"src/main.rs",
		"src/lib.rs",
		"Cargo.toml",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageRust {
		t.Errorf("DetectByExtension() = %v, want [rust]", langs)
	}
}

// TestDetectByExtension_Java は .java ファイルから Java を検出できることを確認する
func TestDetectByExtension_Java(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"src/Main.java",
		"src/Controller.java",
		"pom.xml",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageJava {
		t.Errorf("DetectByExtension() = %v, want [java]", langs)
	}
}

// TestDetectByExtension_Ruby は .rb ファイルから Ruby を検出できることを確認する
func TestDetectByExtension_Ruby(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.rb",
		"app/models/user.rb",
		"Gemfile",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageRuby {
		t.Errorf("DetectByExtension() = %v, want [ruby]", langs)
	}
}

// TestDetectByExtension_Base は該当拡張子がない場合に空スライスを返すことを確認する
func TestDetectByExtension_Base(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"README.md",
		"Makefile",
		"LICENSE",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 0 {
		t.Errorf("DetectByExtension() = %v, want empty", langs)
	}
}

// TestDetectByExtension_EmptyDir は空ディレクトリで空スライスを返すことを確認する
func TestDetectByExtension_EmptyDir(t *testing.T) {
	dir := tempDirWithFiles(t, nil)

	langs := DetectByExtension(dir)
	if len(langs) != 0 {
		t.Errorf("DetectByExtension() = %v, want empty", langs)
	}
}

// TestDetectByExtension_SkipNodeModules は node_modules をスキップすることを確認する
func TestDetectByExtension_SkipNodeModules(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"package.json",
		"src/index.ts",
		"node_modules/some-pkg/index.js",
		"node_modules/some-pkg/dist/main.js",
	})

	langs := DetectByExtension(dir)
	if len(langs) != 1 || langs[0] != LanguageNode {
		t.Errorf("DetectByExtension() = %v, want [node]", langs)
	}
}

// TestDetect_ByConfig は設定ファイルベースの検出をテストする（config優先）
func TestDetect_ByConfig(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"package.json",
		"main.py",
		"main.go",
	})

	result := Detect(dir)
	if len(result.Languages) != 1 || result.Languages[0] != LanguageNode {
		t.Errorf("Detect() = %v, want [node] (configファイル優先)", result.Languages)
	}
}

// TestDetect_ByConfig_Multi は複数の設定ファイルから複数言語を検出する
func TestDetect_ByConfig_Multi(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"package.json",
		"go.mod",
		"main.go",
		"index.ts",
	})

	result := Detect(dir)
	if len(result.Languages) != 2 {
		t.Fatalf("Detect() = %v, want 2 languages", result.Languages)
	}
	if result.TemplateName != "dbox-go-node" {
		t.Errorf("TemplateName = %q, want %q", result.TemplateName, "dbox-go-node")
	}
}

// TestDetect_ByExtension は拡張子ベースの検出をテストする
func TestDetect_ByExtension(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.py",
		"utils.py",
		"tests/test_main.py",
	})

	result := Detect(dir)
	if len(result.Languages) != 1 || result.Languages[0] != LanguagePython {
		t.Errorf("Detect() = %v, want [python]", result.Languages)
	}
	if result.TemplateName != "dbox-python" {
		t.Errorf("TemplateName = %q, want %q", result.TemplateName, "dbox-python")
	}
}

// TestDetect_NoLanguage は該当がない場合に base を返す
func TestDetect_NoLanguage(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"README.md",
		"LICENSE",
	})

	result := Detect(dir)
	if len(result.Languages) != 1 || result.Languages[0] != LanguageBase {
		t.Errorf("Detect() = %v, want [base]", result.Languages)
	}
	if result.TemplateName != "dbox-base" {
		t.Errorf("TemplateName = %q, want %q", result.TemplateName, "dbox-base")
	}
}

// TestTemplateNameForLang は各言語に対応するテンプレート名を確認する
func TestTemplateNameForLang(t *testing.T) {
	tests := []struct {
		lang     Language
		expected string
	}{
		{LanguageNode, "dbox-node"},
		{LanguageGo, "dbox-go"},
		{LanguageRust, "dbox-rust"},
		{LanguagePython, "dbox-python"},
		{LanguageJava, "dbox-java"},
		{LanguageRuby, "dbox-ruby"},
		{LanguageBase, "dbox-base"},
		{Language("unknown"), "dbox-base"},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			got := TemplateNameForLang(tt.lang)
			if got != tt.expected {
				t.Errorf("TemplateNameForLang(%q) = %q, want %q", tt.lang, got, tt.expected)
			}
		})
	}
}

// TestTemplateNameForLangs は複数言語のテンプレート名を確認する
func TestTemplateNameForLangs(t *testing.T) {
	tests := []struct {
		name     string
		langs    []Language
		expected string
	}{
		{"単一言語: node", []Language{LanguageNode}, "dbox-node"},
		{"単一言語: go", []Language{LanguageGo}, "dbox-go"},
		{"複数言語: go+node", []Language{LanguageGo, LanguageNode}, "dbox-go-node"},
		{"複数言語: node+go（逆順）", []Language{LanguageNode, LanguageGo}, "dbox-go-node"},
		{"base のみ", []Language{LanguageBase}, "dbox-base"},
		{"base + 言語", []Language{LanguageBase, LanguageNode}, "dbox-node"},
		{"空", nil, "dbox-base"},
		{"重複除去", []Language{LanguageNode, LanguageNode, LanguageGo}, "dbox-go-node"},
		{"3言語", []Language{LanguageRuby, LanguageGo, LanguageNode}, "dbox-go-node-ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TemplateNameForLangs(tt.langs)
			if got != tt.expected {
				t.Errorf("TemplateNameForLangs(%v) = %q, want %q", tt.langs, got, tt.expected)
			}
		})
	}
}

// TestLangsFromString はカンマ区切り文字列のパースを確認する
func TestLangsFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Language
	}{
		{"単一言語", "node", []Language{LanguageNode}},
		{"複数言語", "go,node", []Language{LanguageGo, LanguageNode}},
		{"スペース混じり", "go, node", []Language{LanguageGo, LanguageNode}},
		{"未知の言語も含む", "node,unknown", []Language{LanguageNode}},
		{"base", "base", []Language{LanguageBase}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LangsFromString(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("LangsFromString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("LangsFromString(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestAllTemplates は全テンプレート名が重複なく含まれていることを確認する
func TestAllTemplates(t *testing.T) {
	templates := AllTemplates()

	expected := []string{"dbox-base", "dbox-go", "dbox-java", "dbox-node", "dbox-python", "dbox-ruby", "dbox-rust"}
	if len(templates) != len(expected) {
		t.Errorf("AllTemplates() の要素数 = %d, want %d", len(templates), len(expected))
	}

	seen := make(map[string]bool)
	for _, tmpl := range templates {
		if seen[tmpl] {
			t.Errorf("テンプレート名 %q が重複しています", tmpl)
		}
		seen[tmpl] = true
	}
}
