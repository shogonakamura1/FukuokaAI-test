'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useSession } from 'next-auth/react'
import ItineraryTimeline from '@/components/ItineraryTimeline'
import MapView from '@/components/MapView'
import { Place, Route } from '@/components/TripPlanner'
import Header from '@/components/Header'

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
    <div className="min-h-screen bg-white">
      <Header />
      <div className="flex h-[calc(100vh-73px)]">
        <div className="w-1/2 overflow-y-auto bg-white border-r border-gray-300">
          <div className="p-0">
            <div className="w-100 d-flex justify-content-center mb-6 pt-4">
              <div className="d-flex align-items-center gap-2" style={{ width: '66.666667%' }}>
                <img 
                  src="/image/travelbag.jpg" 
                  alt="旅程" 
                  className="object-contain"
                  style={{ 
                    width: '50px', 
                    height: '50px', 
                    maxWidth: '50px', 
                    maxHeight: '50px',
                    backgroundColor: 'transparent',
                    mixBlendMode: 'multiply'
                  }}
                />
                <h2 className="text-xl font-medium text-gray-900 mb-0">
                  旅程
                </h2>
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
