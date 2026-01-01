#!/bin/bash

# ルート提案APIテスト用スクリプト（手動でPlace IDを指定）
# 使用方法: ./test_result_api_manual.sh <place_id_1> <place_id_2> [place_id_3 ...]
# 
# このスクリプトは、手動でPlace IDを指定して/result APIをテストします。
# Place IDは、Google Places APIから取得した有効なIDを使用してください。
# 
# 注意: リストの最初の場所が出発地点、最後の場所がゴール地点として自動的に設定されます。
# 中間の場所は経由地点として扱われ、順序が最適化されます。

# APIのベースURL（デフォルトはlocalhost:8080）
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "ルート提案API テスト（手動でPlace IDを指定）"
echo "========================================="
echo ""

# Place IDを引数から取得
if [ $# -lt 2 ]; then
    echo "使用方法: $0 <place_id_1> <place_id_2> [place_id_3 ...]"
    echo ""
    echo "注意:"
    echo "  - 最初の場所 (place_id_1) が【出発地点】として設定されます"
    echo "  - 最後の場所が【ゴール地点】として設定されます"
    echo "  - 中間の場所は【経由地点】として扱われ、順序が最適化されます"
    echo ""
    echo "例（2つの場所）:"
    echo "  $0 ChIJz_VfLQKJRTkRxqxoKTN4xqU ChIJ0_VL5wKJRTkRklWXjPv_sU"
    echo ""
    echo "例（3つ以上の場所）:"
    echo "  $0 place_id_1 place_id_2 place_id_3 place_id_4"
    echo "  → place_id_1が出発、place_id_4がゴール、place_id_2とplace_id_3が経由地点"
    echo ""
    echo "Place IDの取得方法:"
    echo "1. Google Maps PlatformのPlaces APIを使用"
    echo "2. または、/recommend APIのレスポンスから取得（正常に動作する場合）"
    echo ""
    exit 1
fi

# すべてのPlace IDを配列に格納
PLACE_IDS=("$@")
echo "指定されたPlace ID（${#PLACE_IDS[@]}個）:"
for i in "${!PLACE_IDS[@]}"; do
    if [ $i -eq 0 ]; then
        echo "  出発地点: ${PLACE_IDS[$i]}"
    elif [ $i -eq $((${#PLACE_IDS[@]} - 1)) ]; then
        echo "  ゴール地点: ${PLACE_IDS[$i]}"
    else
        echo "  経由地点: ${PLACE_IDS[$i]}"
    fi
done
echo ""

# リクエストボディを作成（jqを使用）
if command -v jq &> /dev/null; then
    PLACE_IDS_JSON=$(printf '%s\n' "${PLACE_IDS[@]}" | jq -R . | jq -s .)
    REQUEST_BODY=$(jq -n --argjson places "$PLACE_IDS_JSON" '{places: $places}' 2>/dev/null)
else
    # jqが利用できない場合、手動でJSONを作成（2つの場合のみ）
    if [ ${#PLACE_IDS[@]} -eq 2 ]; then
        REQUEST_BODY="{\"places\": [\"${PLACE_IDS[0]}\", \"${PLACE_IDS[1]}\"]}"
    else
        echo "エラー: jqがインストールされていないため、3つ以上の場所を指定できません"
        echo "jqをインストールするか、2つの場所のみを指定してください"
        exit 1
    fi
fi

echo "リクエストボディ:"
echo "$REQUEST_BODY" | jq '.' 2>/dev/null || echo "$REQUEST_BODY"
echo ""

# /result APIをテスト
echo "【テスト実行】/result API"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/result" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v | jq '.' 2>/dev/null || cat

echo ""
echo "========================================="
echo "テスト完了"
echo "========================================="

