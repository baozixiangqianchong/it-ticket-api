import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import { api, clearToken, getToken, setToken } from '../api/client'
import type { PublicUser } from '../api/types'

type AuthContextValue = {
  user: PublicUser | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, displayName: string, inviteCode: string) => Promise<void>
  refreshMe: () => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<PublicUser | null>(null)
  const [loading, setLoading] = useState(true)

  const logout = useCallback(() => {
    clearToken()
    setUser(null)
  }, [])

  useEffect(() => {
    const token = getToken()
    if (!token) {
      setLoading(false)
      return
    }
    api
      .me()
      .then(setUser)
      .catch(() => {
        clearToken()
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    const onUnauthorized = () => logout()
    window.addEventListener('auth:unauthorized', onUnauthorized)
    return () => window.removeEventListener('auth:unauthorized', onUnauthorized)
  }, [logout])

  const login = useCallback(async (email: string, password: string) => {
    const out = await api.login(email, password)
    setToken(out.token)
    setUser(out.user)
  }, [])

  const register = useCallback(
    async (email: string, password: string, displayName: string, inviteCode: string) => {
      await api.register(email, password, displayName, inviteCode)
      await login(email, password)
    },
    [login],
  )

  const refreshMe = useCallback(async () => {
    const next = await api.me()
    setUser(next)
  }, [])

  useEffect(() => {
    if (!user) return
    const tick = () => {
      if (document.visibilityState !== 'visible') return
      void refreshMe().catch(() => undefined)
    }
    const id = window.setInterval(tick, 45000)
    document.addEventListener('visibilitychange', tick)
    return () => {
      window.clearInterval(id)
      document.removeEventListener('visibilitychange', tick)
    }
  }, [user, refreshMe])

  const value = useMemo(
    () => ({ user, loading, login, register, refreshMe, logout }),
    [user, loading, login, register, refreshMe, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth 必须在 AuthProvider 内使用')
  }
  return ctx
}
