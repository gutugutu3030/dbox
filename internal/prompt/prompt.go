package prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/mcp"
)

// AgentAnswers はエージェント選択の結果を保持する
type AgentAnswers struct {
	Agent       string
	IsCustom    bool // カスタム入力が選択された場合 true
	CustomAgent string
}

// SelectAgent は利用する AI エージェントを選択するプロンプトを表示する。
// ターミナルがない場合は既定値を返す
func SelectAgent(defaultAgent string) (*AgentAnswers, error) {
	agentOptions := []string{
		"opencode",
		"codex",
		"claude",
		"cursor",
		"github_copilot",
		"（カスタム入力...）",
	}

	var selected string
	agentPrompt := &survey.Select{
		Message: "AI エージェントを選択してください:",
		Options: agentOptions,
		Default: defaultAgent,
	}
	if err := survey.AskOne(agentPrompt, &selected); err != nil {
		return nil, err
	}

	ans := &AgentAnswers{}

	if selected == "（カスタム入力...）" {
		var custom string
		input := &survey.Input{
			Message: "エージェント名を入力してください:",
		}
		if err := survey.AskOne(input, &custom); err != nil {
			return nil, err
		}
		custom = strings.TrimSpace(custom)
		if custom == "" {
			ans.Agent = defaultAgent
		} else {
			ans.Agent = custom
			ans.IsCustom = true
			ans.CustomAgent = custom
		}
	} else {
		ans.Agent = selected
	}

	return ans, nil
}

// MCPAnswers は MCP サーバー選択の結果を保持する
type MCPAnswers struct {
	Servers   []config.MCPServer
	CustomPkg string
}

// customMCPOptionLabel はカスタム MCP モジュール選択肢のラベル
const customMCPOptionLabel = "カスタム npx モジュールを入力..."

// noneMCPOptionLabel は MCP なし選択肢のラベル
const noneMCPOptionLabel = "なし（MCP サーバーを追加しない）"

// SelectMCPServers は MCP サーバーを複数選択するプロンプトを表示する
func SelectMCPServers() (*MCPAnswers, error) {
	// プリセット + 特殊選択肢で選択肢を構築
	options := make([]string, 0, len(mcp.Presets)+2)
	options = append(options, noneMCPOptionLabel)
	for _, p := range mcp.Presets {
		options = append(options, fmt.Sprintf("%s (%s)", p.Name, p.Description))
	}
	options = append(options, customMCPOptionLabel)

	var selected []string
	mcpPrompt := &survey.MultiSelect{
		Message: "MCP サーバーを選択してください（なしの場合は「なし」のみ選択）:",
		Options: options,
	}
	if err := survey.AskOne(mcpPrompt, &selected); err != nil {
		return nil, err
	}

	ans := &MCPAnswers{}

	for _, s := range selected {
		if s == noneMCPOptionLabel {
			// 「なし」が選択された場合は他を無視
			ans.Servers = nil
			return ans, nil
		}
		if s == customMCPOptionLabel {
			// カスタム入力を促す
			var pkg string
			input := &survey.Input{
				Message: "npx パッケージ名を入力してください（例: @myorg/my-mcp）:",
			}
			if err := survey.AskOne(input, &pkg); err != nil {
				return nil, err
			}
			pkg = strings.TrimSpace(pkg)
			if pkg != "" {
				name := extractNameFromPackage(pkg)
				ans.Servers = append(ans.Servers, config.MCPServer{
					Name:    name,
					Package: pkg,
				})
				ans.CustomPkg = pkg
			}
			continue
		}
		// プリセット名を抽出（"GitNexus (説明)" → "GitNexus"）
		name := strings.SplitN(s, " ", 2)[0]
		if preset := mcp.PresetByName(name); preset != nil {
			ans.Servers = append(ans.Servers, preset.ToMCPServer())
		}
	}

	return ans, nil
}

// ConfirmLanguages は検出された言語の確認プロンプトを表示する。
// ユーザーがキャンセルした場合は false を返す
func ConfirmLanguages(langs []string) (bool, error) {
	if len(langs) == 0 {
		return true, nil
	}
	langStr := strings.Join(langs, ", ")
	confirm := &survey.Confirm{
		Message: fmt.Sprintf("言語を検出しました: [%s] このまま進めますか？", langStr),
		Default: true,
	}
	var result bool
	if err := survey.AskOne(confirm, &result); err != nil {
		return false, err
	}
	return result, nil
}

// IsInteractive はターミナルが対話モードかを判定する
func IsInteractive() bool {
	// CI 環境やパイプ経由の場合は非対話と判定
	if os.Getenv("CI") != "" {
		return false
	}
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// extractNameFromPackage は npm パッケージ名から MCP サーバー名を抽出する
func extractNameFromPackage(pkg string) string {
	// @scope/name → name
	// name → name
	if idx := strings.LastIndex(pkg, "/"); idx >= 0 {
		return pkg[idx+1:]
	}
	return pkg
}
