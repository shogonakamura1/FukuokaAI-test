'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useState, type ChangeEvent } from 'react'
// @ts-ignore
import { useForm } from 'react-hook-form'
// @ts-ignore
import { zodResolver } from '@hookform/resolvers/zod'
// @ts-ignore
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
  must_places: z.array(z.string())
    .refine(
      (arr: string[]) => {
        const validPlaces = arr.filter((p: string) => p.trim() !== '')
        return validPlaces.length >= 1 && validPlaces.length <= 5
      },
      { message: '行きたい場所を1つ以上5つ以内で入力してください' }
    ),
  interest_tags: z.array(z.string()).min(1, '少なくとも1つのタグを選択してください'),
  start_place: z.string().min(1, '出発地点を入力してください'),
  goal_place: z.string().min(1, 'ゴール地点を入力してください'),
  free_text: z.string().optional(),
})

type FormData = z.infer<typeof schema>

interface TripFormProps {
  onSubmit: (data: FormData) => void
  loading: boolean
}

export default function TripForm({ onSubmit, loading }: TripFormProps) {
  const [mustPlaces, setMustPlaces] = useState<string[]>([''])

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      must_places: [''],
      interest_tags: [],
      start_place: '',
      goal_place: '',
    },
  })

  const selectedTags = watch('interest_tags') || []

  const addPlaceField = () => {
    if (mustPlaces.length < 5) {
      const updated = [...mustPlaces, '']
      setMustPlaces(updated)
      setValue('must_places', updated, { shouldValidate: true })
    }
  }

  const removePlaceField = (index: number) => {
    const updated = mustPlaces.filter((_: string, i: number) => i !== index)
    setMustPlaces(updated)
    setValue('must_places', updated, { shouldValidate: true })
  }

  const updatePlace = (index: number, value: string) => {
    const updated = [...mustPlaces]
    updated[index] = value
    setMustPlaces(updated)
    setValue('must_places', updated, { shouldValidate: true })
  }

  const toggleTag = (tag: string) => {
    const currentTags = selectedTags || []
    if (currentTags.includes(tag)) {
      const updatedTags = currentTags.filter((t: string) => t !== tag)
      setValue('interest_tags', updatedTags, { shouldValidate: true })
    } else {
      const updatedTags = [...currentTags, tag]
      setValue('interest_tags', updatedTags, { shouldValidate: true })
    }
  }

  const onSubmitForm = (data: FormData) => {
    const validPlaces = data.must_places.filter((p: string) => p.trim() !== '')
    if (validPlaces.length === 0) {
      return
    }
    onSubmit({
      must_places: validPlaces,
      interest_tags: data.interest_tags,
      start_place: data.start_place,
      goal_place: data.goal_place,
      free_text: data.free_text,
    })
  }

  return (
    <form onSubmit={handleSubmit(onSubmitForm)} className="space-y-8">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-3">
          行きたい場所（最大5件）
        </label>
        <div className="space-y-3">
          {mustPlaces.map((place: string, index: number) => (
            <div key={index} className="flex gap-3">
              <input
                type="text"
                value={place}
                onChange={(e: ChangeEvent<HTMLInputElement>) => updatePlace(index, e.target.value)}
                placeholder="例: 太宰府天満宮"
                className="flex-1 px-4 py-2.5 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base transition-all duration-200"
              />
              {mustPlaces.length > 1 && (
                <button
                  type="button"
                  onClick={() => removePlaceField(index)}
                  className="px-4 py-2.5 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors duration-200 text-sm font-medium"
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
              className="w-full px-4 py-2.5 text-sm text-gray-600 border border-dashed border-gray-300 rounded-lg hover:border-gray-400 hover:bg-gray-50 transition-colors duration-200 font-medium"
            >
              + 場所を追加
            </button>
          )}
        </div>
        {errors.must_places && (
          <p className="text-sm mt-2 text-red-600">
            {errors.must_places.message || '行きたい場所を1つ以上入力してください'}
          </p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-3">
          興味タグ（複数選択可）
        </label>
        <div className="grid grid-cols-2 gap-3">
          {interestTags.map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => toggleTag(tag)}
              className={`px-4 py-2.5 rounded-lg text-sm font-medium transition-colors duration-200 ${
                selectedTags.includes(tag)
                  ? 'bg-blue-600 text-white hover:bg-blue-700'
                  : 'bg-white text-gray-700 border border-gray-300 hover:border-gray-400 hover:bg-gray-50'
              }`}
            >
              {tag}
            </button>
          ))}
        </div>
        {errors.interest_tags && (
          <p className="text-sm mt-2 text-red-600">
            少なくとも1つのタグを選択してください
          </p>
        )}
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            出発地点 <span className="text-red-600">*</span>
          </label>
          <input
            {...register('start_place')}
            type="text"
            placeholder="例: 博多駅"
            className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base transition-all duration-200"
          />
          {errors.start_place && (
            <p className="text-sm mt-2 text-red-600">
              {errors.start_place.message}
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            ゴール地点 <span className="text-red-600">*</span>
          </label>
          <input
            {...register('goal_place')}
            type="text"
            placeholder="例: 天神"
            className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base transition-all duration-200"
          />
          {errors.goal_place && (
            <p className="text-sm mt-2 text-red-600">
              {errors.goal_place.message}
            </p>
          )}
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-3">
          追加の希望（任意）
        </label>
        <textarea
          {...register('free_text')}
          placeholder="例: 静かめの古民家カフェが好き"
          className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white resize-none text-base transition-all duration-200"
          rows={3}
        />
      </div>

      <button
        type="submit"
        disabled={loading || selectedTags.length === 0}
        className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200 text-base font-medium"
      >
        {loading ? (
          <span className="flex items-center justify-center gap-2">
            <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span>生成中...</span>
          </span>
        ) : (
          '旅程を生成'
        )}
      </button>
    </form>
  )
}
