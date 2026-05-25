<template>
  <div class="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center z-[100] p-4">
    <div class="bg-white dark:bg-surface-900 rounded-3xl shadow-2xl w-full max-w-sm overflow-hidden">

      <!-- Header / status -->
      <div class="bg-gradient-to-b from-primary-600 to-primary-800 px-8 py-10 text-center">
        <img :src="targetAvatar" class="w-24 h-24 rounded-full object-cover mx-auto mb-4 ring-4 ring-white/30"/>
        <h2 class="text-white text-xl font-bold mb-1">{{ targetName }}</h2>
        <p class="text-primary-200 text-sm">{{ statusText }}</p>
        <!-- Timer when connected -->
        <p v-if="connected" class="text-white/80 text-lg font-mono mt-2">{{ formatTime(elapsed) }}</p>
      </div>

      <!-- Controls -->
      <div class="px-8 py-6 flex items-center justify-center gap-6">
        <!-- Mute -->
        <button @click="toggleMute" :class="['w-12 h-12 rounded-full flex items-center justify-center transition-colors', muted ? 'bg-red-100 dark:bg-red-900/30 text-red-600' : 'bg-gray-100 dark:bg-surface-800 text-gray-600 dark:text-gray-300']">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path v-if="!muted" d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3zM19 10v2a7 7 0 0 1-14 0v-2M12 19v4M8 23h8"/>
            <path v-else d="M1 1l22 22M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23M12 19v4M8 23h8"/>
          </svg>
        </button>

        <!-- Speaker -->
        <button @click="toggleSpeaker" :class="['w-12 h-12 rounded-full flex items-center justify-center transition-colors', speakerOff ? 'bg-red-100 dark:bg-red-900/30 text-red-600' : 'bg-gray-100 dark:bg-surface-800 text-gray-600 dark:text-gray-300']">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
            <path v-if="!speakerOff" d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
            <line v-else x1="23" y1="9" x2="17" y2="15"/>
          </svg>
        </button>

        <!-- Hang up -->
        <button @click="hangUp" class="w-16 h-16 bg-red-500 hover:bg-red-600 rounded-full flex items-center justify-center shadow-lg transition-colors">
          <svg class="w-7 h-7 text-white rotate-135" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12 19.79 19.79 0 0 1 1.62 3.37 2 2 0 0 1 3.6 1.11h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L7.91 8.69a16 16 0 0 0 6 6l.91-.91a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
          </svg>
        </button>
      </div>

      <!-- Hidden audio elements -->
      <audio ref="localAudio"  autoplay muted  class="hidden"/>
      <audio ref="remoteAudio" autoplay        class="hidden"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useChatStore } from '@/stores/chat'

const props = defineProps<{ target: any }>()
const emit  = defineEmits<{ close: [] }>()

const chat = useChatStore()

const localAudio  = ref<HTMLAudioElement>()
const remoteAudio = ref<HTMLAudioElement>()
const muted       = ref(false)
const speakerOff  = ref(false)
const connected   = ref(false)
const elapsed     = ref(0)
const callStatus  = ref<'calling' | 'ringing' | 'connected' | 'ended'>('calling')

let pc: RTCPeerConnection | null = null
let localStream: MediaStream | null = null
let timer: ReturnType<typeof setInterval>

const targetName   = computed(() => props.target?.profile?.first_name || props.target?.username || 'User')
const targetAvatar = computed(() => props.target?.profile?.avatar_url ||
  `https://ui-avatars.com/api/?name=${props.target?.username}&size=96&background=2563eb&color=fff`)

const statusText = computed(() => ({
  calling:   'Calling…',
  ringing:   'Ringing…',
  connected: 'Connected',
  ended:     'Call ended',
}[callStatus.value]))

const ICE_SERVERS: RTCIceServer[] = [
  { urls: import.meta.env.VITE_STUN_URL || 'stun:stun.l.google.com:19302' },
  ...(import.meta.env.VITE_TURN_URL ? [{
    urls: [
      import.meta.env.VITE_TURN_URL,
      ...(import.meta.env.VITE_TURN_URL_TCP ? [import.meta.env.VITE_TURN_URL_TCP] : []),
    ],
    username:   import.meta.env.VITE_TURN_USERNAME || '',
    credential: import.meta.env.VITE_TURN_PASSWORD || '',
  }] : []),
]

async function initCall() {
  localStream = await navigator.mediaDevices.getUserMedia({ audio: true })
  if (localAudio.value) localAudio.value.srcObject = localStream

  pc = new RTCPeerConnection({ iceServers: ICE_SERVERS })

  localStream.getTracks().forEach(t => pc!.addTrack(t, localStream!))

  pc.ontrack = (e) => {
    if (remoteAudio.value) remoteAudio.value.srcObject = e.streams[0]
    callStatus.value = 'connected'
    connected.value  = true
    timer = setInterval(() => elapsed.value++, 1000)
  }

  pc.onicecandidate = (e) => {
    if (e.candidate) {
      chat.sendRTC('rtc_candidate', props.target.id, { candidate: e.candidate })
    }
  }

  pc.onconnectionstatechange = () => {
    if (pc?.connectionState === 'disconnected' || pc?.connectionState === 'failed') {
      hangUp()
    }
  }

  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  chat.sendRTC('rtc_offer', props.target.id, { sdp: offer })
  callStatus.value = 'ringing'
}

async function handleAnswer(sdp: RTCSessionDescriptionInit) {
  if (!pc) return
  await pc.setRemoteDescription(sdp)
}

async function handleCandidate(candidate: RTCIceCandidateInit) {
  if (!pc) return
  try { await pc.addIceCandidate(candidate) } catch {}
}

function handleWSEvent(event: any) {
  if (event.type === 'rtc_answer') handleAnswer(event.payload.sdp)
  if (event.type === 'rtc_candidate') handleCandidate(event.payload.candidate)
  if (event.type === 'rtc_hangup') hangUp()
}

function toggleMute() {
  muted.value = !muted.value
  localStream?.getAudioTracks().forEach(t => { t.enabled = !muted.value })
}

function toggleSpeaker() {
  speakerOff.value = !speakerOff.value
  if (remoteAudio.value) remoteAudio.value.muted = speakerOff.value
}

function hangUp() {
  chat.sendRTC('rtc_hangup', props.target.id, {})
  cleanup()
  emit('close')
}

function cleanup() {
  clearInterval(timer)
  pc?.close(); pc = null
  localStream?.getTracks().forEach(t => t.stop())
  localStream = null
  callStatus.value = 'ended'
}

function formatTime(s: number) {
  const m = Math.floor(s / 60)
  return `${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

onMounted(() => {
  initCall()
  chat.onWSMessage(handleWSEvent)
})

onUnmounted(() => {
  cleanup()
  chat.offWSMessage(handleWSEvent)
})
</script>
