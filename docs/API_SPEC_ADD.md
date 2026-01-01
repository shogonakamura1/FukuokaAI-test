# 場所追加API仕様書

## エンドポイント

```
POST /add/:place_id
```

## 概要

リコメンドされた場所をリストに追加するAPIです。
このAPIは、クライアント側でリストを管理する前提で設計されており、サーバー側では単純に成功を返します。

## リクエスト

### リクエストパスパラメータ

| パラメータ名 | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `place_id` | `string` | 必須 | Google Place ID |

### リクエストヘッダー

```
Content-Type: application/json
```

### リクエストボディ

なし

### リクエスト例

```
POST /add/ChIJz_VfLQKJRTkRxqxoKTN4xqU
Content-Type: application/json
```

## レスポンス

### 成功時 (HTTP 200)

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| `message` | `string` | 成功メッセージ |
| `place_id` | `string` | 追加されたPlace ID |

### レスポンス例

```json
{
  "message": "場所が追加されました",
  "place_id": "ChIJz_VfLQKJRTkRxqxoKTN4xqU"
}
```

## エラーレスポンス

### HTTP 400 Bad Request

#### エラーコード: `INVALID_REQUEST`

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "place_idが指定されていません"
  }
}
```

## 処理フロー

1. ボタンを押した際に、ボタンの場所ID（place_id）を取得
2. サーバー側では単純に成功を返す（実際のリスト管理はクライアント側で行う）

## 注意事項

- このAPIは、クライアント側でリストを管理する前提で設計されています
- サーバー側では状態を保持しません
- 実際のリスト管理はクライアント側の実装に依存します
- Place IDの妥当性は検証しません（存在チェックなどは行いません）

