import httpx
import os
from typing import List, Dict, Any

GOOGLE_MAPS_API_KEY = os.getenv("GOOGLE_MAPS_API_KEY", "")


class RouteService:
    def __init__(self):
        self.api_key = GOOGLE_MAPS_API_KEY
        if not self.api_key:
            raise ValueError("GOOGLE_MAPS_API_KEY is not set")

    async def create_route(
        self,
        start: str,
        waypoints: List[Dict[str, float]],
        mode: str = "driving"
    ) -> Dict[str, Any]:
        """Directions APIでルートを生成"""
        async with httpx.AsyncClient() as client:
            url = "https://maps.googleapis.com/maps/api/directions/json"

            # 出発地点を座標に変換（簡易実装: 博多駅固定）
            if start == "Hakata Station":
                origin = "33.5904,130.4208"  # 博多駅の座標
            else:
                origin = start

            if not waypoints:
                # waypointsがない場合は直接目的地へ
                params = {
                    "origin": origin,
                    "destination": origin,  # 暫定
                    "mode": mode,
                    "key": self.api_key,
                    "language": "ja",
                }
            else:
                # waypointsを文字列に変換
                waypoint_str = "|".join([f"{w['lat']},{w['lng']}" for w in waypoints])
                destination = waypoint_str.split("|")[-1]  # 最後のwaypointを目的地に

                params = {
                    "origin": origin,
                    "destination": destination,
                    "waypoints": waypoint_str,
                    "mode": mode,
                    "key": self.api_key,
                    "language": "ja",
                }

            response = await client.get(url, params=params)
            data = response.json()

            if data.get("status") != "OK" or not data.get("routes"):
                raise Exception(f"Directions API error: {data.get('status')}")

            route = data["routes"][0]
            leg = route["legs"][0] if route.get("legs") else None

            return {
                "polyline": route["overview_polyline"]["points"],
                "legs": [
                    {
                        "duration": leg["duration"]["value"] if leg else 0,
                        "distance": leg["distance"]["value"] if leg else 0,
                    }
                ] if leg else [],
            }


