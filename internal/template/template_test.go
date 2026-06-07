package template

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewBuilder は Builder が正しく作成されることを確認する
func TestNewBuilder(t *testing.T) {
	b := NewBuilder("/tmp/templates", false)
	if b == nil {
		t.Fatal("NewBuilder() が nil を返しました")
	}
	if b.TemplatesDir != "/tmp/templates" {
		t.Errorf("TemplatesDir = %q, want %q", b.TemplatesDir, "/tmp/templates")
	}
	if b.DryRun != false {
		t.Errorf("DryRun = %v, want %v", b.DryRun, false)
	}

	b2 := NewBuilder("/tmp/templates2", true)
	if b2.TemplatesDir != "/tmp/templates2" {
		t.Errorf("TemplatesDir = %q, want %q", b2.TemplatesDir, "/tmp/templates2")
	}
	if b2.DryRun != true {
		t.Errorf("DryRun = %v, want %v", b2.DryRun, true)
	}
}

// TestBuildBase_NoDockerfile は Dockerfile がない場合にエラーを返すことを確認する
func TestBuildBase_NoDockerfile(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	b := NewBuilder(dir, true)
	err = b.BuildBase()
	if err == nil {
		t.Error("BuildBase() は Dockerfile がない場合にエラーを返すべきです")
	}
}

// TestBuildLang_NoDockerfile は言語 Dockerfile がない場合にエラーを返すことを確認する
func TestBuildLang_NoDockerfile(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	b := NewBuilder(dir, true)
	err = b.BuildLang("node")
	if err == nil {
		t.Error("BuildLang() は Dockerfile がない場合にエラーを返すべきです")
	}
}

// TestBuildLang_WithDockerfile は DryRun モードでビルドが正常終了することを確認する
func TestBuildLang_WithDockerfile(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	// base.Dockerfile と node.Dockerfile を作成
	baseDockerfile := `FROM ubuntu:24.04
RUN apt-get update && apt-get install -y git curl
`
	nodeDockerfile := `FROM dbox-base:latest
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs
`

	if err := os.WriteFile(filepath.Join(dir, "base.Dockerfile"), []byte(baseDockerfile), 0644); err != nil {
		t.Fatalf("base.Dockerfile 作成に失敗: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node.Dockerfile"), []byte(nodeDockerfile), 0644); err != nil {
		t.Fatalf("node.Dockerfile 作成に失敗: %v", err)
	}

	// DryRun モードでビルド（エラーが発生しないことを確認）
	b := NewBuilder(dir, true)
	if err := b.BuildLang("node"); err != nil {
		t.Errorf("BuildLang() with DryRun エラー: %v", err)
	}
}

// TestBuildAll は DryRun モードで全ビルドが正常終了することを確認する
func TestBuildAll(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	// 必要な Dockerfile をすべて作成
	dockerfiles := map[string]string{
		"base.Dockerfile":   `FROM ubuntu:24.04`,
		"node.Dockerfile":   `FROM dbox-base:latest`,
		"python.Dockerfile": `FROM dbox-base:latest`,
		"java.Dockerfile":   `FROM dbox-base:latest`,
		"go.Dockerfile":     `FROM dbox-base:latest`,
		"rust.Dockerfile":   `FROM dbox-base:latest`,
		"ruby.Dockerfile":   `FROM dbox-base:latest`,
	}

	for name, content := range dockerfiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("%s 作成に失敗: %v", name, err)
		}
	}

	b := NewBuilder(dir, true)
	if err := b.BuildAll(); err != nil {
		t.Errorf("BuildAll() with DryRun エラー: %v", err)
	}
}

// TestNewComposer は Composer が正しく作成されることを確認する
func TestNewComposer(t *testing.T) {
	b := NewBuilder("/tmp/templates", true)
	c := NewComposer(b)
	if c == nil {
		t.Fatal("NewComposer() が nil を返しました")
	}
	if c.Builder != b {
		t.Error("Builder が正しく設定されていません")
	}
}

// TestCompose_EmptyLangs は空の言語リストでエラーを返すことを確認する
func TestCompose_EmptyLangs(t *testing.T) {
	b := NewBuilder("/tmp/templates", true)
	c := NewComposer(b)

	_, err := c.Compose(nil)
	if err == nil {
		t.Error("Compose(nil) はエラーを返すべきです")
	}

	_, err = c.Compose([]string{})
	if err == nil {
		t.Error("Compose([]string{}) はエラーを返すべきです")
	}
}

// TestCompose_BaseOnly は base のみの場合に dbox-base を返すことを確認する
func TestCompose_BaseOnly(t *testing.T) {
	b := NewBuilder("/tmp/templates", true)
	c := NewComposer(b)

	name, err := c.Compose([]string{"base"})
	if err != nil {
		t.Fatalf("Compose([base]) エラー: %v", err)
	}
	if name != "dbox-base" {
		t.Errorf("Compose([base]) = %q, want %q", name, "dbox-base")
	}
}

// TestCompose_SingleLang はスニペットを使って合成Dockerfileを生成することを確認する
func TestCompose_SingleLang(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	// base.Dockerfile を作成
	baseDockerfile := `FROM ubuntu:24.04
RUN apt-get update && apt-get install -y git curl
`
	if err := os.WriteFile(filepath.Join(dir, "base.Dockerfile"), []byte(baseDockerfile), 0644); err != nil {
		t.Fatalf("base.Dockerfile 作成に失敗: %v", err)
	}

	// snippets ディレクトリと node.snippet を作成
	snippetsDir := filepath.Join(dir, "snippets")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		t.Fatalf("snippets ディレクトリ作成に失敗: %v", err)
	}
	nodeSnippet := `RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs && rm -rf /var/lib/apt/lists/*`
	if err := os.WriteFile(filepath.Join(snippetsDir, "node.snippet"), []byte(nodeSnippet), 0644); err != nil {
		t.Fatalf("node.snippet 作成に失敗: %v", err)
	}

	b := NewBuilder(dir, true)
	c := NewComposer(b)

	// DryRun では HasDockerImage が常に false → ビルド実行
	name, err := c.Compose([]string{"node"})
	if err != nil {
		t.Fatalf("Compose([node]) エラー: %v", err)
	}
	if name != "dbox-node" {
		t.Errorf("Compose([node]) = %q, want %q", name, "dbox-node")
	}
}

// TestCompose_MultiLang は複数言語の合成Dockerfileを生成することを確認する
func TestCompose_MultiLang(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	// base.Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "base.Dockerfile"), []byte("FROM ubuntu:24.04"), 0644); err != nil {
		t.Fatalf("base.Dockerfile 作成に失敗: %v", err)
	}

	// snippets
	snippetsDir := filepath.Join(dir, "snippets")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		t.Fatalf("snippets ディレクトリ作成に失敗: %v", err)
	}

	snippets := map[string]string{
		"node.snippet": `RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs`,
		"go.snippet":   `RUN curl -fsSL https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz`,
	}
	for name, content := range snippets {
		if err := os.WriteFile(filepath.Join(snippetsDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("%s 作成に失敗: %v", name, err)
		}
	}

	b := NewBuilder(dir, true)
	c := NewComposer(b)

	name, err := c.Compose([]string{"node", "go"})
	if err != nil {
		t.Fatalf("Compose([node, go]) エラー: %v", err)
	}
	// ソートされて "dbox-go-node" になるはず
	if name != "dbox-go-node" {
		t.Errorf("Compose([node, go]) = %q, want %q", name, "dbox-go-node")
	}
}

// TestCompose_Dedup は重複言語を除去することを確認する
func TestCompose_Dedup(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "base.Dockerfile"), []byte("FROM ubuntu:24.04"), 0644); err != nil {
		t.Fatalf("base.Dockerfile 作成に失敗: %v", err)
	}

	snippetsDir := filepath.Join(dir, "snippets")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		t.Fatalf("snippets ディレクトリ作成に失敗: %v", err)
	}
	nodeSnippet := `RUN echo node`
	goSnippet := `RUN echo go`
	if err := os.WriteFile(filepath.Join(snippetsDir, "node.snippet"), []byte(nodeSnippet), 0644); err != nil {
		t.Fatalf("node.snippet 作成に失敗: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snippetsDir, "go.snippet"), []byte(goSnippet), 0644); err != nil {
		t.Fatalf("go.snippet 作成に失敗: %v", err)
	}

	b := NewBuilder(dir, true)
	c := NewComposer(b)

	// 重複除去 + ソートにより "dbox-go-node" になる
	name, err := c.Compose([]string{"node", "go", "node"})
	if err != nil {
		t.Fatalf("Compose([node, go, node]) エラー: %v", err)
	}
	if name != "dbox-go-node" {
		t.Errorf("Compose([node, go, node]) = %q, want %q", name, "dbox-go-node")
	}
}

// TestEnsureTemplatesExtracted は埋め込みテンプレートが正しく展開されることを確認する
func TestEnsureTemplatesExtracted(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-extract-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := EnsureTemplatesExtracted(dir); err != nil {
		t.Fatalf("EnsureTemplatesExtracted() エラー: %v", err)
	}

	// 展開されたファイルを確認
	expectedFiles := []string{
		"base.Dockerfile",
		"node.Dockerfile",
		"go.Dockerfile",
		"python.Dockerfile",
		"java.Dockerfile",
		"rust.Dockerfile",
		"ruby.Dockerfile",
		"snippets/node.snippet",
		"snippets/go.snippet",
		"snippets/python.snippet",
		"snippets/java.snippet",
		"snippets/rust.snippet",
		"snippets/ruby.snippet",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("ファイル %s が展開されていません: %v", f, err)
		}
	}

	// 2回目はスキップされること（既存ファイルの確認）
	if err := EnsureTemplatesExtracted(dir); err != nil {
		t.Errorf("2回目の EnsureTemplatesExtracted() がエラー: %v", err)
	}
}

// TestCompose_MissingSnippet はスニペットがない場合にエラーを返すことを確認する
func TestCompose_MissingSnippet(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	// base.Dockerfile だけ作成し、snippets は作成しない
	if err := os.WriteFile(filepath.Join(dir, "base.Dockerfile"), []byte("FROM ubuntu:24.04"), 0644); err != nil {
		t.Fatalf("base.Dockerfile 作成に失敗: %v", err)
	}

	b := NewBuilder(dir, true)
	c := NewComposer(b)

	// go のスニペットがないのでエラーになる
	_, err = c.Compose([]string{"go"})
	if err == nil {
		t.Error("Compose([go]) はスニペットがない場合にエラーを返すべきです")
	}
}
