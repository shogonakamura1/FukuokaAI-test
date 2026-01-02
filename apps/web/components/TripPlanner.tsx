'use client'

import { useState } from 'react'
import { useSession } from 'next-auth/react'
import TripForm from './TripForm'
import ItineraryTimeline from './ItineraryTimeline'
import MapView from './MapView'

export interface Place {
  id?: string
  place_id: string
  name: string
  lat: number
  lng: number
  kind?: 'must' | 'recommended' | 'start' | 'goal'
  stay_minutes?: number
  order_index?: number
  time_range?: string
  reason?: string
  review_summary?: string
  photo_url?: string
  category?: string
  rating?: number
  address?: string
  travel_time_from_previous?: string // 前の地点からの移動時間
}

export interface Route {
  legs: Array<{
    start_location: { lat: number; lng: number }
    end_location: { lat: number; lng: number }
    distance_meters: number
    duration: string
  }>
  distance_meters: number
  duration: string
  optimized_order: number[]
}

export default function TripPlanner() {
  const { data: session } = useSession()
  const [recommendedPlaces, setRecommendedPlaces] = useState<Place[]>([])
  const [selectedPlaceIds, setSelectedPlaceIds] = useState<Set<string>>(new Set())
  const [itinerary, setItinerary] = useState<Place[]>([])
  const [route, setRoute] = useState<Route | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [startPlace, setStartPlace] = useState<string>('')
  const [goalPlace, setGoalPlace] = useState<string>('')
  const [mustPlaces, setMustPlaces] = useState<Place[]>([]) // must_placesを状態として保存

  const apiUrl = (typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_API_URL) || 'http://localhost:8080'

  // 1. リコメンドAPIを呼び出す
  const handleGenerateTrip = async (data: {
    must_places: string[]
    interest_tags: string[]
    start_place: string
    goal_place: string
    free_text?: string
  }) => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/recommend`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          must_places: data.must_places,
          interest_tags: data.interest_tags,
          start_place: data.start_place,
          goal_place: data.goal_place,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error?.message || 'リコメンド取得に失敗しました')
      }

      const result = await response.json()
      const places: Place[] = (result.places || [])
        .slice(0, 4) // 最大4件に制限（念のため）
        .map((place: any) => ({
          place_id: place.place_id,
          name: place.name || '',
          lat: place.lat || 0,
          lng: place.lng || 0,
          photo_url: place.photo_url,
          rating: place.rating,
          review_summary: place.review_summary,
          category: place.category,
          address: place.address,
        }))

      console.log('リコメンド結果:', places.length, '件')
      setRecommendedPlaces(places)
      setSelectedPlaceIds(new Set())
      setItinerary([])
      setRoute(null)
      setStartPlace(data.start_place)
      setGoalPlace(data.goal_place)
      
      // must_placesのplace_idを取得して状態に保存
      // 詳細情報は後で/result APIから取得されるので、ここではplace_idとnameのみ保存
      const mustPlacePromises = data.must_places.map(async (placeName): Promise<Place | null> => {
        const placeId = await getPlaceIdFromName(placeName)
        if (placeId) {
          return {
            place_id: placeId,
            name: placeName,
            lat: 0, // 後で/result APIから取得
            lng: 0, // 後で/result APIから取得
            kind: 'must' as const,
          }
        }
        return null
      })
      
      const mustPlacesResults = await Promise.all(mustPlacePromises)
      const mustPlacesData = mustPlacesResults.filter((p): p is Place => p !== null)
      setMustPlaces(mustPlacesData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  // 2. 候補地の選択/選択解除
  const handleToggleCandidate = (placeId: string) => {
    const newSelected = new Set(selectedPlaceIds)
    if (newSelected.has(placeId)) {
      newSelected.delete(placeId)
    } else {
      newSelected.add(placeId)
    }
    setSelectedPlaceIds(newSelected)
  }

  // 3. 選択した場所をリストに追加（/add/:place_id を呼び出す）
  // 追加後、即座にルート計算を実行して移動時間を表示
  const handleAddSelectedPlaces = async () => {
    if (selectedPlaceIds.size === 0) {
      setError('少なくとも1つの場所を選択してください')
      return
    }

    setLoading(true)
    setError(null)

    try {
      // 各選択した場所に対して /add/:place_id を呼び出す
      const addPromises = Array.from(selectedPlaceIds).map(async (placeId) => {
        const response = await fetch(`${apiUrl}/add/${placeId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        })
        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.error?.message || `場所の追加に失敗しました: ${placeId}`)
        }
        return response.json()
      })

      await Promise.all(addPromises)

      // 選択した場所をitineraryに追加（クライアント側で管理）
      const selectedPlaces = recommendedPlaces.filter(p => selectedPlaceIds.has(p.place_id))
      // must_placesと選択した場所を結合（重複を避ける）
      const existingPlaceIds = new Set(itinerary.map(p => p.place_id))
      const newMustPlaces = mustPlaces.filter(p => !existingPlaceIds.has(p.place_id))
      const newSelectedPlaces = selectedPlaces.filter(p => !existingPlaceIds.has(p.place_id))
      const newItinerary = [...itinerary, ...newMustPlaces, ...newSelectedPlaces]
      setItinerary(newItinerary)
      setSelectedPlaceIds(new Set())
      
      // 即座にルート計算を実行（移動時間を表示するため）
      if (newItinerary.length >= 2 && startPlace && goalPlace) {
        await calculateRouteWithPlaces(newItinerary)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }
  
  // ルート計算の共通処理（handleAddSelectedPlacesとhandleCalculateRouteから使用）
  const calculateRouteWithPlaces = async (placesToCalculate: Place[]) => {
    if (placesToCalculate.length < 2) {
      setError('ルート計算には最低2つの場所が必要です')
      return
    }

    try {
      // start_placeとgoal_placeのplace_idを取得
      console.log('スタート地点取得中:', startPlace)
      const startPlaceId = await getPlaceIdFromName(startPlace)
      console.log('ゴール地点取得中:', goalPlace)
      const goalPlaceId = await getPlaceIdFromName(goalPlace)

      if (!startPlaceId) {
        throw new Error(`スタート地点「${startPlace}」のplace_idを取得できませんでした`)
      }
      if (!goalPlaceId) {
        throw new Error(`ゴール地点「${goalPlace}」のplace_idを取得できませんでした`)
      }

      console.log('取得したplace_id:', { startPlaceId, goalPlaceId, itinerary: placesToCalculate.map(p => p.place_id) })

      // スタート地点、中間地点（must_places + selected places）、ゴール地点の順に配置
      const placeIds = [startPlaceId, ...placesToCalculate.map(p => p.place_id), goalPlaceId]
      console.log('ルート計算リクエスト:', { placeIds })

      const response = await fetch(`${apiUrl}/result`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          places: placeIds,
        }),
      })

      if (!response.ok) {
        let errorMessage = 'ルート計算に失敗しました'
        try {
          const errorData = await response.json()
          errorMessage = errorData.error?.message || errorMessage
          console.error('ルート計算エラー詳細:', JSON.stringify(errorData, null, 2))
          console.error('エラーコード:', errorData.error?.code)
          console.error('エラーメッセージ:', errorData.error?.message)
        } catch (e) {
          // JSONパースに失敗した場合、テキストとして読み取る
          const text = await response.text()
          console.error('ルート計算エラー（テキスト）:', text)
          errorMessage = text || errorMessage
        }
        throw new Error(errorMessage)
      }

      const result = await response.json()
      
      // デバッグ: ルート情報を確認
      console.log('ルート計算結果:', JSON.stringify(result.route, null, 2))
      if (result.route?.legs && result.route.legs.length > 0) {
        console.log('Legs情報:', result.route.legs)
        result.route.legs.forEach((leg: any, idx: number) => {
          console.log(`Leg ${idx}:`, {
            distance_meters: leg.distance_meters,
            distanceMeters: leg.distanceMeters,
            duration: leg.duration,
            allKeys: Object.keys(leg)
          })
        })
      }
      
      // 最適化された順序でplacesを更新（移動時間を含める）
      const optimizedPlaces: Place[] = (result.places || []).map((place: any, index: number) => {
        const placeObj: Place = {
          place_id: place.place_id,
          name: place.name || '',
          lat: place.lat || 0,
          lng: place.lng || 0,
          photo_url: place.photo_url,
          rating: place.rating,
          address: place.address,
        }

        // 最初の地点はスタート地点
        if (index === 0) {
          placeObj.kind = 'start'
        } else if (index === result.places.length - 1) {
          // 最後の地点はゴール地点
          placeObj.kind = 'goal'
        } else {
          // 中間地点は、元のplacesToCalculateからkindを取得
          const originalPlace = placesToCalculate.find(p => p.place_id === place.place_id)
          placeObj.kind = originalPlace?.kind || 'recommended'
        }

        // 移動時間と距離を設定（route.legsから取得）
        if (result.route?.legs && index > 0 && result.route.legs[index - 1]) {
          const leg = result.route.legs[index - 1]
          // durationは "1800s" のような形式なので、分に変換
          const durationSeconds = parseInt(leg.duration?.replace('s', '') || '0', 10)
          const durationMinutes = Math.round(durationSeconds / 60)
          // APIレスポンスはdistance_meters（スネークケース）で返される
          // 念のため両方の形式を確認
          const distanceMeters = leg.distance_meters ?? (leg as any).distanceMeters ?? 0
          console.log(`Place ${index} (${place.name}): distance_meters=${leg.distance_meters}, distanceMeters=${(leg as any).distanceMeters}, final=${distanceMeters}`)
          const distanceKm = distanceMeters > 0 ? (distanceMeters / 1000).toFixed(1) : '0.0'
          placeObj.travel_time_from_previous = `${durationMinutes}分 / ${distanceKm}km`
        }

        return placeObj
      })

      setItinerary(optimizedPlaces)
      setRoute(result.route || null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
      throw err // 呼び出し元でエラーを処理できるように再スロー
    }
  }

  // 場所名からplace_idを取得するヘルパー関数（バックエンドAPI経由）
  const getPlaceIdFromName = async (placeName: string): Promise<string | null> => {
    try {
      const response = await fetch(`${apiUrl}/geocoding`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          place_name: placeName,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        console.error('place_id取得エラー:', errorData.error?.message || 'Unknown error')
        return null
      }

      const data = await response.json()
      return data.place_id || null
    } catch (err) {
      console.error('place_id取得エラー:', err)
      return null
    }
  }

  // 4. ルート計算（/result APIを呼び出す）
  const handleCalculateRoute = async () => {
    setLoading(true)
    setError(null)

    try {
      // must_placesも含めてルート計算
      const allPlaces = [...mustPlaces, ...itinerary.filter(p => p.kind !== 'must')]
      await calculateRouteWithPlaces(allPlaces)
    } catch (err) {
      // エラーはcalculateRouteWithPlaces内で既にsetErrorされている
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    setRecommendedPlaces([])
    setSelectedPlaceIds(new Set())
    setItinerary([])
    setRoute(null)
    setStartPlace('')
    setGoalPlace('')
    setMustPlaces([]) // must_placesもリセット
    setError(null)
  }

  console.log('TripPlanner render - itinerary length:', itinerary.length, 'route:', route ? 'exists' : 'null')
  
  return (
    <div className="flex h-[calc(100vh-80px)]">
      <div className="w-1/2 overflow-y-auto p-4">
        {error && (
          <div className="mb-4 p-4 bg-red-100 text-red-700 rounded">
            {error}
          </div>
        )}
        {recommendedPlaces.length === 0 && itinerary.length === 0 ? (
          <TripForm onSubmit={handleGenerateTrip} loading={loading} />
        ) : (
          <>
            <div className="mb-4 flex gap-2 items-center">
              <button
                onClick={handleReset}
                className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
              >
                新しい旅程を作成
              </button>
            </div>

            {recommendedPlaces.length > 0 && (
              <div className="mb-6">
                <h2 className="text-xl font-bold mb-4">おすすめスポット</h2>
                <div className="space-y-4 mb-4">
                  {recommendedPlaces.map((place, index) => (
                    <div
                      key={place.place_id || `candidate-${index}`}
                      className="bg-white p-4 rounded shadow"
                    >
                      <div className="flex items-start gap-3">
                        <input
                          type="checkbox"
                          checked={selectedPlaceIds.has(place.place_id)}
                          onChange={() => handleToggleCandidate(place.place_id)}
                          className="mt-1"
                        />
                        <div className="flex-1">
                          {place.photo_url && (
                            <img
                              src={place.photo_url}
                              alt={place.name || 'スポット画像'}
                              className="w-full h-48 object-cover rounded mb-3"
                            />
                          )}
                          <h3 className="font-semibold text-lg">{place.name || '名前不明のスポット'}</h3>
                          {place.category && (
                            <span className="text-sm text-gray-600">{place.category}</span>
                          )}
                          {place.review_summary && (
                            <p className="text-xs text-gray-500 mt-1">{place.review_summary}</p>
                          )}
                          {place.rating && (
                            <p className="text-sm text-gray-600 mt-1">評価: {place.rating}</p>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
                <button
                  onClick={handleAddSelectedPlaces}
                  disabled={loading || selectedPlaceIds.size === 0}
                  className="w-full px-6 py-3 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
                >
                  {loading ? '追加中...' : `選択した${selectedPlaceIds.size}件を追加`}
                </button>
              </div>
            )}

            {itinerary.length > 0 && (
              <div className="mb-6">
                <h2 className="text-xl font-bold mb-4">旅程リスト</h2>
                <div className="mb-4 p-3 bg-blue-50 rounded">
                  <p className="text-sm text-gray-700">
                    選択した場所: {itinerary.length}件
                  </p>
                </div>
                <ItineraryTimeline
                  itinerary={itinerary}
                  onReorder={(orderedPlaceIds) => {
                    // 順序変更は後で実装
                    const reordered = orderedPlaceIds
                      .map(id => itinerary.find(p => (p.id || p.place_id) === id))
                      .filter((p): p is Place => p !== undefined)
                    setItinerary(reordered)
                  }}
                />
                <button
                  onClick={handleCalculateRoute}
                  disabled={loading || itinerary.length < 2}
                  className="w-full mt-4 px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
                >
                  {loading ? '計算中...' : 'ルートを計算'}
                </button>
              </div>
            )}
          </>
        )}
      </div>
      <div className="w-1/2">
        <MapView
          itinerary={itinerary}
          route={route}
        />
      </div>
    </div>
  )
}

