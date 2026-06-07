package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
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
func (b *Builder) dockerBuild(tag, dockerfile string) error {
	args := []string{"build", "-t", tag, "-f", dockerfile, b.TemplatesDir}
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

// EnsureNvimConfig は ~/.config/nvim の設定を ~/.config/dbox/nvim/ にコピーする。
// コピー先が既に存在する場合はスキップする
func EnsureNvimConfig() error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	src := cfg.Nvim.ConfigSource
	if src == "" {
		return nil
	}

	dst, err := config.NvimConfigDir()
	if err != nil {
		return err
	}

	// コピー先が既に存在する場合はスキップ
	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	// コピー元が存在しない場合はスキップ
	if _, err := os.Stat(src); err != nil {
		return nil
	}

	// nvim 設定ディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// cp -r でコピー
	cmd := exec.Command("cp", "-r", src, dst)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nvim 設定のコピーに失敗 (%s -> %s): %w", src, dst, err)
	}

	return nil
}

// BuildBase はベースイメージをビルドする。
// base.Dockerfile を元に nvim がインストールされたイメージを作成する
func (b *Builder) BuildBase() error {
	if err := EnsureNvimConfig(); err != nil {
		return err
	}

	dockerfile := filepath.Join(b.TemplatesDir, "base.Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("base.Dockerfile が見つかりません: %w", err)
	}

	tag := "dbox-base:latest"
	if err := b.dockerBuild(tag, dockerfile); err != nil {
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
	if err := b.dockerBuild(tag, dockerfile); err != nil {
		return err
	}

	fmt.Printf("言語イメージ %s をビルドしました\n", tag)
	return nil
}

// SaveTemplate はビルドしたイメージを sbx テンプレートとして保存する
func (b *Builder) SaveTemplate(tag string) error {
	args := []string{"template", "save", tag}
	cmd := exec.Command("sbx", args...)

	if b.DryRun {
		fmt.Printf("[dry-run] sbx %s\n", strings.Join(args, " "))
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("テンプレート %s の保存に失敗: %w", tag, err)
	}

	fmt.Printf("テンプレート %s を保存しました\n", tag)
	return nil
}

// BuildAndSaveAll は全言語のテンプレートをビルドして保存する
func (b *Builder) BuildAndSaveAll() error {
	languages := []string{"node", "python", "java", "go", "rust", "ruby"}

	for _, lang := range languages {
		fmt.Printf("=== %s テンプレートをビルド中 ===\n", lang)
		if err := b.BuildLang(lang); err != nil {
			return err
		}

		tag := fmt.Sprintf("dbox-%s:latest", lang)
		if err := b.SaveTemplate(tag); err != nil {
			return err
		}
	}

	return nil
}
