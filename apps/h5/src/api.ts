function apiBase() {
  if (import.meta.env.VITE_API_BASE) return import.meta.env.VITE_API_BASE
  // Dev: same-origin via Vite proxy (works for localhost and LAN IP).
  if (import.meta.env.DEV) return window.location.origin
  return window.location.origin
}

export type MatchMode = 'video' | 'voice'

export interface AuthResponse {
  token: string
  user: UserProfile
}

export interface UserProfile {
  id: string
  displayName: string
  avatarUrl: string
  bio?: string
  interests?: string[]
  region?: string
  gender?: string
  language?: string
  trustBadge?: boolean
  ageConfirmed: boolean
  membershipPlan?: string
  membershipExpiresAt?: string
  isMember?: boolean
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
  peerProfile?: UserProfile
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

export async function fetchProfile(token: string): Promise<UserProfile> {
  const res = await fetch(`${apiBase()}/api/v1/me`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('读取资料失败')
  const payload = await res.json()
  return payload.user
}

export async function updateProfile(token: string, payload: {
  displayName: string
  bio: string
  interests: string[]
  language: string
  ageConfirmed: boolean
}): Promise<UserProfile> {
  const res = await fetch(`${apiBase()}/api/v1/me`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
  if (!res.ok) throw new Error('保存资料失败')
  const data = await res.json()
  return data.user
}

export async function joinMatch(token: string, payload: {
  mode: MatchMode
  region?: string
  gender?: string
  language?: string
  interests?: string[]
}): Promise<MatchResponse> {
  const res = await fetch(`${apiBase()}/api/v1/match/join`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
  if (res.status === 402) {
    const payload = await res.json().catch(() => null)
    const remaining = typeof payload?.dailyRemaining === 'number' ? payload.dailyRemaining : 0
    throw new Error(`今日免费匹配次数已用完，剩余 ${remaining} 次。开通会员可无限匹配并优先排队`)
  }
  if (res.status === 409) throw new Error('已跳过拉黑用户，请再点一次随机匹配')
  if (!res.ok && res.status !== 202) throw new Error('加入匹配失败，请稍后重试')
  return res.json()
}

export async function fetchDiscoverProfiles(token: string, payload: {
  region: string
  gender: string
}): Promise<UserProfile[]> {
  const params = new URLSearchParams({
    region: payload.region,
    gender: payload.gender
  })
  const res = await fetch(`${apiBase()}/api/v1/discover/profiles?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('读取 Lounge 列表失败')
  const data = await res.json()
  return Array.isArray(data.users) ? data.users : []
}

export async function fetchFollowedUsers(token: string): Promise<UserProfile[]> {
  const res = await fetch(`${apiBase()}/api/v1/users/follows`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('读取关注列表失败')
  const data = await res.json()
  return Array.isArray(data.users) ? data.users : []
}

export async function followUser(token: string, userId: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/follow`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('关注失败')
}

export async function unfollowUser(token: string, userId: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/follow`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('取消关注失败')
}

export async function sendDirectMessage(token: string, userId: string, text: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/messages`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ text })
  })
  if (!res.ok) throw new Error('私信发送失败')
}

export interface CommerceStatus {
  isMember: boolean
  membershipPlan?: string
  membershipExpiresAt?: string
  dailyLimit: number
  dailyUsed: number
  dailyRemaining: number
  priorityQueue: boolean
  gemsBalance: number
}

export interface PaymentOrder {
  id: string
  userId: string
  plan: string
  amount: number
  currency: string
  status: string
  createdAt: string
  paidAt?: string
}

export async function fetchCommerceStatus(token: string): Promise<CommerceStatus> {
  const res = await fetch(`${apiBase()}/api/v1/commerce/status`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('读取会员状态失败')
  return res.json()
}

export async function createPaymentOrder(token: string, plan = 'premium_monthly'): Promise<PaymentOrder> {
  const res = await fetch(`${apiBase()}/api/v1/commerce/orders`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ plan })
  })
  if (!res.ok) throw new Error('创建订单失败')
  const payload = await res.json()
  return payload.order
}

export async function confirmPaymentOrder(token: string, orderId: string): Promise<PaymentOrder> {
  const res = await fetch(`${apiBase()}/api/v1/commerce/orders/${encodeURIComponent(orderId)}/confirm`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('支付确认失败')
  const payload = await res.json()
  return payload.order
}

export async function leaveMatch(token: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/match/leave`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('退出失败，请稍后重试')
}

export async function blockUser(token: string, userId: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/block`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('拉黑失败')
}

export interface BlockedUser {
  user: UserProfile
  createdAt: string
}

export async function fetchBlockedUsers(token: string): Promise<BlockedUser[]> {
  const res = await fetch(`${apiBase()}/api/v1/users/blocks`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('读取拉黑名单失败')
  const payload = await res.json()
  return Array.isArray(payload.users) ? payload.users : []
}

export async function unblockUser(token: string, userId: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/block`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('解除拉黑失败')
}

export async function reportUser(token: string, userId: string, reason = 'inappropriate behavior'): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/users/${encodeURIComponent(userId)}/report`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ reason })
  })
  if (!res.ok) throw new Error('举报失败')
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

export interface PushSubscriptionPayload {
  endpoint: string
  keys: {
    auth: string
    p256dh: string
  }
}

export async function savePushSubscription(token: string, payload: PushSubscriptionPayload): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/push/subscription`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
  if (!res.ok) throw new Error('通知订阅保存失败')
}

export async function sendPushTest(token: string): Promise<void> {
  const res = await fetch(`${apiBase()}/api/v1/push/test`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('服务器测试推送失败')
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

export function vapidPublicKey() {
  return import.meta.env.VITE_VAPID_PUBLIC_KEY || ''
}
