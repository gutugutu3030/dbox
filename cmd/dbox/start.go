package main

import (
	"fmt"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

var startPublish []string
var startNvim bool

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

func init() {
	startCmd.Flags().StringArrayVar(&startPublish, "publish", nil, "ポートを公開 (複数指定可, 例: 8080 または 3000:8080)")
	startCmd.Flags().BoolVar(&startNvim, "nvim", false, "nvim で起動する（デフォルト: sbx run）")
}

// runStart は start コマンドのメイン処理
func runStart(cmd *cobra.Command, args []string) error {
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

	sandboxInfo, err := sb.FindByName(name)
	if err != nil {
		return err
	}

	if sandboxInfo == nil {
		fmt.Printf("サンドボックス %s が見つかりません。新規作成します...\n", name)
		projectCfg, err := config.LoadProjectConfig(".")
		if err != nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: %w", err)
		}
		params := sandbox.CreateParams{
			Name:         name,
			Template:     projectCfg.Template,
			Agent:        projectCfg.Agent,
			Path:         ".",
			Clone:        projectCfg.Clone,
			CPUs:         projectCfg.Resources.CPUs,
			Memory:       projectCfg.Resources.Memory,
			PublishPorts: startPublish,
		}
		if _, err := sb.Create(params); err != nil {
			return err
		}
		if err := publishPorts(sb, name, startPublish); err != nil {
			return err
		}
		if err := syncNvimConfig(sb, name); err != nil {
			return err
		}
		if err := applyNetworkPolicies(sb, name, "."); err != nil {
			return err
		}
	} else if sandboxInfo.Status == "stopped" {
		fmt.Printf("サンドボックス %s は停止中です。起動します...\n", name)
		if err := publishPorts(sb, name, startPublish); err != nil {
			return err
		}
		if err := applyNetworkPolicies(sb, name, "."); err != nil {
			return err
		}
	} else {
		fmt.Printf("サンドボックス %s にアタッチします...\n", name)
		if err := publishPorts(sb, name, startPublish); err != nil {
			return err
		}
		if err := applyNetworkPolicies(sb, name, "."); err != nil {
			return err
		}
	}

	if startNvim {
		return sb.RunCommand(name, "nvim")
	}
	return sb.Run(name)
}
