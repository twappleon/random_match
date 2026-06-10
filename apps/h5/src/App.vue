<template>
  <main class="screen">
    <section v-show="activePage === 'video'" ref="stage" class="call-stage page">
      <div class="remote-video">
        <video ref="remoteVideo" autoplay playsinline></video>
        <div v-if="status !== 'matched'" class="state">{{ stateText }}</div>
      </div>

      <div
        ref="localPreview"
        class="local-preview"
        :style="localPreviewStyle"
        @pointerdown="startPreviewDrag"
      >
        <video ref="localVideo" class="local-video" autoplay playsinline muted></video>
      </div>

      <aside class="stats-panel" aria-label="runtime stats">
        <span>在线 {{ stats.online }}</span>
        <span>等待 {{ stats.waiting }}</span>
        <span>聊天 {{ stats.chatting }}</span>
      </aside>

      <section v-if="status === 'matched' && !chatOpen && !peerCardHidden" class="peer-card" aria-label="peer profile">
        <button class="peer-card-close" type="button" aria-label="hide peer profile" @click="peerCardHidden = true">
          ×
        </button>
        <div class="peer-main">
          <div class="profile-head">
            <div class="avatar">{{ peerInitial }}</div>
            <div>
              <strong>{{ peerDisplayName }}</strong>
              <span>{{ peerBio }}</span>
            </div>
          </div>
          <div class="tags">
            <span v-for="item in peerInterests" :key="item">{{ item }}</span>
          </div>
        </div>
        <div class="safety-actions">
          <button class="report" :disabled="safetyLoading || reportedPeerId === activePeerId" @click="reportPeer">
            {{ reportedPeerId === activePeerId ? '已举报' : '举报' }}
          </button>
          <button class="block" :disabled="safetyLoading" @click="blockPeer">
            拉黑
          </button>
        </div>
      </section>

      <section v-if="chatOpen" class="chat-sheet" aria-label="text chat">
        <div class="chat-header">
          <strong>文字聊天</strong>
          <span>{{ chatHeaderText }}</span>
          <button type="button" aria-label="close text chat" @click="chatOpen = false">收起</button>
        </div>
        <div ref="chatList" class="chat-list">
          <p v-if="chatMessages.length === 0" class="chat-empty">{{ chatEmptyText }}</p>
          <div
            v-for="message in chatMessages"
            :key="message.id"
            class="chat-message"
            :class="{ mine: message.sender === 'self' }"
          >
            <span>{{ message.text }}</span>
          </div>
        </div>
        <form class="chat-form" @submit.prevent="sendChatMessage">
          <input
            v-model="chatDraft"
            maxlength="500"
            autocomplete="off"
            :placeholder="chatInputPlaceholder"
            :disabled="!canUseChat"
          />
          <button :disabled="!canSendChat">发送</button>
        </form>
      </section>
    </section>

    <section v-show="activePage === 'profile'" class="page page-shell">
      <div class="content-card profile-page" aria-label="profile setup">
        <div class="profile-head">
          <div class="avatar">{{ profileInitial }}</div>
          <div>
            <strong>匿名身份</strong>
            <span>用兴趣和状态开始匹配</span>
          </div>
        </div>
        <label>
          昵称
          <input v-model.trim="profileForm.displayName" maxlength="24" placeholder="星球旅人" />
        </label>
        <label>
          简介
          <textarea v-model.trim="profileForm.bio" maxlength="120" rows="3" placeholder="一句话介绍现在的你"></textarea>
        </label>
        <label>
          兴趣标签
          <input v-model="interestsText" placeholder="电影, 音乐, 旅行" />
        </label>
        <label class="age-check">
          <input v-model="profileForm.ageConfirmed" type="checkbox" />
          <span>我已满 18 岁并同意文明视讯</span>
        </label>
        <button class="save-profile" :disabled="savingProfile" @click="saveProfile">
          {{ savingProfile ? '保存中' : '保存资料' }}
        </button>
        <section class="blocked-section" aria-label="blocked users">
          <div class="section-head">
            <div>
              <strong>拉黑名单</strong>
              <span>解除后未来可能再次匹配到对方</span>
            </div>
            <button type="button" :disabled="loadingBlockedUsers" @click="loadBlockedUsers">
              {{ loadingBlockedUsers ? '读取中' : '刷新' }}
            </button>
          </div>
          <p v-if="!loadingBlockedUsers && blockedUsers.length === 0" class="empty-list">目前没有拉黑对象</p>
          <div v-else class="blocked-list">
            <div v-for="item in blockedUsers" :key="item.user.id" class="blocked-user">
              <div class="profile-head">
                <div class="avatar">{{ item.user.displayName.trim().slice(0, 1).toUpperCase() || '星' }}</div>
                <div>
                  <strong>{{ item.user.displayName || '匿名用户' }}</strong>
                  <span>{{ item.user.bio || `拉黑于 ${formatDate(item.createdAt)}` }}</span>
                </div>
              </div>
              <button type="button" :disabled="unblockingUserId === item.user.id" @click="unblockBlockedUser(item.user.id)">
                {{ unblockingUserId === item.user.id ? '解除中' : '解除' }}
              </button>
            </div>
          </div>
        </section>
      </div>
    </section>

    <section v-show="activePage === 'membership'" class="page page-shell">
      <div class="content-card membership-page" aria-label="membership">
        <div class="membership-summary">
          <strong>{{ membershipTitle }}</strong>
          <span>{{ membershipText }}</span>
        </div>
        <div class="benefits">
          <span>无限随机匹配</span>
          <span>进入优先队列</span>
          <span>免费额度用完后继续使用</span>
        </div>
        <button :disabled="paymentLoading || commerceStatus?.isMember" @click="buyMembership">
          {{ paymentButtonText }}
        </button>
      </div>
    </section>

    <nav v-if="activePage === 'video'" class="toolbar" aria-label="match controls">
      <button class="chat-toggle" @click="toggleChat">
        {{ chatOpen ? '收起文字' : '文字' }}
      </button>
      <button class="camera" :disabled="loading || switchingCamera || !localStream" @click="switchCamera">
        {{ switchingCamera ? '切换中' : nextCameraText }}
      </button>
      <button class="primary" :disabled="loading || status === 'waiting'" @click="startMatch">
        {{ actionText }}
      </button>
      <button class="danger" :disabled="leaving || status === 'idle'" @click="leaveCall">
        {{ leaving ? '退出中' : '退出' }}
      </button>
    </nav>

    <nav class="app-nav" aria-label="main navigation">
      <button :class="{ active: activePage === 'video' }" @click="switchPage('video')">视讯</button>
      <button :class="{ active: activePage === 'profile' }" @click="switchPage('profile')">资料</button>
      <button :class="{ active: activePage === 'membership' }" @click="switchPage('membership')">会员</button>
    </nav>

    <button v-if="chatToastText" class="chat-toast" type="button" @click="openChatFromToast">
      <strong>新文字讯息</strong>
      <span>{{ chatToastText }}</span>
    </button>

    <p v-if="errorText" class="error" :class="{ 'error-with-toolbar': activePage === 'video' }" role="alert">{{ errorText }}</p>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { anonymousAuth, blockUser, confirmPaymentOrder, createPaymentOrder, fetchBlockedUsers, fetchCommerceStatus, fetchProfile, fetchStats, iceServers, joinMatch, leaveMatch, reportUser, savePushSubscription, sendPushTest, unblockUser, updateProfile, uploadMatchSnapshot, vapidPublicKey, verifySession, type BlockedUser, type CommerceStatus, type MatchMode, type UserProfile, wsURL } from './api'
import { initAnalytics } from './firebase'

type Status = 'idle' | 'waiting' | 'matched'
type Page = 'video' | 'profile' | 'membership'
type ChatMessage = {
  id: string
  sender: 'self' | 'peer'
  text: string
  createdAt: string
}

const mode = ref<MatchMode>('video')
const activePage = ref<Page>('video')
const status = ref<Status>('idle')
const loading = ref(false)
const leaving = ref(false)
const switchingCamera = ref(false)
const savingProfile = ref(false)
const safetyLoading = ref(false)
const paymentLoading = ref(false)
const loadingBlockedUsers = ref(false)
const reportedPeerId = ref<string | null>(null)
const unblockingUserId = ref<string | null>(null)
const pushStatus = ref<'idle' | 'enabled' | 'blocked' | 'unsupported' | 'unconfigured'>('idle')
const errorText = ref('')
const profile = ref<UserProfile | null>(null)
const peerProfile = ref<UserProfile | null>(null)
const blockedUsers = ref<BlockedUser[]>([])
const commerceStatus = ref<CommerceStatus | null>(null)
const profileForm = ref({
  displayName: '星球旅人',
  bio: '',
  ageConfirmed: false
})
const interestsText = ref('聊天, 电影, 音乐')
const stats = ref({ online: 0, waiting: 0, chatting: 0 })
const statsTimer = ref<number | null>(null)
const token = ref(localStorage.getItem('token') ?? '')
const ws = ref<WebSocket | null>(null)
const wsHeartbeatTimer = ref<number | null>(null)
const closingSocket = ref(false)
const activeRoomId = ref<string | null>(null)
const chatOpen = ref(false)
const chatDraft = ref('')
const chatMessages = ref<ChatMessage[]>([])
const chatToastText = ref('')
const chatToastTimer = ref<number | null>(null)
const localStream = ref<MediaStream | null>(null)
const peer = ref<RTCPeerConnection | null>(null)
const peerDisconnectTimer = ref<number | null>(null)
const activePeerId = ref<string | null>(null)
const peerCardHidden = ref(false)
const pendingCandidates = ref<RTCIceCandidateInit[]>([])
const stage = ref<HTMLElement | null>(null)
const chatList = ref<HTMLElement | null>(null)
const localPreview = ref<HTMLElement | null>(null)
const remoteVideo = ref<HTMLVideoElement | null>(null)
const localVideo = ref<HTMLVideoElement | null>(null)
const previewPosition = ref({ x: 0, y: 0 })
const previewPositioned = ref(false)
const previewDrag = ref<{
  pointerId: number
  startX: number
  startY: number
  originX: number
  originY: number
} | null>(null)
const capturedSnapshotRooms = new Set<string>()
const cameraFacing = ref<'user' | 'environment'>('user')

const stateText = computed(() => {
  if (status.value === 'waiting') return '正在寻找视讯用户…\n请让另一位用户在另一浏览器窗口点击「随机匹配」'
  if (status.value === 'matched') return '已匹配，正在连接视讯…'
  return '点击下方按钮开始视讯匹配'
})

const actionText = computed(() => {
  if (loading.value) return '匹配中'
  if (status.value === 'waiting') return '等待中'
  if (status.value === 'matched') return '已连线'
  return '随机匹配'
})

const localPreviewStyle = computed(() => {
  if (!previewPositioned.value) return {}
  return {
    left: `${previewPosition.value.x}px`,
    top: `${previewPosition.value.y}px`
  }
})

const nextCameraText = computed(() => cameraFacing.value === 'user' ? '后镜头' : '前镜头')
const profileInitial = computed(() => (profileForm.value.displayName || '星').trim().slice(0, 1).toUpperCase())
const peerInitial = computed(() => (peerProfile.value?.displayName || '星').trim().slice(0, 1).toUpperCase())
const peerDisplayName = computed(() => peerProfile.value?.displayName || '对方资料载入中')
const peerBio = computed(() => peerProfile.value?.bio || '对方暂时没有填写简介')
const peerInterests = computed(() => peerProfile.value?.interests?.length ? peerProfile.value.interests : ['随机视讯'])
const membershipTitle = computed(() => commerceStatus.value?.isMember ? '会员已开启' : '免费匹配额度')
const membershipText = computed(() => {
  const status = commerceStatus.value
  if (!status) return '正在读取今日额度'
  if (status.isMember) return `无限匹配 · 优先排队${status.membershipExpiresAt ? ` · 到期 ${formatDate(status.membershipExpiresAt)}` : ''}`
  return `今日剩余 ${status.dailyRemaining}/${status.dailyLimit} 次 · 会员无限匹配并优先排队`
})
const paymentButtonText = computed(() => {
  if (commerceStatus.value?.isMember) return '已是会员'
  return paymentLoading.value ? '开通中' : '$6.99/月 开通'
})
const canSendChat = computed(() => status.value === 'matched' && Boolean(activePeerId.value) && chatDraft.value.trim().length > 0)
const canUseChat = computed(() => status.value === 'matched' && Boolean(activePeerId.value))
const chatEmptyText = computed(() => canUseChat.value ? '开始文字聊天' : '匹配成功后可文字聊天')
const chatInputPlaceholder = computed(() => canUseChat.value ? '输入消息...' : '等待匹配后开始聊天')
const chatHeaderText = computed(() => canUseChat.value ? peerDisplayName.value : '目前尚未连接对象')

initAnalytics()

async function refreshStats() {
  try {
    stats.value = await fetchStats()
  } catch {
    // Keep the last visible values if a short network hiccup happens.
  }
}

function startStatsPolling() {
  void refreshStats()
  statsTimer.value = window.setInterval(refreshStats, 5000)
}

function stopStatsPolling() {
  if (statsTimer.value !== null) {
    window.clearInterval(statsTimer.value)
    statsTimer.value = null
  }
}

function startSocketHeartbeat(socket: WebSocket) {
  stopSocketHeartbeat()
  wsHeartbeatTimer.value = window.setInterval(() => {
    if (ws.value !== socket || socket.readyState !== WebSocket.OPEN) {
      stopSocketHeartbeat()
      return
    }
    socket.send(JSON.stringify({ type: 'ping' }))
  }, 25000)
}

function stopSocketHeartbeat() {
  if (wsHeartbeatTimer.value !== null) {
    window.clearInterval(wsHeartbeatTimer.value)
    wsHeartbeatTimer.value = null
  }
}

async function switchPage(page: Page) {
  activePage.value = page
  if (page !== 'video') chatOpen.value = false
  if (page === 'video') {
    await nextTick()
    ensurePreviewPosition()
  }
  if (page === 'membership') {
    void loadCommerceStatus().catch(() => undefined)
  }
  if (page === 'profile') {
    void loadBlockedUsers().catch(() => undefined)
  }
}

function toggleChat() {
  chatOpen.value = !chatOpen.value
  if (chatOpen.value) {
    clearChatToast()
    scrollChatToBottom()
  }
}

function openChatFromToast() {
  activePage.value = 'video'
  chatOpen.value = true
  clearChatToast()
  scrollChatToBottom()
}

function clearToken() {
  token.value = ''
  localStorage.removeItem('token')
}

async function ensureAuth() {
  if (token.value && (await verifySession(token.value))) return
  clearToken()
  const auth = await anonymousAuth()
  token.value = auth.token
  localStorage.setItem('token', auth.token)
  setProfile(auth.user)
}

function setProfile(nextProfile: UserProfile) {
  profile.value = nextProfile
  profileForm.value = {
    displayName: nextProfile.displayName || '星球旅人',
    bio: nextProfile.bio || '',
    ageConfirmed: Boolean(nextProfile.ageConfirmed)
  }
  interestsText.value = (nextProfile.interests?.length ? nextProfile.interests : ['聊天', '电影', '音乐']).join(', ')
}

async function loadProfile() {
  try {
    await ensureAuth()
    setProfile(await fetchProfile(token.value))
    await loadCommerceStatus()
    await loadBlockedUsers()
  } catch {
    // Profile is refreshed again before matching.
  }
}

async function loadCommerceStatus() {
  await ensureAuth()
  commerceStatus.value = await fetchCommerceStatus(token.value)
}

async function loadBlockedUsers() {
  loadingBlockedUsers.value = true
  try {
    await ensureAuth()
    blockedUsers.value = await fetchBlockedUsers(token.value)
  } finally {
    loadingBlockedUsers.value = false
  }
}

async function unblockBlockedUser(userId: string) {
  if (unblockingUserId.value) return
  unblockingUserId.value = userId
  errorText.value = ''
  try {
    await ensureAuth()
    await unblockUser(token.value, userId)
    blockedUsers.value = blockedUsers.value.filter((item) => item.user.id !== userId)
    errorText.value = '已解除拉黑'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    unblockingUserId.value = null
  }
}

function parsedInterests() {
  const items = interestsText.value
    .split(/[,，]/)
    .map((item) => item.trim())
    .filter(Boolean)
  return Array.from(new Set(items)).slice(0, 6)
}

async function saveProfile() {
  savingProfile.value = true
  errorText.value = ''
  try {
    await persistProfile()
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    savingProfile.value = false
  }
}

async function persistProfile() {
  await ensureAuth()
  const updated = await updateProfile(token.value, {
    displayName: profileForm.value.displayName,
    bio: profileForm.value.bio,
    interests: parsedInterests(),
    ageConfirmed: profileForm.value.ageConfirmed
  })
  setProfile(updated)
}

async function setupPushNotifications(promptUser = false) {
  const publicKey = vapidPublicKey()
  if (!publicKey) {
    pushStatus.value = 'unconfigured'
    return
  }
  if (!('serviceWorker' in navigator) || !('PushManager' in window) || !('Notification' in window)) {
    pushStatus.value = 'unsupported'
    if (promptUser) errorText.value = iosPushHelpText()
    return
  }

  try {
    let permission = Notification.permission
    if (permission === 'default' && promptUser) {
      permission = await Notification.requestPermission()
    }
    if (permission === 'default') return
    if (permission !== 'granted') {
      pushStatus.value = 'blocked'
      if (promptUser) errorText.value = '通知权限未开启，请在浏览器网址列左侧设置里允许通知'
      return
    }

    await ensureAuth()
    const registration = await navigator.serviceWorker.register('/sw.js')
    const existing = await registration.pushManager.getSubscription()
    const subscription = existing ?? await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey)
    })
    const payload = subscription.toJSON()
    if (!payload.endpoint || !payload.keys?.auth || !payload.keys?.p256dh) return
    await savePushSubscription(token.value, {
      endpoint: payload.endpoint,
      keys: {
        auth: payload.keys.auth,
        p256dh: payload.keys.p256dh
      }
    })
    pushStatus.value = 'enabled'
    if (promptUser) {
      await sendPushTest(token.value)
    }
  } catch {
    if (promptUser && !errorText.value) errorText.value = '通知开启失败，请确认使用 HTTPS 并重新整理页面后再试'
  }
}

function iosPushHelpText() {
  if (!/iPad|iPhone|iPod/.test(navigator.userAgent)) return '当前浏览器不支持网页推播通知'
  return 'iPhone 需先用 Safari 分享按钮加入主画面，再从主画面打开后才能开启通知'
}

function urlBase64ToUint8Array(value: string) {
  const padding = '='.repeat((4 - value.length % 4) % 4)
  const base64 = (value + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = window.atob(base64)
  const output = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i += 1) {
    output[i] = raw.charCodeAt(i)
  }
  return output
}

async function startMatch() {
  loading.value = true
  errorText.value = ''
  try {
    await switchPage('video')
    resetCall()
    if (!profileForm.value.ageConfirmed) {
      throw new Error('请先确认已满 18 岁并保存资料')
    }
    await persistProfile()
    await setupPushNotifications(pushStatus.value === 'idle')
    await ensureAuth()
    await openMedia()
    await openSocketWithAuth()
    const result = await joinMatch(token.value, mode.value)
    await loadCommerceStatus()
    status.value = result.status === 'matched' ? 'matched' : 'waiting'
    if (result.status === 'matched' && result.roomId) {
      activeRoomId.value = result.roomId
      activePeerId.value = result.peerId || null
      peerCardHidden.value = false
      peerProfile.value = result.peerProfile || null
      void captureAndUploadSnapshot(result.roomId, result.peerId)
    }
    if (result.status === 'matched' && result.initiator && result.peerId) {
      await createPeer(result.peerId)
    }
    void refreshStats()
  } catch (error) {
    errorText.value = toUserMessage(error)
    void loadCommerceStatus().catch(() => undefined)
  } finally {
    loading.value = false
  }
}

async function buyMembership() {
  if (paymentLoading.value || commerceStatus.value?.isMember) return
  paymentLoading.value = true
  errorText.value = ''
  try {
    await ensureAuth()
    const order = await createPaymentOrder(token.value)
    await confirmPaymentOrder(token.value, order.id)
    await loadCommerceStatus()
    errorText.value = '会员已开通，可无限匹配并优先排队'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    paymentLoading.value = false
  }
}

async function leaveCall() {
  if (leaving.value || status.value === 'idle') return
  leaving.value = true
  errorText.value = ''
  try {
    if (token.value) await leaveMatch(token.value)
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    closeSocket()
    stopLocalMedia()
    resetCall()
    leaving.value = false
    void refreshStats()
  }
}

async function switchCamera() {
  if (switchingCamera.value || !localStream.value) return
  switchingCamera.value = true
  errorText.value = ''
  const nextFacing = cameraFacing.value === 'user' ? 'environment' : 'user'
  const currentFacing = cameraFacing.value
  const oldVideoTracks = localStream.value.getVideoTracks()

  try {
    const nextStream = await openCameraForSwitch(nextFacing, oldVideoTracks)
    const [nextVideoTrack] = nextStream.getVideoTracks()
    if (!nextVideoTrack) throw new Error('没有找到可用的摄像头')

    const currentStream = localStream.value
    const audioTracks = currentStream.getAudioTracks()
    localStream.value = new MediaStream([...audioTracks, nextVideoTrack])
    if (localVideo.value) localVideo.value.srcObject = localStream.value
    cameraFacing.value = nextFacing

    const sender = peer.value?.getSenders().find((item) => item.track?.kind === 'video')
    if (sender) {
      await sender.replaceTrack(nextVideoTrack)
      await setVideoSenderLimits(sender)
    }
    oldVideoTracks.forEach((track) => {
      if (track !== nextVideoTrack) track.stop()
    })
  } catch (error) {
    await restoreCamera(currentFacing)
    errorText.value = toUserMessage(error)
  } finally {
    switchingCamera.value = false
  }
}

async function openCameraForSwitch(facing: 'user' | 'environment', oldVideoTracks: MediaStreamTrack[]) {
  try {
    return await navigator.mediaDevices.getUserMedia({
      video: videoConstraints(facing),
      audio: false
    })
  } catch (error) {
    oldVideoTracks.forEach((track) => track.stop())
    try {
      return await navigator.mediaDevices.getUserMedia({
        video: videoConstraints(facing),
        audio: false
      })
    } catch {
      throw error
    }
  }
}

async function restoreCamera(facing: 'user' | 'environment') {
  if (!localStream.value) return
  try {
    const fallbackStream = await navigator.mediaDevices.getUserMedia({
      video: videoConstraints(facing),
      audio: false
    })
    const [fallbackVideoTrack] = fallbackStream.getVideoTracks()
    if (!fallbackVideoTrack) return

    const audioTracks = localStream.value.getAudioTracks()
    localStream.value = new MediaStream([...audioTracks, fallbackVideoTrack])
    if (localVideo.value) localVideo.value.srcObject = localStream.value

    const sender = peer.value?.getSenders().find((item) => item.track?.kind === 'video')
    if (sender) {
      await sender.replaceTrack(fallbackVideoTrack)
      await setVideoSenderLimits(sender)
    }
  } catch {
    // Keep the original switch error visible to the user.
  }
}

async function openMedia() {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('当前浏览器不支持摄像头/麦克风访问，请使用 HTTPS 或 localhost 打开页面')
  }
  stopLocalMedia()
  localStream.value = await navigator.mediaDevices.getUserMedia({
    video: videoConstraints(cameraFacing.value),
    audio: {
      echoCancellation: true,
      noiseSuppression: true,
      autoGainControl: true
    }
  })
  if (localVideo.value) localVideo.value.srcObject = localStream.value
  await nextTick()
  ensurePreviewPosition()
}

function videoConstraints(facingMode: 'user' | 'environment') {
  return {
    width: { ideal: 480, max: 640 },
    height: { ideal: 640, max: 720 },
    frameRate: { ideal: 15, max: 20 },
    facingMode: { ideal: facingMode }
  }
}

async function openSocketWithAuth() {
  try {
    await openSocket()
  } catch {
    clearToken()
    await ensureAuth()
    await openSocket()
  }
}

function openSocket() {
  if (ws.value?.readyState === WebSocket.OPEN) return Promise.resolve()
  ws.value?.close()

  return new Promise<void>((resolve, reject) => {
    const socket = new WebSocket(wsURL(token.value))
    ws.value = socket

    const failTimer = window.setTimeout(() => {
      reject(new Error('连接信令服务超时，请确认后端服务可访问'))
      socket.close()
    }, 8000)

    socket.onopen = () => {
      window.clearTimeout(failTimer)
      startSocketHeartbeat(socket)
      resolve()
    }

    socket.onerror = () => {
      window.clearTimeout(failTimer)
      reject(new Error('连接信令服务失败，登录可能已过期，请重试'))
    }

    socket.onclose = () => {
      stopSocketHeartbeat()
      if (ws.value === socket) ws.value = null
      if (closingSocket.value) {
        closingSocket.value = false
        return
      }
      if (status.value === 'matched') resetCall('信令连接已断开，请重新匹配')
    }

    socket.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'pong') return
        if (msg.type === 'matched') {
          status.value = 'matched'
          activeRoomId.value = msg.roomId || null
          activePeerId.value = msg.peerId || null
          peerCardHidden.value = false
          peerProfile.value = msg.peerProfile || null
          if (msg.roomId) void captureAndUploadSnapshot(msg.roomId, msg.peerId)
          if (msg.initiator && msg.peerId) await createPeer(msg.peerId)
          return
        }
        if (msg.type === 'offer' && msg.peerId) {
          await acceptOffer(msg.peerId, msg.data)
          return
        }
        if (msg.type === 'answer' && peer.value?.signalingState === 'have-local-offer') {
          await peer.value.setRemoteDescription(msg.data)
          await flushCandidates()
          return
        }
        if (msg.type === 'candidate') {
          await addRemoteCandidate(msg.data)
          return
        }
        if (msg.type === 'chat-message') {
          receiveChatMessage(msg.data)
          return
        }
        if (msg.type === 'peer-left') {
          resetCall('对方已离开，请重新匹配')
        }
      } catch (error) {
        errorText.value = toUserMessage(error)
      }
    }
  })
}

function closeSocket() {
  const socket = ws.value
  ws.value = null
  stopSocketHeartbeat()
  closingSocket.value = Boolean(socket)
  socket?.close()
}

function stopLocalMedia() {
  localStream.value?.getTracks().forEach((track) => track.stop())
  localStream.value = null
  if (localVideo.value) localVideo.value.srcObject = null
}

async function captureAndUploadSnapshot(roomId: string, peerId = '') {
  if (capturedSnapshotRooms.has(roomId)) return
  if (!token.value || !localStream.value?.getVideoTracks().length) return
  capturedSnapshotRooms.add(roomId)

  try {
    await waitForLocalVideoFrame()
    const video = localVideo.value
    if (!video?.videoWidth || !video.videoHeight) return

    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    const context = canvas.getContext('2d')
    if (!context) return
    context.drawImage(video, 0, 0, canvas.width, canvas.height)

    await uploadMatchSnapshot(token.value, {
      roomId,
      peerId,
      mode: 'video',
      image: canvas.toDataURL('image/jpeg', 0.82),
      width: canvas.width,
      height: canvas.height
    })
  } catch {
    capturedSnapshotRooms.delete(roomId)
  }
}

function waitForLocalVideoFrame() {
  const startedAt = performance.now()
  return new Promise<void>((resolve, reject) => {
    const check = () => {
      const video = localVideo.value
      if (video?.videoWidth && video.videoHeight) {
        resolve()
        return
      }
      if (performance.now() - startedAt > 1800) {
        reject(new Error('local video frame timeout'))
        return
      }
      window.requestAnimationFrame(check)
    }
    check()
  })
}

function previewBounds() {
  const stageRect = stage.value?.getBoundingClientRect()
  const previewRect = localPreview.value?.getBoundingClientRect()
  if (!stageRect || !previewRect) return null
  return {
    maxX: Math.max(0, stageRect.width - previewRect.width - 16),
    maxY: Math.max(0, stageRect.height - previewRect.height - 16)
  }
}

function clampPreviewPosition(x: number, y: number) {
  const bounds = previewBounds()
  if (!bounds) return { x, y }
  return {
    x: Math.min(Math.max(16, x), bounds.maxX),
    y: Math.min(Math.max(16, y), bounds.maxY)
  }
}

function ensurePreviewPosition() {
  const bounds = previewBounds()
  if (!bounds) return
  if (!previewPositioned.value) {
    previewPosition.value = { x: bounds.maxX, y: bounds.maxY }
    previewPositioned.value = true
    return
  }
  previewPosition.value = clampPreviewPosition(previewPosition.value.x, previewPosition.value.y)
}

function startPreviewDrag(event: PointerEvent) {
  if (!localPreview.value || !previewPositioned.value) ensurePreviewPosition()
  previewDrag.value = {
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    originX: previewPosition.value.x,
    originY: previewPosition.value.y
  }
  localPreview.value?.setPointerCapture(event.pointerId)
  window.addEventListener('pointermove', dragPreview)
  window.addEventListener('pointerup', stopPreviewDrag)
  window.addEventListener('pointercancel', stopPreviewDrag)
}

function dragPreview(event: PointerEvent) {
  const drag = previewDrag.value
  if (!drag || event.pointerId !== drag.pointerId) return
  previewPosition.value = clampPreviewPosition(
    drag.originX + event.clientX - drag.startX,
    drag.originY + event.clientY - drag.startY
  )
}

function stopPreviewDrag(event: PointerEvent) {
  const drag = previewDrag.value
  if (!drag || event.pointerId !== drag.pointerId) return
  localPreview.value?.releasePointerCapture(event.pointerId)
  previewDrag.value = null
  window.removeEventListener('pointermove', dragPreview)
  window.removeEventListener('pointerup', stopPreviewDrag)
  window.removeEventListener('pointercancel', stopPreviewDrag)
}

function teardownPeer(clearSession = true) {
  clearPeerDisconnectTimer()
  const currentPeer = peer.value
  peer.value = null
  currentPeer?.close()
  activePeerId.value = null
  peerCardHidden.value = false
  pendingCandidates.value = []
  if (clearSession) {
    activeRoomId.value = null
    peerProfile.value = null
    chatOpen.value = false
    clearChatToast()
    chatDraft.value = ''
    chatMessages.value = []
  }
  if (remoteVideo.value) remoteVideo.value.srcObject = null
}

function clearPeerDisconnectTimer() {
  if (peerDisconnectTimer.value !== null) {
    window.clearTimeout(peerDisconnectTimer.value)
    peerDisconnectTimer.value = null
  }
}

function schedulePeerDisconnect(message: string) {
  if (peerDisconnectTimer.value !== null) return
  peerDisconnectTimer.value = window.setTimeout(() => {
    peerDisconnectTimer.value = null
    resetCall(message)
  }, 5000)
}

function resetCall(message = '') {
  teardownPeer()
  status.value = 'idle'
  if (message) errorText.value = message
}

async function reportPeer() {
  if (!activePeerId.value || safetyLoading.value) return
  safetyLoading.value = true
  errorText.value = ''
  try {
    await reportUser(token.value, activePeerId.value, 'user reported during match')
    reportedPeerId.value = activePeerId.value
    errorText.value = '已收到举报'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    safetyLoading.value = false
  }
}

async function blockPeer() {
  if (!activePeerId.value || safetyLoading.value) return
  const confirmed = window.confirm('确定要拉黑并结束当前视讯吗？之后不会再匹配到这个用户。')
  if (!confirmed) return
  safetyLoading.value = true
  errorText.value = ''
  try {
    await blockUser(token.value, activePeerId.value)
    closeSocket()
    stopLocalMedia()
    resetCall('已拉黑对方，不会再匹配到此用户')
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    safetyLoading.value = false
  }
}

async function addRemoteCandidate(candidate: RTCIceCandidateInit) {
  if (!peer.value?.remoteDescription) {
    pendingCandidates.value.push(candidate)
    return
  }
  await peer.value.addIceCandidate(candidate)
}

async function flushCandidates() {
  if (!peer.value?.remoteDescription) return
  for (const candidate of pendingCandidates.value) {
    await peer.value.addIceCandidate(candidate)
  }
  pendingCandidates.value = []
}

async function createPeer(peerId: string) {
  if (activePeerId.value === peerId && peer.value?.localDescription?.type === 'offer') return

  teardownPeer(false)
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  await addLocalTracks(pc)
  if (!isCurrentPeer(pc)) return

  const offer = await pc.createOffer()
  if (!isCurrentPeer(pc)) return
  await pc.setLocalDescription(offer)
  if (!isCurrentPeer(pc)) return
  send({ type: 'offer', peerId, data: pc.localDescription })
}

async function acceptOffer(peerId: string, offer: RTCSessionDescriptionInit) {
  if (activePeerId.value === peerId && peer.value?.localDescription?.type === 'answer') return

  teardownPeer(false)
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  await addLocalTracks(pc)
  if (!isCurrentPeer(pc)) return

  await pc.setRemoteDescription(offer)
  if (!isCurrentPeer(pc)) return
  await flushCandidates()
  if (!isCurrentPeer(pc)) return
  const answer = await pc.createAnswer()
  if (!isCurrentPeer(pc)) return
  await pc.setLocalDescription(answer)
  if (!isCurrentPeer(pc)) return
  send({ type: 'answer', peerId, data: pc.localDescription })
}

function isCurrentPeer(pc: RTCPeerConnection) {
  return peer.value === pc && pc.signalingState !== 'closed'
}

function buildPeer(peerId: string) {
  const pc = new RTCPeerConnection({
    iceServers: iceServers(),
    iceTransportPolicy: import.meta.env.VITE_FORCE_TURN === 'true' ? 'relay' : 'all'
  })
  pc.onicecandidate = (event) => {
    if (!isCurrentPeer(pc)) return
    if (event.candidate) send({ type: 'candidate', peerId, data: event.candidate })
  }
  pc.onconnectionstatechange = () => {
    if (!isCurrentPeer(pc)) return
    if (pc.connectionState === 'connected') {
      clearPeerDisconnectTimer()
      return
    }
    if (pc.connectionState === 'disconnected') {
      schedulePeerDisconnect('对方连接不稳定，请重新匹配')
      return
    }
    if (['failed', 'closed'].includes(pc.connectionState)) {
      resetCall('对方已断线，请重新匹配')
    }
  }
  pc.oniceconnectionstatechange = () => {
    if (!isCurrentPeer(pc)) return
    if (['connected', 'completed'].includes(pc.iceConnectionState)) {
      clearPeerDisconnectTimer()
      return
    }
    if (pc.iceConnectionState === 'disconnected') {
      schedulePeerDisconnect('对方连接不稳定，请重新匹配')
      return
    }
    if (['failed', 'closed'].includes(pc.iceConnectionState)) {
      resetCall('对方连接已中断，请重新匹配')
    }
  }
  pc.ontrack = (event) => {
    if (!isCurrentPeer(pc)) return
    const [stream] = event.streams
    if (remoteVideo.value) remoteVideo.value.srcObject = stream
    event.track.onended = () => {
      if (isCurrentPeer(pc)) resetCall('对方已离开，请重新匹配')
    }
    event.track.onmute = () => {
      if (!isCurrentPeer(pc)) return
      if (pc.connectionState !== 'connected') resetCall('对方媒体已中断，请重新匹配')
    }
  }
  return pc
}

async function addLocalTracks(pc: RTCPeerConnection) {
  const stream = localStream.value
  if (!stream) return
  for (const track of stream.getTracks()) {
    const sender = pc.addTrack(track, stream)
    if (track.kind !== 'video') continue
    await setVideoSenderLimits(sender)
  }
}

async function setVideoSenderLimits(sender: RTCRtpSender) {
  const params = sender.getParameters()
  params.encodings = params.encodings?.length ? params.encodings : [{}]
  params.encodings[0] = {
    ...params.encodings[0],
    maxBitrate: 420_000,
    maxFramerate: 20
  }
  try {
    await sender.setParameters(params)
  } catch {
    // Some browsers reject sender parameter changes before negotiation.
  }
}

function send(message: unknown) {
  if (ws.value?.readyState !== WebSocket.OPEN) {
    errorText.value = '信令连接已断开，请重新匹配'
    return
  }
  ws.value.send(JSON.stringify(message))
}

function sendChatMessage() {
  if (!canSendChat.value || !activePeerId.value) return
  const text = chatDraft.value.trim()
  const message: ChatMessage = {
    id: newMessageId(),
    sender: 'self',
    text,
    createdAt: new Date().toISOString()
  }
  chatDraft.value = ''
  chatMessages.value.push(message)
  scrollChatToBottom()
  send({
    type: 'chat-message',
    peerId: activePeerId.value,
    roomId: activeRoomId.value || undefined,
    data: {
      id: message.id,
      text: message.text,
      createdAt: message.createdAt
    }
  })
}

function receiveChatMessage(data: unknown) {
  if (!data || typeof data !== 'object') return
  const payload = data as { id?: unknown; text?: unknown; createdAt?: unknown }
  if (typeof payload.text !== 'string') return
  const text = payload.text.trim()
  if (!text) return
  chatMessages.value.push({
    id: typeof payload.id === 'string' ? payload.id : newMessageId(),
    sender: 'peer',
    text: truncateText(text, 500),
    createdAt: typeof payload.createdAt === 'string' ? payload.createdAt : new Date().toISOString()
  })
  if (!chatOpen.value) showChatToast(text)
  scrollChatToBottom()
}

function showChatToast(text: string) {
  chatToastText.value = truncateText(text, 42)
  if (chatToastText.value !== text) chatToastText.value += '...'
  if (chatToastTimer.value !== null) window.clearTimeout(chatToastTimer.value)
  chatToastTimer.value = window.setTimeout(clearChatToast, 3500)
}

function clearChatToast() {
  chatToastText.value = ''
  if (chatToastTimer.value !== null) {
    window.clearTimeout(chatToastTimer.value)
    chatToastTimer.value = null
  }
}

function scrollChatToBottom() {
  void nextTick(() => {
    const list = chatList.value
    if (list) list.scrollTop = list.scrollHeight
  })
}

function newMessageId() {
  if (crypto.randomUUID) return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function truncateText(value: string, maxLength: number) {
  const segmenter = typeof Intl !== 'undefined' && 'Segmenter' in Intl
    ? new Intl.Segmenter(undefined, { granularity: 'grapheme' })
    : null
  const chars = segmenter
    ? Array.from(segmenter.segment(value), (item) => item.segment)
    : Array.from(value)
  return chars.length > maxLength ? chars.slice(0, maxLength).join('') : value
}

function toUserMessage(error: unknown) {
  if (error instanceof DOMException && error.name === 'NotAllowedError') {
    return '请允许摄像头和麦克风权限后再开始匹配'
  }
  if (error instanceof DOMException && error.name === 'NotFoundError') {
    return '没有找到可用的摄像头或麦克风，或当前设备没有另一个镜头'
  }
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return '无法连接后端服务，请确认 API 服务已启动且允许当前页面域名'
  }
  if (error instanceof Error) return error.message
  return '操作失败，请稍后重试'
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit'
  }).format(new Date(value))
}

onBeforeUnmount(() => {
  stopStatsPolling()
  stopSocketHeartbeat()
  clearChatToast()
  window.removeEventListener('resize', ensurePreviewPosition)
  window.removeEventListener('pointermove', dragPreview)
  window.removeEventListener('pointerup', stopPreviewDrag)
  window.removeEventListener('pointercancel', stopPreviewDrag)
  closeSocket()
  teardownPeer()
  stopLocalMedia()
})

onMounted(() => {
  window.addEventListener('resize', ensurePreviewPosition)
  startStatsPolling()
  void loadProfile()
  void setupPushNotifications(false)
})
</script>
