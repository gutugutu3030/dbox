package main

import (
	"os"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/spf13/cobra"
)

var dryRun bool

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "dbox",
	Short: "dbox - Docker Sandbox Wrapper CLI",
	Long: `dbox は sbx (Docker Sandboxes) の軽量ラッパーCLIです。
言語検出からテンプレート作成、サンドボックス起動までを一元管理します。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.EnsureGlobalConfigDir()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "実際のコマンドを実行せずに表示のみ行う")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(templateCmd)
}
