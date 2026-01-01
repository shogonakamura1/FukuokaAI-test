# リコメンドAPI仕様書

## エンドポイント

```
POST /recommend
```

## 概要

指定した寄りたい場所と興味タグに基づいて、おすすめの場所を最大10件返すAPIです。
出発地点とゴール地点を含む全ての座標で最小全域木を構築し、各エッジの中点周辺で検索を行います。

## リクエスト

### リクエストヘッダー

```
Content-Type: application/json
```

### リクエストボディ

| フィールド名 | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `must_places` | `string[]` | 必須 | 寄りたい場所のリスト（場所名） |
| `interest_tags` | `string[]` | 必須 | 興味タグのリスト |
| `start_place` | `string` | 任意 | 出発地点（デフォルト: "博多駅"） |
| `goal_place` | `string` | 任意 | ゴール地点（未指定の場合は出発地点と同じ） |

### リクエスト例

```json
{
  "must_places": ["太宰府天満宮", "福岡タワー"],
  "interest_tags": ["カフェ", "神社"],
  "start_place": "博多駅",
  "goal_place": "天神"
}
```

## レスポンス

### 成功時 (HTTP 200)

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `places` | `Place[]` | 推薦場所のリスト（最大10件） |

#### Place オブジェクト

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `place_id` | `string` | Google Place ID |
| `name` | `string` | 場所名 |
| `lat` | `number` | 緯度 |
| `lng` | `number` | 経度 |
| `photo_url` | `string` | 写真URL（存在する場合） |
| `rating` | `number` | 評価（存在する場合） |
| `review_summary` | `string` | レビュー要約（存在する場合） |
| `category` | `string` | カテゴリ（存在する場合） |
| `address` | `string` | 住所（存在する場合） |

### レスポンス例

```json
{
  "places": [
    {
      "place_id": "ChIJ...",
      "name": "スターバックス コーヒー 太宰府天満宮店",
      "lat": 33.5194,
      "lng": 130.5344,
      "photo_url": "https://maps.googleapis.com/maps/api/place/photo?maxwidth=400&photoreference=...&key=...",
      "rating": 4.5,
      "review_summary": "とても素晴らしい場所です...",
      "category": "cafe",
      "address": "福岡県太宰府市..."
    }
  ]
}
```

## エラーレスポンス

### HTTP 400 Bad Request

#### エラーコード: `INVALID_REQUEST`

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "寄りたい場所が指定されていません"
  }
}
```

または

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "興味タグが指定されていません"
  }
}
```

#### エラーコード: `GEOCODING_ERROR`

```json
{
  "error": {
    "code": "GEOCODING_ERROR",
    "message": "出発地点の座標取得に失敗しました: ..."
  }
}
```

### HTTP 500 Internal Server Error

#### エラーコード: `PLACES_API_ERROR`

```json
{
  "error": {
    "code": "PLACES_API_ERROR",
    "message": "Google Places API error: ..."
  }
}
```

#### エラーコード: `CONFIGURATION_ERROR`

```json
{
  "error": {
    "code": "CONFIGURATION_ERROR",
    "message": "GOOGLE_MAPS_API_KEY is not set"
  }
}
```

#### エラーコード: `INTERNAL_ERROR`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "..."
  }
}
```

## 処理フロー

1. 出発地点とゴール地点を指定したのち、それ以外の必ず寄りたい場所の座標を得る
2. 出発地点とゴール地点を含む全ての座標で、それぞれ一番距離が近い組み合わせを作り、全ての点が線でつながるようにする（最小全域木を構築）
3. 全ての枝で、半径が 枝の長さ/√3 となる円内で、興味タグで検索をnearby search APIで検索する
4. その結果を一枚の写真と共にレビューの高い順に合計10件表示する

## 注意事項

- 結果は関連性スコアと評価を考慮してソートされます
- 最大10件まで返されます
- 出発地点が指定されていない場合、デフォルトで「博多駅」が使用されます
- ゴール地点が指定されていない場合、出発地点と同じ場所が使用されます

