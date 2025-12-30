'use client'

import { Place } from './TripPlanner'

interface CandidateListProps {
  candidates: Place[]
  onAdopt: (candidate: Place) => void
}

export default function CandidateList({ candidates, onAdopt }: CandidateListProps) {
  if (candidates.length === 0) {
    return null
  }

  return (
    <div className="mt-8">
      <h2 className="text-xl font-bold mb-4">おすすめスポット</h2>
      <div className="space-y-4">
        {candidates.map((candidate, index) => (
          <div
            key={candidate.place_id || `candidate-${index}`}
            className="bg-white p-4 rounded shadow"
          >
            {candidate.photo_url && (
              <img
                src={candidate.photo_url}
                alt={candidate.name || 'スポット画像'}
                className="w-full h-48 object-cover rounded mb-3"
              />
            )}
            <h3 className="font-semibold text-lg">{candidate.name || '名前不明のスポット'}</h3>
            {candidate.category && (
              <span className="text-sm text-gray-600">{candidate.category}</span>
            )}
            {candidate.reason && (
              <p className="text-sm text-gray-700 mt-2">{candidate.reason}</p>
            )}
            {candidate.review_summary && (
              <p className="text-xs text-gray-500 mt-1">{candidate.review_summary}</p>
            )}
            {!candidate.name && !candidate.reason && !candidate.review_summary && (
              <p className="text-sm text-gray-500 mt-2">詳細情報がありません</p>
            )}
            <button
              onClick={() => onAdopt(candidate)}
              className="mt-3 px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
            >
              採用する
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}


