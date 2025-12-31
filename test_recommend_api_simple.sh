#!/bin/bash

# シンプルなテストコマンド（jqなしでも動作）
# 使用方法: ./test_recommend_api_simple.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "リコメンドAPI テスト (Simple)"
echo "API URL: ${BASE_URL}/recommend"
echo ""

curl -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮", "福岡タワー"],
    "interest_tags": ["カフェ", "神社"]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n"

