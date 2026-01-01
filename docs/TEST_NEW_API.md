# 新しいAPIテスト方法

## 前提条件

1. APIサーバーが起動していること
   ```bash
   cd apps/api
   go run main.go
   # または
   go build -o api . && ./api
   ```

2. 環境変数 `GOOGLE_MAPS_API_KEY` が設定されていること

## テスト方法

### 1. 場所追加API (POST /add/:place_id)

#### シンプルなテストスクリプト（jq不要）
```bash
./test_add_api_simple.sh
```

#### 詳細なテストスクリプト（jqが必要）
```bash
./test_add_api.sh
```

#### curlコマンドを直接実行
```bash
# 基本的なリクエスト
curl -X POST http://localhost:8080/add/ChIJz_VfLQKJRTkRxqxoKTN4xqU \
  -H "Content-Type: application/json"

# レスポンス例:
# {
#   "message": "場所が追加されました",
#   "place_id": "ChIJz_VfLQKJRTkRxqxoKTN4xqU"
# }
```

### 2. ルート提案API (POST /result)

#### シンプルなテストスクリプト（jq不要）
```bash
./test_result_api_simple.sh
```

#### 詳細なテストスクリプト（jqが必要）
```bash
./test_result_api.sh
```

#### curlコマンドを直接実行
```bash
# 基本的なリクエスト（2つの場所で最小限のテスト）
curl -X POST http://localhost:8080/result \
  -H "Content-Type: application/json" \
  -d '{
    "places": [
      "ChIJz_VfLQKJRTkRxqxoKTN4xqU",
      "ChIJ0_VL5wKJRTkRklWXjPv_sU"
    ]
  }'

# レスポンス例:
# {
#   "places": [...],
#   "route": {
#     "legs": [...],
#     "distance_meters": 12345,
#     "duration": "1800s",
#     "optimized_order": [0]
#   }
# }
```

## リクエストパラメータ

### POST /add/:place_id
- `place_id` (パスパラメータ): Google Place ID
  - 例: `ChIJz_VfLQKJRTkRxqxoKTN4xqU` (博多駅)

### POST /result
- `places` (必須): 場所IDのリスト（配列）
  - 例: `["ChIJz_VfLQKJRTkRxqxoKTN4xqU", "ChIJ0_VL5wKJRTkRklWXjPv_sU"]`
  - 最小2つの場所が必要（出発地と目的地）

## テスト用Place ID（福岡の場所）

- 博多駅: `ChIJz_VfLQKJRTkRxqxoKTN4xqU`
- 太宰府天満宮: `ChIJ0_VL5wKJRTkRklWXjPv_sU`

**注意**: これらのPlace IDは例です。実際のテストでは、Google Places APIで取得した正しいPlace IDを使用してください。

## エラーテスト

### 場所追加API
```bash
# place_idが空の場合
curl -X POST http://localhost:8080/add/ \
  -H "Content-Type: application/json"
```

### ルート提案API
```bash
# placesが空の場合
curl -X POST http://localhost:8080/result \
  -H "Content-Type: application/json" \
  -d '{"places": []}'

# リクエストボディなし
curl -X POST http://localhost:8080/result \
  -H "Content-Type: application/json"
```

