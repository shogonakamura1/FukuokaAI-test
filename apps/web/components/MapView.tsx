'use client'

import { useMemo } from 'react'
import { GoogleMap, LoadScript, Marker, Polyline } from '@react-google-maps/api'
import { Place } from './TripPlanner'

interface MapViewProps {
  itinerary: Place[]
  route?: {
    polyline: string
  }
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
    return {
      lat: firstPlace.lat,
      lng: firstPlace.lng,
    }
  }, [itinerary])

  const decodePolyline = (encoded: string) => {
    const poly = []
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
    if (!route?.polyline) return []
    return decodePolyline(route.polyline)
  }, [route])

  const googleMapsApiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || process.env.GOOGLE_MAPS_API_KEY || ''

  if (!googleMapsApiKey) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-100">
        <p>Google Maps APIキーが設定されていません</p>
      </div>
    )
  }

  return (
    <LoadScript googleMapsApiKey={googleMapsApiKey}>
      <GoogleMap
        mapContainerStyle={mapContainerStyle}
        center={mapCenter}
        zoom={12}
      >
        {itinerary.map((place, index) => (
          <Marker
            key={place.id || place.place_id}
            position={{ lat: place.lat, lng: place.lng }}
            label={String(index + 1)}
            title={place.name}
          />
        ))}
        {routePath.length > 0 && (
          <Polyline
            path={routePath}
            options={{
              strokeColor: '#3b82f6',
              strokeWeight: 4,
            }}
          />
        )}
      </GoogleMap>
    </LoadScript>
  )
}

