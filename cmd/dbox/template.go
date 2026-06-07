package main

import (
	"fmt"
	"strings"

	"github.com/gutugutu3030/sbx-template/internal/detect"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"
	"github.com/gutugutu3030/sbx-template/internal/template"

	"github.com/spf13/cobra"
)

var templateBuildLang string

// templateCmd はテンプレート管理のサブコマンド
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Dockerイメージとsbxテンプレートを管理する",
	Long:  `Docker イメージをビルドし、sbx テンプレートとしてロードします。`,
}

func init() {
	templateCmd.AddCommand(templateBuildCmd)
	templateCmd.AddCommand(templateListCmd)
	templateBuildCmd.Flags().StringVarP(&templateBuildLang, "lang", "l", "base", "ビルドする言語 (base, node, go, node,python などカンマ区切りで複数可, all)")
}

// templateBuildCmd はDockerイメージをビルドしsbxテンプレートとしてロードする
var templateBuildCmd = &cobra.Command{
	Use:   "build [--lang=node]",
	Short: "Dockerイメージをビルドしsbxテンプレートとしてロードする",
	Long: `言語指定に応じた Docker イメージをビルドし、sbx テンプレートとしてロードします。
--lang=all で全言語を一括ビルドします。
--lang=node,go のようにカンマ区切りで複数言語を指定すると合成イメージをビルドします。

sbx create --template=<tag> で使用可能になります。`,
	RunE: runTemplateBuild,
}

// runTemplateBuild は template build コマンドのメイン処理
func runTemplateBuild(cmd *cobra.Command, args []string) error {
	templatesDir := findTemplatesDir()
	builder := template.NewBuilder(templatesDir, dryRun)
	sb := sandbox.NewRunner(dryRun)

	if templateBuildLang == "all" {
		languages := []string{"node", "python", "java", "go", "rust", "ruby"}
		for _, lang := range languages {
			fmt.Printf("=== %s テンプレートをビルド中 ===\n", lang)
			if err := builder.BuildLang(lang); err != nil {
				return err
			}
			tag := fmt.Sprintf("dbox-%s:latest", lang)
			if err := sb.TemplateSave(tag); err != nil {
				return err
			}
		}
		return nil
	}

	if templateBuildLang == "base" {
		if err := builder.BuildBase(); err != nil {
			return err
		}
		return sb.TemplateSave("dbox-base:latest")
	}

	langs := detect.LangsFromString(templateBuildLang)
	if len(langs) == 0 {
		return fmt.Errorf("有効な言語が指定されていません: %s", templateBuildLang)
	}

	var tag string
	if len(langs) == 1 {
		lang := string(langs[0])
		if err := builder.BuildLang(lang); err != nil {
			return err
		}
		tag = fmt.Sprintf("dbox-%s:latest", lang)
	} else {
		langStrs := make([]string, len(langs))
		for i, l := range langs {
			langStrs[i] = string(l)
		}
		fmt.Printf("合成イメージをビルドします: %s\n", strings.Join(langStrs, ", "))
		composer := template.NewComposer(builder)
		imageName, err := composer.Compose(langStrs)
		if err != nil {
			return err
		}
		tag = imageName + ":latest"
	}

	if err := sb.TemplateSave(tag); err != nil {
		return err
	}

	fmt.Printf("テンプレート %s のビルドとロードが完了しました\n", tag)
	return nil
}

// templateListCmd はsbxテンプレート一覧を表示する
var templateListCmd = &cobra.Command{
	Use:   "ls",
	Short: "sbxテンプレート一覧を表示する",
	RunE:  runTemplateList,
}

// runTemplateList はsbxテンプレート一覧を表示する
func runTemplateList(cmd *cobra.Command, args []string) error {
	sb := sandbox.NewRunner(dryRun)
	out, err := sb.TemplateList()
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
