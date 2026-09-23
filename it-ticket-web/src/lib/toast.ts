import { ApiError } from '../api/client'

export function errText(err: unknown, fallback: string): string {
  return err instanceof ApiError ? err.message : fallback
}
