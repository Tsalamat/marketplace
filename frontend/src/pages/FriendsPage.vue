<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-4xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-8 flex items-center justify-between">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Друзья</h1>
        <RouterLink to="/map" class="btn-secondary flex items-center gap-2 text-sm">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
          </svg>
          Карта друзей
        </RouterLink>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 bg-gray-100 dark:bg-surface-800 p-1 rounded-xl mb-6 w-fit">
        <button v-for="t in tabs" :key="t.id" @click="activeTab = t.id"
          :class="['px-5 py-2 rounded-lg text-sm font-medium transition-all', activeTab === t.id ? 'bg-white dark:bg-surface-900 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300']">
          {{ t.label }}
          <span v-if="t.id === 'requests' && requests.length" class="ml-1.5 inline-flex items-center justify-center w-5 h-5 text-xs bg-primary-600 text-white rounded-full">{{ requests.length }}</span>
        </button>
      </div>

      <!-- Friends list -->
      <div v-if="activeTab === 'friends'">
        <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div v-for="i in 4" :key="i" class="card p-4 animate-pulse flex items-center gap-4">
            <div class="w-14 h-14 bg-gray-200 dark:bg-surface-800 rounded-full"/>
            <div class="flex-1 space-y-2">
              <div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/3"/>
              <div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/2"/>
            </div>
          </div>
        </div>
        <div v-else-if="!friends.length" class="text-center py-16">
          <svg class="w-16 h-16 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">Нет друзей</h3>
          <p class="text-gray-500 dark:text-gray-400 text-sm">Просматривайте профили и добавляйте друзей!</p>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div v-for="f in friends" :key="f.friendship_id" class="card p-4 flex items-center gap-4 group">
            <div class="relative flex-shrink-0">
              <img :src="avatar(f.user)" class="w-14 h-14 rounded-2xl object-cover"/>
              <div v-if="f.user?.profile?.is_online"
                class="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 rounded-full border-2 border-white dark:border-surface-900"/>
            </div>
            <div class="flex-1 min-w-0">
              <RouterLink :to="`/profile/${f.user?.username}`"
                class="font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400 block truncate">
                {{ f.user?.profile?.first_name }} {{ f.user?.profile?.last_name || f.user?.username }}
              </RouterLink>
              <p class="text-sm text-gray-500 dark:text-gray-400 truncate">{{ f.user?.profile?.university || '@' + f.user?.username }}</p>
              <div class="flex items-center gap-2 mt-1">
                <span v-if="f.user?.profile?.is_online" class="text-xs text-green-600 dark:text-green-400 font-medium">● В сети</span>
                <span v-else class="text-xs text-gray-400">Офлайн</span>
              </div>
            </div>
            <div class="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button @click="startChat(f.user)" class="p-2 btn-ghost rounded-lg" title="Написать">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                </svg>
              </button>
              <button @click="callFriend(f.user)" class="p-2 btn-ghost rounded-lg" title="Аудиозвонок">
                <svg class="w-4 h-4 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12a19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 3.6 1.11h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L7.91 8.69a16 16 0 0 0 6 6l.91-.91a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
                </svg>
              </button>
              <button @click="removeFriend(f)" class="p-2 btn-ghost rounded-lg text-red-500" title="Удалить из друзей">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/>
                  <line x1="23" y1="11" x2="17" y2="11"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Incoming requests -->
      <div v-if="activeTab === 'requests'">
        <div v-if="!requests.length" class="text-center py-16">
          <svg class="w-16 h-16 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/>
          </svg>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">Нет входящих заявок</h3>
        </div>
        <div v-else class="space-y-3">
          <div v-for="r in requests" :key="r.id" class="card p-4 flex items-center gap-4">
            <img :src="avatar(r.requester)" class="w-12 h-12 rounded-2xl object-cover flex-shrink-0"/>
            <div class="flex-1 min-w-0">
              <RouterLink :to="`/profile/${r.requester?.username}`"
                class="font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
                {{ r.requester?.profile?.first_name }} {{ r.requester?.profile?.last_name || r.requester?.username }}
              </RouterLink>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ r.requester?.profile?.university }}</p>
            </div>
            <div class="flex gap-2">
              <button @click="accept(r)" class="btn-primary text-sm px-4 py-2">Принять</button>
              <button @click="reject(r)" class="btn-secondary text-sm px-4 py-2">Отклонить</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Аудиозвонок Modal -->
  <AudioCallModal v-if="callTarget" :target="callTarget" @close="callTarget = null"/>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import AudioCallModal from '@/components/chat/AudioCallModal.vue'
import { useChatStore } from '@/stores/chat'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const router    = useRouter()
const chatStore = useChatStore()
type FriendTab = 'friends' | 'requests'
const activeTab = ref<FriendTab>('friends')
const loading   = ref(false)
const friends   = ref<any[]>([])
const requests  = ref<any[]>([])
const callTarget = ref<any>(null)

const tabs: Array<{ id: FriendTab; label: string }> = [
  { id: 'friends',  label: 'Друзья' },
  { id: 'requests', label: 'Заявки' },
]

function avatar(user: any) {
  return user?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${user?.username}&background=2563eb&color=fff&size=56`
}

async function load() {
  loading.value = true
  try {
    const [fl, rl] = await Promise.all([
      api.get('/api/v1/friends'),
      api.get('/api/v1/friends/requests'),
    ])
    friends.value  = fl.data
    requests.value = rl.data
  } finally { loading.value = false }
}

async function accept(r: any) {
  await api.put(`/api/v1/friends/${r.id}/accept`)
  toast.success('Заявка принята!')
  load()
}

async function reject(r: any) {
  await api.put(`/api/v1/friends/${r.id}/reject`)
  requests.value = requests.value.filter(x => x.id !== r.id)
}

async function removeFriend(f: any) {
  if (!confirm('Удалить из друзей?')) return
  await api.delete(`/api/v1/friends/${f.user?.id}`)
  friends.value = friends.value.filter(x => x.friendship_id !== f.friendship_id)
  toast.success('Друг удалён')
}

async function startChat(user: any) {
  const chat = await chatStore.getOrCreateDirect(user.id)
  await chatStore.openChat(chat)
  router.push('/chat')
}

function callFriend(user: any) {
  callTarget.value = user
}

onMounted(load)
</script>
