declare module 'next-auth/react' {
  export function useSession(): {
    data: {
      user?: {
        id?: string
        email?: string | null
        name?: string | null
        image?: string | null
      }
    } | null
    status: 'loading' | 'authenticated' | 'unauthenticated'
  }

  export function signIn(provider?: string, options?: any): Promise<void>
  export function signOut(options?: any): Promise<void>
}

