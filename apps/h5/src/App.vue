<template>
  <main class="screen">
    <section ref="stage" class="call-stage">
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
    </section>

    <nav class="toolbar" aria-label="match controls">
      <button :class="{ active: mode === 'video' }" :aria-pressed="mode === 'video'" :disabled="loading || status !== 'idle'" @click="selectMode('video')">
        视讯{{ mode === 'video' ? '中' : '' }}
      </button>
      <button :class="{ active: mode === 'voice' }" :aria-pressed="mode === 'voice'" :disabled="loading || status !== 'idle'" @click="selectMode('voice')">
        语音{{ mode === 'voice' ? '中' : '' }}
      </button>
      <button class="primary" :disabled="loading || status === 'waiting'" @click="startMatch">
        {{ actionText }}
      </button>
      <button class="danger" :disabled="leaving || status === 'idle'" @click="leaveCall">
        {{ leaving ? '退出中' : '退出' }}
      </button>
    </nav>

    <p v-if="errorText" class="error" role="alert">{{ errorText }}</p>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { anonymousAuth, fetchStats, iceServers, joinMatch, leaveMatch, verifySession, type MatchMode, wsURL } from './api'
import { initAnalytics } from './firebase'

type Status = 'idle' | 'waiting' | 'matched'

const mode = ref<MatchMode>('video')
const status = ref<Status>('idle')
const loading = ref(false)
const leaving = ref(false)
const errorText = ref('')
const stats = ref({ online: 0, waiting: 0, chatting: 0 })
const statsTimer = ref<number | null>(null)
const token = ref(localStorage.getItem('token') ?? '')
const ws = ref<WebSocket | null>(null)
const closingSocket = ref(false)
const localStream = ref<MediaStream | null>(null)
const peer = ref<RTCPeerConnection | null>(null)
const activePeerId = ref<string | null>(null)
const pendingCandidates = ref<RTCIceCandidateInit[]>([])
const stage = ref<HTMLElement | null>(null)
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

const stateText = computed(() => {
  const modeText = mode.value === 'video' ? '视讯' : '语音'
  if (status.value === 'waiting') return `正在寻找${modeText}用户…\n请让另一位用户在另一浏览器窗口点击「随机匹配」`
  if (status.value === 'matched') return `已匹配，正在连接${modeText}…`
  return `已选择${modeText}匹配，点击下方按钮开始`
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
}

async function startMatch() {
  loading.value = true
  errorText.value = ''
  try {
    resetCall()
    await ensureAuth()
    await openMedia()
    await openSocketWithAuth()
    const result = await joinMatch(token.value, mode.value)
    status.value = result.status === 'matched' ? 'matched' : 'waiting'
    if (result.status === 'matched' && result.initiator && result.peerId) {
      await createPeer(result.peerId)
    }
    void refreshStats()
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    loading.value = false
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

async function selectMode(nextMode: MatchMode) {
  if (mode.value === nextMode || loading.value) return
  mode.value = nextMode
  errorText.value = ''

  if (!localStream.value) return

  try {
    await openMedia()
  } catch (error) {
    errorText.value = toUserMessage(error)
  }
}

async function openMedia() {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('当前浏览器不支持摄像头/麦克风访问，请使用 HTTPS 或 localhost 打开页面')
  }
  stopLocalMedia()
  localStream.value = await navigator.mediaDevices.getUserMedia({
    video: mode.value === 'video',
    audio: true
  })
  if (localVideo.value) localVideo.value.srcObject = localStream.value
  await nextTick()
  ensurePreviewPosition()
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
      resolve()
    }

    socket.onerror = () => {
      window.clearTimeout(failTimer)
      reject(new Error('连接信令服务失败，登录可能已过期，请重试'))
    }

    socket.onclose = () => {
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
        if (msg.type === 'matched') {
          status.value = 'matched'
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
  closingSocket.value = Boolean(socket)
  socket?.close()
}

function stopLocalMedia() {
  localStream.value?.getTracks().forEach((track) => track.stop())
  localStream.value = null
  if (localVideo.value) localVideo.value.srcObject = null
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

function teardownPeer() {
  peer.value?.close()
  peer.value = null
  activePeerId.value = null
  pendingCandidates.value = []
  if (remoteVideo.value) remoteVideo.value.srcObject = null
}

function resetCall(message = '') {
  teardownPeer()
  status.value = 'idle'
  if (message) errorText.value = message
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

  teardownPeer()
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  localStream.value?.getTracks().forEach((track) => pc.addTrack(track, localStream.value!))

  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  send({ type: 'offer', peerId, data: pc.localDescription })
}

async function acceptOffer(peerId: string, offer: RTCSessionDescriptionInit) {
  if (activePeerId.value === peerId && peer.value?.localDescription?.type === 'answer') return

  teardownPeer()
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  localStream.value?.getTracks().forEach((track) => pc.addTrack(track, localStream.value!))

  await pc.setRemoteDescription(offer)
  await flushCandidates()
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  send({ type: 'answer', peerId, data: pc.localDescription })
}

function buildPeer(peerId: string) {
  const pc = new RTCPeerConnection({
    iceServers: iceServers(),
    iceTransportPolicy: import.meta.env.VITE_FORCE_TURN === 'true' ? 'relay' : 'all'
  })
  pc.onicecandidate = (event) => {
    if (event.candidate) send({ type: 'candidate', peerId, data: event.candidate })
  }
  pc.onconnectionstatechange = () => {
    if (['disconnected', 'failed', 'closed'].includes(pc.connectionState)) {
      resetCall('对方已断线，请重新匹配')
    }
  }
  pc.oniceconnectionstatechange = () => {
    if (['disconnected', 'failed', 'closed'].includes(pc.iceConnectionState)) {
      resetCall('对方连接已中断，请重新匹配')
    }
  }
  pc.ontrack = (event) => {
    const [stream] = event.streams
    if (remoteVideo.value) remoteVideo.value.srcObject = stream
    event.track.onended = () => resetCall('对方已离开，请重新匹配')
    event.track.onmute = () => {
      if (pc.connectionState !== 'connected') resetCall('对方媒体已中断，请重新匹配')
    }
  }
  return pc
}

function send(message: unknown) {
  if (ws.value?.readyState !== WebSocket.OPEN) {
    errorText.value = '信令连接已断开，请重新匹配'
    return
  }
  ws.value.send(JSON.stringify(message))
}

function toUserMessage(error: unknown) {
  if (error instanceof DOMException && error.name === 'NotAllowedError') {
    return '请允许摄像头和麦克风权限后再开始匹配'
  }
  if (error instanceof DOMException && error.name === 'NotFoundError') {
    return '没有找到可用的摄像头或麦克风'
  }
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return '无法连接后端服务，请确认 API 服务已启动且允许当前页面域名'
  }
  if (error instanceof Error) return error.message
  return '操作失败，请稍后重试'
}

onBeforeUnmount(() => {
  stopStatsPolling()
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
})
</script>
