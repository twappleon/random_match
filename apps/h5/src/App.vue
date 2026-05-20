<template>
  <main class="screen">
    <section class="call-stage">
      <div class="remote-video">
        <video ref="remoteVideo" autoplay playsinline></video>
        <div v-if="status !== 'matched'" class="state">{{ stateText }}</div>
      </div>

      <video ref="localVideo" class="local-video" autoplay playsinline muted></video>
    </section>

    <nav class="toolbar" aria-label="match controls">
      <button :class="{ active: mode === 'video' }" :aria-pressed="mode === 'video'" @click="selectMode('video')">
        视讯{{ mode === 'video' ? '中' : '' }}
      </button>
      <button :class="{ active: mode === 'voice' }" :aria-pressed="mode === 'voice'" @click="selectMode('voice')">
        语音{{ mode === 'voice' ? '中' : '' }}
      </button>
      <button class="primary" :disabled="loading" @click="startMatch">
        {{ loading ? '匹配中' : '随机匹配' }}
      </button>
    </nav>

    <p v-if="errorText" class="error" role="alert">{{ errorText }}</p>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { anonymousAuth, joinMatch, verifySession, type MatchMode, wsURL } from './api'
import { initAnalytics } from './firebase'

type Status = 'idle' | 'waiting' | 'matched'

const mode = ref<MatchMode>('video')
const status = ref<Status>('idle')
const loading = ref(false)
const errorText = ref('')
const token = ref(localStorage.getItem('token') ?? '')
const ws = ref<WebSocket | null>(null)
const localStream = ref<MediaStream | null>(null)
const peer = ref<RTCPeerConnection | null>(null)
const activePeerId = ref<string | null>(null)
const pendingCandidates = ref<RTCIceCandidateInit[]>([])
const remoteVideo = ref<HTMLVideoElement | null>(null)
const localVideo = ref<HTMLVideoElement | null>(null)

const stateText = computed(() => {
  const modeText = mode.value === 'video' ? '视讯' : '语音'
  if (status.value === 'waiting') return `正在寻找${modeText}用户…\n请让另一位用户在另一浏览器窗口点击「随机匹配」`
  if (status.value === 'matched') return `已匹配，正在连接${modeText}…`
  return `已选择${modeText}匹配，点击下方按钮开始`
})

initAnalytics()

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
    await ensureAuth()
    await openMedia()
    await openSocketWithAuth()
    const result = await joinMatch(token.value, mode.value)
    status.value = result.status === 'matched' ? 'matched' : 'waiting'
    if (result.status === 'matched' && result.initiator && result.peerId) {
      await createPeer(result.peerId)
    }
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    loading.value = false
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
  localStream.value?.getTracks().forEach((track) => track.stop())
  localStream.value = await navigator.mediaDevices.getUserMedia({
    video: mode.value === 'video',
    audio: true
  })
  if (localVideo.value) localVideo.value.srcObject = localStream.value
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
        }
      } catch (error) {
        errorText.value = toUserMessage(error)
      }
    }
  })
}

function teardownPeer() {
  peer.value?.close()
  peer.value = null
  activePeerId.value = null
  pendingCandidates.value = []
  if (remoteVideo.value) remoteVideo.value.srcObject = null
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
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  })
  pc.onicecandidate = (event) => {
    if (event.candidate) send({ type: 'candidate', peerId, data: event.candidate })
  }
  pc.ontrack = (event) => {
    if (remoteVideo.value) remoteVideo.value.srcObject = event.streams[0]
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
  ws.value?.close()
  teardownPeer()
  localStream.value?.getTracks().forEach((track) => track.stop())
})
</script>
