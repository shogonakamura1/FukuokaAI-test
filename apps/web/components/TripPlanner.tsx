'use client'

import { useState } from 'react'
import { useSession } from 'next-auth/react'
import TripForm from './TripForm'
import ItineraryTimeline from './ItineraryTimeline'
import MapView from './MapView'
import CandidateList from './CandidateList'

export interface Place {
  id?: string
  place_id: string
  name: string
  lat: number
  lng: number
  kind?: 'must' | 'recommended' | 'start'
  stay_minutes?: number
  order_index?: number
  time_range?: string
  reason?: string
  review_summary?: string
  photo_url?: string
  category?: string
}

export interface Trip {
  trip_id: string
  share_id: string
  itinerary: Place[]
  candidates: Place[]
  route?: {
    polyline: string
  }
}

export default function TripPlanner() {
  const { data: session } = useSession()
  const [trip, setTrip] = useState<Trip | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleGenerateTrip = async (data: {
    must_places: string[]
    interest_tags: string[]
    free_text?: string
  }) => {
    if (!session?.user?.id) {
      setError('ログインが必要です')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/v1/trips`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-Id': session.user.id,
        },
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error?.message || '旅程生成に失敗しました')
      }

      const tripData = await response.json()
      setTrip(tripData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  const handleAdoptCandidate = async (candidate: Place) => {
    if (!trip || !session?.user?.id) return

    setLoading(true)
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      // TODO: 候補を旅程に追加するAPIを実装
      // 暫定的にフロントエンドで追加
      const newPlace: Place = {
        ...candidate,
        id: `temp-${Date.now()}`,
        kind: 'recommended',
        stay_minutes: 60,
        order_index: trip.itinerary.length,
      }
      setTrip({
        ...trip,
        itinerary: [...trip.itinerary, newPlace],
        candidates: trip.candidates.filter(c => c.place_id !== candidate.place_id),
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  const handleRecompute = async (orderedPlaceIds: string[], stayMinutesMap?: Record<string, number>) => {
    if (!trip || !session?.user?.id) return

    setLoading(true)
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/v1/trips/${trip.trip_id}/recompute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-Id': session.user.id,
        },
        body: JSON.stringify({
          ordered_place_ids: orderedPlaceIds,
          stay_minutes_map: stayMinutesMap,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error?.message || '再計算に失敗しました')
      }

      const updatedTrip = await response.json()
      setTrip({
        ...trip,
        itinerary: updatedTrip.itinerary,
        route: updatedTrip.route,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex h-[calc(100vh-80px)]">
      <div className="w-1/2 overflow-y-auto p-4">
        {error && (
          <div className="mb-4 p-4 bg-red-100 text-red-700 rounded">
            {error}
          </div>
        )}
        {!trip ? (
          <TripForm onSubmit={handleGenerateTrip} loading={loading} />
        ) : (
          <>
            <div className="mb-4 flex gap-2 items-center">
              <button
                onClick={() => setTrip(null)}
                className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
              >
                新しい旅程を作成
              </button>
              {trip.share_id && (
                <div className="flex-1">
                  <label className="text-sm text-gray-600">共有URL:</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      readOnly
                      value={typeof window !== 'undefined' ? `${window.location.origin}/share/${trip.share_id}` : ''}
                      className="flex-1 px-2 py-1 border rounded text-sm"
                    />
                    <button
                      onClick={() => {
                        if (typeof window !== 'undefined') {
                          navigator.clipboard.writeText(`${window.location.origin}/share/${trip.share_id}`)
                          alert('共有URLをコピーしました')
                        }
                      }}
                      className="px-3 py-1 bg-blue-500 text-white rounded text-sm hover:bg-blue-600"
                    >
                      コピー
                    </button>
                  </div>
                </div>
              )}
            </div>
            <ItineraryTimeline
              itinerary={trip.itinerary}
              onReorder={handleRecompute}
            />
            {trip.candidates.length > 0 && (
              <CandidateList
                candidates={trip.candidates}
                onAdopt={handleAdoptCandidate}
              />
            )}
          </>
        )}
      </div>
      <div className="w-1/2">
        <MapView
          itinerary={trip?.itinerary || []}
          route={trip?.route}
        />
      </div>
    </div>
  )
}

