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
        max_candidates: int = 20,
        priority_locations: Optional[List[Dict[str, float]]] = None
    ) -> List[Dict[str, Any]]:
        """ルート上の推薦スポットを探索"""
        import sys
        print(f"[DEBUG] find_recommended_places called with priority_locations: {priority_locations}", file=sys.stderr, flush=True)
        
        # Polylineをデコード
        route_points = self.decode_polyline(route_polyline)
        print(f"[DEBUG] Route points decoded: {len(route_points)} points", file=sys.stderr, flush=True)

        # 優先検索ポイントを決定
        sample_points = []
        
        # 優先位置（must_places）の周辺を重点的に検索
        if priority_locations:
            import sys
            print(f"[DEBUG] Generating priority points around {len(priority_locations)} priority locations", file=sys.stderr, flush=True)
            # 各優先位置の中心から直接検索（最も重要）
            for priority_loc in priority_locations:
                # 優先位置の中心を最優先ポイントとして追加
                sample_points.append({
                    "lat": priority_loc["lat"],
                    "lng": priority_loc["lng"],
                    "priority": True,
                    "radius": 2000  # 中心から2000m以内を検索
                })
                # さらに周辺からもポイントを生成（補完的）
                for radius in [500, 1000, 1500]:
                    # 4方向からポイントを生成（円周上）
                    for angle in range(0, 360, 90):
                        import math
                        lat_offset = (radius / 111000) * math.cos(math.radians(angle))
                        lng_offset = (radius / 111000) * math.sin(math.radians(angle)) / math.cos(math.radians(priority_loc["lat"]))
                        sample_points.append({
                            "lat": priority_loc["lat"] + lat_offset,
                            "lng": priority_loc["lng"] + lng_offset,
                            "priority": True,
                            "radius": 1500  # 周辺ポイントからは1500m以内を検索
                        })
        
        # ルート上からも数点をサンプリング（優先度は低い）
        route_sample_count = max(3, 10 - len(sample_points) // 8)  # 優先ポイントが少ない場合はルートからも取得
        step = max(1, len(route_points) // route_sample_count) if route_points else 0
        for i in range(0, len(route_points), step):
            if len(sample_points) >= 50:  # 最大50ポイント
                break
            sample_points.append({
                "lat": route_points[i]["lat"],
                "lng": route_points[i]["lng"],
                "priority": False
            })

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
            # 優先ポイントから先に検索
            priority_points = [p for p in sample_points if p.get("priority", False)]
            non_priority_points = [p for p in sample_points if not p.get("priority", False)]
            import sys
            print(f"[DEBUG] Sample points: {len(priority_points)} priority, {len(non_priority_points)} non-priority", file=sys.stderr, flush=True)
            
            # 優先ポイントから検索
            for point in priority_points:
                for tag in interest_tags:
                    keyword = keyword_map.get(tag, tag)
                    url = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
                    # 優先ポイントの場合は指定された半径を使用、それ以外は1500m
                    search_radius = point.get("radius", 1500)
                    params = {
                        "location": f"{point['lat']},{point['lng']}",
                        "radius": search_radius,
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
                                # 優先ポイントからの結果には優先度スコアを付与
                                all_candidates.append({
                                    "place_id": place_id,
                                    "name": result["name"],
                                    "lat": result["geometry"]["location"]["lat"],
                                    "lng": result["geometry"]["location"]["lng"],
                                    "priority_score": 10,  # 優先ポイントからの結果
                                })
                    
                    # 十分な候補が見つかったら早期終了
                    if len(all_candidates) >= max_candidates * 2:
                        break
                if len(all_candidates) >= max_candidates * 2:
                    break
            
            # 優先ポイントで十分な候補が見つからなかった場合のみ、非優先ポイントから検索
            if len(all_candidates) < max_candidates:
                for point in non_priority_points:
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
                                        "priority_score": 1,  # 非優先ポイントからの結果
                                    })
                        
                        if len(all_candidates) >= max_candidates * 2:
                            break
                    if len(all_candidates) >= max_candidates * 2:
                        break

        import sys
        print(f"[DEBUG] Total candidates found: {len(all_candidates)}", file=sys.stderr, flush=True)
        
        # 優先度スコアでソートし、重複排除
        # 優先度が高い順、距離が近い順にソート
        if priority_locations:
            print(f"[DEBUG] Sorting by priority and distance to {len(priority_locations)} priority locations", file=sys.stderr, flush=True)
            for candidate in all_candidates:
                # 最も近い優先位置までの距離を計算
                min_distance = float('inf')
                for priority_loc in priority_locations:
                    distance = self._calculate_distance(
                        candidate["lat"], candidate["lng"],
                        priority_loc["lat"], priority_loc["lng"]
                    )
                    min_distance = min(min_distance, distance)
                candidate["distance_to_priority"] = min_distance
                # 優先度スコアに距離の逆数を加算（近いほど高スコア）
                candidate["priority_score"] += max(0, 10 - (min_distance / 100))
            
            # 優先度スコアと距離でソート
            all_candidates.sort(key=lambda x: (-x.get("priority_score", 0), x.get("distance_to_priority", float('inf'))))
        else:
            # 優先位置がない場合は優先度スコアのみでソート
            all_candidates.sort(key=lambda x: -x.get("priority_score", 0))

        # 重複排除と上位N件に絞る
        unique_candidates = []
        seen = set()
        for candidate in all_candidates:
            if candidate["place_id"] not in seen:
                seen.add(candidate["place_id"])
                # priority_scoreとdistance_to_priorityを削除（レスポンスに含めない）
                priority_score = candidate.pop("priority_score", 0)
                distance = candidate.pop("distance_to_priority", None)
                if distance is not None:
                    import sys
                    print(f"[DEBUG] Candidate: {candidate.get('name', 'Unknown')} - Priority: {priority_score}, Distance: {distance:.0f}m", file=sys.stderr, flush=True)
                unique_candidates.append(candidate)
                if len(unique_candidates) >= max_candidates:
                    break

        import sys
        print(f"[DEBUG] Returning {len(unique_candidates)} unique candidates", file=sys.stderr, flush=True)
        return unique_candidates
    
    def _calculate_distance(self, lat1: float, lng1: float, lat2: float, lng2: float) -> float:
        """2点間の距離を計算（メートル単位）"""
        import math
        R = 6371000  # 地球の半径（メートル）
        dlat = math.radians(lat2 - lat1)
        dlng = math.radians(lng2 - lng1)
        a = math.sin(dlat / 2) ** 2 + math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) * math.sin(dlng / 2) ** 2
        c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))
        return R * c


