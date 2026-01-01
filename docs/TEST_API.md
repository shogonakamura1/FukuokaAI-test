# API テスト方法

## 前提条件

1. APIサーバーが起動していること
   ```bash
   cd apps/api
   go run main.go
   # または
   go build -o api . && ./api
   ```

2. 環境変数 `GOOGLE_MAPS_API_KEY` が設定されていること
   ```bash
   export GOOGLE_MAPS_API_KEY="your-api-key"
   ```

## テスト方法

### 方法1: テストスクリプトを使用（推奨）

#### 詳細なテストスクリプト（jqが必要）
```bash
./test_recommend_api.sh
```

#### シンプルなテストスクリプト（jq不要）
```bash
./test_recommend_api_simple.sh
```

### 方法2: curlコマンドを直接実行

#### 基本的なリクエスト
```bash
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮", "福岡タワー"],
    "interest_tags": ["カフェ", "神社"]
  }'
```

#### 出発地点・ゴール地点を指定
```bash
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮", "福岡タワー"],
    "interest_tags": ["カフェ", "レストラン"],
    "start_place": "博多駅",
    "goal_place": "天神"
  }'
```

#### レスポンスを整形して表示（jqが必要）
```bash
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮"],
    "interest_tags": ["カフェ", "神社"]
  }' | jq '.'
```

## リクエストパラメータ

- `must_places` (必須): 寄りたい場所のリスト
  - 例: `["太宰府天満宮", "福岡タワー"]`
- `interest_tags` (必須): 興味タグのリスト
  - 例: `["カフェ", "神社", "観光"]`
- `start_place` (オプション): 出発地点（デフォルト: "博多駅"）
  - 例: `"博多駅"`
- `goal_place` (オプション): ゴール地点
  - 例: `"天神"`

## レスポンス例

```json
{
  "places": [
    {
      "place_id": "ChIJ...",
      "name": "カフェ名",
      "lat": 33.5904,
      "lng": 130.4208,
      "photo_url": "https://maps.googleapis.com/...",
      "rating": 4.5,
      "review_summary": "とても良いカフェです...",
      "category": "cafe",
      "address": "福岡県..."
    }
  ]
}
```

## エラーレスポンス例

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "寄りたい場所が指定されていません"
  }
}
```

## トラブルシューティング

### APIサーバーが起動していない場合
```bash
cd apps/api
go run main.go
```

### ポートが既に使用されている場合
```bash
export PORT=8081
go run main.go
```

### Google Maps API キーが設定されていない場合
```bash
export GOOGLE_MAPS_API_KEY="your-api-key"
```

### jqがインストールされていない場合（macOS）
```bash
brew install jq
```

