'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useState, useEffect } from 'react'
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
  // マーカーアイコンの設定を保持する状態
  const [iconConfig, setIconConfig] = useState<any>(null)

  // 地図の中心位置を決定（旅程の最初の場所、またはデフォルト位置）
  let mapCenter = defaultCenter
  if (itinerary.length > 0) {
    const firstPlace = itinerary[0]
    if (typeof firstPlace.lat === 'number' && typeof firstPlace.lng === 'number' && 
        !isNaN(firstPlace.lat) && !isNaN(firstPlace.lng) &&
        isFinite(firstPlace.lat) && isFinite(firstPlace.lng)) {
      mapCenter = {
        lat: firstPlace.lat,
        lng: firstPlace.lng,
      }
    }
  }

  // ルートの経路を生成（各legのstart_locationとend_locationを結ぶ）
  const routePath: Array<{ lat: number; lng: number }> = []
  if (route?.legs && route.legs.length > 0) {
    route.legs.forEach((leg, index) => {
      if (index === 0) {
        routePath.push(leg.start_location)
      }
      routePath.push(leg.end_location)
    })
  }

  // Google Maps APIキーを取得
  const googleMapsApiKey = (typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY) || 
                           (typeof process !== 'undefined' && process.env?.GOOGLE_MAPS_API_KEY) || 
                           ''

  // デバッグ: APIキーの状態を確認（開発環境のみ）
  useEffect(() => {
    if (process.env.NODE_ENV === 'development') {
      console.log('Google Maps API Key exists:', !!googleMapsApiKey)
      console.log('API Key length:', googleMapsApiKey ? googleMapsApiKey.length : 0)
      console.log('API Key starts with:', googleMapsApiKey ? googleMapsApiKey.substring(0, 10) + '...' : 'N/A')
    }
  }, [googleMapsApiKey])

  // APIキーが設定されていない場合はエラーメッセージを表示
  if (!googleMapsApiKey) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-100">
        <p>Google Maps APIキーが設定されていません</p>
        <p className="text-sm text-gray-500 mt-2">
          環境変数 NEXT_PUBLIC_GOOGLE_MAPS_API_KEY を設定してください
        </p>
      </div>
    )
  }

  // マーカーのリストを生成（有効な緯度経度を持つ場所のみ）
  const markers = itinerary
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
    }))

  // ポリライン（ルート線）のスタイル設定
  const polylineOptions = {
    strokeColor: '#3b82f6',
    strokeWeight: 4,
  }

  // Google Maps APIが読み込まれた時に呼ばれるコールバック
  const handleLoad = () => {
    console.log('Google Maps API loaded successfully')
    // APIが完全に読み込まれるまで少し待つ
    setTimeout(() => {
      if (typeof window !== 'undefined' && (window as any).google?.maps) {
        const google = (window as any).google
        const mappinUrl = `${window.location.origin}/image/mappin.png`
        
        try {
          // SizeとPointが利用可能な場合、アイコンのサイズを20px x 20pxに設定
          if (google.maps.Size && typeof google.maps.Size === 'function' &&
              google.maps.Point && typeof google.maps.Point === 'function') {
            setIconConfig({
              url: mappinUrl,
              scaledSize: new google.maps.Size(20, 20),
              anchor: new google.maps.Point(10, 20),
            })
            console.log('Map icon configured with Size and Point')
          } else {
            // Size/Pointが利用できない場合は、URLだけを設定
            setIconConfig({ url: mappinUrl })
            console.log('Map icon configured with URL only')
          }
        } catch (e) {
          // エラーが発生した場合も、URLだけを設定
          console.error('Error setting icon config:', e)
          setIconConfig({ url: mappinUrl })
        }
      } else {
        console.error('Google Maps API not available after load')
      }
    }, 100)
  }
  
  // エラーハンドラー
  const handleError = (error: Error) => {
    console.error('Google Maps API loading error:', error)
    console.error('Error message:', error.message)
    console.error('Error stack:', error.stack)
  }

  // コンポーネントがマウントされた時に、初期アイコンを設定（URLだけ）
  useEffect(() => {
    if (!iconConfig) {
      const mappinUrl = typeof window !== 'undefined' 
        ? `${window.location.origin}/image/mappin.png`
        : '/image/mappin.png'
      setIconConfig({ url: mappinUrl })
    }
  }, [])

  return (
    <LoadScript 
      googleMapsApiKey={googleMapsApiKey}
      loadingElement={<div className="w-full h-full flex items-center justify-center">読み込み中...</div>}
      onLoad={handleLoad}
      onError={handleError}
    >
      <GoogleMap
        mapContainerStyle={mapContainerStyle}
        center={mapCenter}
        zoom={12}
        options={{
          disableDefaultUI: false,
          zoomControl: true,
        }}
      >
        {markers.map((marker) => (
          <Marker
            key={marker.id}
            position={marker.position}
            label={marker.label}
            title={marker.title}
            icon={iconConfig || { url: '/image/mappin.png' }}
          />
        ))}
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

