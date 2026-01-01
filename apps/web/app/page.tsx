'use client'

import { useSession, signIn, signOut } from 'next-auth/react'
import RecommendPage from '@/components/RecommendPage'

export default function Home() {
  const { data: session, status } = useSession()

  if (status === 'loading') {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-400 text-base">読み込み中...</div>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center px-4">
        <div className="max-w-md w-full">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-semibold text-gray-900 mb-3 tracking-tight">
              福岡観光
            </h1>
            <h2 className="text-4xl font-semibold text-gray-900 mb-6 tracking-tight">
              旅程作成AI
            </h2>
            <p className="text-gray-600 text-base leading-relaxed">
              Googleアカウントでログインして、あなただけの旅程を作成しましょう
            </p>
          </div>
          <button
            onClick={() => signIn('google')}
            className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors duration-200 text-base font-medium"
          >
            Googleでログイン
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex justify-between items-center">
            <h1 className="text-lg font-semibold text-gray-900">
              福岡観光 旅程作成AI
            </h1>
            <nav className="flex items-center gap-4">
              <div className="text-sm text-gray-600">
                {session.user?.email}
              </div>
              <button
                onClick={() => signOut()}
                className="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors duration-200"
              >
                ログアウト
              </button>
            </nav>
          </div>
        </div>
      </header>
      <main>
        <RecommendPage />
      </main>
    </div>
  )
}


