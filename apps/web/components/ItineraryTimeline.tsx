'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useState, useEffect, useMemo } from 'react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Card } from 'react-bootstrap'
import { Place } from './TripPlanner'

interface ItineraryTimelineProps {
  itinerary: Place[]
  onReorder: (orderedPlaceIds: string[], stayMinutesMap?: Record<string, number>) => void
}

interface SortableItemProps {
  place: Place
  index: number
  startTime: string
  endTime: string
  travelTime?: string
  travelDistance?: string
  isLast: boolean
}

// 滞在時間を計算（デフォルトは90分）
const getStayMinutes = (place: Place): number => {
  if (place.stay_minutes) {
    return place.stay_minutes
  }
  // デフォルトの滞在時間（90分）
  return 90
}

// 分を時間文字列に変換
const minutesToTime = (minutes: number): string => {
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return `${hours.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}`
}

// 時間範囲を計算
const calculateTimeRanges = (places: Place[]): Array<{ start: string; end: string; travelTime?: string; travelDistance?: string }> => {
  if (places.length === 0) return []
  
  const startHour = 10 // デフォルト開始時刻 10:00
  const ranges: Array<{ start: string; end: string; travelTime?: string; travelDistance?: string }> = []
  
  let currentMinutes = startHour * 60
  
  places.forEach((place, index) => {
    // 移動時間と距離（前の場所から）
    let travelMinutes = 0
    let travelDistance = ''
    if (index > 0 && place.travel_time_from_previous) {
      // "50分 / 18km" または "50分" の形式をパース
      const timeMatch = place.travel_time_from_previous.match(/(\d+)分/)
      const distanceMatch = place.travel_time_from_previous.match(/([\d.]+)km/)
      if (timeMatch) {
        travelMinutes = parseInt(timeMatch[1], 10)
      }
      if (distanceMatch) {
        travelDistance = `${distanceMatch[1]}km`
      }
    }
    
    const startTime = minutesToTime(currentMinutes + travelMinutes)
    currentMinutes += travelMinutes
    
    const stayMinutes = getStayMinutes(place)
    const endTime = minutesToTime(currentMinutes + stayMinutes)
    currentMinutes += stayMinutes
    
    ranges.push({
      start: startTime,
      end: endTime,
      travelTime: index > 0 && travelMinutes > 0 ? `${travelMinutes}分` : undefined,
      travelDistance: travelDistance || undefined,
    })
  })
  
  return ranges
}

function SortableItem({ place, index, startTime, endTime, travelTime, travelDistance, isLast }: SortableItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: place.id || place.place_id || `place-${index}` })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const stayMinutes = getStayMinutes(place)
  const stayHours = Math.floor(stayMinutes / 60)
  const stayMins = stayMinutes % 60
  const stayDuration = stayHours > 0 
    ? `${stayHours}時間${stayMins > 0 ? stayMins + '分' : '0分'}`
    : `${stayMins}分`

  return (
    <>
      {/* 移動時間と距離（前のカードから） */}
      {travelTime && (
        <div
          className="w-100 d-flex justify-content-center mb-3"
          style={{ width: "100%" }}
        >
          <div
            className="d-flex align-items-center"
            style={{ width: "66.666667%", maxWidth: "66.666667%" }}
          >
            {/* 縦線 */}
            <div
              className="bg-gray-300 h-12 me-4 flex-shrink-0"
              style={{ width: "1px" }}
            ></div>
            {/* 移動情報 */}
            <div className="d-flex align-items-center gap-2">
              <div className="d-flex" style={{ height: '80px' }}>
                <div className="vr"></div>
              </div>
              <img
                src="/image/caricon.png"
                alt="移動時間"
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
              <span className="text-sm text-secondary">
                移動時間{travelTime}
                {travelDistance && ` ${travelDistance}`}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* 場所カード */}
      <div
        ref={setNodeRef}
        style={{ ...style, width: "100%" }}
        {...attributes}
        {...listeners}
        className="w-100 d-flex justify-content-center"
      >
        <Card
          className="mb-3 cursor-move border"
          style={{ width: "66.666667%", maxWidth: "66.666667%" }}
        >
          <Card.Body>
            {/* 場所名 */}
            <Card.Title className="mb-2">
              {place.name || "名前不明のスポット"}
            </Card.Title>

            {/* 場所情報 */}
            <Card.Text className="d-flex align-items-center gap-2 mb-0">
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
              <span className="text-muted">
                {place.address || "Fukuoka city"}
              </span>
            </Card.Text>
          </Card.Body>
        </Card>
      </div>
    </>
  );
}

export default function ItineraryTimeline({ itinerary, onReorder }: ItineraryTimelineProps) {
  const [items, setItems] = useState(itinerary)

  // itineraryが変更されたらitemsを更新
  useEffect(() => {
    setItems(itinerary)
  }, [itinerary])

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  // 時間範囲を計算
  const timeRanges = useMemo(() => calculateTimeRanges(items), [items])

  const handleDragEnd = (event: { active: { id: string | number }; over: { id: string | number } | null }) => {
    const { active, over } = event

    if (over && active.id !== over.id) {
      setItems((items: Place[]) => {
        const oldIndex = items.findIndex((item: Place) => {
          const itemId = item.id || item.place_id
          return String(itemId) === String(active.id)
        })
        const newIndex = items.findIndex((item: Place) => {
          const itemId = item.id || item.place_id
          return String(itemId) === String(over.id)
        })
        const newItems = arrayMove(items, oldIndex, newIndex)
        
        // 再計算を実行
        const orderedIds = newItems
          .map((item: Place) => item.id || item.place_id)
          .filter((id: string | undefined): id is string => typeof id === 'string' && id.length > 0)
        onReorder(orderedIds)
        
        return newItems
      })
    }
  }

  if (itinerary.length === 0) {
    return null
  }

  return (
    <div className="w-100 d-flex flex-column align-items-center pb-4">
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={items.map((item: Place, index: number) => item.id || item.place_id || `item-${index}`)}
          strategy={verticalListSortingStrategy}
        >
          {items.map((place: Place, index: number) => {
            const itemKey = place.id || place.place_id || `place-${index}`
            const timeRange = timeRanges[index] || { start: '10:00', end: '11:30' }
            const isLast = index === items.length - 1
            
            // Reactのkeyプロパティは特別なプロパティなので、型定義に含まれない
            // @ts-ignore - Reactのkeyプロパティは型定義に含まれない
            return (
              <SortableItem
                key={itemKey}
                place={place}
                index={index}
                startTime={timeRange.start}
                endTime={timeRange.end}
                travelTime={timeRange.travelTime}
                travelDistance={timeRange.travelDistance}
                isLast={isLast}
              />
            )
          })}
        </SortableContext>
      </DndContext>
    </div>
  )
}
