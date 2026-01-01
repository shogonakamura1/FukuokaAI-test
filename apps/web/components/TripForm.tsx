'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useState, type ChangeEvent } from 'react'
// @ts-ignore
import { useForm } from 'react-hook-form'
// @ts-ignore
import { zodResolver } from '@hookform/resolvers/zod'
// @ts-ignore
import { z } from 'zod'
// @ts-ignore
import { Card } from 'react-bootstrap'

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
    <Card className="p-4">
      <Card.Body className="d-flex flex-column align-items-center">
        <form onSubmit={handleSubmit(onSubmitForm)} className="space-y-4" style={{ width: '100%', maxWidth: '500px' }}>
        {/* 必ず行きたい場所 */}
        <div className="w-100 my-4">
          <div className="d-flex align-items-center mb-2">
            <label className="text-base font-medium text-gray-900 me-3" style={{ minWidth: '180px' }}>
              必ず行きたい場所は?
            </label>
          </div>
          <div className="space-y-3 w-100">
            {mustPlaces.map((place: string, index: number) => (
              <div key={index} className="flex gap-3">
                <input
                  type="text"
                  value={place}
                  onChange={(e: ChangeEvent<HTMLInputElement>) =>
                    updatePlace(index, e.target.value)
                  }
                  placeholder="例: 太宰府天満宮"
                  className="flex-1 px-4 py-2.5 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base"
                />
                {index === mustPlaces.length - 1 && mustPlaces.length < 5 && (
                  <button
                    type="button"
                    onClick={addPlaceField}
                    className="px-4 py-2.5 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition-colors duration-200 text-sm font-medium"
                  >
                    追加
                  </button>
                )}
                {mustPlaces.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removePlaceField(index)}
                    className="px-4 py-2.5 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded transition-colors duration-200 text-sm font-medium"
                  >
                    削除
                  </button>
                )}
              </div>
            ))}
          </div>
          {errors.must_places && (
            <p className="text-sm mt-2 text-red-600 w-100" style={{ marginLeft: '183px' }}>
              {errors.must_places.message ||
                "行きたい場所を1つ以上入力してください"}
            </p>
          )}
        </div>

        {/* 出発地点 */}
        <div className="w-100 my-4">
          <div className="d-flex align-items-center">
            <label className="text-base font-medium text-gray-900 me-3" style={{ minWidth: '180px' }}>
              出発地点は?
            </label>
            <input
              {...register("start_place")}
              type="text"
              placeholder="例: 博多駅"
              className="flex-grow-1 px-4 py-2.5 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base"
            />
          </div>
          {errors.start_place && (
            <p className="text-sm mt-2 text-red-600" style={{ marginLeft: '183px' }}>
              {errors.start_place.message}
            </p>
          )}
        </div>

        {/* 到着地点 */}
        <div className="w-100 my-4">
          <div className="d-flex align-items-center">
            <label className="text-base font-medium text-gray-900 me-3" style={{ minWidth: '180px' }}>
              到着地点は?
            </label>
            <input
              {...register("goal_place")}
              type="text"
              placeholder="例: 天神"
              className="flex-grow-1 px-4 py-2.5 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white text-base"
            />
          </div>
          {errors.goal_place && (
            <p className="text-sm mt-2 text-red-600" style={{ marginLeft: '183px' }}>
              {errors.goal_place.message}
            </p>
          )}
        </div>

        {/* 興味あるタグ */}
        <div className="w-100 my-4">
          <div className="d-flex align-items-center mb-3">
            <label className="text-base font-medium text-gray-900 me-3" style={{ minWidth: '180px' }}>
              あなたの興味あるタグは?
            </label>
          </div>
          <div className="d-flex flex-wrap gap-2 justify-content-center w-100">
            {interestTags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => toggleTag(tag)}
                className={`px-4 py-2 rounded border transition-colors duration-200 text-base ${
                  selectedTags.includes(tag)
                    ? 'bg-blue-600 text-white border-blue-600'
                    : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-100'
                }`}
              >
                {tag}
              </button>
            ))}
          </div>
          {errors.interest_tags && (
            <p className="text-sm mt-2 text-red-600 text-center w-100">
              少なくとも1つのタグを選択してください
            </p>
          )}
        </div>

        {/* 作成開始ボタン */}
        <div className="d-flex justify-content-center pt-4 w-100 my-4">
          <button
            type="submit"
            disabled={loading || selectedTags.length === 0}
            className="px-12 py-3 text-white rounded hover:opacity-90 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200 text-base font-medium"
            style={{ 
              backgroundColor: loading || selectedTags.length === 0 ? '#ffb3ba' : '#ff6b7a',
              color: '#8b0000',
              minWidth: '200px'
            }}
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <svg
                  className="animate-spin h-5 w-5"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  ></circle>
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                <span>生成中...</span>
              </span>
            ) : (
              "作成開始"
            )}
          </button>
        </div>
        </form>
      </Card.Body>
    </Card>
  );
}
