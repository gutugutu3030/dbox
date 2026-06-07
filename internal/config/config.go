package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// コマンド名（設定ディレクトリ名として使用）
const CommandName = "dbox"

// GlobalConfig は ~/.config/dbox/config.yaml の構造体
type GlobalConfig struct {
	DefaultAgent string         `yaml:"default_agent"`
	Template     TemplateConfig `yaml:"template"`
}

// TemplateConfig はテンプレートに関する既定値を定義
type TemplateConfig struct {
	Registry string `yaml:"registry"`
}

// ProjectConfig はプロジェクトルートの .dbox.yaml の構造体
type ProjectConfig struct {
	Version     int            `yaml:"version"`
	Agent       string         `yaml:"agent"`
	Langs       []string       `yaml:"langs"`
	Template    string         `yaml:"template"`
	SandboxName string         `yaml:"sandbox_name"`
	Clone       bool           `yaml:"clone"`
	Resources   ResourceConfig `yaml:"resources"`
	Network     NetworkConfig  `yaml:"network"`
}

// NetworkConfig はサンドボックスのネットワークポリシー設定を定義
type NetworkConfig struct {
	AllowedDomains []string `yaml:"allowed_domains"`
}

// AgentDefaultDomains はエージェント別のデフォルト許可ドメインを返す
func AgentDefaultDomains(agent string) []string {
	switch agent {
	case "opencode":
		return []string{"opencode.ai:443"}
	default:
		return nil
	}
}

// ResourceConfig はサンドボックスのリソース制限を定義
type ResourceConfig struct {
	CPUs   int    `yaml:"cpus"`
	Memory string `yaml:"memory"`
}

// DefaultGlobalConfig はグローバル設定の既定値を返す
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		DefaultAgent: "opencode",
		Template: TemplateConfig{
			Registry: "docker/sandbox-templates",
		},
	}
}

// DefaultProjectConfig はプロジェクト設定の既定値を返す
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Version: 2,
		Agent:   "opencode",
		Langs:   []string{"base"},
		Clone:   false,
		Resources: ResourceConfig{
			CPUs:   0,
			Memory: "",
		},
	}
}

// MergeDomains は allowed_domains のユーザー指定とエージェント既定値をマージする
func MergeDomains(cfg *ProjectConfig) []string {
	seen := make(map[string]struct{})
	var result []string

	// エージェント既定値を先に追加
	for _, d := range AgentDefaultDomains(cfg.Agent) {
		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
			result = append(result, d)
		}
	}

	// ユーザー指定を追加（重複排除）
	for _, d := range cfg.Network.AllowedDomains {
		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
			result = append(result, d)
		}
	}

	return result
}

// GlobalConfigDir はグローバル設定ディレクトリのパスを返す
func GlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory の取得に失敗: %w", err)
	}
	return filepath.Join(home, ".config", CommandName), nil
}

// GlobalConfigPath はグローバル設定ファイルのパスを返す
func GlobalConfigPath() (string, error) {
	dir, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// EnsureGlobalConfigDir はグローバル設定ディレクトリを作成する
func EnsureGlobalConfigDir() error {
	dir, err := GlobalConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadGlobalConfig はグローバル設定ファイルを読み込む。
// ファイルが存在しない場合は既定値を返す
func LoadGlobalConfig() (*GlobalConfig, error) {
	cfg := DefaultGlobalConfig()
	path, err := GlobalConfigPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("グローバル設定の読み込みに失敗: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("グローバル設定のパースに失敗: %w", err)
	}
	return cfg, nil
}

// SaveGlobalConfig はグローバル設定ファイルを保存する
func SaveGlobalConfig(cfg *GlobalConfig) error {
	if err := EnsureGlobalConfigDir(); err != nil {
		return err
	}

	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("グローバル設定のシリアライズに失敗: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("グローバル設定の書き込みに失敗: %w", err)
	}
	return nil
}

// FindProjectConfig は指定されたディレクトリから .dbox.yaml を探す
func FindProjectConfig(dir string) (string, error) {
	path := filepath.Join(dir, ".dbox.yaml")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

// LoadProjectConfig はプロジェクト設定ファイルを読み込む。
// ファイルが存在しない場合は ErrNotExist を返す
func LoadProjectConfig(dir string) (*ProjectConfig, error) {
	path, err := FindProjectConfig(dir)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("プロジェクト設定の読み込みに失敗: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("プロジェクト設定のパースに失敗: %w", err)
	}

	// version 1 の設定（単一言語 lang: node）を version 2 に変換
	if cfg.Version == 1 {
		cfg.Version = 2
	}

	return &cfg, nil
}

// SaveProjectConfig は指定ディレクトリに .dbox.yaml を保存する
func SaveProjectConfig(dir string, cfg *ProjectConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("プロジェクト設定のシリアライズに失敗: %w", err)
	}

	path := filepath.Join(dir, ".dbox.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("プロジェクト設定の書き込みに失敗: %w", err)
	}
	return nil
}

