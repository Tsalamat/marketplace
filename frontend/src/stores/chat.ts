import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/axios'
import { useAuthStore } from './auth'

export interface ChatMessage {
  id: string
  chat_id: string
  sender_id: string
  content: string
  file_url?: string
  file_name?: string
  file_type?: string
  is_read: boolean
  created_at: string
  sender?: any
}

export interface Chat {
  id: string
  order_id?: string
  is_group: boolean
  name?: string
  participants: any[]
  last_message?: ChatMessage
  unread_count: number
  updated_at: string
}

export const useChatStore = defineStore('chat', () => {
  const chats = ref<Chat[]>([])
  const activeChat = ref<Chat | null>(null)
  const messages = ref<ChatMessage[]>([])
  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const typingUsers = ref<Record<string, boolean>>({})

  const totalUnread = computed(() =>
    chats.value.reduce((sum, c) => sum + c.unread_count, 0)
  )

  function connectWS() {
    const token = localStorage.getItem('access_token')
    if (!token || ws.value?.readyState === WebSocket.OPEN) return

    const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'
    const socket = new WebSocket(`${wsUrl}/ws/chat?token=${token}`)

    socket.onopen = () => {
      isConnected.value = true
    }

    socket.onmessage = (e) => {
      const event = JSON.parse(e.data)
      handleWSEvent(event)
    }

    socket.onclose = () => {
      isConnected.value = false
      // Reconnect after 3s
      setTimeout(connectWS, 3000)
    }

    socket.onerror = () => socket.close()

    ws.value = socket
  }

  const wsListeners = new Set<(e: any) => void>()

  function onWSMessage(fn: (e: any) => void) { wsListeners.add(fn) }
  function offWSMessage(fn: (e: any) => void) { wsListeners.delete(fn) }

  function handleWSEvent(event: any) {
    // Notify external listeners (map, calls, etc.)
    wsListeners.forEach(fn => fn(event))

    switch (event.type) {
      case 'message':
        onNewMessage(event.chat_id, event.payload)
        break
      case 'typing':
        typingUsers.value[event.payload.user_id] = event.payload.is_typing
        setTimeout(() => { delete typingUsers.value[event.payload.user_id] }, 3000)
        break
      case 'online':
      case 'offline':
        break
    }
  }

  function onNewMessage(chatId: string, message: ChatMessage) {
    if (activeChat.value?.id === chatId) {
      messages.value.push(message)
    }
    // Update last message & unread in chat list
    const chat = chats.value.find(c => c.id === chatId)
    if (chat) {
      chat.last_message = message
      const auth = useAuthStore()
      if (activeChat.value?.id !== chatId && message.sender_id !== auth.user?.id) {
        chat.unread_count++
      }
    }
  }

  function sendMessage(chatId: string, content: string, file?: { url: string; name: string; type: string }) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return

    ws.value.send(JSON.stringify({
      type: 'message',
      payload: { chat_id: chatId, content, ...file },
    }))
  }

  function sendTyping(chatId: string, isTyping: boolean) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return
    ws.value.send(JSON.stringify({
      type: 'typing',
      payload: { chat_id: chatId, is_typing: isTyping },
    }))
  }

  async function fetchChats() {
    const { data } = await api.get('/api/v1/chat')
    chats.value = data
    return data
  }

  async function fetchMessages(chatId: string, page = 1) {
    const { data } = await api.get(`/api/v1/chat/${chatId}/messages`, { params: { page, limit: 50 } })
    if (page === 1) messages.value = data
    else messages.value = [...data, ...messages.value]
    return data
  }

  async function openChat(chat: Chat) {
    activeChat.value = chat
    chat.unread_count = 0
    await fetchMessages(chat.id)

    // Mark as read
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'read', payload: { chat_id: chat.id } }))
    }
  }

  async function getOrCreateDirect(userId: string) {
    const { data } = await api.post(`/api/v1/chat/direct/${userId}`)
    const existing = chats.value.find(c => c.id === data.id)
    if (!existing) chats.value.unshift(data)
    return data
  }

  function sendLocation(lat: number, lng: number) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return
    ws.value.send(JSON.stringify({ type: 'location', payload: { lat, lng } }))
  }

  function sendRTC(type: string, targetId: string, payload: any) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return
    ws.value.send(JSON.stringify({ type, payload: { ...payload, target_id: targetId } }))
  }

  function disconnectWS() {
    ws.value?.close()
    ws.value = null
  }

  return {
    chats, activeChat, messages, isConnected, typingUsers, totalUnread,
    connectWS, disconnectWS, sendMessage, sendTyping, sendLocation, sendRTC,
    onWSMessage, offWSMessage,
    fetchChats, fetchMessages, openChat, getOrCreateDirect,
  }
})
