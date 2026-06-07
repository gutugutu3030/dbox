package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/detect"
	"github.com/gutugutu3030/sbx-template/internal/mcp"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"
	"github.com/gutugutu3030/sbx-template/internal/template"

	"github.com/spf13/cobra"
)

var initAgent string
var initLang string
var initPublish []string
var initNoInteractive bool
var initNoMCP bool
var initMCP []string

// initCmd はプロジェクトの初期化とサンドボックス作成を行う
var initCmd = &cobra.Command{
	Use:   "init [--agent=opencode] [--lang=auto] [path]",
	Short: "プロジェクトを初期化しサンドボックスを作成する",
	Long: `指定ディレクトリの使用言語を自動検出し、.dbox.yaml を作成した上で
sbx create を実行してサンドボックスを作成します。
--agent, --mcp フラグを省略すると対話的に選択できます。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&initAgent, "agent", "a", "", "使用するAIエージェント (opencode, codex, claude など)")
	initCmd.Flags().StringVarP(&initLang, "lang", "l", "auto", "使用言語 (auto:自動検出, node, go, node,python などカンマ区切りで複数指定可)")
	initCmd.Flags().StringArrayVar(&initPublish, "publish", nil, "ポートを公開 (複数指定可, 例: 8080 または 3000:8080)")
	initCmd.Flags().BoolVarP(&initNoInteractive, "yes", "y", false, "すべての入力を既定値で進める（非対話モード）")
	initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "MCP サーバーを追加しない")
	initCmd.Flags().StringArrayVar(&initMCP, "mcp", nil, "MCP サーバーを指定（カンマ区切り, 例: gitnexus,context7）")
}

// runInit は init コマンドのメイン処理
func runInit(cmd *cobra.Command, args []string) error {
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("パスの解決に失敗: %w", err)
	}

	langs, err := resolveLanguages(absDir, initLang)
	if err != nil {
		return err
	}
	langStrs := make([]string, len(langs))
	for i, l := range langs {
		langStrs[i] = string(l)
	}

	// 対話的プロンプト
	agent, mcpServers, err := runInteractivePrompts(absDir, langStrs)
	if err != nil {
		return err
	}

	// Node.js が必要な場合は langs に自動追加
	if mcp.HasNodeJS(mcpServers) && !hasLang(langStrs, "node") {
		fmt.Println("MCP サーバーに Node.js が必要なため、node を言語に追加します")
		langStrs = append(langStrs, "node")
		langs = append(langs, detect.LanguageNode)
	}

	templateName := detect.TemplateNameForLangs(langs)

	if err := ensureTemplate(langs); err != nil {
		return err
	}

	sandboxName := generateSandboxName(agent, absDir)

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
		MCP: config.MCPConfig{
			Servers: mcpServers,
		},
	}

	// 既存の allowed_domains を引き継ぐ
	if existingCfg, err := config.LoadProjectConfig(absDir); err == nil {
		if len(existingCfg.Network.AllowedDomains) > 0 {
			projectCfg.Network.AllowedDomains = existingCfg.Network.AllowedDomains
		}
	}

	// MCP サーバーのドメインを allowed_domains にマージ
	projectCfg.Network.AllowedDomains = mcp.MergeMCPDomains(mcpServers, projectCfg.Network.AllowedDomains)

	if err := config.SaveProjectConfig(absDir, projectCfg); err != nil {
		return fmt.Errorf(".dbox.yaml の保存に失敗: %w", err)
	}
	fmt.Printf(".dbox.yaml を作成しました (agent=%s, langs=%v)\n", agent, langStrs)
	if len(mcpServers) > 0 {
		names := make([]string, len(mcpServers))
		for i, s := range mcpServers {
			names[i] = s.Name
		}
		fmt.Printf("  MCP サーバー: %v\n", names)
	}

	sb := sandbox.NewRunner(dryRun)

	existing, err := sb.FindByName(sandboxName)
	if err != nil {
		return fmt.Errorf("サンドボックスの確認に失敗: %w", err)
	}
	if existing != nil {
		fmt.Printf("既存のサンドボックス %s を削除中...\n", sandboxName)
		if err := sb.Stop(sandboxName); err != nil {
			fmt.Fprintf(os.Stderr, "警告: サンドボックスの停止に失敗: %v\n", err)
		}
		if err := sb.Remove(sandboxName); err != nil {
			return fmt.Errorf("既存のサンドボックス %s の削除に失敗: %w", sandboxName, err)
		}
	}

	params := sandbox.CreateParams{
		Name:         sandboxName,
		Template:     projectCfg.Template,
		Agent:        agent,
		Path:         absDir,
		Clone:        true,
		CPUs:         projectCfg.Resources.CPUs,
		Memory:       projectCfg.Resources.Memory,
		PublishPorts: initPublish,
	}

	if _, err := sb.Create(params); err != nil {
		return fmt.Errorf("サンドボックスの作成に失敗: %w", err)
	}

	if err := sb.WaitForExec(sandboxName, 60*time.Second); err != nil {
		return err
	}

	// MCP サーバーのセットアップ
	if err := setupMCPServers(sb, sandboxName, absDir, mcpServers); err != nil {
		return err
	}

	if err := publishPorts(sb, sandboxName, initPublish); err != nil {
		return err
	}

	if err := applyNetworkPoliciesFromFile(sb, sandboxName, absDir); err != nil {
		return err
	}

	fmt.Printf("サンドボックス %s を作成しました\n", sandboxName)
	fmt.Printf("  dbox start でサンドボックスを起動できます\n")
	return nil
}

// resolveLanguages は言語指定または自動検出により言語を決定する
func resolveLanguages(dir, langFlag string) ([]detect.Language, error) {
	if langFlag != "auto" {
		langs := detect.LangsFromString(langFlag)
		if len(langs) == 0 {
			return nil, fmt.Errorf("有効な言語が指定されていません: %s", langFlag)
		}
		return langs, nil
	}

	result := detect.Detect(dir)
	fmt.Printf("言語検出結果: %v\n", result.Languages)
	return result.Languages, nil
}

// resolveAgent はエージェント名を決定する
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

// ensureTemplate はテンプレートの存在確認とビルドを行う
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

// findTemplatesDir はテンプレートディレクトリのパスを解決する
func findTemplatesDir() string {
	execPath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "..", "templates")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	cwd, _ := os.Getwd()
	candidate := filepath.Join(cwd, "templates")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}

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

// publishPorts はサンドボックスのポートを公開する
func publishPorts(sb sandbox.SandboxOperator, sandboxName string, ports []string) error {
	for _, p := range ports {
		fmt.Printf("ポート %s を公開中...\n", p)
		if err := sb.PortPublish(sandboxName, p); err != nil {
			return fmt.Errorf("ポート %s の公開に失敗: %w", p, err)
		}
	}
	return nil
}

// applyNetworkPolicies はドメインリストを許可するポリシーを適用する
func applyNetworkPolicies(sb sandbox.SandboxOperator, sandboxName string, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	fmt.Printf("ネットワークポリシーを適用中 (許可ドメイン: %v)...\n", domains)
	if err := sb.PolicyAllowNetwork(sandboxName, domains); err != nil {
		return fmt.Errorf("ネットワークポリシーの適用に失敗: %w", err)
	}
	return nil
}

// applyNetworkPoliciesFromFile は設定ファイルに基づきポリシーを適用する
func applyNetworkPoliciesFromFile(sb sandbox.SandboxOperator, sandboxName, dir string) error {
	projectCfg, err := config.LoadProjectConfig(dir)
	if err != nil {
		return fmt.Errorf(".dbox.yaml の読み込みに失敗: %w", err)
	}
	return applyNetworkPolicies(sb, sandboxName, config.MergeDomains(projectCfg))
}
