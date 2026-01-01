'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useMemo } from 'react'
// @ts-ignore
import { GoogleMap, LoadScript, Marker, Polyline } from '@react-google-maps/api'
import { Place } from './TripPlanner'

interface RouteLeg {
  start_location: { lat: number; lng: number }
  end_location: { lat: number; lng: number }
  distance_meters: number
  duration: string
}

interface Route {
  legs: RouteLeg[]
  distance_meters: number
  duration: string
  optimized_order: number[]
}

interface MapViewProps {
  itinerary: Place[]
  route?: Route | null
}

const mapContainerStyle = {
  width: '100%',
  height: '100%',
}

const defaultCenter = {
  lat: 33.5904,
  lng: 130.4017,
}

export default function MapView({ itinerary, route }: MapViewProps) {
  const mapCenter = useMemo(() => {
    if (itinerary.length === 0) {
      return defaultCenter
    }
    const firstPlace = itinerary[0]
    // latとlngが数値であることを確認
    if (typeof firstPlace.lat === 'number' && typeof firstPlace.lng === 'number' && 
        !isNaN(firstPlace.lat) && !isNaN(firstPlace.lng) &&
        isFinite(firstPlace.lat) && isFinite(firstPlace.lng)) {
      return {
        lat: firstPlace.lat,
        lng: firstPlace.lng,
      }
    }
    return defaultCenter
  }, [itinerary])

  const decodePolyline = (encoded: string): Array<{ lat: number; lng: number }> => {
    const poly: Array<{ lat: number; lng: number }> = []
    let index = 0
    const len = encoded.length
    let lat = 0
    let lng = 0

    while (index < len) {
      let b
      let shift = 0
      let result = 0
      do {
        b = encoded.charCodeAt(index++) - 63
        result |= (b & 0x1f) << shift
        shift += 5
      } while (b >= 0x20)
      const dlat = ((result & 1) !== 0 ? ~(result >> 1) : (result >> 1))
      lat += dlat

      shift = 0
      result = 0
      do {
        b = encoded.charCodeAt(index++) - 63
        result |= (b & 0x1f) << shift
        shift += 5
      } while (b >= 0x20)
      const dlng = ((result & 1) !== 0 ? ~(result >> 1) : (result >> 1))
      lng += dlng

      poly.push({ lat: lat * 1e-5, lng: lng * 1e-5 })
    }
    return poly
  }

  const routePath = useMemo(() => {
    if (!route?.legs || route.legs.length === 0) return []
    // legsから経路を生成（各legのstart_locationとend_locationを結ぶ）
    const path: Array<{ lat: number; lng: number }> = []
    route.legs.forEach((leg, index) => {
      if (index === 0) {
        // 最初のlegはstart_locationから開始
        path.push(leg.start_location)
      }
      // end_locationを追加
      path.push(leg.end_location)
    })
    return path
  }, [route])

  // Next.jsではprocess.env.NEXT_PUBLIC_*はクライアント側でも利用可能
  const googleMapsApiKey = (typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY) || 
                           (typeof process !== 'undefined' && process.env?.GOOGLE_MAPS_API_KEY) || 
                           ''

  if (!googleMapsApiKey) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-100">
        <p>Google Maps APIキーが設定されていません</p>
      </div>
    )
  }

  // マーカーのリストをメモ化してパフォーマンスを改善
  const markers = useMemo(() => {
    // 画像ファイルのパス（Next.jsではpublicディレクトリがルートになる）
    const iconUrl = '/image/mappin.png'
    
    return itinerary
      .filter(place => 
        typeof place.lat === 'number' && typeof place.lng === 'number' &&
        !isNaN(place.lat) && !isNaN(place.lng) &&
        isFinite(place.lat) && isFinite(place.lng)
      )
      .map((place, index) => ({
        id: place.id || place.place_id || `marker-${index}`,
        position: { lat: place.lat, lng: place.lng },
        label: String(index + 1),
        title: place.name,
        iconUrl,
      }))
  }, [itinerary])

  // ポリラインのオプションをメモ化
  const polylineOptions = useMemo(() => ({
    strokeColor: '#3b82f6',
    strokeWeight: 4,
  }), [])

  return (
    <LoadScript 
      googleMapsApiKey={googleMapsApiKey}
      loadingElement={<div className="w-full h-full flex items-center justify-center">読み込み中...</div>}
    >
      <GoogleMap
        mapContainerStyle={mapContainerStyle}
        center={mapCenter}
        zoom={12}
        options={{
          // パフォーマンス改善のためのオプション
          disableDefaultUI: false,
          zoomControl: true,
        }}
      >
        {markers.map((marker) => {
          return (
            <Marker
              key={marker.id}
              position={marker.position}
              label={marker.label}
              title={marker.title}
              icon={marker.iconUrl}
            />
          )
        })}
        {routePath.length > 0 && (
          <Polyline
            path={routePath}
            options={polylineOptions}
          />
        )}
      </GoogleMap>
    </LoadScript>
  )
}

