---
name: coding-conventions
description: >
  Use when the user asks to follow, apply, or check coding conventions,
  coding rules, code style, comment style, file size limits, test rules,
  or naming conventions. Also when generating or editing any source code
  file. This skill defines project-independent coding standards. Use
  ALWAYS during code generation or editing to ensure compliance with
  comment rules, file size limits, test coverage, and code style.
---

# Coding Conventions Skill

このスキルはリポジトリ非依存のコーディング規則を定義します。
AGENTS.md が存在する場合は AGENTS.md の設定がこのスキルより優先されます。

## コメント規則

- コメントは **日本語** で記載すること
- コメントは **関数/メソッド単位** で記載すること（宣言直上に配置）
- 関数/メソッドの長さが **100行を超える場合** は、内部にも処理ブロック単位でコメントを挟むこと
- `TODO` コメントは `TODO: <内容>` の形式で記載し、担当者と課題を明確にすること

### 良い例

```java
/**
 * ユーザー情報をデータベースから取得する。
 * 存在しない場合は Optional.empty() を返す。
 */
public Optional<User> findById(String id) {
```

### 悪い例

```java
public Optional<User> findById(String id) { // ユーザー検索
```

## ファイルサイズ規則

- 1ファイルの最大行数は **400行** とする
- 400行を超える場合は、適切な責務でファイル分割すること
- 分割の単位は「責任」（例: コマンド定義、ビジネスロジック、入出力）とすること

## テスト規則

- すべての公開関数・メソッドには **単体テスト（ユニットテスト）** を作成すること
- テストファイルは対象ファイルと同じディレクトリに配置すること
- テストファイル名は `{対象ファイル名}.test.{ext}` または `{対象ファイル名}_test.{ext}` とする（言語の慣習に従う）
- **テーブル駆動テスト**（パラメタライズドテスト）を推奨する
- テストの後処理（クリーンアップ）はフレームワークの機構を活用して確実に行うこと

## コードスタイル

- その言語の標準フォーマッタに従うこと
- 未使用の import や変数は削除すること（リンターで警告が出ない状態を維持）
- **エラーハンドリングは必ず行い**、エラーを無視しないこと
- エラーメッセージ・ログは日本語で記述すること
- マジックナンバー・マジック文字列は定数として定義すること

## ファイル分割の方針

400行を超えた場合の分割単位の指針:

| ファイル/クラス名 | 責務 |
|---|---|
| `{domain}.{ext}` / `{domain}.ts` | エンティティ・モデル・型定義 |
| `{domain}Service.{ext}` | ビジネスロジック |
| `{domain}Controller.{ext}` | リクエスト/レスポンス処理（HTTP） |
| `{domain}Repository.{ext}` | データアクセス・永続化 |
| `errors.{ext}` | エラー定義 |
| `validation.{ext}` | バリデーションロジック |

## コード生成時のチェックリスト

- [ ] すべての公開関数/メソッドに日本語のコメントがあるか
- [ ] ファイルの行数が400行以下か
- [ ] 未使用の import や変数がないか（リンターが通るか）
- [ ] エラー/例外を無視していないか
- [ ] テーブル駆動テスト（パラメタライズドテスト）が書かれているか（新規関数の場合）
- [ ] コメントが日本語で書かれているか
- [ ] マジックナンバー/マジック文字列が定数定義されているか

## 言語別マッピング

| 言語 | フォーマッタ | テーブル駆動テストの手法 |
|---|---|---|
| Go | `gofmt` / `go fmt` | 匿名構造体スライス + `range` |
| Java | `google-java-format` / `spotless` | `@ParameterizedTest` + `@CsvSource` / `@MethodSource` |
| TypeScript | `prettier` / `deno fmt` | `test.each`（Vitest/Jest） |
| Python | `ruff format` / `black` | `@pytest.mark.parametrize` |
| Rust | `rustfmt` | `#[test]` + マクロ or `rstest` |

## 補足

このスキルはあくまでデフォルト規則です。リポジトリルートに `AGENTS.md`（または `AGENTS.ja.md`）が存在する場合は、その内容をこのスキルより優先して適用してください。

グローバル適用したい場合は以下の場所に配置してください:

```bash
cp -r .opencode/skills/coding-conventions ~/.config/opencode/skills/coding-conventions
```
