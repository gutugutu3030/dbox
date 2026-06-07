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

// Result は言語検出の結果を保持する
type Result struct {
	Language     Language
	TemplateName string
	Confidence   float64 // 0.0 ~ 1.0
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

// TemplateNameForLang は言語に対応するテンプレート名を返す
func TemplateNameForLang(lang Language) string {
	for _, rule := range detectionRules {
		if rule.Language == lang {
			return rule.TemplateName
		}
	}
	return "dbox-base"
}

// DetectByConfig は設定ファイルの存在に基づいて言語を検出する
func DetectByConfig(files []os.FileInfo) (Language, float64) {
	for _, rule := range detectionRules {
		for _, file := range files {
			for _, cfgFile := range rule.ConfigFiles {
				if file.Name() == cfgFile {
					return rule.Language, 1.0
				}
			}
		}
	}
	return LanguageBase, 0.0
}

// DetectByExtension は拡張子に基づいて言語を検出する。
// 指定ディレクトリを走査し、一番多く出現した拡張子に対応する言語を返す
func DetectByExtension(root string) (Language, float64) {
	extCount := make(map[string]int)
	totalFiles := 0

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// node_modules, .git などのディレクトリはスキップ
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
		return LanguageBase, 0.0
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
		return LanguageBase, 0.0
	}

	// 最高スコアの言語を特定
	bestLang := LanguageBase
	bestScore := 0
	for lang, score := range langScore {
		if score > bestScore {
			bestLang = lang
			bestScore = score
		}
	}

	confidence := float64(bestScore) / float64(totalFiles)
	return bestLang, confidence
}

// Detect は設定ファイルと拡張子の両方を使って言語を検出する。
// 設定ファイルによる検出を最優先する
func Detect(root string) Result {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Result{Language: LanguageBase, TemplateName: "dbox-base", Confidence: 0.0}
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
	lang, confidence := DetectByConfig(infos)
	if lang != LanguageBase {
		return Result{
			Language:     lang,
			TemplateName: TemplateNameForLang(lang),
			Confidence:   confidence,
		}
	}

	// 拡張子ベースの検出（フォールバック）
	lang, confidence = DetectByExtension(root)
	return Result{
		Language:     lang,
		TemplateName: TemplateNameForLang(lang),
		Confidence:   confidence,
	}
}
