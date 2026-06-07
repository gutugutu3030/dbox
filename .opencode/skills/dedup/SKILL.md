---
name: dedup
description: >
  Use when the user asks to detect, find, remove, or refactor code duplication
  (duplicate code, duplicated logic, copy-pasted code, DRY, deduplicate,
  dedup, unify duplicates, consolidate code, similar functions, identical
  blocks, repeated patterns). Use ONLY for duplication detection and
  consolidation — not for general refactoring or code review.
---

# Deduplication Skill

このスキルはコードベース内の重複を検出し、適切な責務のファイルに統一する手順を定義します。

## 重複検出手順

1. **重複検出ツールの実行**:
   - 言語固有のツール（各言語の重複検出ツール一覧を参照）
   - 汎用的な方法: 以下のようなパターンを `grep` / `rg` で検索
     - 同一または類似の関数/メソッド定義
     - コピーペーストされたエラーハンドリング/例外処理ブロック
     - 同一の型・インターフェース定義
     - 同じロジックが複数ファイルに散らばっていないか
   - `jscpd`（多言語対応）などの専用ツールも検討

2. **重複の種類を分類**:
   - **完全一致**: まったく同じコードが複数箇所に存在
   - **類似ロジック**: 一部の定数や変数名だけが異なる
   - **責務の混在**: 共通処理が本来あるべき場所ではなく各所に散らばっている

## 統一・集約ルール

1. **適切な配置先の決定**:
   - 共通ユーティリティ → `util.{ext}` / `helpers.{ext}`
   - 共通モデル → `model.{ext}` / `types.{ext}`
   - 共通エラーハンドリング → `errors.{ext}` / `exceptions.{ext}`
   - 共通バリデーション → `validation.{ext}` / `validators.{ext}`

2. **統一時のルール**:
   - 統一後のファイルは `coding-conventions` スキルの **ファイルサイズ規則（400行上限）** に従うこと
   - 400行を超える場合はさらに適切な責務で分割すること
   - 重複元のコードは削除し、新しい共通関数/メソッドを呼び出すように変更すること
   - 削除したファイルがある場合は参照（import / require / using 等）をすべて更新すること
   - 同じ責務のコードがすでに存在する場合は、そちらに統合すること（新規ファイルは原則作らない）

3. **テストの移動**:
   - 重複元のテストも対象のファイルに移動すること
   - 移動後もすべての既存テストがパスすることを確認すること
   - 移動元のテストファイルに残った空のテストは削除すること
   - パラメタライズドテストの形式を推奨（`coding-conventions` スキル準拠）

4. **変更検証**:
   - プロジェクトのビルド/コンパイルが通ることを確認
   - リンターで警告が出ないことを確認
   - テストを実行してすべてパスすることを確認
   - 不要になった import / require / using が削除されていることを確認

## 重複チェックリスト

- [ ] `grep` や `rg` で同一関数・定数・変数が複数ファイルに定義されていないか
- [ ] エラーメッセージ/ログ文字列の重複（typo リスク）
- [ ] 同一ロジックのコピーペースト（3行以上のブロック）
- [ ] テストコードが本番コードの移動に追従しているか
- [ ] 統一後のコードがファイルサイズ上限（400行）を超えていないか

## 言語別: 重複検出ツール

| 言語 | ツール | コマンド例 |
|---|---|---|
| Go | `gocyclo`, `goreporter` | `gocyclo -over 10 .` |
| Java | `pmd-cpd`, `jscpd` | `pmd cpd --minimum-tokens 100 --language java --files src/` |
| TypeScript | `jscpd`, `ts-query`, `repomix` | `jscpd --pattern "src/**/*.ts"` |
| Python | `pylint --duplicate-code`, `jscpd` | `pylint --disable=all --enable=duplicate-code src/` |
| Rust | `rustfmt`, `cargo-dups`（手動） | `cargo install cargo-dups && cargo dups` |

## 言語別: 類似関数/型の検出コマンド

```bash
# 全言語共通: 類似文字列リテラルの検出
rg -n '"[^"]{20,}"' src/ | sort

# Go: 関数定義の重複チェック
rg -n "func.*(" --type go src/ | sort | uniq -d

# TypeScript/JavaScript: 関数/メソッド定義の重複チェック
rg -n "(function |=>|const .* = .*=>|async )" --type ts src/ | sort | uniq -d

# Java: メソッド定義の重複チェック
rg -n "(public|private|protected).*\(" --type java src/ | sort | uniq -d

# Python: 関数定義の重複チェック
rg -n "^def " --type py src/ | sort | uniq -d

# Rust: 関数定義の重複チェック
rg -n "^fn " --type rust src/ | sort | uniq -d

# 定数/列挙型の重複チェック
rg -n "(const |enum |static final)" src/ | sort | uniq -d
```

## 注意事項

- **重複除去は過剰に行わないこと**: 2箇所以下の使用であれば、責務的に分離すべきでない場合もある
- **3箇所以上**で同じロジックが登場したら共通化を検討する
- 重複除去によってコードの可読性が下がる場合は、より適切な命名や分割を検討する
- 重複を除去したら必ずテストを実行して回帰がないことを確認する
