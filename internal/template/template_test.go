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
RUN apt-get update && apt-get install -y neovim git curl
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

// TestBuildAndSaveAll は DryRun モードで全ビルドが正常終了することを確認する
func TestBuildAndSaveAll(t *testing.T) {
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
	if err := b.BuildAndSaveAll(); err != nil {
		t.Errorf("BuildAndSaveAll() with DryRun エラー: %v", err)
	}
}

// TestEnsureNvimConfig は nvim 設定ディレクトリが存在しない場合にスキップされることを確認する
func TestEnsureNvimConfig(t *testing.T) {
	// HOME を一時ディレクトリに設定
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpHome, err := os.MkdirTemp("", "dbox-nvim-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)

	// nvim 設定がない状態ではスキップされる
	if err := EnsureNvimConfig(); err != nil {
		t.Errorf("EnsureNvimConfig() エラー: %v", err)
	}
}

// TestSaveTemplate_DryRun は DryRun モードでテンプレート保存が正常終了することを確認する
func TestSaveTemplate_DryRun(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbox-template-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	defer os.RemoveAll(dir)

	b := NewBuilder(dir, true)
	if err := b.SaveTemplate("dbox-test:latest"); err != nil {
		t.Errorf("SaveTemplate() with DryRun エラー: %v", err)
	}
}
