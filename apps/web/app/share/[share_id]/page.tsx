'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import MapView from '@/components/MapView'
import { Place } from '@/components/TripPlanner'

export default function SharePage() {
  const params = useParams()
  const shareId = params.share_id as string
  const [trip, setTrip] = useState<any>(null)
  const [itinerary, setItinerary] = useState<Place[]>([])
  const [route, setRoute] = useState<{ polyline: string } | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchShare = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
        const response = await fetch(`${apiUrl}/v1/shares/${shareId}`)

        if (!response.ok) {
          throw new Error('共有された旅程が見つかりません')
        }

        const data = await response.json()
        setTrip(data.trip)
        setItinerary(data.itinerary || [])
        if (data.route) {
          setRoute(data.route)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'エラーが発生しました')
      } finally {
        setLoading(false)
      }
    }

    if (shareId) {
      fetchShare()
    }
  }, [shareId])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div>読み込み中...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-red-500">{error}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      <header className="bg-white shadow-sm p-4">
        <div className="max-w-7xl mx-auto">
          <h1 className="text-2xl font-bold">{trip?.title || '福岡観光 旅程'}</h1>
        </div>
      </header>
      <div className="flex h-[calc(100vh-80px)]">
        <div className="w-1/2 overflow-y-auto p-4">
          <h2 className="text-xl font-bold mb-4">旅程タイムライン</h2>
          {itinerary.length === 0 ? (
            <p>旅程がありません</p>
          ) : (
            <div className="space-y-2">
              {itinerary.map((place, index) => (
                <div key={place.id || place.place_id || `place-${index}`} className="bg-white p-4 rounded shadow">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-semibold">{place.name}</h3>
                      {place.time_range && (
                        <p className="text-sm text-gray-600">{place.time_range}</p>
                      )}
                      {place.reason && (
                        <p className="text-sm text-gray-500 mt-1">{place.reason}</p>
                      )}
                    </div>
                    <span className="text-xs bg-blue-100 px-2 py-1 rounded">
                      {place.kind === 'must' ? '必須' : place.kind === 'start' ? '出発' : 'おすすめ'}
                    </span>
                  </div>
                  {place.photo_url && (
                    <img
                      src={place.photo_url}
                      alt={place.name}
                      className="w-full h-32 object-cover rounded mt-2"
                    />
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
        <div className="w-1/2">
          <MapView itinerary={itinerary} route={route} />
        </div>
      </div>
    </div>
  )
}


