package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
)

// InstallCommand は MCP サーバーのインストールコマンドを構築する
func InstallCommand(servers []config.MCPServer) string {
	if len(servers) == 0 {
		return ""
	}
	var pkgs []string
	for _, s := range servers {
		if s.Package != "" {
			pkgs = append(pkgs, s.Package)
		}
	}
	if len(pkgs) == 0 {
		return ""
	}
	return "npm install -g " + strings.Join(pkgs, " ")
}

// OpenCodeMCPConfig は opencode 用の MCP 設定マップを生成する
type OpenCodeMCPConfig map[string]OpenCodeMCPServer

// OpenCodeMCPServer は opencode 設定内の個別 MCP サーバー定義
type OpenCodeMCPServer struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// GenerateOpenCodeConfig は opencode の MCP 設定を生成する
func GenerateOpenCodeConfig(servers []config.MCPServer) OpenCodeMCPConfig {
	cfg := make(OpenCodeMCPConfig)
	for _, s := range servers {
		if s.Name == "" || s.Package == "" {
			continue
		}
		cfg[s.Name] = OpenCodeMCPServer{
			Type:    "local",
			Command: "npx",
			Args:    []string{"-y", s.Package},
		}
	}
	return cfg
}

// OpenCodeConfigPath は .opencode/config.json のパスを返す
func OpenCodeConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".opencode", "config.json")
}

// EnsureOpenCodeDir は .opencode ディレクトリを作成する
func EnsureOpenCodeDir(projectDir string) error {
	dir := filepath.Join(projectDir, ".opencode")
	return os.MkdirAll(dir, 0755)
}

// WriteOpenCodeConfig は opencode の MCP 設定を .opencode/config.json に書き込む
func WriteOpenCodeConfig(projectDir string, mcpCfg OpenCodeMCPConfig) error {
	if len(mcpCfg) == 0 {
		return nil
	}
	dir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf(".opencode ディレクトリの作成に失敗: %w", err)
	}

	path := filepath.Join(dir, "config.json")

	// 既存の設定を読み込む
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("既存の opencode 設定のパースに失敗: %w", err)
		}
	}

	// MCP 設定をマージ
	mcpSection := map[string]interface{}{
		"servers": mcpCfg,
	}
	existing["mcp"] = mcpSection

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("opencode 設定のシリアライズに失敗: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("opencode 設定の書き込みに失敗: %w", err)
	}
	return nil
}

// MergeMCPDomains は MCP サーバーのドメインリストを allowed_domains にマージする。
// 重複は除去される
func MergeMCPDomains(servers []config.MCPServer, existing []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, d := range existing {
		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
			result = append(result, d)
		}
	}
	for _, s := range servers {
		for _, d := range s.Domains {
			if _, ok := seen[d]; !ok {
				seen[d] = struct{}{}
				result = append(result, d)
			}
		}
	}
	return result
}

// HasNodeJS は MCP サーバーが Node.js のランタイムを必要とするか判定する
func HasNodeJS(servers []config.MCPServer) bool {
	for _, s := range servers {
		if s.Package != "" {
			return true
		}
	}
	return false
}
