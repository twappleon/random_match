const API_BASE = import.meta.env.VITE_API_BASE || (import.meta.env.DEV ? 'http://localhost:8080' : window.location.origin)

export type MatchMode = 'video' | 'voice'

export interface AuthResponse {
  token: string
  user: {
    id: string
    displayName: string
    avatarUrl: string
  }
}

export async function anonymousAuth(): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/api/v1/auth/anonymous`, { method: 'POST' })
  if (!res.ok) throw new Error('匿名登录失败，请确认后端服务已启动')
  return res.json()
}

export interface MatchResponse {
  status: 'waiting' | 'matched'
  roomId?: string
  peerId?: string
  initiator?: boolean
}

export async function joinMatch(token: string, mode: MatchMode, region = 'global'): Promise<MatchResponse> {
  const res = await fetch(`${API_BASE}/api/v1/match/join`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ mode, region })
  })
  if (!res.ok && res.status !== 202) throw new Error('加入匹配失败，请稍后重试')
  return res.json()
}

export function wsURL(token: string) {
  const url = new URL(API_BASE)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/api/v1/ws'
  url.searchParams.set('token', token)
  return url.toString()
}
