package main

import (
	"fmt"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpCmd は MCP サーバーの管理を行うサブコマンド
var mcpCmd = &cobra.Command{
	Use:   "mcp <command>",
	Short: "MCP サーバーを管理する",
	Long: `サンドボックスで使用する MCP サーバーの追加、削除、一覧表示を行います。
dbox init 時に選択した MCP サーバーは .dbox.yaml に保存され、
サンドボックス作成時に自動でインストールされます。`,
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name> <package>",
	Short: "MCP サーバーを追加する",
	Long:  `.dbox.yaml に MCP サーバーを追加します。`,
	Args:  cobra.ExactArgs(2),
	RunE:  runMCPAdd,
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "MCP サーバーを削除する",
	Long:  `.dbox.yaml から MCP サーバーを削除します。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMCPRemove,
}

var mcpListCmd = &cobra.Command{
	Use:   "ls",
	Short: "MCP サーバー一覧を表示する",
	Long:  `.dbox.yaml に登録されている MCP サーバーと利用可能なプリセットを表示します。`,
	RunE:  runMCPList,
}

var mcpAddDomain []string
var mcpDir string

func init() {
	mcpAddCmd.Flags().StringArrayVar(&mcpAddDomain, "domain", nil, "MCP サーバーが必要とするネットワークドメイン（複数指定可）")
	mcpCmd.PersistentFlags().StringVar(&mcpDir, "dir", ".", "プロジェクトディレクトリ（.dbox.yaml の場所）")
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpCmd.AddCommand(mcpListCmd)
}

// runMCPAdd は MCP サーバーを追加する
func runMCPAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	pkg := args[1]

	dir := mcpDir
	projectCfg, err := config.LoadProjectConfig(dir)
	if err != nil {
		return fmt.Errorf(".dbox.yaml の読み込みに失敗: %w", err)
	}

	// 重複チェック
	for _, s := range projectCfg.MCP.Servers {
		if s.Name == name {
			return fmt.Errorf("MCP サーバー %s は既に登録されています", name)
		}
	}

	server := config.MCPServer{
		Name:    name,
		Package: pkg,
		Domains: mcpAddDomain,
	}
	projectCfg.MCP.Servers = append(projectCfg.MCP.Servers, server)

	if err := config.SaveProjectConfig(dir, projectCfg); err != nil {
		return fmt.Errorf(".dbox.yaml の保存に失敗: %w", err)
	}

	fmt.Printf("MCP サーバー %s (%s) を追加しました\n", name, pkg)
	if len(mcpAddDomain) > 0 {
		fmt.Printf("  許可ドメイン: %v\n", mcpAddDomain)
	}
	return nil
}

// runMCPRemove は MCP サーバーを削除する
func runMCPRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	dir := mcpDir
	projectCfg, err := config.LoadProjectConfig(dir)
	if err != nil {
		return fmt.Errorf(".dbox.yaml の読み込みに失敗: %w", err)
	}

	found := false
	servers := make([]config.MCPServer, 0, len(projectCfg.MCP.Servers))
	for _, s := range projectCfg.MCP.Servers {
		if s.Name == name {
			found = true
			continue
		}
		servers = append(servers, s)
	}
	if !found {
		return fmt.Errorf("MCP サーバー %s は登録されていません", name)
	}

	projectCfg.MCP.Servers = servers
	if err := config.SaveProjectConfig(dir, projectCfg); err != nil {
		return fmt.Errorf(".dbox.yaml の保存に失敗: %w", err)
	}

	fmt.Printf("MCP サーバー %s を削除しました\n", name)
	return nil
}

// runMCPList は MCP サーバー一覧を表示する
func runMCPList(cmd *cobra.Command, args []string) error {
	dir := mcpDir

	fmt.Println("利用可能な MCP プリセット:")
	fmt.Println(strings.Repeat("-", 60))
	for _, p := range mcp.Presets {
		fmt.Printf("  %-15s %s\n", p.Name, p.Description)
		if len(p.Domains) > 0 {
			fmt.Printf("  %-15s ドメイン: %v\n", "", p.Domains)
		}
	}

	projectCfg, err := config.LoadProjectConfig(dir)
	if err != nil {
		fmt.Println("\n(.dbox.yaml が存在しないため、登録済み MCP は表示できません)")
		return nil
	}

	if len(projectCfg.MCP.Servers) == 0 {
		fmt.Println("\n登録済みの MCP サーバーはありません")
		return nil
	}

	fmt.Println("\n登録済みの MCP サーバー:")
	fmt.Println(strings.Repeat("-", 60))
	for _, s := range projectCfg.MCP.Servers {
		fmt.Printf("  %-15s %s\n", s.Name, s.Package)
		if len(s.Domains) > 0 {
			fmt.Printf("  %-15s ドメイン: %v\n", "", s.Domains)
		}
	}
	return nil
}
