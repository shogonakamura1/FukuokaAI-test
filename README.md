# 福岡観光 旅程作成AI Webアプリ（MVP）

福岡観光の旅程を自動生成するAI Webアプリケーションです。ユーザーが行きたい場所と興味タグを入力すると、AIが最適なルートとおすすめスポットを提案します。

## アーキテクチャ

- **フロントエンド**: Next.js 14+ (App Router)
- **バックエンドAPI**: Go (Gin)
- **MLサービス**: Python (FastAPI)
- **データベース**: SQLite
- **外部API**: Google Maps Platform, OpenAI

## 機能

- Googleログイン認証
- 行きたい場所（最大5件）の入力
- 興味タグによる推薦
- ルート上のおすすめスポット提案（写真+理由+レビュー要約付き）
- 30分単位の旅程タイムライン生成
- ドラッグ&ドロップによる順序変更
- 共有URLでの閲覧（ログイン不要）
  - 生成された旅程は `/share/[share_id]` で誰でも閲覧可能

## クイックスタート

**最も簡単な方法（Docker Compose）:**

```bash
# 1. 環境変数を設定
cp .env.example .env
# .envファイルを編集してAPIキーを設定

# 2. 起動
docker-compose up --build

# 3. ブラウザで http://localhost:3000 にアクセス
```

詳細な手順は [SETUP.md](./SETUP.md) を参照してください。

## セットアップ

### 前提条件

- Docker & Docker Compose（推奨）
- または Node.js 20+, Go 1.21+, Python 3.11+
- Google Maps Platform API キー
- OpenAI API キー
- Google OAuth クライアントID/シークレット（NextAuth用）

### 1. 環境変数の設定

`.env.example`をコピーして`.env`を作成し、必要な値を設定してください。

```bash
cp .env.example .env
```

`.env`ファイルに以下を設定：
- `GOOGLE_MAPS_API_KEY`: Google Maps Platform API キー
- `OPENAI_API_KEY`: OpenAI API キー
- `OPENAI_MODEL`: 使用するモデル（デフォルト: gpt-4o-mini）
- `NEXTAUTH_SECRET`: NextAuth用のシークレット（`openssl rand -base64 32`で生成可能）
- `GOOGLE_CLIENT_ID`: Google OAuth クライアントID
- `GOOGLE_CLIENT_SECRET`: Google OAuth クライアントシークレット

### 2. Docker Composeで起動

```bash
docker-compose up --build
```

初回起動時は、各サービスのビルドと依存関係のインストールに時間がかかります。

### 3. アクセス

- Webアプリ: http://localhost:3000
- Go API: http://localhost:8080
- Python MLサービス: http://localhost:8000

### 4. 使用方法

1. ブラウザで http://localhost:3000 にアクセス
2. Googleアカウントでログイン
3. 「行きたい場所」を入力（最大5件）
4. 「興味タグ」を選択（複数選択可）
5. 「旅程を生成」ボタンをクリック
6. おすすめスポットが表示されたら、「採用する」ボタンで旅程に追加
7. 旅程タイムラインでドラッグ&ドロップで順序を変更
8. 共有URLをコピーして他の人と共有

## ローカル開発（Dockerなし）

### Web (Next.js)

```bash
cd apps/web
npm install
npm run dev
```

### API (Go)

```bash
cd apps/api
go mod download
go run main.go
```

### ML (Python)

```bash
cd apps/ml
pip install -r requirements.txt
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

## プロジェクト構造

```
.
├── apps/
│   ├── web/          # Next.js フロントエンド
│   ├── api/          # Go REST API
│   └── ml/           # Python MLサービス
├── data/             # SQLiteデータベースファイル
├── docs/             # 設計ドキュメント
├── docker-compose.yml
├── .env.example
└── README.md
```

## API仕様

詳細は `docs/api.md` を参照してください。

## データベーススキーマ

詳細は `docs/database.md` を参照してください。

## トラブルシューティング

### APIキーエラー

- Google Maps APIキーが正しく設定されているか確認
- 必要なAPIが有効化されているか確認（Places API, Directions API, Maps JavaScript API）

### データベースエラー

- `data/`ディレクトリの書き込み権限を確認
- SQLiteファイルが正しく作成されているか確認

### 認証エラー

- NextAuthのシークレットが設定されているか確認
- Google OAuthのリダイレクトURIが正しく設定されているか確認（`http://localhost:3000/api/auth/callback/google`）

## ライセンス

MIT

