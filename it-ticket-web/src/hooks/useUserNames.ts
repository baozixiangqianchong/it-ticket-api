import { useEffect, useState } from 'react'

import { api } from '../api/client'
import { useAuth } from '../auth/AuthContext'

export function useUserNames() {
  const { user } = useAuth()
  const [names, setNames] = useState<Record<number, string>>({})

  useEffect(() => {
    if (!user) return
    const mine: Record<number, string> = { [user.id]: user.display_name }
    setNames((prev) => ({ ...mine, ...prev }))
    if (user.role !== 'admin') return
    api
      .listUsers()
      .then((page) => {
        const next: Record<number, string> = { ...mine }
        for (const item of page.items ?? []) {
          next[item.id] = item.display_name
        }
        setNames(next)
      })
      .catch(() => undefined)
  }, [user])

  return names
}

export function userLabel(
  id: number | null | undefined,
  names: Record<number, string>,
  empty = '未指派',
): string {
  if (id == null) return empty
  return names[id] ?? `用户 #${id}`
}
