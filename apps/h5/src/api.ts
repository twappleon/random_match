function apiBase() {
  if (import.meta.env.VITE_API_BASE) return import.meta.env.VITE_API_BASE
  // Dev: same-origin via Vite proxy (works for localhost and LAN IP).
  if (import.meta.env.DEV) return window.location.origin
  return window.location.origin
}

export type MatchMode = 'video' | 'voice'

export interface AuthResponse {
  token: string
  user: {
    id: string
    displayName: string
    avatarUrl: string
  }
}

export async function verifySession(token: string): Promise<boolean> {
  const res = await fetch(`${apiBase()}/api/v1/auth/session`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.ok
}

export async function anonymousAuth(): Promise<AuthResponse> {
  const res = await fetch(`${apiBase()}/api/v1/auth/anonymous`, { method: 'POST' })
  if (!res.ok) throw new Error('匿名登录失败，请确认后端服务已启动')
  return res.json()
}

export interface MatchResponse {
  status: 'waiting' | 'matched'
  roomId?: string
  peerId?: string
  initiator?: boolean
}

export interface StatsResponse {
  online: number
  waiting: number
  chatting: number
}

export async function fetchStats(): Promise<StatsResponse> {
  const res = await fetch(`${apiBase()}/api/v1/stats`)
  if (!res.ok) throw new Error('统计数据读取失败')
  return res.json()
}

export async function joinMatch(token: string, mode: MatchMode, region = 'global'): Promise<MatchResponse> {
  const res = await fetch(`${apiBase()}/api/v1/match/join`, {
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

export async function leaveMatch(token: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/match/leave`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('退出失败，请稍后重试')
}

export interface SnapshotPayload {
  roomId: string
  peerId?: string
  mode: MatchMode
  image: string
  width: number
  height: number
}

export async function uploadMatchSnapshot(token: string, payload: SnapshotPayload): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/match/snapshot`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
  if (!res.ok) throw new Error('截图上传失败')
}

export function wsURL(token: string) {
  const url = new URL(apiBase())
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/api/v1/ws'
  url.searchParams.set('token', token)
  return url.toString()
}

export function iceServers(): RTCIceServer[] {
  const fallback: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }]
  const raw = import.meta.env.VITE_ICE_SERVERS
  if (!raw) return fallback
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) && parsed.length > 0 ? parsed : fallback
  } catch {
    return fallback
  }
}
