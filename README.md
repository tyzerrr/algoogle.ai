# algoogle.ai

algoogle.ai is a local, AI-assisted coding interview practice app for SWE interview training.

ローカルで動く、AI対話型のコーディング面接トレーニングアプリです。

LeetCodeの解答暗記ではなく、AI面接官との会話を通じて、問題理解、制約、全探索、最適化、エッジケース、計算量、説明力を鍛えることを目的にしています。

## 特徴

- 毎日のおすすめ問題を表示
- ARAI60の60問をカード形式で表示
- 問題ごとに一発OK、フォローアップ込みOK、再挑戦OK、要復習を管理
- 問題詳細画面で、実装前からAI面接官と方針を会話
- AI providerをCodex CLI / Claude Code CLIからattemptごとに選択
- Google / Meta / Amazon / Generic の面接プリセットを切り替え
- Real Interview Modeで、ローカル実行なし・補完なし・方針説明必須の練習
- WhiteBoardで疑似コード、Mermaid sequence diagram、データ構造、不変条件、状態遷移、計算量比較をAI面接官と議論
- Light / Dark / Netflix の3テーマを切り替え
- 30秒以上沈黙した場合、AI面接官が発話や不変条件の説明を促す
- Monaco Editorとローカル `workspace/` のPythonファイルを双方向同期
- ローカルのテストケースをsubprocessで実行
- Codex CLIまたはClaude Code CLIをサブプロセスとして起動し、AIチャットとAIレビューを実行
- Submit後に採用判定、scorecard、shadow evaluator notes、mini roundsを表示
- attempts、chat history、follow-up questions、learning notes、weakness signalsをSQLiteに保存

## 問題セット

問題リストは新井康平氏の「[コーディング面接対策のために解きたいLeetCode 60問](https://1kohei1.com/leetcode/)」をもとにしています。

LeetCode本文の丸写しではなく、アプリ内では面接練習用カードとしてタイトル、カテゴリ、タグ、公式問題へのリンク、スターターコードを管理します。

## 技術スタック

- Frontend: Next.js, TypeScript, Monaco Editor, pnpm
- Backend: Go, chi, SQLite
- Code runner: Python subprocess
- AI interviewer / reviewer: Codex CLI or Claude Code CLI subprocess
- Local runtime: Docker Compose

## AI CLI providerについて

AIチャットとAIレビューは、APIキーをアプリに保存せず、ログイン済みのCLIをサブプロセスとして起動します。

問題詳細画面の `AI provider` からattemptごとに選択できます。

```bash
codex login
codex exec --help
claude auth
claude -p --help
```

Docker ComposeではLinux版の `@openai/codex` と `@anthropic-ai/claude-code` をAPIコンテナに入れ、ホストの `${HOME}/.codex` と `${HOME}/.claude` を読み取り専用でマウントします。

## Credential管理

このリポジトリはcredentialを含めない前提です。

- OpenAI API keyやAnthropic API keyはアプリ設定として使わず、ログイン済みCLIセッションを利用します
- `.env` と `.env.*` は `.gitignore` で除外しています
- Docker Composeでは `${HOME}/.codex` と `${HOME}/.claude` を読み取り専用でマウントします
- `~/.codex`、`~/.claude`、API key、token、private keyはcommitしないでください

## ローカル起動

Makefileを使う場合:

```bash
make up
```

直接Docker Composeを使う場合:

```bash
docker compose up --build
```

- Web: http://localhost:3000
- API: http://localhost:8000
- Health: http://localhost:8000/health

## NeoVim同期

attemptを開くと、APIが `workspace/` にPythonファイルを作ります。画面の `NeoVim sync` 行に表示されるコマンドで編集できます。

```bash
nvim ./workspace/<problem-id>-<attempt-id>.py
```

NeoVimで保存すると、アプリ上のMonaco Editorへ自動反映されます。アプリ側で編集した内容も同じファイルへ保存されます。

`workspace/` は `.gitignore` 済みです。練習中のコードやメモがcommitされないようにしています。

## WhiteBoard Discussion

実際のcoding interviewで白板を使う場面を、PC上ではWhiteBoardとして再現します。

問題詳細画面の `WhiteBoard` から、AI面接官に「今WhiteBoardを使うべきか」を判断させられます。AIは会話、コード、過去のミス、過去のWhiteBoardを見たうえで、次のいずれかを選びます。

- `Pseudo`: 実装前の疑似コード
- `Mermaid`: `sequenceDiagram` による操作順序や面接官との確認フロー
- `Data`: データ構造、保持する値、操作コスト
- `Invariant`: 初期化、維持、終了条件
- `State`: 状態遷移と不正遷移
- `Cost`: 複数解法の時間・空間・trade-off比較

WhiteBoardはattemptごとにSQLiteへ保存されます。次回同じ問題を解くとき、AI面接官は過去のWhiteBoardと間違え方を参照し、同じ説明の繰り返しではなく、より厳しいフォローアップや新しい白板課題を出します。

## Real Interview Mode

問題詳細画面では、上部のInterview ControlからAI provider、会社プリセット、練習モードを選べます。

- `Codex`: Codex CLI subprocessで面接官/レビュワーを起動
- `Claude Code`: Claude Code CLI subprocessで面接官/レビュワーを起動

- `Google`: 曖昧さ、制約、証明、不変条件、深いフォローアップを重視
- `Meta`: 実装速度、簡潔な説明、dry run、追加問題への対応を重視
- `Amazon`: trade-off、顧客影響のあるedge case、行動面の説明も確認
- `Generic`: 総合的なcoding interview練習

`Real` ではローカル実行ボタンを無効化し、Monaco Editorの補完も抑制します。面接官は、実装前にbrute force、最適化方針、データ構造、不変条件、edge caseを確認します。

`Practice` ではローカル実行と補完を使えるので、学習初期や復習に向いています。

## Theme Modes

画面右上からテーマを切り替えられます。

- `Light`: 明るい通常テーマ
- `Dark`: 長時間練習向けの暗色テーマ
- `Netflix`: 黒と赤を基調にした集中モード

## Review Loop

AIレビューは単なる正誤判定ではなく、実際の面接後フィードバックに近い形で保存します。

- `hire_recommendation`: Strong Hire / Hire / Lean Hire / Lean No Hire / No Hire
- `scorecard`: correctness、communication、complexity、code qualityなどを1-4で評価
- `shadow_notes`: 面接官が裏で残す観察メモ
- `mini_rounds`: 追加coding follow-up、behavioral、system/product thinkingの質問
- `weakness signals`: 次回の問題やフォローアップに使う弱点履歴

復習画面では、過去のfollow-up、覚えておくべきミス、苦手パターン、弱点グラフを見られます。

## ホストで開発する場合

API:

```bash
cd apps/api
go run ./cmd/api
```

Web:

```bash
cd apps/web
corepack pnpm install
corepack pnpm dev
```

## 環境変数

主な設定は `.env.example` を参照してください。

- `DATABASE_URL`: SQLiteの保存先
- `PYTHON_BIN`: Python実行コマンド
- `CODEX_CLI_PATH`: Codex CLI実行ファイル
- `CODEX_MODEL`: 必要な場合だけCodex CLIに渡すモデル名
- `CODEX_WORKDIR`: Codex CLIの作業ディレクトリ
- `CODEX_CLI_TIMEOUT_SECONDS`: Codex CLIのタイムアウト秒数
- `CLAUDE_CLI_PATH`: Claude Code CLI実行ファイル
- `CLAUDE_MODEL`: 必要な場合だけClaude Code CLIに渡すモデル名
- `CLAUDE_WORKDIR`: Claude Code CLIの作業ディレクトリ
- `CLAUDE_CLI_TIMEOUT_SECONDS`: Claude Code CLIのタイムアウト秒数
- `CODE_WORKSPACE_DIR`: APIが同期ファイルを読み書きするディレクトリ
- `CODE_WORKSPACE_PUBLIC_DIR`: Webに表示する同期ファイルのパス
- `NEXT_PUBLIC_API_BASE_URL`: Webから参照するAPI URL

## テストとCI

Go APIにはユニットテストがあります。

```bash
cd apps/api
go test ./...
```

Webはpnpmで型チェックとビルドを行います。

```bash
cd apps/web
corepack pnpm typecheck
corepack pnpm build
```

GitHub Actionsでは、Go APIのテスト、Next.jsの型チェック/ビルド、Docker image buildを実行します。

## MVPでできること

- 今日のおすすめ問題を見る
- 問題一覧を見る
- ARAI60を10問ずつページングして見る
- 問題詳細でPythonコードを書く
- Monaco EditorまたはNeoVimで編集する
- Google / Meta / Amazon風のReal Interview Modeで練習する
- ローカルテストケースを実行する
- Codex CLIまたはClaude Code CLI経由のAI面接官と会話する
- WhiteBoardで疑似コード、Mermaid、データ構造、不変条件、状態遷移、計算量比較を保存しながら議論する
- Codex CLIまたはClaude Code CLI経由のAIレビューをJSON構造で保存する
- Light / Dark / Netflix modeを切り替える
- scorecard、採用判定、フォローアップ質問、過去のミス、弱点signalをSQLiteに保存し、次回の面接官プロンプトへ反映する
