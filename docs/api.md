# API仕様書

## エンドポイント一覧

### POST /v1/trips

旅程を生成します。

**リクエスト**
```json
{
  "must_places": ["太宰府天満宮", "糸島のカフェ"],
  "interest_tags": ["カフェ", "景色"],
  "free_text": "静かめの古民家カフェが好き"
}
```

**レスポンス**
```json
{
  "trip_id": "uuid",
  "share_id": "uuid",
  "itinerary": [
    {
      "id": "uuid",
      "place_id": "google_place_id",
      "name": "スポット名",
      "lat": 33.123,
      "lng": 130.456,
      "kind": "must",
      "stay_minutes": 60,
      "order_index": 0,
      "time_range": "10:00-11:00",
      "reason": "おすすめ理由",
      "review_summary": "レビュー要約"
    }
  ],
  "candidates": [
    {
      "place_id": "google_place_id",
      "name": "スポット名",
      "lat": 33.123,
      "lng": 130.456,
      "category": "カフェ",
      "photo_url": "https://...",
      "reason": "おすすめ理由",
      "review_summary": "レビュー要約"
    }
  ],
  "route": {
    "polyline": "encoded_polyline_string"
  }
}
```

### POST /v1/trips/:trip_id/recompute

旅程の順序を再計算します。

**リクエスト**
```json
{
  "ordered_place_ids": ["uuid1", "uuid2", "uuid3"],
  "stay_minutes_map": {
    "uuid1": 60,
    "uuid2": 90
  }
}
```

**レスポンス**
```json
{
  "itinerary": [...],
  "route": {
    "polyline": "encoded_polyline_string"
  }
}
```

### GET /v1/shares/:share_id

共有された旅程を取得します（ログイン不要）。

**レスポンス**
```json
{
  "trip": {
    "id": "uuid",
    "title": "旅程タイトル",
    "start_time": "10:00"
  },
  "itinerary": [...],
  "route": {
    "polyline": "encoded_polyline_string"
  }
}
```

## エラーレスポンス

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ",
    "details": "詳細情報（オプション）"
  }
}
```

## ステータスコード

- 200: 成功
- 400: リクエストエラー
- 401: 認証エラー
- 404: リソースが見つからない
- 500: サーバーエラー


