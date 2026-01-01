'use client'

import { useSession } from 'next-auth/react'

export default function Header() {
  const { data: session } = useSession()

  return (
    <header className="bg-light border-bottom">
      <div className="py-3 px-5">
        <div className="d-flex justify-content-between align-items-center">
          <h1 className="h5 mb-0 text-dark">Fukuoka AI</h1>
          <div className="d-flex align-items-center gap-2">
            <div
              className="rounded-circle bg-secondary d-flex align-items-center justify-content-center"
              style={{ width: "32px", height: "32px" }}
            >
              <span className="small text-white">G</span>
            </div>
            <span className="small text-secondary">Google アカウント</span>
          </div>
        </div>
      </div>
    </header>
  );
}

