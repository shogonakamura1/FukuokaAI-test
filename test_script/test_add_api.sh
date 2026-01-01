#!/bin/bash

# 場所追加APIテスト用スクリプト
# 使用方法: ./test_add_api.sh

# APIのベースURL（デフォルトはlocalhost:8080）
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "場所追加API テスト"
echo "========================================="
echo "API URL: ${BASE_URL}/add/:place_id"
echo ""

# テストケース1: 基本的なリクエスト
echo "【テストケース1】基本的なリクエスト（place_idを指定）"
echo "----------------------------------------"
PLACE_ID="ChIJz_VfLQKJRTkRxqxoKTN4xqU"  # 博多駅のPlace ID（例）
curl -X POST "${BASE_URL}/add/${PLACE_ID}" \
  -H "Content-Type: application/json" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v

echo ""
echo ""

# テストケース2: place_idが空の場合（エラーテスト）
echo "【テストケース2】エラーテスト（place_idが空）"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/add/" \
  -H "Content-Type: application/json" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -v

echo ""
echo "========================================="
echo "テスト完了"
echo "========================================="

