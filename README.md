# algoogle.ai

algoogle.ai is a local, AI-assisted coding interview practice app for SWE interview training.

ローカルで動く、AI対話型のコーディング面接トレーニングアプリです。

LeetCodeの解答暗記ではなく、AI面接官との会話を通じて、問題理解、制約、全探索、最適化、エッジケース、計算量、説明力を鍛えることを目的にしています。

## 特徴

- 毎日のおすすめ問題を表示
- ARAI60の60問をカード形式で表示
- 問題ごとに一発OK、フォローアップ込みOK、再挑戦OK、要復習を管理
- 問題詳細画面で、実装前からAI面接官と方針を会話
- Monaco Editorとローカル `workspace/` のPythonファイルを双方向同期
- ローカルのテストケースをsubprocessで実行
- Codex CLIをサブプロセスとして起動し、AIチャットとAIレビューを実行
- attempts、chat history、follow-up questions、learning notesをSQLiteに保存

## 問題セット

問題リストは新井康平氏の「[コーディング面接対策のために解きたいLeetCode 60問](https://1kohei1.com/leetcode/)」をもとにしています。

LeetCode本文の丸写しではなく、アプリ内では面接練習用カードとしてタイトル、カテゴリ、タグ、公式問題へのリンク、スターターコードを管理します。

## 技術スタック

- Frontend: Next.js, TypeScript, Monaco Editor, pnpm
- Backend: Go, chi, SQLite
- Code runner: Python subprocess
- AI interviewer / reviewer: Codex CLI subprocess
- Local runtime: Docker Compose

## Codex CLIについて

AIチャットとAIレビューは、OpenAI APIキーではなく `codex exec` をサブプロセスとして起動します。

ローカル実行では、すでにログイン済みのCodex CLIを使います。

```bash
codex login
codex exec --help
```

Docker ComposeではLinux版の `@openai/codex` をAPIコンテナに入れ、ホストの `${HOME}/.codex` を読み取り専用でマウントします。APIキーは発行しません。

## Credential管理

このリポジトリはcredentialを含めない前提です。

- OpenAI API keyは使わず、Codex CLIのログイン済みセッションを利用します
- `.env` と `.env.*` は `.gitignore` で除外しています
- Docker Composeでは `${HOME}/.codex` を読み取り専用でマウントします
- `~/.codex`、API key、token、private keyはcommitしないでください

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
- ローカルテストケースを実行する
- Codex CLI経由のAI面接官と会話する
- Codex CLI経由のAIレビューをJSON構造で保存する
- フォローアップ質問と過去のミスをSQLiteに保存し、次回の面接官プロンプトへ反映する
