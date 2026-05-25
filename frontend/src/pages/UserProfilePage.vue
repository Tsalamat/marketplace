<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div v-if="loading" class="max-w-4xl mx-auto px-4 pt-28 text-center">
      <div class="w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
    </div>
    <div v-else-if="user" class="max-w-4xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <!-- Cover -->
      <div class="relative h-48 rounded-2xl overflow-hidden mt-6 mb-16 bg-gradient-to-r from-primary-500 to-blue-600">
        <img v-if="user.profile?.cover_url"  class="w-full h-full object-cover" />
        <div class="absolute bottom-0 left-6 translate-y-1/2">
          <img :src="avatar" class="w-24 h-24 rounded-2xl object-cover border-4 border-white dark:border-surface-950 shadow-lg" />
        </div>
      </div>

      <!-- Info -->
      <div class="flex items-start justify-between mb-6 flex-wrap gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ fullName || user.username }}</h1>
          <p class="text-gray-500 dark:text-gray-400">@{{ user.username }}</p>
          <p v-if="user.profile?.tagline" class="text-primary-600 dark:text-primary-400 font-medium mt-1">{{ user.profile.tagline }}</p>
        </div>
        <div v-if="auth.isAuthenticated && auth.user?.username !== user.username" class="flex gap-2 flex-wrap">
          <button @click="follow" class="btn-secondary text-sm">
            {{ isFollowing ? 'Отписаться' : 'Подписаться' }}
          </button>
          <button @click="addFriend" :disabled="friendStatus === 'pending' || friendStatus === 'friends'" class="btn-secondary text-sm flex items-center gap-1.5">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <line x1="19" y1="8" x2="19" y2="14"/>
              <line x1="22" y1="11" x2="16" y2="11"/>
            </svg>
            {{ friendLabel }}
          </button>
          <button @click="message" class="btn-primary text-sm flex items-center gap-1.5">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            Message
          </button>
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-3 gap-4 mb-6">
        <div class="card p-4 text-center">
          <div class="text-xl font-bold text-gray-900 dark:text-white">{{ user.profile?.completed_jobs || 0 }}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400">Выполнено</div>
        </div>
        <div class="card p-4 text-center">
          <div class="text-xl font-bold text-gray-900 dark:text-white flex items-center justify-center gap-1">
            <svg class="w-4 h-4 text-yellow-400" viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
            {{ (user.profile?.rating || 0).toFixed(1) }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">Rating ({{ user.profile?.total_reviews || 0 }})</div>
        </div>
        <div class="card p-4 text-center">
          <div class="text-xl font-bold text-gray-900 dark:text-white">{{ services.length }}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400">Услуги</div>
        </div>
      </div>

      <!-- Bio & Info -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-1 space-y-4">
          <div class="card p-5 space-y-3">
            <div v-if="user.profile?.university" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/></svg>
              <span>{{ user.profile.university }}</span>
            </div>
            <div v-if="user.profile?.department" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
              <span>{{ user.profile.department }}</span>
            </div>
            <div v-if="user.profile?.location" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
              <span>{{ user.profile.location }}</span>
            </div>
            <div v-if="user.profile?.languages?.length" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
              <span>{{ user.profile.languages.join(', ') }}</span>
            </div>
            <div v-if="user.profile?.github_url" class="flex items-center gap-2 text-sm">
              <svg class="w-4 h-4 flex-shrink-0 text-gray-600 dark:text-gray-400" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              <a :href="user.profile.github_url" target="_blank" class="text-primary-600 dark:text-primary-400 hover:underline truncate">GitHub</a>
            </div>
            <div v-if="user.profile?.linkedin_url" class="flex items-center gap-2 text-sm">
              <svg class="w-4 h-4 flex-shrink-0 text-blue-600" viewBox="0 0 24 24" fill="currentColor"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6zM2 9h4v12H2z"/><circle cx="4" cy="4" r="2"/></svg>
              <a :href="user.profile.linkedin_url" target="_blank" class="text-primary-600 dark:text-primary-400 hover:underline truncate">LinkedIn</a>
            </div>
          </div>

          <div v-if="user.profile?.skills?.length" class="card p-5">
            <h3 class="font-semibold text-gray-900 dark:text-white mb-3 text-sm">Навыки</h3>
            <div class="flex flex-wrap gap-2">
              <span v-for="s in user.profile.skills" :key="s" class="badge-gray">{{ s }}</span>
            </div>
          </div>
        </div>

        <div class="lg:col-span-2 space-y-4">
          <div v-if="user.profile?.bio" class="card p-5">
            <h3 class="font-semibold text-gray-900 dark:text-white mb-2">О себе</h3>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed">{{ user.profile.bio }}</p>
          </div>

          <div class="card p-5">
            <h3 class="font-semibold text-gray-900 dark:text-white mb-4">Услуги</h3>
            <div v-if="services.length === 0" class="text-center py-6 text-gray-400 text-sm">Нет услуг</div>
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <RouterLink v-for="s in services" :key="s.id" :to="`/gig/${s.slug}`"
                class="border border-gray-100 dark:border-surface-800 rounded-xl overflow-hidden hover:shadow-card-hover transition-shadow">
                <div v-if="s.gallery?.length" class="aspect-video">
                  <img :src="s.gallery[0]" class="w-full h-full object-cover" />
                </div>
                <div class="p-3">
                  <h4 class="font-medium text-gray-900 dark:text-white text-sm mb-2 line-clamp-2">{{ s.title }}</h4>
                  <div class="flex items-center justify-between">
                    <span class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1">
                      <svg class="w-3 h-3 text-yellow-400" viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                      {{ s.rating.toFixed(1) }} ({{ s.total_reviews }})
                    </span>
                    <span class="text-sm font-bold text-primary-600 dark:text-primary-400">From ${{ minPrice(s) }}</span>
                  </div>
                </div>
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="max-w-xl mx-auto px-4 pt-32 text-center">
      <svg class="w-16 h-16 text-gray-300 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Пользователь не найден</h2>
      <RouterLink to="/marketplace" class="btn-primary">На маркетплейс</RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()
const chat   = useChatStore()

const user         = ref<any>(null)
const services     = ref<any[]>([])
const loading      = ref(false)
const isFollowing  = ref(false)
const friendStatus = ref<'none' | 'pending' | 'friends'>('none')

const fullName = computed(() => {
  const p = user.value?.profile
  return p ? `${p.first_name || ''} ${p.last_name || ''}`.trim() : ''
})

const avatar = computed(() =>
  user.value?.profile?.avatar_url ||
  `https://ui-avatars.com/api/?name=${user.value?.username}&size=96&background=3b82f6&color=fff`
)

const friendLabel = computed(() => {
  if (friendStatus.value === 'friends') return 'Friends'
  if (friendStatus.value === 'pending') return 'Request Sent'
  return 'Add Friend'
})

function minPrice(s: any) {
  if (!s.packages?.length) return 0
  return Math.min(...s.packages.map((p: any) => p.price))
}

async function follow() {
  if (!auth.isAuthenticated) { router.push('/login'); return }
  try {
    const { data } = await api.post(`/api/v1/social/follow/${user.value.id}`)
    isFollowing.value = data.following
  } catch {}
}

async function addFriend() {
  if (!auth.isAuthenticated) { router.push('/login'); return }
  if (friendStatus.value !== 'none') return
  try {
    await api.post(`/api/v1/friends/${user.value.id}`)
    friendStatus.value = 'pending'
    toast.success('Заявка в друзья отправлена!')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to send request')
  }
}

async function message() {
  if (!auth.isAuthenticated) { router.push('/login'); return }
  const c = await chat.getOrCreateDirect(user.value.id)
  await chat.openChat(c)
  router.push('/chat')
}

async function checkFriendStatus() {
  if (!auth.isAuthenticated || !user.value) return
  try {
    const [friendsRes, requestsRes] = await Promise.all([
      api.get('/api/v1/friends'),
      api.get('/api/v1/friends/requests'),
    ])
    const friends = friendsRes.data as any[]
    const requests = requestsRes.data as any[]
    if (friends.find((f: any) => f.user?.id === user.value.id)) {
      friendStatus.value = 'friends'
    } else if (requests.find((r: any) => r.requester_id === user.value.id || r.addressee_id === user.value.id)) {
      friendStatus.value = 'pending'
    }
  } catch {}
}

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await api.get(`/api/v1/users/${route.params.username}`)
    user.value   = data
    services.value = data.services || []
    await checkFriendStatus()
  } catch {
    user.value = null
  } finally {
    loading.value = false
  }
})
</script>
