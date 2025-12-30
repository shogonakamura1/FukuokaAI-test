from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional, Dict, Any
import os
from dotenv import load_dotenv

from services.places import PlacesService
from services.llm import LLMService
from services.route import RouteService

load_dotenv()

app = FastAPI(title="Fukuoka AI ML Service")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

places_service = PlacesService()
llm_service = LLMService()
route_service = RouteService()


class RecommendRequest(BaseModel):
    start: str
    must_places: List[str]
    interest_tags: List[str]
    free_text: Optional[str] = None


class RecomputeRouteRequest(BaseModel):
    start: str
    waypoints: List[Dict[str, float]]


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.post("/recommend")
async def recommend(request: RecommendRequest):
    try:
        # 1. must_placesを解決
        must_places = []
        for place_name in request.must_places:
            place = await places_service.search_place(place_name)
            if place:
                must_places.append(place)
            else:
                # 見つからない場合はスキップ（エラーにはしない）
                continue

        if not must_places:
            raise HTTPException(
                status_code=400,
                detail="指定された場所が見つかりませんでした"
            )

        # 2. ルート生成
        route_result = await route_service.create_route(
            start="Hakata Station",
            waypoints=[{"lat": p["lat"], "lng": p["lng"]} for p in must_places]
        )

        # 3. ルート上の推薦スポット探索
        # must_placesの座標を渡して、近くを優先的に検索
        candidates = await places_service.find_recommended_places(
            route_polyline=route_result["polyline"],
            interest_tags=request.interest_tags,
            exclude_place_ids=[p["place_id"] for p in must_places],
            priority_locations=[{"lat": p["lat"], "lng": p["lng"]} for p in must_places]
        )

        # 4. Place Detailsで詳細取得
        for candidate in candidates:
            details = await places_service.get_place_details(candidate["place_id"])
            if details:
                candidate.update(details)

        # 5. LLMで理由とレビュー要約を生成
        if candidates:
            llm_results = await llm_service.generate_recommendations(
                candidates=candidates,
                interest_tags=request.interest_tags,
                free_text=request.free_text
            )
            for i, result in enumerate(llm_results):
                if i < len(candidates):
                    candidates[i]["reason"] = result.get("reason", "")
                    candidates[i]["review_summary"] = result.get("review_summary", "")

        # 6. 初期旅程を作成（must_placesを順序通りに）
        initial_itinerary = []
        for i, place in enumerate(must_places):
            initial_itinerary.append({
                "id": f"place_{i}",
                "place_id": place["place_id"],
                "name": place["name"],
                "lat": place["lat"],
                "lng": place["lng"],
                "kind": "must",
                "stay_minutes": 60,
                "order_index": i,
            })

        return {
            "must": must_places,
            "candidates": candidates[:20],  # 最大20件
            "initial_itinerary": initial_itinerary,
            "route": {
                "polyline": route_result["polyline"],
                "legs": route_result.get("legs", [])
            }
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/recompute-route")
async def recompute_route(request: RecomputeRouteRequest):
    try:
        route_result = await route_service.create_route(
            start=request.start,
            waypoints=request.waypoints
        )
        return {
            "route": {
                "polyline": route_result["polyline"],
                "legs": route_result.get("legs", [])
            }
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)


