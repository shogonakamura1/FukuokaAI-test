declare namespace NodeJS {
  interface ProcessEnv {
    NEXT_PUBLIC_GOOGLE_MAPS_API_KEY?: string
    GOOGLE_MAPS_API_KEY?: string
    [key: string]: string | undefined
  }
}

// Next.jsではprocess.env.NEXT_PUBLIC_*はクライアント側でも利用可能
declare var process: {
  env: NodeJS.ProcessEnv
} | undefined

