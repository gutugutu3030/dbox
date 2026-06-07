# dbox

`dbox` は [Docker Sandboxes (`sbx`)](https://docs.docker.com/desktop/features/sandbox/) の軽量ラッパーCLIです。

![おもしろ画像](./description.png)

## 概要

`dbox` は以下の課題を解決します：

1. **sbx の使い方を忘れる** → 単純なコマンドインターフェースでラップ
2. **リポジトリによって必要な環境が変わる** → 言語検出からテンプレート作成まで自動化

### 機能

| コマンド | 説明 |
|----------|------|
| `dbox init` | 言語検出 → テンプレート自動作成 → `sbx create` |
| `dbox start` | サンドボックスを起動（デフォルトで nvim が開く） |
| `dbox stop` | サンドボックスを停止 |
| `dbox exec` | サンドボックス内でコマンド実行 |
| `dbox template build` | 言語別 Docker テンプレートをビルド |
| `dbox template ls` | テンプレート一覧を表示 |

---

## インストール

### 前提条件

- Go 1.24 以上
- Docker Desktop（sbx が利用可能なこと）

### ビルド

```bash
# リポジトリをクローン
git clone <repository-url>
cd sbx-template

# ビルド
go build -o dbox ./cmd/dbox/

# パスの通ったディレクトリに配置
mv dbox /usr/local/bin/
```

### 開発モード

```bash
# ホットリロード（要: air）
go run ./cmd/dbox/
```

---

## クイックスタート

```bash
# プロジェクトの初期化（カレントディレクトリの言語を自動検出）
dbox init

# エージェントと言語を指定
dbox init --agent=codex --lang=python

# 特定のディレクトリを指定
dbox init --agent=opencode ./my-project

# サンドボックスを起動（nvim が開く）
dbox start

# サンドボックス内でコマンド実行
dbox exec "node --version"

# サンドボックスを停止
dbox stop
```

### dry-run モード

実際のコマンドを実行せずに動作確認ができます：

```bash
dbox init --dry-run
dbox start --dry-run
```

---

## コマンドリファレンス

### `dbox init [path]`

プロジェクトを初期化し、サンドボックスを作成します。

| フラグ | 短縮 | 既定値 | 説明 |
|--------|------|--------|------|
| `--agent` | `-a` | グローバル設定の値 | 使用するAIエージェント（opencode, codex, claude 等） |
| `--lang` | `-l` | `auto` | 使用言語（auto で自動検出） |
| `--dry-run` | `-n` | `false` | 実際のコマンドを実行せず表示のみ |

**言語検出ロジック**（優先順位順）：

1. 設定ファイルの存在検出（`package.json` → Node, `go.mod` → Go, `Cargo.toml` → Rust 等）
2. 拡張子の出現頻度による推定（`.ts/.js` → Node, `.py` → Python, `.go` → Go 等）
3. 該当なし → base（最小構成）

### `dbox start [sandbox-name]`

サンドボックスを起動します。

| 状態 | 動作 |
|------|------|
| サンドボックスが存在しない | 新規作成してからアタッチ |
| 停止中 | 起動してからアタッチ |
| 実行中 | そのままアタッチ |

### `dbox stop [sandbox-name]`

サンドボックスを停止します。

### `dbox exec <command>`

サンドボックス内でコマンドを実行します。

### `dbox template build --lang=<lang>`

テンプレートをビルドして保存します。

| フラグ | 短縮 | 既定値 | 説明 |
|--------|------|--------|------|
| `--lang` | `-l` | `base` | ビルドする言語（all で全言語一括） |

### `dbox template ls`

登録済みのテンプレート一覧を表示します。

---

## 設定ファイル

### グローバル設定: `~/.config/dbox/config.yaml`

```yaml
default_agent: opencode
nvim:
  config_source: ~/.config/nvim
template:
  registry: docker/sandbox-templates
```

### プロジェクト設定: `.dbox.yaml`（プロジェクトルートに自動生成）

```yaml
version: 1
agent: opencode
lang: node
template: dbox-node
sandbox_name: dbox-opencode-my-project
clone: true
resources:
  cpus: 0      # 0 = auto
  memory: 50%  # ホストメモリの50%
```

---

## プロジェクト構造

```
.
├── cmd/dbox/
│   ├── main.go       # エントリポイント、rootCmd、main()
│   ├── init.go       # init コマンド（言語検出、テンプレート確認、sbx create）
│   ├── start.go      # start コマンド（attach/再起動/新規作成）
│   ├── stop.go       # stop コマンド
│   ├── exec.go       # exec コマンド
│   └── template.go   # template build/ls コマンド
├── internal/
│   ├── config/       # 設定ファイルの読み書き（yaml パース）
│   ├── detect/       # 言語検出ロジック（設定ファイル + 拡張子）
│   ├── sandbox/      # sbx コマンドラッパー
│   └── template/     # Docker イメージビルド、テンプレート保存
├── templates/
│   ├── base.Dockerfile      # Ubuntu 24.04 + nvim + git + curl
│   ├── node.Dockerfile      # Node.js
│   ├── python.Dockerfile    # Python 3 + pip
│   ├── java.Dockerfile      # OpenJDK 21 + Maven
│   ├── go.Dockerfile        # Go
│   ├── rust.Dockerfile      # Rust
│   └── ruby.Dockerfile      # Ruby
├── go.mod
└── plan.md           # 設計書
```

---

## テスト

```bash
# 全テスト実行
go test ./... -v

# カバレッジ計測
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## テンプレートカスタマイズ

テンプレートは以下の優先順位で探索されます：

1. 実行ファイルからの相対パス（`../templates/`）
2. カレントディレクトリ（`./templates/`）
3. グローバル設定ディレクトリ（`~/.config/dbox/templates/`）

独自の `base.Dockerfile` や `<lang>.Dockerfile` を任意の場所に配置することで、
デフォルトのテンプレートを上書きできます。
