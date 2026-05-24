# prompt-cli

プロンプトをローカルに保存し、再利用・編集を簡単にするCLIツール。

[Faceted Prompting](https://nrslib.com/faceted-prompting) の手法に基づき、プロンプトを **Persona**・**Policy**・**Instruction**・**Output Contract** の4つのファセットに分解して管理する。

各ファセットは **`@part-name`** でパーツを参照するか、テキストをそのまま入力してインライン指定できる。

## インストール

```bash
go install github.com/kazki/prompt-cli@latest
```

または、リポジトリをクローンしてビルド:

```bash
git clone https://github.com/kazki/prompt-cli.git
cd prompt-cli
go build -o prompt-cli
```

## 使い方

### パーツの管理

パーツは各ファセットの再利用可能なコンテンツ。`~/.prompt-cli/parts/<facet>/<name>.md` に保存される。

```bash
# パーツの作成（TUIエディタが開く）
prompt-cli part add persona my-engineer

# パーツの一覧
prompt-cli part list              # 全ファセット
prompt-cli part list persona      # 特定ファセットのみ

# パーツの編集・削除
prompt-cli part edit persona my-engineer
prompt-cli part delete persona my-engineer
```

### テンプレートの作成

テンプレートは各ファセットの設定を定義する。`~/.prompt-cli/templates/<name>/template.yaml` に保存される。

```bash
# TUIで対話的に作成
prompt-cli init my-template
```

### テンプレートの編集

```bash
# TUIで対話的に編集
prompt-cli edit review
```

TUIでは各ファセットにテキストエリアが表示される。`@` を入力するとパーツ補完が起動し、絞り込み・選択できる。`@` なしで直接テキストを入力するとインラインとして保存される。

| キー | 操作 |
|------|------|
| Tab / Shift+Tab | フィールド切り替え |
| `@` + 入力 | パーツ検索・絞り込み |
| ↑↓ / Enter | 候補の選択・確定 |
| Esc | 補完中: 補完をクリア / 通常: キャンセル |
| Ctrl+S | 保存 |

`template.yaml` の例:

```yaml
persona: '@code-reviewer'
policy: '@japanese'
instruction: このPRをレビューしてください
output-contract: '@markdown'
```

### テンプレートの実行

```bash
prompt-cli run review
```

### テンプレートの一覧・削除

```bash
prompt-cli list
prompt-cli delete <name>
```

## ビルトインパーツ

インストール直後から使えるパーツが用意されている。`part list` で `(builtin)` と表示される。同名のユーザーパーツを作成するとビルトインを上書きできる。

| ファセット | パーツ名 | 内容 |
|---|---|---|
| persona | `software-engineer` | シニアソフトウェアエンジニア |
| persona | `code-reviewer` | コードレビュアー |
| persona | `technical-writer` | 技術ドキュメント専門家 |
| policy | `japanese` | 日本語で回答（技術用語は英語併記） |
| policy | `concise` | 簡潔に回答 |
| policy | `step-by-step` | 段階的に思考 |
| instruction | `code-review` | コードレビュー |
| instruction | `summarize` | 要約 |
| instruction | `explain` | 解説 |
| output-contract | `markdown` | Markdown形式で出力 |
| output-contract | `json` | JSON形式で出力 |

## 出力例

```bash
prompt-cli init review
# TUIで persona=@code-reviewer, policy=@japanese, output-contract=@markdown を設定
prompt-cli run review
```

```xml
<persona>
あなたは経験豊富なコードレビュアーです。
バグ、セキュリティリスク、パフォーマンス問題、可読性の観点からコードを分析します。
</persona>

<policy>
すべての回答を日本語で行ってください。
技術用語は必要に応じて英語を併記してください。
</policy>

<instruction>
このPRのセキュリティリスクを重点的にレビューしてください
</instruction>

<output-contract>
出力はMarkdown形式で記述してください。
見出し、箇条書き、コードブロックなどを適切に使用してください。
</output-contract>
```

## ライセンス

MIT
