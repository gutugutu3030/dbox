package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/mcp"
	"github.com/gutugutu3030/sbx-template/internal/prompt"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"
)

// runInteractivePrompts は init 時の対話的入力を処理する
func runInteractivePrompts(dir string, langs []string) (string, []config.MCPServer, error) {
	// 非対話モードの場合は既定値/フラグ値を返す
	if initNoInteractive || !prompt.IsInteractive() {
		agent, _ := resolveAgent(initAgent)
		mcpServers := parseMCPFlags()
		return agent, mcpServers, nil
	}

	// 言語確認
	if len(langs) > 0 {
		ok, err := prompt.ConfirmLanguages(langs)
		if err != nil {
			return "", nil, err
		}
		if !ok {
			fmt.Println("初期化をキャンセルしました")
			os.Exit(0)
		}
	}

	// エージェント選択（フラグ指定がない場合のみ）
	agent := initAgent
	if agent == "" {
		agentAns, err := prompt.SelectAgent("opencode")
		if err != nil {
			return "", nil, fmt.Errorf("エージェント選択に失敗: %w", err)
		}
		agent = agentAns.Agent
	}

	// MCP サーバー選択（フラグ指定がない場合のみ）
	mcpServers := parseMCPFlags()
	if !initNoMCP && len(initMCP) == 0 {
		mcpAns, err := prompt.SelectMCPServers()
		if err != nil {
			return "", nil, fmt.Errorf("MCP サーバー選択に失敗: %w", err)
		}
		mcpServers = mcpAns.Servers
	}

	return agent, mcpServers, nil
}

// parseMCPFlags は --mcp フラグと --no-mcp フラグを解釈する
func parseMCPFlags() []config.MCPServer {
	if initNoMCP {
		return nil
	}
	if len(initMCP) == 0 {
		return nil
	}

	var servers []config.MCPServer
	seen := make(map[string]bool)
	for _, name := range initMCP {
		parts := strings.Split(name, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true

			// プリセット名として解釈
			if preset := mcp.PresetByName(p); preset != nil {
				servers = append(servers, preset.ToMCPServer())
				continue
			}
			// npx パッケージ名として解釈
			servers = append(servers, config.MCPServer{
				Name:    p,
				Package: p,
			})
		}
	}
	return servers
}

// setupMCPServers はサンドボックス内に MCP サーバーをインストールし、
// opencode の設定ファイルを生成する
func setupMCPServers(sb sandbox.SandboxOperator, sandboxName, projectDir string, servers []config.MCPServer) error {
	if len(servers) == 0 {
		return nil
	}

	// MCP パッケージのインストール
	cmd := mcp.InstallCommand(servers)
	if cmd != "" {
		fmt.Println("MCP サーバーをインストール中...")
		out, err := sb.Exec(sandboxName, cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: MCP パッケージのインストールに失敗: %v\n", err)
			if out != "" {
				fmt.Fprintf(os.Stderr, "出力: %s\n", out)
			}
			return nil
		}
		fmt.Printf("MCP パッケージをインストールしました: %s\n", out)
	}

	// opencode 設定を生成
	mcpCfg := mcp.GenerateOpenCodeConfig(servers)
	if err := mcp.WriteOpenCodeConfig(projectDir, mcpCfg); err != nil {
		fmt.Fprintf(os.Stderr, "警告: opencode 設定の書き込みに失敗: %v\n", err)
		return nil
	}
	fmt.Println(".opencode/config.json を更新しました")
	return nil
}

// hasLang は文字列スライスに特定の言語が含まれているか判定する
func hasLang(langs []string, lang string) bool {
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}
