#!/bin/bash

# ルート提案APIテスト用スクリプト（有効なPlace IDを使用）
# 使用方法: ./test_result_api_with_place_ids.sh
# 
# 注意: このスクリプトは、まず/recommend APIを使って有効なPlace IDを取得し、
# そのPlace IDを使って/result APIをテストします。

# APIのベースURL（デフォルトはlocalhost:8080）
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "ルート提案API テスト（有効なPlace IDを使用）"
echo "========================================="
echo ""

# ステップ1: /recommend APIを使って有効なPlace IDを取得
echo "【ステップ1】/recommend APIでPlace IDを取得"
echo "----------------------------------------"
RECOMMEND_RESPONSE=$(curl -s -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["博多駅"],
    "interest_tags": ["カフェ"]
  }')

echo "$RECOMMEND_RESPONSE" | jq '.' 2>/dev/null || echo "$RECOMMEND_RESPONSE"
echo ""

# Place IDを抽出（jqが利用可能な場合）
if command -v jq &> /dev/null; then
    # placesがnullでないことを確認
    PLACES_COUNT=$(echo "$RECOMMEND_RESPONSE" | jq '.places | length' 2>/dev/null)
    if [ "$PLACES_COUNT" = "null" ] || [ "$PLACES_COUNT" = "0" ]; then
        echo "警告: /recommend APIからPlace IDを取得できませんでした"
        echo "レスポンス: $RECOMMEND_RESPONSE"
        echo ""
        echo "/recommend APIが正常に動作していないため、手動でPlace IDを指定してください。"
        echo "使用方法: ./test_result_api_manual.sh <place_id_1> <place_id_2>"
        echo ""
        echo "または、直接curlコマンドを使用:"
        echo "curl -X POST \"${BASE_URL}/result\" \\"
        echo "  -H \"Content-Type: application/json\" \\"
        echo "  -d '{\"places\": [\"PLACE_ID_1\", \"PLACE_ID_2\"]}'"
        exit 1
    fi
    
    PLACE_ID_ARRAY=$(echo "$RECOMMEND_RESPONSE" | jq -c '[.places[].place_id] | .[0:2]' 2>/dev/null)
    
    if [ -z "$PLACE_ID_ARRAY" ] || [ "$PLACE_ID_ARRAY" = "[]" ] || [ "$PLACE_ID_ARRAY" = "null" ]; then
        echo "エラー: Place IDを抽出できませんでした"
        echo "レスポンス: $RECOMMEND_RESPONSE"
        exit 1
    fi
    
    echo "取得したPlace ID（最初の2つ）:"
    echo "$PLACE_ID_ARRAY"
    echo ""
    
    # ステップ2: 取得したPlace IDを使って/result APIをテスト
    echo "【ステップ2】取得したPlace IDで/result APIをテスト"
    echo "----------------------------------------"
    REQUEST_BODY=$(echo "$PLACE_ID_ARRAY" | jq -c '{places: .}' 2>/dev/null)
    if [ -z "$REQUEST_BODY" ] || [ "$REQUEST_BODY" = "null" ]; then
        echo "エラー: リクエストボディの作成に失敗しました"
        exit 1
    fi
    echo "リクエストボディ: $REQUEST_BODY"
    echo ""
    curl -X POST "${BASE_URL}/result" \
      -H "Content-Type: application/json" \
      -d "$REQUEST_BODY" \
      -w "\n\nHTTP Status: %{http_code}\n" \
      -v | jq '.' 2>/dev/null || cat
else
    echo "警告: jqがインストールされていないため、Place IDの自動抽出ができません"
    echo "以下の手順で手動でテストしてください:"
    echo "1. 上記のレスポンスからplace_idを2つコピー"
    echo "2. 以下のコマンドを実行（PLACE_ID_1とPLACE_ID_2を置き換えてください）:"
    echo ""
    echo "curl -X POST \"${BASE_URL}/result\" \\"
    echo "  -H \"Content-Type: application/json\" \\"
    echo "  -d '{\"places\": [\"PLACE_ID_1\", \"PLACE_ID_2\"]}'"
fi

echo ""
echo "========================================="
echo "テスト完了"
echo "========================================="

