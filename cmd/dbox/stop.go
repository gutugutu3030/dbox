package main

import (
	"fmt"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

// stopCmd はサンドボックスを停止する
var stopCmd = &cobra.Command{
	Use:   "stop [sandbox-name]",
	Short: "サンドボックスを停止する",
	Long:  `.dbox.yaml に基づいてサンドボックスを停止します。`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStop,
}

// runStop は stop コマンドのメイン処理
func runStop(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)

	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		projectCfg, err := config.LoadProjectConfig(".")
		if err != nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: %w", err)
		}
		name = projectCfg.SandboxName
	}

	if name == "" {
		return fmt.Errorf("サンドボックス名が指定されていません")
	}

	return sb.Stop(name)
}
