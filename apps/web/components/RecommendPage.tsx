'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import TripForm from './TripForm'
import { Place } from './TripPlanner'

export default function RecommendPage() {
  const router = useRouter()
  const [recommendedPlaces, setRecommendedPlaces] = useState<Place[]>([])
  const [selectedPlaceIds, setSelectedPlaceIds] = useState<Set<string>>(new Set())
  const [mustPlaces, setMustPlaces] = useState<Place[]>([])
  const [startPlace, setStartPlace] = useState<string>('')
  const [goalPlace, setGoalPlace] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

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
        }))

      setRecommendedPlaces(places)
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

        // 移動時間を設定
        if (result.route?.legs && index > 0 && result.route.legs[index - 1]) {
          const leg = result.route.legs[index - 1]
          const durationSeconds = parseInt(leg.duration?.replace('s', '') || '0', 10)
          const durationMinutes = Math.round(durationSeconds / 60)
          placeObj.travel_time_from_previous = `${durationMinutes}分`
        }

        return placeObj
      })

      // sessionStorageに保存して旅程ページへ遷移
      sessionStorage.setItem('itinerary', JSON.stringify(optimizedPlaces))
      sessionStorage.setItem('route', JSON.stringify(result.route || null))
      router.push('/itinerary')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto px-6 py-12">
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 text-sm">{error}</p>
          </div>
        )}

        {recommendedPlaces.length === 0 ? (
          <div className="max-w-2xl mx-auto">
            <div className="bg-white rounded-lg border border-gray-200 p-8">
              <div className="mb-8">
                <h1 className="text-3xl font-semibold text-gray-900 mb-3">
                  旅程を作成
                </h1>
                <p className="text-gray-600 leading-relaxed">
                  あなたの好みや興味に合わせて、最適な旅程をAIが提案します。行きたい場所や興味のあるタグを選択して、素敵な旅の計画を立てましょう。
                </p>
              </div>
              <TripForm onSubmit={handleGenerateTrip} loading={loading} />
            </div>
          </div>
        ) : (
          <div className="animate-fade-in">
            <div className="mb-8">
              <h2 className="text-2xl font-semibold text-gray-900 mb-2">
                おすすめスポット
              </h2>
              <p className="text-gray-600">気になる場所を選択して、旅程に追加してください</p>
            </div>
            
            <div className="grid md:grid-cols-2 gap-6 mb-8">
              {recommendedPlaces.map((place, index) => (
                <div
                  key={place.place_id || `candidate-${index}`}
                  className="bg-white rounded-lg border border-gray-200 overflow-hidden hover:border-gray-300 transition-colors duration-200"
                >
                  {place.photo_url && (
                    <div className="aspect-video overflow-hidden bg-gray-100">
                      <img
                        src={place.photo_url}
                        alt={place.name || 'スポット画像'}
                        className="w-full h-full object-cover"
                      />
                    </div>
                  )}
                  <div className="p-6">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">
                          {place.name || '名前不明のスポット'}
                        </h3>
                        {place.category && (
                          <span className="inline-block text-xs font-medium text-gray-600 bg-gray-100 px-2 py-1 rounded">
                            {place.category}
                          </span>
                        )}
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer ml-4 flex-shrink-0">
                        <input
                          type="checkbox"
                          checked={selectedPlaceIds.has(place.place_id)}
                          onChange={() => handleToggleCandidate(place.place_id)}
                          className="sr-only peer"
                        />
                        <div className="w-5 h-5 border-2 border-gray-300 rounded peer-checked:bg-blue-600 peer-checked:border-blue-600 transition-colors duration-200 flex items-center justify-center">
                          {selectedPlaceIds.has(place.place_id) && (
                            <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                            </svg>
                          )}
                        </div>
                      </label>
                    </div>
                    
                    {place.review_summary && (
                      <p className="text-sm text-gray-600 leading-relaxed mb-3 line-clamp-2">
                        {place.review_summary}
                      </p>
                    )}
                    
                    {place.rating && (
                      <div className="flex items-center gap-2 pt-3 border-t border-gray-100">
                        <div className="flex items-center gap-0.5">
                          {Array.from({ length: 5 }).map((_, i) => (
                            <svg
                              key={i}
                              className={`w-4 h-4 ${i < Math.floor(place.rating) ? 'text-yellow-400' : 'text-gray-300'}`}
                              fill="currentColor"
                              viewBox="0 0 20 20"
                            >
                              <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                            </svg>
                          ))}
                        </div>
                        <span className="text-sm font-medium text-gray-700">{place.rating}</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
            
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <button
                onClick={handleAddSelectedPlaces}
                disabled={loading || selectedPlaceIds.size === 0}
                className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200 text-base font-medium"
              >
                {loading ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>処理中...</span>
                  </span>
                ) : (
                  <span>選択した{selectedPlaceIds.size}件を追加して旅程を表示</span>
                )}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
