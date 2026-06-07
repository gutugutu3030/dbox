package template

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	assets "github.com/gutugutu3030/sbx-template"
)

// Builder はDockerテンプレートのビルドを担当する
type Builder struct {
	// DryRun が true の場合、実際のコマンドは実行しない
	DryRun bool
	// TemplatesDir はDockerfileが配置されているディレクトリ
	TemplatesDir string
}

// NewBuilder は Builder を作成する
func NewBuilder(templatesDir string, dryRun bool) *Builder {
	return &Builder{
		DryRun:       dryRun,
		TemplatesDir: templatesDir,
	}
}

// dockerBuild は docker build を実行する
func (b *Builder) dockerBuild(tag, dockerfile, contextDir string) error {
	args := []string{"build", "-t", tag, "-f", dockerfile, contextDir}
	cmd := exec.Command("docker", args...)

	if b.DryRun {
		fmt.Printf("[dry-run] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build (tag=%s) に失敗: %w", tag, err)
	}
	return nil
}

// BuildBase はベースイメージをビルドする
func (b *Builder) BuildBase() error {
	dockerfile := filepath.Join(b.TemplatesDir, "base.Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("base.Dockerfile が見つかりません: %w", err)
	}

	tag := "dbox-base:latest"
	if err := b.dockerBuild(tag, dockerfile, b.TemplatesDir); err != nil {
		return err
	}

	fmt.Printf("ベースイメージ %s をビルドしました\n", tag)
	return nil
}

// BuildLang は言語ごとのイメージをビルドする。
// base.Dockerfile を先にビルドし、その上に言語別のレイヤーを追加する
func (b *Builder) BuildLang(lang string) error {
	if err := b.BuildBase(); err != nil {
		return err
	}

	dockerfile := filepath.Join(b.TemplatesDir, lang+".Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("%s.Dockerfile が見つかりません: %w", lang, err)
	}

	tag := fmt.Sprintf("dbox-%s:latest", lang)
	if err := b.dockerBuild(tag, dockerfile, b.TemplatesDir); err != nil {
		return err
	}

	fmt.Printf("言語イメージ %s をビルドしました\n", tag)
	return nil
}

// BuildAll は全言語のDockerイメージをビルドする
func (b *Builder) BuildAll() error {
	languages := []string{"node", "python", "java", "go", "rust", "ruby"}

	for _, lang := range languages {
		fmt.Printf("=== %s イメージをビルド中 ===\n", lang)
		if err := b.BuildLang(lang); err != nil {
			return err
		}
	}

	return nil
}

// Composer は複数言語のスニペットを合成したイメージをビルドする
type Composer struct {
	Builder *Builder
}

// NewComposer は Composer を作成する
func NewComposer(builder *Builder) *Composer {
	return &Composer{Builder: builder}
}

// Compose は指定された言語群の合成Dockerイメージをビルドする。
// 呼び出し元が sbx へのロードを行う
func (c *Composer) Compose(langs []string) (string, error) {
	if len(langs) == 0 {
		return "", fmt.Errorf("言語が指定されていません")
	}

	if len(langs) == 1 && langs[0] == "base" {
		return "dbox-base", nil
	}

	filtered := make([]string, 0, len(langs))
	for _, lang := range langs {
		if lang != "base" {
			filtered = append(filtered, lang)
		}
	}
	if len(filtered) == 0 {
		return "dbox-base", nil
	}

	sort.Strings(filtered)
	uniq := make([]string, 0, len(filtered))
	seen := make(map[string]bool)
	for _, lang := range filtered {
		if !seen[lang] {
			seen[lang] = true
			uniq = append(uniq, lang)
		}
	}

	imageName := "dbox-" + strings.Join(uniq, "-")
	tag := imageName + ":latest"

	if err := c.Builder.BuildBase(); err != nil {
		return "", err
	}

	dockerfile, contextDir, err := c.generateCompositeDockerfile(uniq)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(contextDir)

	if err := c.Builder.dockerBuild(tag, dockerfile, contextDir); err != nil {
		return "", err
	}

	return imageName, nil
}

// generateCompositeDockerfile は複数言語のスニペットを合成した Dockerfile を生成する
func (c *Composer) generateCompositeDockerfile(langs []string) (dockerfile string, contextDir string, err error) {
	contextDir, err = os.MkdirTemp("", "dbox-compose-*")
	if err != nil {
		return "", "", fmt.Errorf("一時ディレクトリ作成に失敗: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("FROM dbox-base:latest\n\n")

	for _, lang := range langs {
		snippetPath := filepath.Join(c.Builder.TemplatesDir, "snippets", lang+".snippet")
		data, err := os.ReadFile(snippetPath)
		if err != nil {
			os.RemoveAll(contextDir)
			return "", "", fmt.Errorf("%s のスニペットが見つかりません: %w", lang, err)
		}
		sb.WriteString(fmt.Sprintf("# %s\n%s\n\n", lang, string(data)))
	}

	sb.WriteString("WORKDIR /workspace\n")

	dockerfile = filepath.Join(contextDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte(sb.String()), 0644); err != nil {
		os.RemoveAll(contextDir)
		return "", "", fmt.Errorf("Dockerfile 書き込みに失敗: %w", err)
	}

	return dockerfile, contextDir, nil
}

// ensureFile は埋め込みFSからファイルを読み込み、書き出し先にコピーする。
// 書き出し先が既に存在する場合はスキップする
func ensureFile(fsys fs.FS, src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	data, err := fs.ReadFile(fsys, src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// EnsureTemplatesExtracted は埋め込みテンプレートを指定ディレクトリにコピーする。
// 既に存在するファイルはスキップする
func EnsureTemplatesExtracted(dstDir string) error {
	entries, err := assets.Templates.ReadDir("templates")
	if err != nil {
		return fmt.Errorf("埋め込みテンプレートの読み取りに失敗: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "snippets" {
				snippetEntries, err := assets.Templates.ReadDir("templates/snippets")
				if err != nil {
					return fmt.Errorf("埋め込みスニペットの読み取りに失敗: %w", err)
				}
				snipDst := filepath.Join(dstDir, "snippets")
				if err := os.MkdirAll(snipDst, 0755); err != nil {
					return err
				}
				for _, se := range snippetEntries {
					src := "templates/snippets/" + se.Name()
					dst := filepath.Join(snipDst, se.Name())
					if err := ensureFile(assets.Templates, src, dst); err != nil {
						return fmt.Errorf("スニペット %s の展開に失敗: %w", se.Name(), err)
					}
				}
			}
			continue
		}
		src := "templates/" + entry.Name()
		dst := filepath.Join(dstDir, entry.Name())
		if err := ensureFile(assets.Templates, src, dst); err != nil {
			return fmt.Errorf("テンプレート %s の展開に失敗: %w", entry.Name(), err)
		}
	}

	return nil
}
