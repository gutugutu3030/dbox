package main

import (
	"fmt"
	"time"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

var startPublish []string

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
}

// runStart は start コマンドのメイン処理
func runStart(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)

	var name string
	var projectCfg *config.ProjectConfig
	if len(args) > 0 {
		name = args[0]
	} else {
		var err error
		projectCfg, err = config.LoadProjectConfig(".")
		if err != nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: %w", err)
		}
		name = projectCfg.SandboxName
	}

	if name == "" {
		return fmt.Errorf("サンドボックス名が指定されていません")
	}

	return doStart(sb, name, startPublish, projectCfg)
}

// doStart は runStart の実体。テスト可能にするため SandboxOperator と ProjectConfig を受け取る
func doStart(sb sandbox.SandboxOperator, name string, ports []string, projectCfg *config.ProjectConfig) error {
	sandboxInfo, err := sb.FindByName(name)
	if err != nil {
		return err
	}

	if sandboxInfo == nil {
		fmt.Printf("サンドボックス %s が見つかりません。新規作成します...\n", name)
		if projectCfg == nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: 設定が指定されていません")
		}
		params := sandbox.CreateParams{
			Name:         name,
			Template:     projectCfg.Template,
			Agent:        projectCfg.Agent,
			Path:         ".",
			Clone:        projectCfg.Clone,
			CPUs:         projectCfg.Resources.CPUs,
			Memory:       projectCfg.Resources.Memory,
			PublishPorts: ports,
		}
		if _, err := sb.Create(params); err != nil {
			return err
		}
		// 作成直後は DinD 初期化中で exec が使えない場合があるため待機
		if err := sb.WaitForExec(name, 60*time.Second); err != nil {
			return err
		}
		if err := publishPorts(sb, name, ports); err != nil {
			return err
		}
		domains := config.MergeDomains(projectCfg)
		if err := applyNetworkPolicies(sb, name, domains); err != nil {
			return err
		}
		return sb.Run(name)
	}

	if sandboxInfo.Status == "stopped" {
		fmt.Printf("サンドボックス %s は停止中です。起動します...\n", name)
		if projectCfg == nil {
			return fmt.Errorf(".dbox.yaml が見つかりません: 設定が指定されていません")
		}
		domains := config.MergeDomains(projectCfg)
		if err := applyNetworkPolicies(sb, name, domains); err != nil {
			return err
		}
		// ポート公開は停止中には適用できないため、起動後のアタッチ時に実行する
		return sb.Run(name)
	}

	// running: コンテナは起動済みだが exec 準備ができるまで待機
	fmt.Printf("サンドボックス %s にアタッチします...\n", name)
	if projectCfg == nil {
		return fmt.Errorf(".dbox.yaml が見つかりません: 設定が指定されていません")
	}
	if err := sb.WaitForExec(name, 60*time.Second); err != nil {
		return err
	}
	if err := publishPorts(sb, name, ports); err != nil {
		return err
	}
	domains := config.MergeDomains(projectCfg)
	if err := applyNetworkPolicies(sb, name, domains); err != nil {
		return err
	}
	return sb.Run(name)
}
