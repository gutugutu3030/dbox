# dbox - Docker Sandbox Wrapper CLI

## 概要

`dbox` は `sbx` (Docker Sandboxes) の軽量ラッパーCLI。
言語検出 → テンプレート自動ビルド → サンドボックス起動 を一連のコマンドで実現する。

## 背景・課題

1. sbx の使い方を忘れる
2. リポジトリによって必要な環境が変わる（node, java など）

## 解決策

- `init` で言語を自動検出し、テンプレートからサンドボックスを作成
- nvim のインストール・設定をテンプレートに組み込み
- coding agent (opencode, codex など) を引数で指定
- 独自設定を `~/.config/dbox/` に保存

---

## コマンド仕様

```
dbox init [--agent=opencode] [--lang=auto] [path]
dbox start [sandbox-name]
dbox stop [sandbox-name]
dbox template build [--lang=node]
dbox exec <command>
```

### `dbox init`

1. `path` のファイル走査 → 言語検出
2. 言語対応テンプレート存在確認 (`sbx template ls`)
   - なければ `dbox template build --lang=<lang>` 自動実行
3. `.dbox.yaml` 生成
4. `sbx create --name=<name> --template=sbxw-<lang> <agent> <path>` 実行
5. サンドボックス名を `.dbox.yaml` に保存

### `dbox start`

1. `.dbox.yaml` 読み込み
2. `sbx ls` で sandbox_name の状態確認
   - running → `sbx run <name>` で attach
   - stopped → `sbx start <name>` してから attach
   - 不存在 → `sbx create` 実行してから attach

### `dbox stop`

1. `.dbox.yaml` からサンドボックス名取得
2. `sbx stop <name>` 実行

### `dbox template build`

1. `~/.config/dbox/nvim/` に nvim 設定コピー (初回のみ)
2. `templates/base.Dockerfile` でビルド → `dbox-base` タグ付け
3. `templates/<lang>.Dockerfile` でビルド → `dbox-<lang>` タグ付け
4. `sbx template save` で保存

### `dbox exec`

1. `.dbox.yaml` からサンドボックス名取得
2. `sbx exec <name> <command>` 実行

---

## ファイル構成

### プロジェクト (dbox 自体)

```
sbx-template/
├── cmd/
│   └── dbox/
│       └── main.go
├── internal/
│   ├── config/        # 設定読み込み・書き込み
│   ├── detect/        # 言語検出
│   ├── template/      # テンプレートビルド
│   └── sandbox/       # sbx コマンドラッパー
├── templates/
│   ├── base.Dockerfile
│   ├── node.Dockerfile
│   ├── python.Dockerfile
│   ├── java.Dockerfile
│   ├── go.Dockerfile
│   ├── rust.Dockerfile
│   └── ruby.Dockerfile
├── go.mod
└── plan.md
```

### グローバル設定

```
~/.config/dbox/
├── config.yaml         # 既定値
├── nvim/               # nvim 設定コピー元
└── templates/          # カスタム Dockerfile (オプション)
```

### プロジェクト設定

```
<project-root>/
└── .dbox.yaml          # プロジェクト固有設定
```

---

## 設定ファイル仕様

### `~/.config/dbox/config.yaml` (グローバル既定値)

```yaml
default_agent: opencode
nvim:
  config_source: ~/.config/nvim
template:
  registry: docker/sandbox-templates
```

### `.dbox.yaml` (プロジェクト設定)

```yaml
version: 1
agent: opencode
lang: node
template: dbox-node
sandbox_name: dbox-sbx-template
resources:
  cpus: 0    # 0 = auto
  memory: 50%
```

---

## 言語検出テーブル

| ファイル | 言語 | テンプレート |
|----------|------|-------------|
| `package.json` | node | dbox-node |
| `go.mod` | go | dbox-go |
| `Cargo.toml` | rust | dbox-rust |
| `requirements.txt`, `pyproject.toml`, `Pipfile` | python | dbox-python |
| `pom.xml`, `build.gradle` | java | dbox-java |
| `Gemfile` | ruby | dbox-ruby |
| 上記なし | base | dbox-base |

---

## Dockerfile

### `templates/base.Dockerfile`

```dockerfile
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y \
    neovim git curl \
    && rm -rf /var/lib/apt/lists/*
COPY nvim-config/ /root/.config/nvim/
```

### `templates/node.Dockerfile`

```dockerfile
FROM dbox-base:latest
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs
```

---

## 実装方針

- 言語: Go
- nvim 反映: テンプレートビルド時に `~/.config/nvim` をコピー
- 設定管理: プロジェクト (.dbox.yaml) + グローバル (~/.config/dbox/) の両方
- start 時の既存サンドボックス: attach (既存に接続)
- 1 プロジェクト 1 サンドボックス

---

## 実装フェーズ

### Phase 1: 基盤

- [ ] Go プロジェクト初期化 (`go mod init`)
- [ ] `internal/config/` - 設定ファイルの読み書き
- [ ] `internal/detect/` - 言語検出ロジック
- [ ] `internal/sandbox/` - sbx コマンド実行ラッパー
- [ ] `cmd/dbox/main.go` - CLI エントリポイント
- [ ] `dbox init` 実装

### Phase 2: テンプレート

- [ ] `templates/` 配下の Dockerfile 作成
- [ ] `internal/template/` - テンプレートビルド処理
- [ ] `dbox template build` 実装

### Phase 3: 起動・停止

- [ ] `dbox start` 実装 (attach 含む)
- [ ] `dbox stop` 実装
- [ ] `dbox exec` 実装

### Phase 4: nvim 統合

- [ ] nvim 設定の `~/.config/dbox/nvim/` へのコピー処理
- [ ] base.Dockerfile への nvim 設定組み込み

---

## 依存ライブラリ

- `github.com/spf13/cobra` - CLI フレームワーク
- `gopkg.in/yaml.v3` - YAML パース
- `os/exec` - sbx コマンド実行
