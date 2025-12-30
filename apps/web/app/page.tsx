'use client'

// @ts-ignore - モジュールは存在するが、型定義が見つからない場合がある
import { useSession, signIn, signOut } from 'next-auth/react'
import TripPlanner from '@/components/TripPlanner'

export default function Home() {
  const { data: session, status } = useSession()

  if (status === 'loading') {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div>読み込み中...</div>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen">
        <h1 className="text-3xl font-bold mb-4">福岡観光 旅程作成AI</h1>
        <p className="mb-8">Googleアカウントでログインしてください</p>
        <button
          onClick={() => signIn('google')}
          className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
        >
          Googleでログイン
        </button>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      <header className="bg-white shadow-sm p-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold">福岡観光 旅程作成AI</h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{session.user?.email}</span>
            <button
              onClick={() => signOut()}
              className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
            >
              ログアウト
            </button>
          </div>
        </div>
      </header>
      <main>
        <TripPlanner />
      </main>
    </div>
  )
}


