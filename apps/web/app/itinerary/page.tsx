'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useSession } from 'next-auth/react'
import ItineraryTimeline from '@/components/ItineraryTimeline'
import MapView from '@/components/MapView'
import { Place, Route } from '@/components/TripPlanner'

export default function ItineraryPage() {
  const router = useRouter()
  const { data: session } = useSession()
  const [itinerary, setItinerary] = useState<Place[]>([])
  const [route, setRoute] = useState<Route | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // sessionStorageから旅程データを取得
    const storedItinerary = sessionStorage.getItem('itinerary')
    const storedRoute = sessionStorage.getItem('route')

    if (storedItinerary) {
      try {
        setItinerary(JSON.parse(storedItinerary))
      } catch (e) {
        console.error('Failed to parse itinerary:', e)
        setError('旅程データの読み込みに失敗しました')
      }
    } else {
      setError('旅程データが見つかりません')
    }

    if (storedRoute) {
      try {
        setRoute(JSON.parse(storedRoute))
      } catch (e) {
        console.error('Failed to parse route:', e)
      }
    }

    setLoading(false)
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-3 border-gray-300 border-t-blue-600 mx-auto mb-4"></div>
          <div className="text-gray-600 text-base">読み込み中...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="text-center max-w-md">
          <div className="text-red-600 mb-6 text-lg font-medium bg-red-50 p-6 rounded-lg border border-red-200">{error}</div>
          <button
            onClick={() => router.push('/')}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors duration-200 text-base font-medium"
          >
            ホームに戻る
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-50">
        <div className="max-w-full mx-auto px-6 py-4">
          <div className="flex justify-between items-center">
            <h1 className="text-lg font-semibold text-gray-900">
              旅程タイムライン
            </h1>
            <nav className="flex items-center gap-4">
              <div className="text-sm text-gray-600">
                {session?.user?.email}
              </div>
              <button
                onClick={() => router.push('/')}
                className="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors duration-200"
              >
                新しい旅程を作成
              </button>
            </nav>
          </div>
        </div>
      </header>
      <div className="flex h-[calc(100vh-73px)]">
        <div className="w-1/2 overflow-y-auto bg-white border-r border-gray-200">
          <div className="p-6">
            <div className="mb-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-3">
                旅程リスト
              </h2>
              <div className="text-sm text-gray-600">
                選択した場所: <span className="font-medium text-gray-900">{itinerary.length}件</span>
              </div>
            </div>
            <ItineraryTimeline
              itinerary={itinerary}
              onReorder={(orderedPlaceIds) => {
                const reordered = orderedPlaceIds
                  .map(id => itinerary.find(p => (p.id || p.place_id) === id))
                  .filter((p): p is Place => p !== undefined)
                setItinerary(reordered)
                sessionStorage.setItem('itinerary', JSON.stringify(reordered))
              }}
            />
          </div>
        </div>
        <div className="w-1/2 bg-gray-50">
          <MapView itinerary={itinerary} route={route} />
        </div>
      </div>
    </div>
  )
}
