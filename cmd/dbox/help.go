package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// helpCmd は組み込みの help コマンドを置き換える
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "dbox のヘルプを表示する",
	Long: `dbox の使い方と全コマンドの一覧を表示します。
サブコマンドを指定すると、そのコマンドの詳細ヘルプを表示します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHelp,
}

// runHelp は help コマンドのメイン処理
func runHelp(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		targetCmd, _, err := rootCmd.Find(args)
		if err != nil {
			return fmt.Errorf("不明なコマンドです: %s\n  dbox help でコマンド一覧を表示できます", args[0])
		}
		return targetCmd.Help()
	}
	printHelpOverview()
	return nil
}

func init() {
	// 引数なしの dbox で currated なヘルプを表示
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		printHelpOverview()
		return nil
	}
	// 組み込みの help コマンドを独自のものに置き換え
	rootCmd.SetHelpCommand(helpCmd)
}

// printHelpOverview は currated なヘルプ一覧を表示する
func printHelpOverview() {
	fmt.Print(helpHeader)
	fmt.Print(helpUsage)
	fmt.Print(helpCommands)
	fmt.Print(helpExamples)
	fmt.Print(helpFooter)
}

const helpHeader = `  dbox - Docker Sandbox Wrapper CLI
  ================================
  dbox は sbx (Docker Sandboxes) の軽量ラッパーCLIです。
  言語検出 → テンプレート自動ビルド → サンドボックス起動 を
  一連のコマンドで実現します。

`

const helpUsage = `  使い方:
    dbox <command> [flags] [args]
    dbox help [command]     # 詳細ヘルプ

`

const helpCommands = `  コマンド一覧:
    init      プロジェクトを初期化しサンドボックスを作成する
    start     サンドボックスを起動する（デフォルト: nvim 起動）
    stop      サンドボックスを停止する
    exec      サンドボックス内でコマンドを実行する
    template  テンプレートを管理する (build, ls)
    help      このヘルプを表示する

  グローバルフラグ:
    -n, --dry-run  実際のコマンドを実行せずに表示のみ行う
    -h, --help     ヘルプを表示する

`

const helpExamples = `  主な使用例:
    # カレントディレクトリの言語を自動検出し初期化
    dbox init

    # エージェントと言語を指定
    dbox init --agent=codex --lang=python

    # 特定のディレクトリを指定
    dbox init --agent=opencode ./my-project

    # サンドボックスを起動（nvim が開く）
    dbox start

    # サンドボックス内でコマンド実行
    dbox exec "node --version"

    # 全テンプレートをビルド
    dbox template build --lang=all

    # 詳細は dbox help <command> または dbox <command> --help

`

const helpFooter = `  設定ファイル:
    ~/.config/dbox/config.yaml    グローバル設定
    .dbox.yaml                    プロジェクト設定（自動生成）

  詳しくは README.md を参照してください。
`

func init() {
	// 各コマンドに使用例を追加して --help を充実させる
	initCmd.Example = strings.TrimSpace(`
  # カレントディレクトリの言語を自動検出
  dbox init

  # エージェントと言語を指定
  dbox init --agent=codex --lang=python

  # 特定のディレクトリで初期化
  dbox init --agent=opencode ./my-project

  # dry-run モードで確認
  dbox init --dry-run`)

	startCmd.Example = strings.TrimSpace(`
  # カレントディレクトリのサンドボックスを起動
  dbox start

  # サンドボックス名を指定
  dbox start dbox-opencode-my-project

  # dry-run モードで確認
  dbox start --dry-run`)

	stopCmd.Example = strings.TrimSpace(`
  # カレントディレクトリのサンドボックスを停止
  dbox stop

  # サンドボックス名を指定
  dbox stop dbox-opencode-my-project`)

	execCmd.Example = strings.TrimSpace(`
  # サンドボックス内でコマンド実行
  dbox exec "node --version"
  dbox exec "ls -la /workspace"
  dbox exec "npm test"`)

	templateBuildCmd.Example = strings.TrimSpace(`
  # ベースイメージをビルド
  dbox template build

  # Node.js テンプレートをビルド
  dbox template build --lang=node

  # 全言語を一括ビルド
  dbox template build --lang=all

  # dry-run モードで確認
  dbox template build --lang=node --dry-run`)
}
