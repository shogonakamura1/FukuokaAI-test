# データベーススキーマ

## テーブル一覧

### users

ユーザー情報を保存します。

| カラム名 | 型 | 説明 |
|---------|-----|------|
| id | TEXT PRIMARY KEY | NextAuthのsub等 |
| created_at | TEXT | 作成日時 (ISO8601) |

### trips

旅程情報を保存します。

| カラム名 | 型 | 説明 |
|---------|-----|------|
| id | TEXT PRIMARY KEY | UUID |
| user_id | TEXT | ユーザーID (users.idへの外部キー) |
| title | TEXT | 旅程タイトル |
| start_time | TEXT | 開始時刻 (例: "10:00") |
| created_at | TEXT | 作成日時 (ISO8601) |

**インデックス**
- `user_id` にインデックス

### trip_places

旅程に含まれるスポット情報を保存します。

| カラム名 | 型 | 説明 |
|---------|-----|------|
| id | TEXT PRIMARY KEY | UUID |
| trip_id | TEXT | 旅程ID (trips.idへの外部キー) |
| place_id | TEXT | Google Place ID |
| name | TEXT | スポット名 |
| lat | REAL | 緯度 |
| lng | REAL | 経度 |
| kind | TEXT | 種類 (must/recommended/start) |
| stay_minutes | INTEGER | 滞在時間（分） |
| order_index | INTEGER | 順序 |
| reason | TEXT | おすすめ理由（LLM生成） |
| review_summary | TEXT | レビュー要約（LLM生成） |
| photo_url | TEXT | 写真URL |

**インデックス**
- `trip_id` にインデックス

### shares

共有情報を保存します。

| カラム名 | 型 | 説明 |
|---------|-----|------|
| share_id | TEXT PRIMARY KEY | UUID |
| trip_id | TEXT UNIQUE | 旅程ID (trips.idへの外部キー) |
| created_at | TEXT | 作成日時 (ISO8601) |

## リレーション

- `trips.user_id` → `users.id`
- `trip_places.trip_id` → `trips.id`
- `shares.trip_id` → `trips.id`


