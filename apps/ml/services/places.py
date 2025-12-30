import httpx
import os
from typing import List, Dict, Optional, Any
import math

GOOGLE_MAPS_API_KEY = os.getenv("GOOGLE_MAPS_API_KEY", "")


class PlacesService:
    def __init__(self):
        self.api_key = GOOGLE_MAPS_API_KEY
        if not self.api_key:
            raise ValueError("GOOGLE_MAPS_API_KEY is not set")

    async def search_place(self, query: str) -> Optional[Dict[str, Any]]:
        """Places Text Searchで場所を検索"""
        async with httpx.AsyncClient() as client:
            url = "https://maps.googleapis.com/maps/api/place/textsearch/json"
            params = {
                "query": f"{query} 福岡",
                "key": self.api_key,
                "language": "ja",
            }
            response = await client.get(url, params=params)
            data = response.json()

            if data.get("status") == "OK" and data.get("results"):
                result = data["results"][0]
                return {
                    "place_id": result["place_id"],
                    "name": result["name"],
                    "lat": result["geometry"]["location"]["lat"],
                    "lng": result["geometry"]["location"]["lng"],
                }
            return None

    async def get_place_details(self, place_id: str) -> Optional[Dict[str, Any]]:
        """Place Detailsで詳細情報を取得"""
        async with httpx.AsyncClient() as client:
            url = "https://maps.googleapis.com/maps/api/place/details/json"
            params = {
                "place_id": place_id,
                "key": self.api_key,
                "language": "ja",
                "fields": "name,photo,review,type",
            }
            response = await client.get(url, params=params)
            data = response.json()

            if data.get("status") == "OK" and data.get("result"):
                result = data["result"]
                details = {}

                # 写真URLを取得
                if result.get("photos"):
                    photo_ref = result["photos"][0].get("photo_reference")
                    if photo_ref:
                        details["photo_url"] = (
                            f"https://maps.googleapis.com/maps/api/place/photo"
                            f"?maxwidth=400&photoreference={photo_ref}&key={self.api_key}"
                        )

                # カテゴリを取得
                if result.get("types"):
                    types = result["types"]
                    # 最初の有効なタイプを使用
                    category_map = {
                        "cafe": "カフェ",
                        "restaurant": "レストラン",
                        "tourist_attraction": "観光地",
                        "shopping_mall": "ショッピング",
                        "park": "公園",
                        "temple": "寺社",
                    }
                    for t in types:
                        if t in category_map:
                            details["category"] = category_map[t]
                            break
                    if "category" not in details:
                        details["category"] = types[0] if types else "その他"

                # レビュー（最大3件）
                if result.get("reviews"):
                    details["reviews"] = result["reviews"][:3]

                return details
            return None

    def decode_polyline(self, encoded: str) -> List[Dict[str, float]]:
        """Polylineをデコードして座標リストに変換"""
        points = []
        index = 0
        lat = 0
        lng = 0

        while index < len(encoded):
            # 緯度
            shift = 0
            result = 0
            while True:
                b = ord(encoded[index]) - 63
                index += 1
                result |= (b & 0x1f) << shift
                shift += 5
                if b < 0x20:
                    break
            dlat = ~(result >> 1) if (result & 1) else (result >> 1)
            lat += dlat

            # 経度
            shift = 0
            result = 0
            while True:
                b = ord(encoded[index]) - 63
                index += 1
                result |= (b & 0x1f) << shift
                shift += 5
                if b < 0x20:
                    break
            dlng = ~(result >> 1) if (result & 1) else (result >> 1)
            lng += dlng

            points.append({"lat": lat * 1e-5, "lng": lng * 1e-5})

        return points

    async def find_recommended_places(
        self,
        route_polyline: str,
        interest_tags: List[str],
        exclude_place_ids: List[str],
        max_candidates: int = 20
    ) -> List[Dict[str, Any]]:
        """ルート上の推薦スポットを探索"""
        # Polylineをデコード
        route_points = self.decode_polyline(route_polyline)

        # ルート上から数点をサンプリング（最大10点）
        sample_points = []
        step = max(1, len(route_points) // 10)
        for i in range(0, len(route_points), step):
            sample_points.append(route_points[i])
            if len(sample_points) >= 10:
                break

        # 各サンプルポイントでNearby Searchを実行
        all_candidates = []
        seen_place_ids = set(exclude_place_ids)

        # タグをキーワードに変換
        keyword_map = {
            "カフェ": "cafe",
            "屋台": "food",
            "景色": "viewpoint",
            "寺社": "temple",
            "買い物": "shopping",
            "グルメ": "restaurant",
            "自然": "park",
        }

        async with httpx.AsyncClient() as client:
            for point in sample_points:
                for tag in interest_tags:
                    keyword = keyword_map.get(tag, tag)
                    url = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
                    params = {
                        "location": f"{point['lat']},{point['lng']}",
                        "radius": 1500,
                        "type": keyword if keyword in ["cafe", "restaurant", "park", "temple"] else None,
                        "keyword": keyword if keyword not in ["cafe", "restaurant", "park", "temple"] else None,
                        "key": self.api_key,
                        "language": "ja",
                    }
                    # typeがNoneの場合は削除
                    if params["type"] is None:
                        del params["type"]
                    if params["keyword"] is None:
                        del params["keyword"]

                    response = await client.get(url, params=params)
                    data = response.json()

                    if data.get("status") == "OK" and data.get("results"):
                        for result in data["results"]:
                            place_id = result["place_id"]
                            if place_id not in seen_place_ids:
                                seen_place_ids.add(place_id)
                                all_candidates.append({
                                    "place_id": place_id,
                                    "name": result["name"],
                                    "lat": result["geometry"]["location"]["lat"],
                                    "lng": result["geometry"]["location"]["lng"],
                                })

        # 重複排除と上位N件に絞る
        unique_candidates = []
        seen = set()
        for candidate in all_candidates:
            if candidate["place_id"] not in seen:
                seen.add(candidate["place_id"])
                unique_candidates.append(candidate)
                if len(unique_candidates) >= max_candidates:
                    break

        return unique_candidates


