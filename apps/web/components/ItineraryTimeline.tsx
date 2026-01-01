'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useState, useEffect } from 'react'
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
import { Place } from './TripPlanner'

interface ItineraryTimelineProps {
  itinerary: Place[]
  onReorder: (orderedPlaceIds: string[], stayMinutesMap?: Record<string, number>) => void
}

interface SortableItemProps {
  place: Place
  index: number
}

function SortableItem({ place, index }: SortableItemProps) {
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

  const getBadgeStyle = (kind?: string) => {
    switch (kind) {
      case 'must':
        return 'bg-amber-100 text-amber-800 border-amber-200'
      case 'start':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'goal':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getBadgeText = (kind?: string) => {
    switch (kind) {
      case 'must':
        return '必須'
      case 'start':
        return '出発'
      case 'goal':
        return '到着'
      default:
        return 'おすすめ'
    }
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="bg-white rounded-lg border border-gray-200 p-5 mb-3 hover:border-gray-300 transition-all duration-200 cursor-move"
      {...attributes}
      {...listeners}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-4 mb-3">
            <h3 className="text-lg font-semibold text-gray-900">
              {place.name || '名前不明のスポット'}
            </h3>
            <span className={`text-xs px-2.5 py-1 rounded border ${getBadgeStyle(place.kind)} font-medium flex-shrink-0`}>
              {getBadgeText(place.kind)}
            </span>
          </div>
          
          {place.travel_time_from_previous && (
            <div className="flex items-center gap-2 mb-3 text-sm text-gray-600">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>移動時間: <span className="font-medium text-gray-900">{place.travel_time_from_previous}</span></span>
            </div>
          )}
          
          {place.time_range && (
            <p className="text-sm text-gray-600 mb-2">{place.time_range}</p>
          )}
          
          {place.reason && (
            <p className="text-sm text-gray-600 mt-3 leading-relaxed">{place.reason}</p>
          )}
          
          {place.address && (
            <p className="text-sm text-gray-500 mt-2">{place.address}</p>
          )}
          
          {!place.name && !place.time_range && !place.reason && (
            <p className="text-sm text-gray-500 mt-2">詳細情報がありません</p>
          )}
        </div>
      </div>
    </div>
  )
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
    <div>
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
            // Reactのkeyプロパティは特別なプロパティなので、型定義に含まれない
            // @ts-ignore - Reactのkeyプロパティは型定義に含まれない
            return <SortableItem place={place} index={index} key={itemKey} />
          })}
        </SortableContext>
      </DndContext>
    </div>
  )
}
