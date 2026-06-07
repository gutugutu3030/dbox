package main

import (
	"fmt"

	"github.com/gutugutu3030/sbx-template/internal/sandbox"
	"github.com/gutugutu3030/sbx-template/internal/template"

	"github.com/spf13/cobra"
)

var templateBuildLang string

// templateCmd はテンプレート管理のサブコマンド
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "テンプレートを管理する",
	Long:  `Docker テンプレートをビルド・保存します。`,
}

func init() {
	templateCmd.AddCommand(templateBuildCmd)
	templateCmd.AddCommand(templateListCmd)
	templateBuildCmd.Flags().StringVarP(&templateBuildLang, "lang", "l", "base", "ビルドする言語 (base, node, python, go, rust, java, ruby, all)")
}

// templateBuildCmd はテンプレートをビルドする
var templateBuildCmd = &cobra.Command{
	Use:   "build [--lang=node]",
	Short: "テンプレートをビルドする",
	Long: `言語指定に応じた Docker イメージをビルドし、sbx テンプレートとして保存します。
--lang=all で全言語を一括ビルドします。`,
	RunE: runTemplateBuild,
}

// runTemplateBuild は template build コマンドのメイン処理
func runTemplateBuild(cmd *cobra.Command, args []string) error {
	templatesDir := findTemplatesDir()
	builder := template.NewBuilder(templatesDir, dryRun)

	if templateBuildLang == "all" {
		return builder.BuildAndSaveAll()
	}

	if templateBuildLang == "base" {
		if err := builder.BuildBase(); err != nil {
			return err
		}
	} else {
		if err := builder.BuildLang(templateBuildLang); err != nil {
			return err
		}
	}

	tag := fmt.Sprintf("dbox-%s:latest", templateBuildLang)
	return builder.SaveTemplate(tag)
}

// templateListCmd はテンプレート一覧を表示する
var templateListCmd = &cobra.Command{
	Use:   "ls",
	Short: "テンプレート一覧を表示する",
	RunE:  runTemplateList,
}

// runTemplateList はテンプレート一覧を表示する
func runTemplateList(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)
	out, err := sb.TemplateList()
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
