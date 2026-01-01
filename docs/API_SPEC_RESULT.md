# ルート提案API仕様書

## エンドポイント

```
POST /result
```

## 概要

指定した場所のリストから最適なルートを計算し、経由地の順序を最適化するAPIです。
Google Maps Routes APIを使用してルートを計算し、交通状況を考慮した最適化を行います。

## リクエスト

### リクエストヘッダー

```
Content-Type: application/json
```

### リクエストボディ

| フィールド名 | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `places` | `string[]` | 必須 | 場所ID（Google Place ID）のリスト（最低2つ必要） |

**重要**: 
- リストの**最初の場所**が**出発地点（origin）**として設定されます
- リストの**最後の場所**が**ゴール地点（destination）**として設定されます
- 中間の場所は**経由地点（intermediates）**として扱われ、順序が最適化されます

### リクエスト例

```json
{
  "places": [
    "ChIJz_VfLQKJRTkRxqxoKTN4xqU",
    "ChIJ0_VL5wKJRTkRklWXjPv_sU"
  ]
}
```

3つ以上の場所の場合:

```json
{
  "places": [
    "ChIJz_VfLQKJRTkRxqxoKTN4xqU",
    "ChIJ0_VL5wKJRTkRklWXjPv_sU",
    "ChIJ3_something_else"
  ]
}
```

この場合、最初が出発地点、最後がゴール地点、中間が経由地点として最適化されます。

## レスポンス

### 成功時 (HTTP 200)

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `places` | `Place[]` | 最適化された順序の場所リスト |
| `route` | `Route` | ルート情報 |

#### Place オブジェクト

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `place_id` | `string` | Google Place ID |
| `name` | `string` | 場所名 |
| `lat` | `number` | 緯度 |
| `lng` | `number` | 経度 |
| `photo_url` | `string` | 写真URL（存在する場合） |
| `rating` | `number` | 評価（存在する場合） |
| `address` | `string` | 住所（存在する場合） |

#### Route オブジェクト

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `legs` | `RouteLeg[]` | ルートの区間情報のリスト |
| `distance_meters` | `number` | 総距離（メートル単位） |
| `duration` | `string` | 総所要時間（例: "3600s"） |
| `optimized_order` | `number[]` | 最適化された経由地点の順序（インデックスの配列） |

#### RouteLeg オブジェクト

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `start_location` | `Coordinate` | 開始地点の座標 |
| `end_location` | `Coordinate` | 終了地点の座標 |
| `distance_meters` | `number` | 区間の距離（メートル単位） |
| `duration` | `string` | 区間の所要時間（例: "1800s"） |

#### Coordinate オブジェクト

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `lat` | `number` | 緯度 |
| `lng` | `number` | 経度 |

### レスポンス例

```json
{
  "places": [
    {
      "place_id": "ChIJz_VfLQKJRTkRxqxoKTN4xqU",
      "name": "博多駅",
      "lat": 33.5904,
      "lng": 130.4208,
      "rating": 4.5,
      "address": "福岡県福岡市博多区..."
    },
    {
      "place_id": "ChIJ0_VL5wKJRTkRklWXjPv_sU",
      "name": "太宰府天満宮",
      "lat": 33.5194,
      "lng": 130.5344,
      "rating": 4.8,
      "address": "福岡県太宰府市..."
    }
  ],
  "route": {
    "legs": [
      {
        "start_location": {
          "lat": 33.5904,
          "lng": 130.4208
        },
        "end_location": {
          "lat": 33.5194,
          "lng": 130.5344
        },
        "distance_meters": 12345,
        "duration": "1800s"
      }
    ],
    "distance_meters": 12345,
    "duration": "1800s",
    "optimized_order": [0]
  }
}
```

## エラーレスポンス

### HTTP 400 Bad Request

#### エラーコード: `INVALID_REQUEST`

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "場所リストが指定されていません"
  }
}
```

または

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "有効な場所が見つかりませんでした"
  }
}
```

または

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "リクエストの形式が不正です: ..."
  }
}
```

### HTTP 500 Internal Server Error

#### エラーコード: `ROUTES_API_ERROR`

```json
{
  "error": {
    "code": "ROUTES_API_ERROR",
    "message": "ルート計算に失敗しました: ..."
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

1. 行きたい場所リストを取得
2. 各場所のPlace IDから詳細情報（座標など）を取得（Place Details API）
3. リストの最初の場所を出発地点、最後の場所をゴール地点として設定
4. 中間の場所を経由地点として設定
5. Google Maps Routes APIでルートを計算（経由地順最適化を有効化）
   - `travelMode`: "DRIVE"
   - `routingPreference`: "TRAFFIC_AWARE"
   - `optimizeWaypointOrder`: true
   - `departureTime`: 現在時刻から1時間後（デフォルト）
6. 最適化された順序に従って場所リストを並び替え
7. ルート情報と最適化された場所リストを返す

## 使用しているGoogle Maps API

- **Place Details API**: 場所の詳細情報（座標など）を取得
- **Routes API (v2)**: ルート計算と経由地順最適化

## 注意事項

- 最低2つの場所（出発地点とゴール地点）が必要です
- Place IDは有効なGoogle Place IDである必要があります
- 経由地点の順序は最適化されますが、出発地点とゴール地点の順序は変更されません
- `optimized_order`は経由地点のみの順序を表します（出発地点とゴール地点は含まれません）
- ルート計算は交通状況を考慮します（TRAFFIC_AWARE）
- 移動手段は車（DRIVE）を想定しています

