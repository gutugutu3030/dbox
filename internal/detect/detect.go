package detect

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Language は検出された言語を表す
type Language string

const (
	LanguageNode   Language = "node"
	LanguageGo     Language = "go"
	LanguageRust   Language = "rust"
	LanguagePython Language = "python"
	LanguageJava   Language = "java"
	LanguageRuby   Language = "ruby"
	LanguageBase   Language = "base"
)

// LangConfig は言語ごとの検出ルールを定義
type LangConfig struct {
	Language     Language
	ConfigFiles  []string // 設定ファイル（例: package.json）
	Extensions   []string // 拡張子（例: .js）
	TemplateName string   // 対応テンプレート名
}

// 言語検出ルール一覧（優先順位順）
var detectionRules = []LangConfig{
	{
		Language:     LanguageNode,
		ConfigFiles:  []string{"package.json"},
		Extensions:   []string{".js", ".jsx", ".ts", ".tsx"},
		TemplateName: "dbox-node",
	},
	{
		Language:     LanguageGo,
		ConfigFiles:  []string{"go.mod"},
		Extensions:   []string{".go"},
		TemplateName: "dbox-go",
	},
	{
		Language:     LanguageRust,
		ConfigFiles:  []string{"Cargo.toml"},
		Extensions:   []string{".rs"},
		TemplateName: "dbox-rust",
	},
	{
		Language:     LanguagePython,
		ConfigFiles:  []string{"requirements.txt", "pyproject.toml", "Pipfile", "setup.py"},
		Extensions:   []string{".py"},
		TemplateName: "dbox-python",
	},
	{
		Language:     LanguageJava,
		ConfigFiles:  []string{"pom.xml", "build.gradle", "build.gradle.kts"},
		Extensions:   []string{".java"},
		TemplateName: "dbox-java",
	},
	{
		Language:     LanguageRuby,
		ConfigFiles:  []string{"Gemfile"},
		Extensions:   []string{".rb"},
		TemplateName: "dbox-ruby",
	},
}

// MultiResult は多言語検出の結果を保持する
type MultiResult struct {
	Languages    []Language
	TemplateName string
}

// AllTemplates は定義済みの全テンプレート名を返す
func AllTemplates() []string {
	names := make([]string, 0, len(detectionRules)+1)
	for _, rule := range detectionRules {
		names = append(names, rule.TemplateName)
	}
	names = append(names, "dbox-base")
	sort.Strings(names)
	return names
}

// TemplateNameForLang は単一言語に対応するテンプレート名を返す
func TemplateNameForLang(lang Language) string {
	for _, rule := range detectionRules {
		if rule.Language == lang {
			return rule.TemplateName
		}
	}
	return "dbox-base"
}

// TemplateNameForLangs は複数言語に対応するテンプレート名を返す。
// 言語をアルファベット順にソートし "dbox-<lang1>-<lang2>" の形式で返す
func TemplateNameForLangs(langs []Language) string {
	if len(langs) == 0 {
		return "dbox-base"
	}

	// base 以外の言語だけ抽出
	filtered := make([]Language, 0, len(langs))
	for _, lang := range langs {
		if lang != LanguageBase {
			filtered = append(filtered, lang)
		}
	}
	if len(filtered) == 0 {
		return "dbox-base"
	}

	// アルファベット順にソート
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i] < filtered[j]
	})

	// 重複除去
	uniq := make([]Language, 0, len(filtered))
	seen := make(map[Language]bool)
	for _, lang := range filtered {
		if !seen[lang] {
			seen[lang] = true
			uniq = append(uniq, lang)
		}
	}

	names := make([]string, len(uniq))
	for i, lang := range uniq {
		names[i] = string(lang)
	}
	return "dbox-" + strings.Join(names, "-")
}

// DetectByConfig は設定ファイルの存在に基づいて言語を検出する。
// マッチした言語を全て返す（base は含まない）
func DetectByConfig(files []os.FileInfo) []Language {
	var detected []Language
	for _, rule := range detectionRules {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			for _, cfgFile := range rule.ConfigFiles {
				if file.Name() == cfgFile {
					detected = append(detected, rule.Language)
					break
				}
			}
		}
	}
	return detected
}

// DetectByExtension は拡張子に基づいて言語を検出する。
// 閾値（全ファイルの5%）以上のスコアを持つ言語を全て返す
func DetectByExtension(root string) []Language {
	extCount := make(map[string]int)
	totalFiles := 0

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == ".git" || base == "target" ||
				base == "vendor" || base == ".venv" || base == "__pycache__" ||
				strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			return nil
		}
		extCount[ext]++
		totalFiles++
		return nil
	})

	if totalFiles == 0 {
		return nil
	}

	// 言語ごとに該当拡張子の出現数を集計
	langScore := make(map[Language]int)
	for _, rule := range detectionRules {
		score := 0
		for _, ext := range rule.Extensions {
			score += extCount[ext]
		}
		if score > 0 {
			langScore[rule.Language] = score
		}
	}

	if len(langScore) == 0 {
		return nil
	}

	// 閾値（5%）以上のスコアを持つ言語を抽出
	threshold := 0.05
	var detected []Language
	for lang, score := range langScore {
		if float64(score)/float64(totalFiles) >= threshold {
			detected = append(detected, lang)
		}
	}

	return detected
}

// Detect は設定ファイルと拡張子の両方を使って言語を検出する。
// 設定ファイルによる検出を最優先し、なければ拡張子ベースにフォールバックする
func Detect(root string) MultiResult {
	entries, err := os.ReadDir(root)
	if err != nil {
		return MultiResult{
			Languages:    []Language{LanguageBase},
			TemplateName: "dbox-base",
		}
	}

	// os.FileInfo のスライスに変換
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}

	// 設定ファイルベースの検出（優先）
	langs := DetectByConfig(infos)
	if len(langs) == 0 {
		// 拡張子ベースの検出（フォールバック）
		langs = DetectByExtension(root)
	}

	if len(langs) == 0 {
		langs = []Language{LanguageBase}
	}

	return MultiResult{
		Languages:    langs,
		TemplateName: TemplateNameForLangs(langs),
	}
}

// LangsFromString はカンマ区切りの言語文字列を Language スライスに変換する
func LangsFromString(s string) []Language {
	parts := strings.Split(s, ",")
	langs := make([]Language, 0, len(parts))
	seen := make(map[Language]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lang := Language(p)
		if seen[lang] {
			continue
		}
		seen[lang] = true
		if TemplateNameForLang(lang) != "dbox-base" || lang == LanguageBase {
			langs = append(langs, lang)
		}
	}
	return langs
}
