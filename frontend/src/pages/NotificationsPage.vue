<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-2xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-8 flex items-center justify-between">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Notifications</h1>
        <button v-if="notifs.length" @click="markAllRead" class="btn-ghost text-sm">Mark all read</button>
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="i in 5" :key="i" class="card p-4 animate-pulse flex gap-4">
          <div class="w-10 h-10 bg-gray-200 dark:bg-surface-800 rounded-full flex-shrink-0"/>
          <div class="flex-1"><div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-3/4 mb-2"/><div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/2"/></div>
        </div>
      </div>

      <div v-else-if="notifs.length === 0" class="text-center py-24">
        <svg class="w-14 h-14 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">All caught up!</h3>
        <p class="text-gray-500 dark:text-gray-400">No notifications right now</p>
      </div>

      <div v-else class="space-y-2">
        <div v-for="n in notifs" :key="n.id"
          :class="['card p-4 flex items-start gap-4 cursor-pointer hover:shadow-card-hover transition-all', !n.is_read ? 'border-l-4 border-primary-600' : '']"
          @click="markRead(n)">
          <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0', typeColor(n.type)]">
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path :d="typeIcon(n.type)"/>
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-gray-900 dark:text-white text-sm">{{ n.title }}</p>
            <p class="text-gray-600 dark:text-gray-400 text-sm mt-0.5">{{ n.body }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ timeAgo(n.created_at) }}</p>
          </div>
          <div v-if="!n.is_read" class="w-2 h-2 rounded-full bg-primary-600 mt-1.5 flex-shrink-0"/>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { formatDistanceToNow } from 'date-fns'
import AppNavbar from '@/components/common/AppNavbar.vue'
import api from '@/api/axios'

const notifs  = ref<any[]>([])
const loading = ref(false)

const icons: Record<string, string> = {
  order:          'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
  message:        'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
  comment:        'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
  review:         'M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z',
  follow:         'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z',
  like:           'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z',
  friend_request: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8zM23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75',
  system:         'M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0',
}
const colors: Record<string, string> = {
  order:          'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400',
  message:        'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400',
  comment:        'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400',
  review:         'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400',
  follow:         'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400',
  like:           'bg-red-100 dark:bg-red-900/30 text-red-500 dark:text-red-400',
  friend_request: 'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400',
  system:         'bg-gray-100 dark:bg-surface-800 text-gray-500 dark:text-gray-400',
}

function typeIcon(t: string) { return icons[t] || icons.system }
function typeColor(t: string) { return colors[t] || colors.system }
function timeAgo(d: string) {
  try { return formatDistanceToNow(new Date(d), { addSuffix: true }) } catch { return d }
}

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/api/v1/notifications')
    // Backend returns { data: [], unread_count: N }
    notifs.value = Array.isArray(data) ? data : (data.data ?? [])
  } finally {
    loading.value = false
  }
}

async function markRead(n: any) {
  if (n.is_read) return
  try {
    await api.patch(`/api/v1/notifications/${n.id}/read`)
    n.is_read = true
  } catch {}
}

async function markAllRead() {
  try {
    await api.post('/api/v1/notifications/read-all')
    notifs.value.forEach(n => n.is_read = true)
  } catch {}
}

onMounted(load)
</script>
