package mcp

import (
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
)

// Preset は MCP サーバーのプリセット定義を表す
type Preset struct {
	Name        string
	Package     string
	Domains     []string
	Description string
}

// Presets は定義済みの MCP サーバープリセット一覧
var Presets = []Preset{
	{
		Name:        "GitNexus",
		Package:     "@anthropic/gitnexus",
		Domains:     []string{"api.github.com:443"},
		Description: "Git操作の強化（リポジトリ管理、コードレビュー等）",
	},
	{
		Name:        "Context7",
		Package:     "@upstash/context7-mcp",
		Domains:     []string{"api.context7.com:443"},
		Description: "コンテキスト削減（必要なコード断片のみを取得）",
	},
	{
		Name:        "Fetch",
		Package:     "@anthropic/mcp-fetch",
		Domains:     nil, // 全ドメイン許可（sbx の既定のネットワークポリシーに委譲）
		Description: "Webページの取得と解析",
	},
}

// PresetByName はプリセット名から Preset を検索する（大文字小文字を区別しない）
func PresetByName(name string) *Preset {
	lower := strings.ToLower(name)
	for _, p := range Presets {
		if strings.ToLower(p.Name) == lower {
			return &p
		}
	}
	return nil
}

// ToMCPServer はプリセットを MCPServer に変換する
func (p *Preset) ToMCPServer() config.MCPServer {
	return config.MCPServer{
		Name:    p.Name,
		Package: p.Package,
		Domains: p.Domains,
	}
}
