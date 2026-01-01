#!/bin/bash

# リコメンドAPIテスト用スクリプト
# 使用方法: ./test_recommend_api.sh

# APIのベースURL（デフォルトはlocalhost:8080）
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "リコメンドAPI テスト"
echo "========================================="
echo "API URL: ${BASE_URL}/recommend"
echo ""

# テストケース1: 基本的なリクエスト
echo "【テストケース1】基本的なリクエスト"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮", "福岡タワー"],
    "interest_tags": ["カフェ", "神社"]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.'

echo ""
echo ""

# テストケース2: 出発地点とゴール地点を指定
echo "【テストケース2】出発地点・ゴール地点指定"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮", "福岡タワー"],
    "interest_tags": ["カフェ", "レストラン"],
    "start_place": "博多駅",
    "goal_place": "天神"
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.'

echo ""
echo ""

# テストケース3: 複数の興味タグ
echo "【テストケース3】複数の興味タグ"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": ["太宰府天満宮"],
    "interest_tags": ["カフェ", "神社", "観光", "ショッピング"]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.'

echo ""
echo ""

# テストケース4: エラーテスト（必須パラメータなし）
echo "【テストケース4】エラーテスト（必須パラメータなし）"
echo "----------------------------------------"
curl -X POST "${BASE_URL}/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "must_places": []
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.'

echo ""
echo "========================================="
echo "テスト完了"
echo "========================================="

