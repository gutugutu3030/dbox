package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/sandbox"

	"github.com/spf13/cobra"
)

// pruneCmd は dbox 関連リソースを一掃する
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "dbox 関連の全リソースを一掃する",
	Long: `dbox- で始まる全サンドボックス・sbxテンプレート・Dockerイメージを
停止・削除して初期状態に戻します。`,
	RunE: runPrune,
}

// runPrune は prune コマンドのメイン処理
func runPrune(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)
	pruned := 0

	// 1. dbox- サンドボックスを停止して削除
	if n, err := pruneSandboxes(sb); err != nil {
		fmt.Fprintf(os.Stderr, "警告: サンドボックス削除中にエラー: %v\n", err)
	} else {
		pruned += n
	}

	// 2. dbox- sbx テンプレートを削除
	if n, err := pruneTemplates(sb); err != nil {
		fmt.Fprintf(os.Stderr, "警告: テンプレート削除中にエラー: %v\n", err)
	} else {
		pruned += n
	}

	// 3. dbox- Docker イメージを削除
	if n, err := pruneDockerImages(sb); err != nil {
		fmt.Fprintf(os.Stderr, "警告: Dockerイメージ削除中にエラー: %v\n", err)
	} else {
		pruned += n
	}

	if pruned == 0 {
		fmt.Println("削除対象の dbox リソースはありません")
	} else {
		fmt.Printf("%d 個のリソースを削除しました\n", pruned)
	}
	return nil
}

// pruneSandboxes は dbox- サンドボックスを停止・削除する
func pruneSandboxes(sb *sandbox.Runner) (int, error) {
	sandboxes, err := sb.List()
	if err != nil {
		return 0, fmt.Errorf("サンドボックス一覧の取得に失敗: %w", err)
	}

	var count int
	for _, s := range sandboxes {
		if !strings.HasPrefix(s.Sandbox, "dbox-") {
			continue
		}
		if s.Status != "stopped" {
			fmt.Printf("サンドボックス %s を停止中...\n", s.Sandbox)
			if err := sb.Stop(s.Sandbox); err != nil {
				fmt.Fprintf(os.Stderr, "警告: %s の停止に失敗: %v\n", s.Sandbox, err)
			}
		}
		fmt.Printf("サンドボックス %s を削除中...\n", s.Sandbox)
		if err := sb.Remove(s.Sandbox); err != nil {
			return count, fmt.Errorf("%s の削除に失敗: %w", s.Sandbox, err)
		}
		count++
	}
	return count, nil
}

// pruneTemplates は dbox- sbx テンプレートを削除する
func pruneTemplates(sb *sandbox.Runner) (int, error) {
	out, err := sb.TemplateList()
	if err != nil {
		return 0, fmt.Errorf("テンプレート一覧の取得に失敗: %w", err)
	}

	var count int
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "REPOSITORY") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		repo := fields[0]
		tag := fields[1]
		if !strings.HasPrefix(repo, "dbox-") {
			continue
		}
		fullName := repo + ":" + tag
		fmt.Printf("テンプレート %s を削除中...\n", fullName)
		if err := sb.TemplateRemove(fullName); err != nil {
			fmt.Fprintf(os.Stderr, "警告: テンプレート %s の削除に失敗: %v\n", fullName, err)
			continue
		}
		count++
	}
	return count, nil
}

// pruneDockerImages は dbox- Docker イメージを削除する
func pruneDockerImages(sb *sandbox.Runner) (int, error) {
	if sb.DryRun {
		fmt.Println("[dry-run] docker rmi で dbox-* イメージを削除")
		return 0, nil
	}

	cmd := exec.Command("docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}}")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("Dockerイメージ一覧の取得に失敗: %w", err)
	}

	var count int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// docker image ls の出力は "dbox-go:latest" 形式
		repo := strings.Split(line, ":")[0]
		if !strings.HasPrefix(repo, "dbox-") {
			continue
		}
		fmt.Printf("Dockerイメージ %s を削除中...\n", line)
		rmCmd := exec.Command("docker", "rmi", "-f", line)
		rmCmd.Stderr = os.Stderr
		if err := rmCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "警告: Dockerイメージ %s の削除に失敗: %v\n", line, err)
			continue
		}
		count++
	}
	return count, nil
}
