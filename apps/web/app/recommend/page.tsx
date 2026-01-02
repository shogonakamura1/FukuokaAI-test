'use client'

import { useSession, signIn } from 'next-auth/react'
import RecommendPage from '@/components/RecommendPage'
import Header from '@/components/Header'

export default function RecommendRoute() {
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
      <div className="min-h-screen bg-white" style={{width: '75%'}}>
        <Header />
        <div className="flex items-center justify-center min-h-[calc(100vh-73px)] px-4">
          <div className="w-full max-w-md">
            <div className="bg-gray-200 rounded-lg p-12 flex flex-col items-center">
              <button
                onClick={() => signIn('google')}
                className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-8 py-4 rounded-lg transition-colors duration-200 text-base font-medium"
              >
                <div className="text-center">
                  <div className="text-sm mb-1">Google アカウントで</div>
                  <div className="text-base font-semibold">ログイン</div>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <Header />
      <main className="w-full flex items-center min-h-[calc(100vh-73px)]">
        <div className="w-full flex justify-center">
          <RecommendPage />
        </div>
      </main>
    </div>
  )
}

