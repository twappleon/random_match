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
      <button :class="{ active: mode === 'video' }" @click="mode = 'video'">视讯</button>
      <button :class="{ active: mode === 'voice' }" @click="mode = 'voice'">语音</button>
      <button class="primary" :disabled="loading" @click="startMatch">
        {{ loading ? '匹配中' : '随机匹配' }}
      </button>
    </nav>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { anonymousAuth, joinMatch, type MatchMode, wsURL } from './api'
import { initAnalytics } from './firebase'

type Status = 'idle' | 'waiting' | 'matched'

const mode = ref<MatchMode>('video')
const status = ref<Status>('idle')
const loading = ref(false)
const token = ref(localStorage.getItem('token') ?? '')
const ws = ref<WebSocket | null>(null)
const localStream = ref<MediaStream | null>(null)
const peer = ref<RTCPeerConnection | null>(null)
const remoteVideo = ref<HTMLVideoElement | null>(null)
const localVideo = ref<HTMLVideoElement | null>(null)

const stateText = computed(() => {
  if (status.value === 'waiting') return '正在寻找在线用户'
  return '准备开始随机交友'
})

initAnalytics()

async function ensureAuth() {
  if (token.value) return
  const auth = await anonymousAuth()
  token.value = auth.token
  localStorage.setItem('token', auth.token)
}

async function startMatch() {
  loading.value = true
  try {
    await ensureAuth()
    await openMedia()
    openSocket()
    const result = await joinMatch(token.value, mode.value)
    status.value = result.status === 'matched' ? 'matched' : 'waiting'
  } finally {
    loading.value = false
  }
}

async function openMedia() {
  localStream.value?.getTracks().forEach((track) => track.stop())
  localStream.value = await navigator.mediaDevices.getUserMedia({
    video: mode.value === 'video',
    audio: true
  })
  if (localVideo.value) localVideo.value.srcObject = localStream.value
}

function openSocket() {
  if (ws.value && ws.value.readyState === WebSocket.OPEN) return
  ws.value = new WebSocket(wsURL(token.value))
  ws.value.onmessage = async (event) => {
    const msg = JSON.parse(event.data)
    if (msg.type === 'matched') {
      status.value = 'matched'
      if (msg.initiator) await createPeer(msg.peerId)
    }
    if (msg.type === 'offer') await acceptOffer(msg.peerId, msg.data)
    if (msg.type === 'answer') await peer.value?.setRemoteDescription(msg.data)
    if (msg.type === 'candidate') await peer.value?.addIceCandidate(msg.data)
  }
}

async function createPeer(peerId: string) {
  peer.value = buildPeer(peerId)
  localStream.value?.getTracks().forEach((track) => peer.value?.addTrack(track, localStream.value!))
  const offer = await peer.value.createOffer()
  await peer.value.setLocalDescription(offer)
  send({ type: 'offer', peerId, data: offer })
}

async function acceptOffer(peerId: string, offer: RTCSessionDescriptionInit) {
  peer.value = buildPeer(peerId)
  localStream.value?.getTracks().forEach((track) => peer.value?.addTrack(track, localStream.value!))
  await peer.value.setRemoteDescription(offer)
  const answer = await peer.value.createAnswer()
  await peer.value.setLocalDescription(answer)
  send({ type: 'answer', peerId, data: answer })
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
  ws.value?.send(JSON.stringify(message))
}

onBeforeUnmount(() => {
  ws.value?.close()
  peer.value?.close()
  localStream.value?.getTracks().forEach((track) => track.stop())
})
</script>
