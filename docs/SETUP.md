# ローカル環境での起動手順

このドキュメントでは、ローカル環境でアプリケーションを起動・確認する方法を詳しく説明します。

## 方法1: Docker Composeを使用（推奨）

最も簡単な方法です。すべてのサービスが自動的に起動します。

### ステップ1: 環境変数の設定

プロジェクトルートに`.env`ファイルを作成します：

```bash
# プロジェクトルートで実行
cp .env.example .env
```

`.env`ファイルを編集して、以下の値を設定してください：

```bash
# Google Maps Platform
GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here

# OpenAI
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-4o-mini

# NextAuth
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your_nextauth_secret_here

# Google OAuth (for NextAuth)
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
```

#### 各APIキーの取得方法

1. **Google Maps Platform API キー**
   - https://console.cloud.google.com/ にアクセス
   - プロジェクトを作成または選択
   - 「APIとサービス」→「認証情報」からAPIキーを作成
   - 以下のAPIを有効化（重要）：
     - **Places API** (Places API (New) または Places API)
     - **Routes API** (Routes API (v2)) - ルート計算に必要
     - **Maps JavaScript API** - 地図表示に必要
   - 注意: Routes APIが有効化されていないと、ルート計算時に403エラーが発生します

2. **OpenAI API キー**
   - https://platform.openai.com/api-keys にアクセス
   - アカウントを作成してAPIキーを取得

3. **NextAuth Secret**
   - 以下のコマンドで生成：
   ```bash
   openssl rand -base64 32
   ```

4. **Google OAuth クライアントID/シークレット**
   - https://console.cloud.google.com/ にアクセス
   - 「APIとサービス」→「認証情報」→「OAuth 2.0 クライアント ID」を作成
   - 承認済みのリダイレクト URI に以下を追加：
     - `http://localhost:3000/api/auth/callback/google`

### ステップ2: Docker Composeで起動

```bash
# プロジェクトルートで実行
docker-compose up --build
```

初回起動時は、各サービスのビルドと依存関係のインストールに時間がかかります（5-10分程度）。

### ステップ3: アプリケーションにアクセス

- **Webアプリ**: http://localhost:3000
- **Go API**: http://localhost:8080
- **Python MLサービス**: http://localhost:8000

### ステップ4: 動作確認

1. ブラウザで http://localhost:3000 にアクセス
2. 「Googleでログイン」ボタンをクリック
3. Googleアカウントでログイン
4. 旅程作成フォームが表示されることを確認

## 方法2: Dockerなしで個別に起動

各サービスを個別のターミナルで起動する方法です。開発時に便利です。

### 前提条件

- Node.js 20以上
- Go 1.21以上
- Python 3.11以上
- SQLite3

### ステップ1: 環境変数の設定

プロジェクトルートに`.env`ファイルを作成（方法1と同じ）

### ステップ2: 各サービスを起動

#### ターミナル1: Python MLサービス

```bash
cd apps/ml
pip install -r requirements.txt
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

#### ターミナル2: Go API

```bash
cd apps/api
go mod download
go run main.go
```

#### ターミナル3: Next.js Webアプリ

```bash
cd apps/web
npm install
npm run dev
```

### ステップ3: アプリケーションにアクセス

- **Webアプリ**: http://localhost:3000
- **Go API**: http://localhost:8080
- **Python MLサービス**: http://localhost:8000

## トラブルシューティング

### ポートが既に使用されている

以下のコマンドでポートの使用状況を確認：

```bash
# macOS/Linux
lsof -i :3000
lsof -i :8080
lsof -i :8000

# 使用中のプロセスを終了
kill -9 <PID>
```

### Docker Composeでエラーが発生する

```bash
# コンテナとボリュームをクリーンアップ
docker-compose down -v

# 再度起動
docker-compose up --build
```

### Google Maps APIキーエラー

- Google Cloud ConsoleでAPIが有効化されているか確認
- APIキーに制限がかかっていないか確認
- ブラウザのコンソールでエラーメッセージを確認

### NextAuth認証エラー

- `.env`ファイルの`NEXTAUTH_SECRET`が設定されているか確認
- Google OAuthのリダイレクトURIが正しく設定されているか確認
- `http://localhost:3000/api/auth/callback/google` が承認済みリダイレクトURIに含まれているか確認

### データベースエラー

```bash
# dataディレクトリの権限を確認
ls -la data/

# 必要に応じて権限を変更
chmod 755 data/
```

### ログの確認

#### Docker Composeの場合

```bash
# すべてのサービスのログ
docker-compose logs

# 特定のサービスのログ
docker-compose logs web
docker-compose logs api
docker-compose logs ml
```

#### 個別起動の場合

各ターミナルにログが表示されます。

## 動作確認チェックリスト

- [ ] Webアプリが http://localhost:3000 で起動している
- [ ] Googleログインが動作する
- [ ] 旅程作成フォームが表示される
- [ ] 「行きたい場所」を入力できる
- [ ] 「興味タグ」を選択できる
- [ ] 「旅程を生成」ボタンで旅程が生成される
- [ ] おすすめスポットが表示される
- [ ] 地図にマーカーとルートが表示される
- [ ] ドラッグ&ドロップで順序変更ができる
- [ ] 共有URLが表示される

## 次のステップ

アプリケーションが正常に起動したら、以下を試してみてください：

1. 複数の場所を入力して旅程を生成
2. おすすめスポットを採用して旅程に追加
3. 順序を変更して再計算
4. 共有URLをコピーして別のブラウザで開く

