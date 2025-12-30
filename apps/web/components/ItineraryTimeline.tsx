'use client'

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
} from '@dnd-kit/sortable'
import {
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Place } from './TripPlanner'

interface ItineraryTimelineProps {
  itinerary: Place[]
  onReorder: (orderedPlaceIds: string[], stayMinutesMap?: Record<string, number>) => void
}

function SortableItem({ place, index }: { place: Place; index: number }) {
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

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="bg-white p-4 mb-2 rounded shadow cursor-move"
      {...attributes}
      {...listeners}
    >
      <div className="flex justify-between items-start">
        <div className="flex-1">
          <h3 className="font-semibold">{place.name || '名前不明のスポット'}</h3>
          {place.time_range && (
            <p className="text-sm text-gray-600">{place.time_range}</p>
          )}
          {place.reason && (
            <p className="text-sm text-gray-500 mt-1">{place.reason}</p>
          )}
          {!place.name && !place.time_range && !place.reason && (
            <p className="text-sm text-gray-500 mt-1">詳細情報がありません</p>
          )}
        </div>
        <span className="text-xs bg-blue-100 px-2 py-1 rounded ml-2">
          {place.kind === 'must' ? '必須' : place.kind === 'start' ? '出発' : 'おすすめ'}
        </span>
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

  const handleDragEnd = (event: any) => {
    const { active, over } = event

    if (active.id !== over.id) {
      setItems((items) => {
        const oldIndex = items.findIndex(item => (item.id || item.place_id) === active.id)
        const newIndex = items.findIndex(item => (item.id || item.place_id) === over.id)
        const newItems = arrayMove(items, oldIndex, newIndex)
        
        // 再計算を実行
        const orderedIds = newItems.map(item => item.id || item.place_id)
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
      <h2 className="text-xl font-bold mb-4">旅程タイムライン</h2>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={items.map((item, index) => item.id || item.place_id || `item-${index}`)}
          strategy={verticalListSortingStrategy}
        >
          {items.map((place, index) => (
            <SortableItem key={place.id || place.place_id || `place-${index}`} place={place} index={index} />
          ))}
        </SortableContext>
      </DndContext>
    </div>
  )
}


