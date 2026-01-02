'use client'

import { useState, useEffect, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { Place } from './TripPlanner'

export default function RecommendPage() {
  console.log('=== RecommendPage component rendered ===')
  const router = useRouter()
  const [recommendedPlaces, setRecommendedPlaces] = useState<Place[]>([])
  const [selectedPlaceIds, setSelectedPlaceIds] = useState<Set<string>>(new Set())
  const [mustPlaces, setMustPlaces] = useState<Place[]>([])
  const [startPlace, setStartPlace] = useState<string>('')
  const [goalPlace, setGoalPlace] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [maxPossibleScore, setMaxPossibleScore] = useState<number>(100) // 理論的最大スコア（デフォルト100）

  const apiUrl = (typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_API_URL) || 'http://localhost:8080'

  // 場所名からplace_idを取得するヘルパー関数
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

  // リコメンドAPIを呼び出す
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
        .slice(0, 4)
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
          relevance_score: place.relevance_score,
        }))

      setRecommendedPlaces(places)
      // 理論的最大スコアを保存（デフォルト値100を設定）
      setMaxPossibleScore(result.max_possible_score || 100)
      setStartPlace(data.start_place)
      setGoalPlace(data.goal_place)

      // must_placesのplace_idを取得して状態に保存
      const mustPlacePromises = data.must_places.map(async (placeName): Promise<Place | null> => {
        const placeId = await getPlaceIdFromName(placeName)
        if (placeId) {
          return {
            place_id: placeId,
            name: placeName,
            lat: 0,
            lng: 0,
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

  // sessionStorageからフォームデータを読み取り、APIを呼び出す
  useEffect(() => {
    const loadFormDataAndFetch = async () => {
      const formDataStr = sessionStorage.getItem('formData')
      if (!formDataStr) {
        setError('フォームデータが見つかりません')
        return
      }

      try {
        const formData = JSON.parse(formDataStr)
        await handleGenerateTrip(formData)
      } catch (err) {
        console.error('フォームデータの読み込みエラー:', err)
        setError('フォームデータの読み込みに失敗しました')
      }
    }

    loadFormDataAndFetch()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 候補地の選択/選択解除
  const handleToggleCandidate = (placeId: string) => {
    const newSelected = new Set(selectedPlaceIds)
    if (newSelected.has(placeId)) {
      newSelected.delete(placeId)
    } else {
      newSelected.add(placeId)
    }
    setSelectedPlaceIds(newSelected)
  }

  // 選択した場所を追加して旅程ページへ遷移
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

      // 選択した場所をitineraryに追加
      const selectedPlaces = recommendedPlaces.filter(p => selectedPlaceIds.has(p.place_id))
      const newItinerary = [...mustPlaces, ...selectedPlaces]

      // start_placeとgoal_placeのplace_idを取得
      const startPlaceId = await getPlaceIdFromName(startPlace)
      const goalPlaceId = await getPlaceIdFromName(goalPlace)

      if (!startPlaceId) {
        throw new Error(`スタート地点「${startPlace}」のplace_idを取得できませんでした`)
      }
      if (!goalPlaceId) {
        throw new Error(`ゴール地点「${goalPlace}」のplace_idを取得できませんでした`)
      }

      // ルート計算
      const placeIds = [startPlaceId, ...newItinerary.map(p => p.place_id), goalPlaceId]
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
        } catch (e) {
          const text = await response.text()
          errorMessage = text || errorMessage
        }
        throw new Error(errorMessage)
      }

      const result = await response.json()

      // デバッグ: ルート情報を確認
      console.log('ルート計算結果 (RecommendPage):', JSON.stringify(result.route, null, 2))
      if (result.route?.legs && result.route.legs.length > 0) {
        console.log('Legs情報 (RecommendPage):', result.route.legs)
        result.route.legs.forEach((leg: any, idx: number) => {
          console.log(`Leg ${idx} (RecommendPage):`, {
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

        if (index === 0) {
          placeObj.kind = 'start'
        } else if (index === result.places.length - 1) {
          placeObj.kind = 'goal'
        } else {
          const originalPlace = newItinerary.find(p => p.place_id === place.place_id)
          placeObj.kind = originalPlace?.kind || 'recommended'
        }

        // 移動時間と距離を設定
        if (result.route?.legs && index > 0 && result.route.legs[index - 1]) {
          const leg = result.route.legs[index - 1]
          const durationSeconds = parseInt(leg.duration?.replace('s', '') || '0', 10)
          const durationMinutes = Math.round(durationSeconds / 60)
          // APIレスポンスはdistance_meters（スネークケース）で返される
          // 念のため両方の形式を確認
          const distanceMeters = (leg as any).distance_meters ?? (leg as any).distanceMeters ?? 0
          console.log(`Place ${index} (${place.name}) (RecommendPage): distance_meters=${(leg as any).distance_meters}, distanceMeters=${(leg as any).distanceMeters}, final=${distanceMeters}`)
          const distanceKm = distanceMeters > 0 ? (distanceMeters / 1000).toFixed(1) : '0.0'
          placeObj.travel_time_from_previous = `${durationMinutes}分 / ${distanceKm}km`
        }

        return placeObj
      })

      // sessionStorageに保存して旅程ページへ遷移
      console.log('=== RecommendPage: Saving to sessionStorage and navigating to /itinerary ===')
      console.log('Itinerary places:', optimizedPlaces.length)
      sessionStorage.setItem('itinerary', JSON.stringify(optimizedPlaces))
      sessionStorage.setItem('route', JSON.stringify(result.route || null))
      console.log('Navigating to /itinerary...')
      router.push('/itinerary')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  // 各場所のマッチ度を100%表示に変換（理論的最大スコアを使って絶対評価）
  const matchScores = useMemo(() => {
    const scores: Record<string, number> = {}
    
    if (maxPossibleScore <= 0) {
      // 理論的最大スコアが無効な場合は全て70%を表示
      recommendedPlaces.forEach((place) => {
        scores[place.place_id] = 70
      })
      return scores
    }
    
    // 理論的最大スコアを100%として、各スコアをパーセンテージに変換（絶対評価）
    recommendedPlaces.forEach((place) => {
      const score = place.relevance_score || 0
      // 理論的最大スコアを100%として計算
      const percentage = Math.round((score / maxPossibleScore) * 100)
      // 0%以上100%以下に制限
      scores[place.place_id] = Math.min(100, Math.max(0, percentage))
    })
    
    return scores
  }, [recommendedPlaces, maxPossibleScore])

  const getNumberBadgeColor = (index: number) => {
    const colors = [
      'bg-yellow-500',  // 1
      'bg-green-500',   // 2
      'bg-orange-500',  // 3
      'bg-blue-500',    // 4
    ]
    return index < colors.length ? colors[index] : 'bg-gray-400'
  }

  return (
    <div className="bg-white">
      {error && (
        <div className="max-w-4xl mx-auto px-6 pt-6">
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 text-sm">{error}</p>
          </div>
        </div>
      )}

      {loading ? (
        <div className="px-6 py-12 text-center">
          <div className="text-gray-400 text-base">読み込み中...</div>
        </div>
      ) : recommendedPlaces.length > 0 ? (
        <div className="py-5">
          <div className="mb-10">
            <h2 className="text-center">
              AIはここを提案します
            </h2>
          </div>

          <div className="space-y-6 mb-8 flex flex-col items-center">
            {recommendedPlaces.map((place, index) => (
              <div
                key={place.place_id || `candidate-${index}`}
                className=""
              >
                <div className="flex mx-auto" style={{ width: "600px" }}>
                  {/* 写真 */}
                  <div className="w-32 h-32 flex-shrink-0 bg-gray-200">
                    {place.photo_url ? (
                      <img
                        src={place.photo_url}
                        alt={place.name || "スポット画像"}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-400 text-xs">
                        写真
                      </div>
                    )}
                  </div>

                  {/* 情報 */}
                  <div className="flex-1 p-4 flex flex-col justify-between">
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-2">
                        {place.name || "名前不明のスポット"}
                      </h3>

                      {/* 場所 */}
                      <div className="flex items-center gap-1 mb-2">
                        <img
                          src="/image/mappin.png"
                          alt="場所"
                          className="object-contain"
                          style={{
                            width: "20px",
                            height: "20px",
                            maxWidth: "20px",
                            maxHeight: "20px",
                            backgroundColor: "transparent",
                            mixBlendMode: "multiply",
                          }}
                        />
                        <span className="text-sm text-gray-600">Fukuoka</span>
                      </div>

                      {/* タグ */}
                      <div className="flex gap-2 mb-2">
                        {place.category && (
                          <span className="px-2 py-1 bg-gray-100 text-xs text-gray-700 rounded">
                            {place.category}
                          </span>
                        )}
                        {place.rating && (
                          <span className="px-2 py-1 bg-gray-100 text-xs text-gray-700 rounded">
                            ⭐ {place.rating}
                          </span>
                        )}
                      </div>
                    </div>

                    {/* マッチ度と選択ボタン */}
                    <div className="flex items-center justify-between mt-2">
                      <span className="text-sm text-gray-600">
                        マッチ度: {matchScores[place.place_id] || 70}%
                      </span>
                      <button
                        onClick={() => handleToggleCandidate(place.place_id)}
                        className={`px-4 py-2 rounded text-sm font-medium transition-colors duration-200 ${
                          selectedPlaceIds.has(place.place_id)
                            ? "bg-blue-600 text-white hover:bg-blue-700"
                            : "bg-gray-200 text-gray-700 hover:bg-gray-300"
                        }`}
                      >
                        {selectedPlaceIds.has(place.place_id)
                          ? "選択済み"
                          : "選択"}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="text-center">
            <button
              onClick={handleAddSelectedPlaces}
              disabled={loading || selectedPlaceIds.size === 0}
              className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200 text-base font-medium"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg
                    className="animate-spin h-5 w-5"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <span>処理中...</span>
                </span>
              ) : (
                <span>
                  選択した{selectedPlaceIds.size}件を追加して旅程を表示
                </span>
              )}
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
