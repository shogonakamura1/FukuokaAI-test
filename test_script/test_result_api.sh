#!/bin/bash

# ルート提案APIテスト用スクリプト
# 使用方法: ./test_result_api.sh

# APIのベースURL（デフォルトはlocalhost:8080）
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "ルート提案API テスト"
echo "========================================="
echo "API URL: ${BASE_URL}/result"
echo ""

# テストケース1: 基本的なリクエスト（有効なPlace IDが必要）
echo "【テストケース1】基本的なリクエスト（2つの場所）"
echo "----------------------------------------"
echo "注意:"
echo "  - 実際に有効なPlace IDが必要です"
echo "  - 最初の場所が【出発地点】、最後の場所が【ゴール地点】として設定されます"
echo "以下の方法でPlace IDを取得してから使用してください:"
echo "  curl -X POST \"${BASE_URL}/recommend\" \\"
echo "    -H \"Content-Type: application/json\" \\"
echo "    -d '{\"must_places\": [\"博多駅\"], \"interest_tags\": [\"カフェ\"]}' | jq '.places[].place_id'"
echo ""
echo "または、test_result_api_with_place_ids.sh を使用してください（自動でPlace IDを取得します）"
echo ""
echo "ここでは、例として無効なPlace IDを使用しています（エラーになる可能性があります）:"
curl -X POST "${BASE_URL}/result" \
  -H "Content-Type: application/json" \
  -d '{
    "places": [
      "ChIJz_VfLQKJRTkRxqxoKTN4xqU",
      "ChIJ0_VL5wKJRTkRklWXjPv_sU"
    ]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v | jq '.' 2>/dev/null || cat

echo ""
echo ""

# テストケース2: エラーテスト（placesが空）
echo "【テストケース2】エラーテスト（placesが空）"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/result" \
  -H "Content-Type: application/json" \
  -d '{
    "places": []
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v | jq '.' 2>/dev/null || cat

echo ""
echo ""

# テストケース3: エラーテスト（リクエストボディなし）
echo "【テストケース3】エラーテスト（リクエストボディなし）"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/result" \
  -H "Content-Type: application/json" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v | jq '.' 2>/dev/null || cat

echo ""
echo "========================================="
echo "テスト完了"
echo "========================================="

