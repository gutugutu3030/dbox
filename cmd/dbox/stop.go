package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

var stopAll bool

// stopCmd はサンドボックスを停止する
var stopCmd = &cobra.Command{
	Use:   "stop [sandbox-name]",
	Short: "サンドボックスを停止する",
	Long: `.dbox.yaml に基づいてサンドボックスを停止します。
--all を指定すると dbox- で始まる全サンドボックスを停止します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStop,
}

func init() {
	stopCmd.Flags().BoolVarP(&stopAll, "all", "a", false, "dbox- で始まる全サンドボックスを停止する")
}

// runStop は stop コマンドのメイン処理
func runStop(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)

	if stopAll {
		return stopAllSandboxes(sb)
	}

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

// stopAllSandboxes は dbox- で始まる全サンドボックスを停止する
func stopAllSandboxes(sb *sandbox.Runner) error {
	sandboxes, err := sb.List()
	if err != nil {
		return fmt.Errorf("サンドボックス一覧の取得に失敗: %w", err)
	}

	var stopped int
	for _, s := range sandboxes {
		if !strings.HasPrefix(s.Sandbox, "dbox-") {
			continue
		}
		if s.Status == "stopped" {
			continue
		}
		fmt.Printf("サンドボックス %s を停止中...\n", s.Sandbox)
		if err := sb.Stop(s.Sandbox); err != nil {
			fmt.Fprintf(os.Stderr, "警告: %s の停止に失敗: %v\n", s.Sandbox, err)
			continue
		}
		stopped++
	}

	if stopped == 0 {
		fmt.Println("停止対象の dbox サンドボックスはありません")
	} else {
		fmt.Printf("%d 個のサンドボックスを停止しました\n", stopped)
	}
	return nil
}
