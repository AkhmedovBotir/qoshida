import { API_URL, REQUEST_TIMEOUT_MS } from '../constants/config'
import { notifyApiError, notifyApiSuccess } from './snack'

type ApiOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  signal?: AbortSignal
  silent?: boolean
}

type Envelope<T> = {
  success: boolean
  data?: T
  error?: string
}

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

const AUTH_REFRESH = '/api/v1/seller-auth/refresh'
const SKIP_REFRESH = /\/(login|logout|refresh|check-phone|send-code|resend-code|verify-code|set-password)$/

let refreshPromise: Promise<boolean> | null = null

async function tryRefresh(): Promise<boolean> {
  if (refreshPromise) return refreshPromise
  refreshPromise = (async () => {
    try {
      const response = await fetch(`${API_URL}${AUTH_REFRESH}`, {
        method: 'POST',
        credentials: 'include',
        headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
        body: '{}',
      })
      const payload = (await response.json()) as Envelope<unknown>
      return response.ok && payload.success
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()
  return refreshPromise
}

export async function api<T>(path: string, options: ApiOptions = {}, retried = false): Promise<T> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  try {
    const response = await fetch(`${API_URL}${path}`, {
      method: options.method ?? 'GET',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
        ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
      signal: options.signal ?? controller.signal,
    })

    let payload: Envelope<T>
    try {
      payload = (await response.json()) as Envelope<T>
    } catch {
      const parseErr = new ApiError(`So'rov xato: ${response.status}`, response.status)
      notifyApiError(path, parseErr, options.silent)
      throw parseErr
    }
    if (response.status === 401 && !retried && !SKIP_REFRESH.test(path)) {
      if (await tryRefresh()) {
        return api<T>(path, options, true)
      }
    }
    if (!response.ok || !payload.success || payload.data === undefined) {
      const apiErr = new ApiError(payload.error ?? `So'rov xato: ${response.status}`, response.status)
      notifyApiError(path, apiErr, options.silent)
      throw apiErr
    }

    notifyApiSuccess(path, options.method ?? 'GET', options.silent)
    return payload.data
  } catch (err) {
    if (err instanceof ApiError) throw err
    const wrapped = new ApiError(err instanceof Error && err.name === 'AbortError' ? 'So‘rov vaqti tugadi' : 'Tarmoq xatosi', 0)
    notifyApiError(path, wrapped, options.silent)
    throw wrapped
  } finally {
    window.clearTimeout(timer)
  }
}
