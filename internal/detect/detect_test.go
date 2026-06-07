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

	lang, confidence := DetectByConfig(infos)
	if lang != LanguageNode {
		t.Errorf("DetectByConfig() = %q, want %q", lang, LanguageNode)
	}
	if confidence != 1.0 {
		t.Errorf("confidence = %f, want 1.0", confidence)
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

	lang, _ := DetectByConfig(infos)
	if lang != LanguagePython {
		t.Errorf("DetectByConfig() = %q, want %q", lang, LanguagePython)
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

	lang, _ := DetectByConfig(infos)
	if lang != LanguageGo {
		t.Errorf("DetectByConfig() = %q, want %q", lang, LanguageGo)
	}
}

// TestDetectByConfig_Empty は該当ファイルがない場合に LanguageBase を返すことを確認する
func TestDetectByConfig_Empty(t *testing.T) {
	dir := tempDirWithFiles(t, []string{"README.md", "LICENSE"})
	entries, _ := os.ReadDir(dir)
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, info)
	}

	lang, _ := DetectByConfig(infos)
	if lang != LanguageBase {
		t.Errorf("DetectByConfig() = %q, want %q", lang, LanguageBase)
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

	lang, confidence := DetectByExtension(dir)
	if lang != LanguageNode {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageNode)
	}
	if confidence <= 0.5 {
		t.Errorf("confidence = %f, want > 0.5", confidence)
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

	lang, _ := DetectByExtension(dir)
	if lang != LanguagePython {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguagePython)
	}
}

// TestDetectByExtension_Go は .go ファイルから Go を検出できることを確認する
func TestDetectByExtension_Go(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.go",
		"internal/handler.go",
		"go.mod",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageGo {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageGo)
	}
}

// TestDetectByExtension_Rust は .rs ファイルから Rust を検出できることを確認する
func TestDetectByExtension_Rust(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"src/main.rs",
		"src/lib.rs",
		"Cargo.toml",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageRust {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageRust)
	}
}

// TestDetectByExtension_Java は .java ファイルから Java を検出できることを確認する
func TestDetectByExtension_Java(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"src/Main.java",
		"src/Controller.java",
		"pom.xml",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageJava {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageJava)
	}
}

// TestDetectByExtension_Ruby は .rb ファイルから Ruby を検出できることを確認する
func TestDetectByExtension_Ruby(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"main.rb",
		"app/models/user.rb",
		"Gemfile",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageRuby {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageRuby)
	}
}

// TestDetectByExtension_Base は該当拡張子がない場合に LanguageBase を返すことを確認する
func TestDetectByExtension_Base(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"README.md",
		"Makefile",
		"LICENSE",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageBase {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageBase)
	}
}

// TestDetectByExtension_EmptyDir は空ディレクトリで LanguageBase を返すことを確認する
func TestDetectByExtension_EmptyDir(t *testing.T) {
	dir := tempDirWithFiles(t, nil)

	lang, _ := DetectByExtension(dir)
	if lang != LanguageBase {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageBase)
	}
}

// TestDetectByExtension_SkipNodeModules は node_modules をスキップすることを確認する
func TestDetectByExtension_SkipNodeModules(t *testing.T) {
	dir := tempDirWithFiles(t, []string{
		"package.json",
		"src/index.ts",
		// node_modules 以下はスキャンされない
		"node_modules/some-pkg/index.js",
		"node_modules/some-pkg/dist/main.js",
	})

	lang, _ := DetectByExtension(dir)
	if lang != LanguageNode {
		t.Errorf("DetectByExtension() = %q, want %q", lang, LanguageNode)
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
	if result.Language != LanguageNode {
		t.Errorf("Detect() = %q, want %q (configファイル優先)", result.Language, LanguageNode)
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
	if result.Language != LanguagePython {
		t.Errorf("Detect() = %q, want %q", result.Language, LanguagePython)
	}
	if result.TemplateName != "dbox-python" {
		t.Errorf("TemplateName = %q, want %q", result.TemplateName, "dbox-python")
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
