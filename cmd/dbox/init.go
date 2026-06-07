package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/detect"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"
	"github.com/gutugutu3030/sbx-template/internal/template"

	"github.com/spf13/cobra"
)

var initAgent string
var initLang string

// initCmd はプロジェクトの初期化とサンドボックス作成を行う
var initCmd = &cobra.Command{
	Use:   "init [--agent=opencode] [--lang=auto] [path]",
	Short: "プロジェクトを初期化しサンドボックスを作成する",
	Long: `指定ディレクトリの使用言語を自動検出し、.dbox.yaml を作成した上で
sbx create を実行してサンドボックスを作成します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&initAgent, "agent", "a", "", "使用するAIエージェント (opencode, codex, claude など)")
	initCmd.Flags().StringVarP(&initLang, "lang", "l", "auto", "使用言語 (auto:自動検出, node, python, go など)")
}

// runInit は init コマンドのメイン処理
func runInit(cmd *cobra.Command, args []string) error {
	// 対象ディレクトリを決定
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("パスの解決に失敗: %w", err)
	}

	// 言語を検出または指定された言語を使用
	lang, err := resolveLanguage(absDir, initLang)
	if err != nil {
		return err
	}

	// エージェント名を決定
	agent, err := resolveAgent(initAgent)
	if err != nil {
		return err
	}

	// テンプレートの存在確認、なければビルド
	if err := ensureTemplate(lang); err != nil {
		return err
	}

	// サンドボックス名を生成
	sandboxName := generateSandboxName(agent, absDir)

	// プロジェクト設定を保存
	projectCfg := &config.ProjectConfig{
		Version:     1,
		Agent:       agent,
		Lang:        string(lang),
		Template:    detect.TemplateNameForLang(lang),
		SandboxName: sandboxName,
		Clone:       true,
		Resources: config.ResourceConfig{
			CPUs:   0,
			Memory: "50%",
		},
	}

	if err := config.SaveProjectConfig(absDir, projectCfg); err != nil {
		return fmt.Errorf(".dbox.yaml の保存に失敗: %w", err)
	}
	fmt.Printf(".dbox.yaml を作成しました (agent=%s, lang=%s)\n", agent, lang)

	// sbx create を実行
	sb := sandbox.NewRunner(dryRun)
	params := sandbox.CreateParams{
		Name:     sandboxName,
		Template: projectCfg.Template,
		Agent:    agent,
		Path:     absDir,
		Clone:    true,
		CPUs:     projectCfg.Resources.CPUs,
		Memory:   projectCfg.Resources.Memory,
	}

	if _, err := sb.Create(params); err != nil {
		return fmt.Errorf("サンドボックスの作成に失敗: %w", err)
	}

	fmt.Printf("サンドボックス %s を作成しました\n", sandboxName)
	fmt.Printf("  dbox start でサンドボックスを起動できます\n")
	return nil
}

// resolveLanguage は言語指定または自動検出により言語を決定する
func resolveLanguage(dir, langFlag string) (detect.Language, error) {
	if langFlag != "auto" {
		lang := detect.Language(langFlag)
		if detect.TemplateNameForLang(lang) == "dbox-base" && lang != detect.LanguageBase {
			return "", fmt.Errorf("不明な言語です: %s", langFlag)
		}
		return lang, nil
	}

	// 自動検出
	result := detect.Detect(dir)
	fmt.Printf("言語検出結果: %s (確信度: %.0f%%)\n", result.Language, result.Confidence*100)
	return result.Language, nil
}

// resolveAgent はエージェント名を決定する。
// 指定がなければグローバル設定の既定値を使用する
func resolveAgent(agentFlag string) (string, error) {
	if agentFlag != "" {
		return agentFlag, nil
	}

	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		return "", err
	}
	return globalCfg.DefaultAgent, nil
}

// ensureTemplate は言語に対応するテンプレートが存在するか確認し、
// なければ自動ビルドする
func ensureTemplate(lang detect.Language) error {
	tmplName := detect.TemplateNameForLang(lang)
	sb := sandbox.NewRunner(dryRun)

	exists, err := sb.HasTemplate(tmplName)
	if err != nil {
		return fmt.Errorf("テンプレート確認に失敗: %w", err)
	}
	if exists {
		fmt.Printf("テンプレート %s は既に存在します\n", tmplName)
		return nil
	}

	fmt.Printf("テンプレート %s が見つかりません。ビルドを開始します...\n", tmplName)

	templatesDir := findTemplatesDir()
	builder := template.NewBuilder(templatesDir, dryRun)

	if lang == detect.LanguageBase {
		if err := builder.BuildBase(); err != nil {
			return err
		}
	} else {
		if err := builder.BuildLang(string(lang)); err != nil {
			return err
		}
	}

	tag := tmplName + ":latest"
	if err := builder.SaveTemplate(tag); err != nil {
		return err
	}

	return nil
}

// generateSandboxName はサンドボックス名を生成する
func generateSandboxName(agent, dir string) string {
	base := filepath.Base(dir)
	return fmt.Sprintf("dbox-%s-%s", agent, base)
}

// findTemplatesDir は組み込みテンプレートディレクトリのパスを解決する
func findTemplatesDir() string {
	// 実行ファイルからの相対パスを試す
	execPath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "..", "templates")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	// カレントディレクトリからの相対パスを試す
	cwd, _ := os.Getwd()
	candidate := filepath.Join(cwd, "templates")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}

	// グローバル設定ディレクトリ
	globalDir, _ := config.GlobalConfigDir()
	candidate = filepath.Join(globalDir, "templates")
	os.MkdirAll(candidate, 0755)
	return candidate
}
