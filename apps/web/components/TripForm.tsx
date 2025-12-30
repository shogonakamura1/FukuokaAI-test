'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const interestTags = [
  'カフェ',
  '屋台',
  '景色',
  '寺社',
  '買い物',
  'グルメ',
  '自然',
]

const schema = z.object({
  must_places: z.array(z.string().min(1)).min(1).max(5),
  interest_tags: z.array(z.string()).min(1),
  free_text: z.string().optional(),
})

type FormData = z.infer<typeof schema>

interface TripFormProps {
  onSubmit: (data: FormData) => void
  loading: boolean
}

export default function TripForm({ onSubmit, loading }: TripFormProps) {
  const [mustPlaces, setMustPlaces] = useState<string[]>([''])
  const [selectedTags, setSelectedTags] = useState<string[]>([])

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const addPlaceField = () => {
    if (mustPlaces.length < 5) {
      setMustPlaces([...mustPlaces, ''])
    }
  }

  const removePlaceField = (index: number) => {
    setMustPlaces(mustPlaces.filter((_, i) => i !== index))
  }

  const updatePlace = (index: number, value: string) => {
    const updated = [...mustPlaces]
    updated[index] = value
    setMustPlaces(updated)
  }

  const toggleTag = (tag: string) => {
    if (selectedTags.includes(tag)) {
      setSelectedTags(selectedTags.filter(t => t !== tag))
    } else {
      setSelectedTags([...selectedTags, tag])
    }
  }

  const onSubmitForm = (data: FormData) => {
    const validPlaces = mustPlaces.filter(p => p.trim() !== '')
    if (validPlaces.length === 0) {
      return
    }
    onSubmit({
      must_places: validPlaces,
      interest_tags: selectedTags,
      free_text: data.free_text,
    })
  }

  return (
    <form onSubmit={handleSubmit(onSubmitForm)} className="space-y-6">
      <div>
        <label className="block text-sm font-medium mb-2">
          行きたい場所（最大5件）
        </label>
        {mustPlaces.map((place, index) => (
          <div key={index} className="flex gap-2 mb-2">
            <input
              type="text"
              value={place}
              onChange={(e) => updatePlace(index, e.target.value)}
              placeholder="例: 太宰府天満宮"
              className="flex-1 px-3 py-2 border rounded"
            />
            {mustPlaces.length > 1 && (
              <button
                type="button"
                onClick={() => removePlaceField(index)}
                className="px-3 py-2 bg-red-500 text-white rounded hover:bg-red-600"
              >
                削除
              </button>
            )}
          </div>
        ))}
        {mustPlaces.length < 5 && (
          <button
            type="button"
            onClick={addPlaceField}
            className="mt-2 px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          >
            場所を追加
          </button>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">
          興味タグ（複数選択可）
        </label>
        <div className="flex flex-wrap gap-2">
          {interestTags.map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => toggleTag(tag)}
              className={`px-4 py-2 rounded ${
                selectedTags.includes(tag)
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-200 hover:bg-gray-300'
              }`}
            >
              {tag}
            </button>
          ))}
        </div>
        {errors.interest_tags && (
          <p className="text-red-500 text-sm mt-1">
            少なくとも1つのタグを選択してください
          </p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">
          追加の希望（任意）
        </label>
        <textarea
          {...register('free_text')}
          placeholder="例: 静かめの古民家カフェが好き"
          className="w-full px-3 py-2 border rounded"
          rows={3}
        />
      </div>

      <button
        type="submit"
        disabled={loading || selectedTags.length === 0}
        className="w-full px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
      >
        {loading ? '生成中...' : '旅程を生成'}
      </button>
    </form>
  )
}


