package main

import (
	"fmt"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

// execCmd はサンドボックス内でコマンドを実行する
var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "サンドボックス内でコマンドを実行する",
	Long:  `.dbox.yaml に基づいてサンドボックス内でコマンドを実行します。`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExec,
}

// runExec は exec コマンドのメイン処理
func runExec(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)

	projectCfg, err := config.LoadProjectConfig(".")
	if err != nil {
		return fmt.Errorf(".dbox.yaml が見つかりません: %w", err)
	}

	command := strings.Join(args, " ")
	out, err := sb.Exec(projectCfg.SandboxName, command)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}
