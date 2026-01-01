# ルート提案API (/result) テスト方法

## 問題

`/result` APIをテストする際、「有効な場所が見つかりませんでした」というエラーが発生する場合があります。
これは、使用したPlace IDが無効または存在しない場合に発生します。

## 解決方法

### 方法1: 自動でPlace IDを取得するスクリプトを使用（推奨）

`test_result_api_with_place_ids.sh` スクリプトを使用すると、自動で`/recommend` APIから有効なPlace IDを取得してテストします。

```bash
./test_script/test_result_api_with_place_ids.sh
```

このスクリプトは:
1. `/recommend` APIを実行してPlace IDを取得
2. 取得したPlace IDを使って`/result` APIをテスト

### 方法2: 手動でPlace IDを取得

1. `/recommend` APIでPlace IDを取得:

```bash
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["博多駅"],
    "interest_tags": ["カフェ"]
  }' | jq '.places[].place_id'
```

2. 取得したPlace IDを2つ選んで`/result` APIをテスト:

```bash
curl -X POST http://localhost:8080/result \
  -H "Content-Type: application/json" \
  -d '{
    "places": [
      "取得したplace_id_1",
      "取得したplace_id_2"
    ]
  }'
```

### 方法3: 既知の有効なPlace IDを使用

実際に動作確認済みのPlace IDを使用する場合は、`/recommend` APIのレスポンスから取得したPlace IDを使用してください。

## テストスクリプト一覧

- `test_script/test_result_api_with_place_ids.sh` - 自動でPlace IDを取得してテスト（推奨）
- `test_script/test_result_api.sh` - エラーケースも含む詳細なテスト（Place IDは手動で指定）
- `test_script/test_result_api_simple.sh` - シンプルな説明スクリプト

## 注意事項

- Place IDはGoogle Places APIで取得した有効なIDである必要があります
- 例として使用したPlace ID（`ChIJz_VfLQKJRTkRxqxoKTN4xqU`など）は無効な可能性があります
- APIコール数を最小限にするため、2つの場所（出発地と目的地）でテストすることを推奨します

