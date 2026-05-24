# prompt-cli

プロンプトをローカルに保存し、再利用・編集を簡単にするCLIツール。

[Faceted Prompting](https://nrslib.com/faceted-prompting) の手法に基づき、プロンプトを **Persona**・**Policy**・**Instruction**・**Output Contract** の4つのファセットに分解して管理する。

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

### テンプレートの作成

```bash
prompt-cli init <name>
```

`~/.prompt-cli/templates/<name>/` にテンプレートを作成し、TUIエディタを起動する。

### テンプレートの編集

```bash
prompt-cli edit <name>
```

TUIエディタでPersona・Policy・Instruction・Output Contractを編集する。

| キー | 操作 |
|------|------|
| Tab / Shift+Tab | フィールド切り替え |
| Ctrl+S | 保存 |
| Esc | キャンセル |

### テンプレートの実行

```bash
prompt-cli run <name>
```

各ファセットをXMLタグで結合したプロンプトを標準出力に出力する。

### テンプレートの一覧

```bash
prompt-cli list
```

### テンプレートの削除

```bash
prompt-cli delete <name>
```

## 出力例

```bash
prompt-cli run code-reviewer
```

```xml
<persona>
あなたは経験豊富なコードレビュアーです。
</persona>

<policy>
建設的で具体的なフィードバックを心がけてください。
</policy>

<instruction>
コードの品質・可読性・保守性を確認してください。
</instruction>

<output-contract>
問題点と改善案を箇条書きで出力してください。
</output-contract>
```

## ライセンス

MIT
