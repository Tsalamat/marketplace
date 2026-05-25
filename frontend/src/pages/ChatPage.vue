<template>
  <div class="min-h-screen flex flex-col bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="flex flex-1 pt-16 overflow-hidden" style="height:calc(100vh - 64px)">

      <!-- Chat List -->
      <aside class="w-80 flex-shrink-0 border-r border-gray-200 dark:border-surface-800 bg-white dark:bg-surface-900 flex flex-col">
        <div class="p-4 border-b border-gray-100 dark:border-surface-800">
          <div class="flex items-center justify-between mb-3">
            <h2 class="font-bold text-gray-900 dark:text-white text-lg">Сообщения</h2>
            <div class="flex items-center gap-1">
              <div :class="['w-2 h-2 rounded-full', chatStore.isConnected ? 'bg-green-500' : 'bg-red-400']"/>
              <span class="text-xs text-gray-400">{{ chatStore.isConnected ? 'Онлайн' : 'Не в сети' }}</span>
            </div>
          </div>
          <input v-model="chatSearch" type="text" placeholder="Поиск…" class="input text-sm py-2"/>
        </div>

        <div class="flex-1 overflow-y-auto">
          <div v-if="!chatStore.chats.length" class="p-8 text-center">
            <svg class="w-12 h-12 text-gray-300 dark:text-surface-700 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            <p class="text-sm text-gray-400">Нет диалогов</p>
          </div>
          <button v-for="c in filteredChats" :key="c.id" @click="openChat(c)"
            :class="['w-full flex items-center gap-3 px-4 py-3.5 hover:bg-gray-50 dark:hover:bg-surface-800/60 transition-colors text-left border-l-2', chatStore.activeChat?.id === c.id ? 'border-primary-600 bg-primary-50/50 dark:bg-primary-900/10' : 'border-transparent']">
            <div class="relative flex-shrink-0">
              <img :src="otherAvatar(c)" class="w-11 h-11 rounded-full object-cover ring-1 ring-gray-200 dark:ring-surface-700"/>
              <div v-if="isOtherOnline(c)" class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-white dark:border-surface-900"/>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <span class="font-semibold text-sm text-gray-900 dark:text-white truncate">{{ otherName(c) }}</span>
                <span class="text-xs text-gray-400 flex-shrink-0 ml-2">{{ formatTime(c.last_message?.created_at) }}</span>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
                {{ c.last_message?.content || 'Нет сообщений' }}
              </p>
            </div>
            <div v-if="c.unread_count" class="flex-shrink-0 w-5 h-5 bg-primary-600 rounded-full text-white text-[10px] font-bold flex items-center justify-center">
              {{ c.unread_count > 9 ? '9+' : c.unread_count }}
            </div>
          </button>
        </div>
      </aside>

      <!-- Chat Area -->
      <div class="flex-1 flex flex-col min-w-0">
        <!-- No chat selected -->
        <div v-if="!chatStore.activeChat" class="flex-1 flex items-center justify-center">
          <div class="text-center">
            <svg class="w-16 h-16 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">Выберите диалог</h3>
            <p class="text-gray-400 text-sm">Выберите чат слева, чтобы начать общение</p>
          </div>
        </div>

        <template v-else>
          <!-- Header -->
          <div class="flex items-center gap-3 px-5 py-3.5 border-b border-gray-200 dark:border-surface-800 bg-white dark:bg-surface-900 flex-shrink-0">
            <img :src="otherAvatar(chatStore.activeChat)" class="w-9 h-9 rounded-full object-cover"/>
            <div class="flex-1">
              <p class="font-semibold text-gray-900 dark:text-white text-sm">{{ otherName(chatStore.activeChat) }}</p>
              <p class="text-xs text-gray-400 flex items-center gap-1">
                <span v-if="isOtherOnline(chatStore.activeChat)" class="w-1.5 h-1.5 bg-green-500 rounded-full"/>
                <span v-if="isOtherTyping" class="text-primary-500 font-medium">Печатает…</span>
                <span v-else-if="isOtherOnline(chatStore.activeChat)" class="text-green-600 dark:text-green-400">В сети</span>
                <span v-else>Не в сети</span>
              </p>
            </div>
            <!-- Call button -->
            <button @click="startCall" class="p-2 btn-ghost rounded-lg text-green-500" title="Голосовой звонок">
              <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12 19.79 19.79 0 0 1 1.62 3.37 2 2 0 0 1 3.6 1.11h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L7.91 8.69a16 16 0 0 0 6 6l.91-.91a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
              </svg>
            </button>
          </div>

          <!-- Сообщения -->
          <div ref="messagesEl" class="flex-1 overflow-y-auto px-5 py-4 space-y-3 bg-gray-50/50 dark:bg-surface-950">
            <div v-for="msg in chatStore.messages" :key="msg.id"
              :class="['flex gap-3', isMyMsg(msg) ? 'flex-row-reverse' : '']">
              <img v-if="!isMyMsg(msg)" :src="msgAvatar(msg)" class="w-7 h-7 rounded-full object-cover flex-shrink-0 self-end"/>
              <div :class="['max-w-xs lg:max-w-md px-4 py-2.5 rounded-2xl text-sm shadow-sm', isMyMsg(msg) ? 'bg-primary-600 text-white rounded-tr-sm' : 'bg-white dark:bg-surface-800 text-gray-900 dark:text-gray-100 rounded-tl-sm']">
                <!-- File attachment -->
                <div v-if="msg.file_url" class="mb-2">
                  <a v-if="msg.file_type?.startsWith('image')" :href="msg.file_url" target="_blank">
                    <img :src="msg.file_url" class="rounded-xl max-w-full max-h-48 object-cover"/>
                  </a>
                  <a v-else :href="msg.file_url" target="_blank"
                    :class="['flex items-center gap-2 text-xs underline', isMyMsg(msg) ? 'text-white/80' : 'text-primary-600 dark:text-primary-400']">
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
                    </svg>
                    {{ msg.file_name || 'Attachment' }}
                  </a>
                </div>
                <p v-if="msg.content" class="leading-relaxed whitespace-pre-wrap">{{ msg.content }}</p>
                <p :class="['text-[10px] mt-1', isMyMsg(msg) ? 'text-white/50 text-right' : 'text-gray-400']">
                  {{ formatTimeFull(msg.created_at) }}
                  <span v-if="isMyMsg(msg)">{{ msg.is_read ? ' ✓✓' : ' ✓' }}</span>
                </p>
              </div>
            </div>
          </div>

          <!-- Input -->
          <div class="px-4 py-3 border-t border-gray-200 dark:border-surface-800 bg-white dark:bg-surface-900 flex-shrink-0">
            <!-- File preview -->
            <div v-if="pendingFile" class="flex items-center gap-2 px-3 py-2 bg-gray-50 dark:bg-surface-800 rounded-xl mb-2">
              <img v-if="pendingFile.type.startsWith('image')" :src="pendingFile.preview" class="w-10 h-10 rounded-lg object-cover"/>
              <svg v-else class="w-8 h-8 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              </svg>
              <div class="flex-1 min-w-0">
                <p class="text-xs font-medium text-gray-700 dark:text-gray-300 truncate">{{ pendingFile.name }}</p>
                <p class="text-xs text-gray-400">{{ formatBytes(pendingFile.size) }}</p>
              </div>
              <button @click="pendingFile = null" class="text-gray-400 hover:text-red-500 transition-colors">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>

            <div class="flex items-end gap-2">
              <!-- File upload -->
              <label class="flex-shrink-0 p-2 btn-ghost rounded-lg cursor-pointer" title="Прикрепить файл">
                <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/>
                </svg>
                <input type="file" class="hidden" accept="image/*,video/*,application/pdf,.doc,.docx,.zip" @change="onFileSelect"/>
              </label>

              <textarea v-model="messageInput"
                @keydown.enter.exact.prevent="send"
                @input="onTyping"
                rows="1"
                placeholder="Написать сообщение… (Enter — отправить)"
                class="input flex-1 resize-none py-2.5"
                style="min-height:42px;max-height:120px"/>

              <button @click="send"
                :disabled="(!messageInput.trim() && !pendingFile) || sending"
                class="flex-shrink-0 p-2.5 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl transition-colors">
                <svg v-if="!sending" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
                </svg>
                <svg v-else class="w-5 h-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>

  <!-- Incoming call modal -->
  <div v-if="incomingCall" class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
    <div class="bg-white dark:bg-surface-900 rounded-3xl shadow-2xl w-80 overflow-hidden">
      <div class="bg-gradient-to-b from-green-500 to-green-700 px-8 py-10 text-center">
        <div class="w-20 h-20 rounded-full bg-white/20 flex items-center justify-center mx-auto mb-4">
          <svg class="w-10 h-10 text-white animate-pulse" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12 19.79 19.79 0 0 1 1.62 3.37 2 2 0 0 1 3.6 1.11h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L7.91 8.69a16 16 0 0 0 6 6l.91-.91a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
          </svg>
        </div>
        <h3 class="text-white text-xl font-bold">Входящий звонок</h3>
        <p class="text-green-100 text-sm mt-1">{{ incomingCall?.username }}</p>
      </div>
      <div class="flex divide-x divide-gray-100 dark:divide-surface-800">
        <button @click="rejectCall" class="flex-1 py-4 text-red-500 font-semibold text-sm hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">Отклонить</button>
        <button @click="acceptCall" class="flex-1 py-4 text-green-600 font-semibold text-sm hover:bg-green-50 dark:hover:bg-green-900/20 transition-colors">Принять</button>
      </div>
    </div>
  </div>

  <AudioCallModal v-if="callTarget" :target="callTarget" @close="callTarget = null"/>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { format, isToday, formatDistanceToNow } from 'date-fns'
import AppNavbar from '@/components/common/AppNavbar.vue'
import AudioCallModal from '@/components/chat/AudioCallModal.vue'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'

const chatStore = useChatStore()
const authStore = useAuthStore()

const chatSearch   = ref('')
const messageInput = ref('')
const messagesEl   = ref<HTMLDivElement>()
const sending      = ref(false)
const pendingFile  = ref<{ file: File; name: string; size: number; type: string; preview?: string } | null>(null)
const callTarget   = ref<any>(null)
const incomingCall = ref<any>(null)

let typingTimer: ReturnType<typeof setTimeout>

const filteredChats = computed(() => {
  const q = chatSearch.value.toLowerCase()
  if (!q) return chatStore.chats
  return chatStore.chats.filter(c => otherName(c).toLowerCase().includes(q))
})

const isOtherTyping = computed(() => {
  if (!chatStore.activeChat) return false
  const other = getOtherParticipant(chatStore.activeChat)
  return other ? !!chatStore.typingUsers[other.user_id] : false
})

function getOtherParticipant(chat: any) {
  return chat.participants?.find((p: any) => p.user_id !== authStore.user?.id)
}
function otherAvatar(chat: any) {
  const p = getOtherParticipant(chat)?.user
  return p?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${p?.username || '?'}&size=44&background=2563eb&color=fff`
}
function otherName(chat: any) {
  const p = getOtherParticipant(chat)?.user
  if (!p) return chat.name || 'Chat'
  const prof = p.profile
  return (prof?.first_name ? `${prof.first_name} ${prof.last_name || ''}`.trim() : p.username) || p.username
}
function isOtherOnline(chat: any) {
  return getOtherParticipant(chat)?.user?.profile?.is_online === true
}
function isMyMsg(msg: any)  { return msg.sender_id === authStore.user?.id }
function msgAvatar(msg: any) {
  return msg.sender?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${msg.sender?.username}&size=28&background=2563eb&color=fff`
}
function formatTime(d?: string) {
  if (!d) return ''
  const dt = new Date(d)
  return isToday(dt) ? format(dt, 'HH:mm') : formatDistanceToNow(dt, { addSuffix: true })
}
function formatTimeFull(d: string) {
  return format(new Date(d), 'HH:mm')
}
function formatBytes(b: number) {
  if (b < 1024) return b + ' B'
  if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB'
  return (b / 1024 / 1024).toFixed(1) + ' MB'
}

async function openChat(chat: any) {
  await chatStore.openChat(chat)
  await nextTick()
  scrollBottom()
}

async function send() {
  const content = messageInput.value.trim()
  if (!content && !pendingFile.value) return
  if (!chatStore.activeChat) return
  sending.value = true
  try {
    if (pendingFile.value) {
      const fd = new FormData()
      fd.append('file', pendingFile.value.file)
      const { data: up } = await api.post('/api/v1/upload/file', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      chatStore.sendMessage(chatStore.activeChat.id, content, {
        url: up.url, name: up.name, type: up.type,
      })
      pendingFile.value = null
    } else {
      chatStore.sendMessage(chatStore.activeChat.id, content)
    }
    messageInput.value = ''
    await nextTick(); scrollBottom()
  } finally { sending.value = false }
}

function onTyping() {
  if (!chatStore.activeChat) return
  chatStore.sendTyping(chatStore.activeChat.id, true)
  clearTimeout(typingTimer)
  typingTimer = setTimeout(() => chatStore.sendTyping(chatStore.activeChat!.id, false), 2000)
}

function onFileSelect(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  const preview = file.type.startsWith('image') ? URL.createObjectURL(file) : undefined
  pendingFile.value = { file, name: file.name, size: file.size, type: file.type, preview }
}

function startCall() {
  if (!chatStore.activeChat) return
  callTarget.value = getOtherParticipant(chatStore.activeChat)?.user
}

function handleWSEvent(event: any) {
  if (event.type === 'rtc_call') {
    incomingCall.value = event.payload
  }
}
function acceptCall() {
  callTarget.value = incomingCall.value
  incomingCall.value = null
}
function rejectCall() {
  chatStore.sendRTC('rtc_hangup', incomingCall.value?.sender_id, {})
  incomingCall.value = null
}

function scrollBottom() {
  if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
}

watch(() => chatStore.messages.length, () => nextTick(scrollBottom))

onMounted(async () => {
  chatStore.connectWS()
  await chatStore.fetchChats()
  chatStore.onWSMessage(handleWSEvent)
})
onUnmounted(() => chatStore.offWSMessage(handleWSEvent))
</script>
