package main

import (
	"fmt"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

// startCmd はサンドボックスを起動する
var startCmd = &cobra.Command{
	Use:   "start [sandbox-name]",
	Short: "サンドボックスを起動する",
	Long: `.dbox.yaml に基づいてサンドボックスを起動します。
サンドボックスが既に存在する場合はアタッチ、停止中の場合は開始、
存在しない場合は新規作成します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

// runStart は start コマンドのメイン処理
func runStart(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)

	// サンドボックス名を決定
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

	sandboxInfo, err := sb.FindByName(name)
	if err != nil {
		return err
	}

	switch {
	case sandboxInfo == nil:
		// サンドボックスが存在しない場合、新規作成
		fmt.Printf("サンドボックス %s が見つかりません。新規作成します...\n", name)
		projectCfg, err := config.LoadProjectConfig(".")
		if err != nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: %w", err)
		}
		params := sandbox.CreateParams{
			Name:     name,
			Template: projectCfg.Template,
			Agent:    projectCfg.Agent,
			Path:     ".",
			Clone:    projectCfg.Clone,
			CPUs:     projectCfg.Resources.CPUs,
			Memory:   projectCfg.Resources.Memory,
		}
		if _, err := sb.Create(params); err != nil {
			return err
		}
		return sb.Run(name)

	case sandboxInfo.Status == "stopped":
		// 停止中の場合、起動してからアタッチ
		fmt.Printf("サンドボックス %s は停止中です。起動します...\n", name)
		if err := sb.Start(name); err != nil {
			return err
		}
		return sb.Run(name)

	default:
		// running またはその他の状態 → アタッチ
		fmt.Printf("サンドボックス %s にアタッチします...\n", name)
		return sb.Run(name)
	}
}
