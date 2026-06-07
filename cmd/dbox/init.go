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
	initCmd.Flags().StringVarP(&initLang, "lang", "l", "auto", "使用言語 (auto:自動検出, node, go, node,python などカンマ区切りで複数指定可)")
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
	langs, err := resolveLanguages(absDir, initLang)
	if err != nil {
		return err
	}

	// エージェント名を決定
	agent, err := resolveAgent(initAgent)
	if err != nil {
		return err
	}

	// テンプレート名を生成
	templateName := detect.TemplateNameForLangs(langs)

	// テンプレートの存在確認、なければビルド
	if err := ensureTemplate(langs); err != nil {
		return err
	}

	// サンドボックス名を生成
	sandboxName := generateSandboxName(agent, absDir)

	// 言語名を文字列配列に変換
	langStrs := make([]string, len(langs))
	for i, l := range langs {
		langStrs[i] = string(l)
	}

	// プロジェクト設定を保存
	projectCfg := &config.ProjectConfig{
		Version:     2,
		Agent:       agent,
		Langs:       langStrs,
		Template:    templateName,
		SandboxName: sandboxName,
		Clone:       true,
		Resources: config.ResourceConfig{
			CPUs:   0,
			Memory: "",
		},
	}

	if err := config.SaveProjectConfig(absDir, projectCfg); err != nil {
		return fmt.Errorf(".dbox.yaml の保存に失敗: %w", err)
	}
	fmt.Printf(".dbox.yaml を作成しました (agent=%s, langs=%v)\n", agent, langStrs)

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

// resolveLanguages は言語指定または自動検出により言語を決定する
func resolveLanguages(dir, langFlag string) ([]detect.Language, error) {
	if langFlag != "auto" {
		// カンマ区切りで複数言語指定に対応
		langs := detect.LangsFromString(langFlag)
		if len(langs) == 0 {
			return nil, fmt.Errorf("有効な言語が指定されていません: %s", langFlag)
		}
		return langs, nil
	}

	// 自動検出
	result := detect.Detect(dir)
	fmt.Printf("言語検出結果: %v\n", result.Languages)
	return result.Languages, nil
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

// ensureTemplate は言語に対応するテンプレートが sbx に存在するか確認し、
// なければ Docker イメージをビルドして sbx にロードする
func ensureTemplate(langs []detect.Language) error {
	tmplName := detect.TemplateNameForLangs(langs)
	tag := tmplName + ":latest"
	templatesDir := findTemplatesDir()
	builder := template.NewBuilder(templatesDir, dryRun)
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

	if len(langs) == 1 {
		lang := langs[0]
		if lang == detect.LanguageBase {
			if err := builder.BuildBase(); err != nil {
				return err
			}
		} else {
			if err := builder.BuildLang(string(lang)); err != nil {
				return err
			}
		}
	} else {
		langStrs := make([]string, len(langs))
		for i, l := range langs {
			langStrs[i] = string(l)
		}
		composer := template.NewComposer(builder)
		if _, err := composer.Compose(langStrs); err != nil {
			return err
		}
	}

	return sb.TemplateSave(tag)
}

// generateSandboxName はサンドボックス名を生成する
func generateSandboxName(agent, dir string) string {
	base := filepath.Base(dir)
	return fmt.Sprintf("dbox-%s-%s", agent, base)
}

// findTemplatesDir はテンプレートディレクトリのパスを解決する。
// 既存のテンプレートディレクトリがない場合は、グローバル設定ディレクトリに
// 埋め込みテンプレートを展開して使用する
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

	// グローバル設定ディレクトリに展開
	globalDir, _ := config.GlobalConfigDir()
	candidate = filepath.Join(globalDir, "templates")
	if err := os.MkdirAll(candidate, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "警告: テンプレートディレクトリ作成に失敗: %v\n", err)
		return candidate
	}
	if err := template.EnsureTemplatesExtracted(candidate); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 組み込みテンプレートの展開に失敗: %v\n", err)
	}
	return candidate
}
