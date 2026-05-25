<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-2xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-6 flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Сообщество</h1>
        <button v-if="auth.isAuthenticated" @click="showCreate = !showCreate" class="btn-primary">
          + Написать
        </button>
      </div>

      <!-- Создать пост -->
      <div v-if="showCreate && auth.isAuthenticated" class="card p-5 mb-6 animate-fade-in">
        <div class="flex items-start gap-3">
          <img :src="myAvatar" class="w-10 h-10 rounded-full object-cover flex-shrink-0" />
          <div class="flex-1">
            <textarea v-model="newPost" rows="3" class="input resize-none w-full"
              placeholder="Что у вас нового?" />
            <div class="flex justify-end gap-2 mt-3">
              <button @click="showCreate = false; newPost = ''" class="btn-secondary text-sm">Отмена</button>
              <button @click="createPost" :disabled="!newPost.trim() || posting" class="btn-primary text-sm">
                {{ posting ? 'Публикация...' : 'Опубликовать' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Лента -->
      <div v-if="loading" class="space-y-4">
        <div v-for="i in 3" :key="i" class="card p-5 animate-pulse">
          <div class="flex gap-3 mb-4">
            <div class="w-10 h-10 bg-gray-200 dark:bg-surface-800 rounded-full"/>
            <div class="flex-1"><div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/3 mb-2"/><div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/4"/></div>
          </div>
          <div class="h-4 bg-gray-200 dark:bg-surface-800 rounded mb-2"/>
          <div class="h-4 bg-gray-200 dark:bg-surface-800 rounded w-3/4"/>
        </div>
      </div>

      <div v-else class="space-y-4">
        <div v-for="post in posts" :key="post.id" class="card overflow-hidden">
          <div class="p-5">
            <!-- Автор -->
            <div class="flex items-center gap-3 mb-4">
              <img :src="postAvatar(post)" class="w-10 h-10 rounded-full object-cover" />
              <div class="flex-1">
                <RouterLink :to="`/profile/${post.author?.username}`"
                  class="font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400 text-sm">
                  {{ post.author?.profile?.first_name || post.author?.username }}
                </RouterLink>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ post.author?.profile?.university }} · {{ timeAgo(post.created_at) }}
                </p>
              </div>
              <button v-if="auth.user?.id === post.author_id || auth.isAdmin"
                @click="deletePost(post)" class="p-1.5 btn-ghost rounded-lg text-red-400 hover:text-red-500">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/>
                  <path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/>
                </svg>
              </button>
            </div>

            <!-- Текст -->
            <p class="text-gray-800 dark:text-gray-200 leading-relaxed mb-4 whitespace-pre-wrap">{{ post.content }}</p>

            <!-- Картинки -->
            <div v-if="post.images?.length" class="mb-4 rounded-xl overflow-hidden">
              <img :src="post.images[0]" class="w-full object-cover max-h-64" />
            </div>

            <!-- Действия -->
            <div class="flex items-center gap-5 pt-3 border-t border-gray-100 dark:border-surface-800">
              <button @click="toggleLike(post)"
                :class="['flex items-center gap-1.5 text-sm transition-colors', post.is_liked ? 'text-red-500' : 'text-gray-500 dark:text-gray-400 hover:text-red-500']">
                <svg class="w-4 h-4" viewBox="0 0 24 24" :fill="post.is_liked ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2">
                  <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
                </svg>
                {{ post.likes_count || 0 }}
              </button>

              <button @click="toggleComments(post)"
                class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                </svg>
                {{ post.comments_count || 0 }} {{ post.comments_count === 1 ? 'комментарий' : 'комментариев' }}
              </button>
            </div>
          </div>

          <!-- Комментарии -->
          <div v-if="post._showComments" class="border-t border-gray-100 dark:border-surface-800 bg-gray-50/50 dark:bg-surface-950/50">
            <div v-if="post._loadingComments" class="p-4 text-center">
              <div class="w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto"/>
            </div>
            <div v-else>
              <!-- Список комментариев -->
              <div v-if="!post._comments?.length" class="px-5 py-4 text-sm text-gray-400 text-center">
                Комментариев пока нет. Будьте первым!
              </div>
              <div v-else class="divide-y divide-gray-100 dark:divide-surface-800">
                <div v-for="c in post._comments" :key="c.id" class="px-5 py-3 flex gap-3">
                  <img :src="commentAvatar(c)" class="w-7 h-7 rounded-full object-cover flex-shrink-0 mt-0.5"/>
                  <div>
                    <div class="flex items-baseline gap-2">
                      <span class="text-xs font-semibold text-gray-900 dark:text-white">{{ c.author?.username }}</span>
                      <span class="text-xs text-gray-400">{{ timeAgo(c.created_at) }}</span>
                    </div>
                    <p class="text-sm text-gray-700 dark:text-gray-300 mt-0.5 leading-relaxed">{{ c.content }}</p>
                  </div>
                </div>
              </div>

              <!-- Поле ввода комментария -->
              <div v-if="auth.isAuthenticated" class="px-4 py-3 flex gap-2 border-t border-gray-100 dark:border-surface-800">
                <img :src="myAvatar" class="w-7 h-7 rounded-full object-cover flex-shrink-0 mt-1"/>
                <div class="flex-1 flex gap-2">
                  <input v-model="post._newComment" @keyup.enter="submitComment(post)"
                    class="input flex-1 text-sm py-2" placeholder="Написать комментарий…" />
                  <button @click="submitComment(post)"
                    :disabled="!post._newComment?.trim() || post._submittingComment"
                    class="p-2 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white rounded-xl transition-colors">
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="!posts.length && !loading" class="text-center py-16">
          <svg class="w-16 h-16 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
          </svg>
          <p class="text-gray-500 dark:text-gray-400">Постов пока нет. Будьте первым!</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'
import { formatDistanceToNow } from 'date-fns'
import { ru } from 'date-fns/locale'
import { toast } from 'vue3-toastify'

const auth    = useAuthStore()
const posts   = ref<any[]>([])
const loading = ref(false)
const posting = ref(false)
const showCreate = ref(false)
const newPost = ref('')

const myAvatar = computed(() =>
  auth.user?.profile?.avatar_url ||
  `https://ui-avatars.com/api/?name=${auth.user?.username}&size=40&background=3b82f6&color=fff`
)

function postAvatar(p: any) {
  return p.author?.profile?.avatar_url ||
    `https://ui-avatars.com/api/?name=${p.author?.username}&size=40&background=3b82f6&color=fff`
}

function commentAvatar(c: any) {
  return c.author?.profile?.avatar_url ||
    `https://ui-avatars.com/api/?name=${c.author?.username}&size=28&background=6366f1&color=fff`
}

function timeAgo(d: string) {
  try { return formatDistanceToNow(new Date(d), { addSuffix: true, locale: ru }) } catch { return d }
}

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/api/v1/posts?limit=20')
    posts.value = (data || []).map((p: any) => ({
      ...p,
      _showComments: false,
      _loadingComments: false,
      _comments: [],
      _newComment: '',
      _submittingComment: false,
    }))
  } finally {
    loading.value = false
  }
}

async function createPost() {
  if (!newPost.value.trim()) return
  posting.value = true
  try {
    const { data } = await api.post('/api/v1/posts', { content: newPost.value })
    posts.value.unshift({ ...data, _showComments: false, _loadingComments: false, _comments: [], _newComment: '', _submittingComment: false })
    newPost.value = ''
    showCreate.value = false
    toast.success('Пост опубликован!')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Ошибка публикации')
  } finally {
    posting.value = false
  }
}

async function deletePost(post: any) {
  if (!confirm('Удалить пост?')) return
  try {
    await api.delete(`/api/v1/posts/${post.id}`)
    posts.value = posts.value.filter(p => p.id !== post.id)
    toast.success('Пост удалён')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Ошибка удаления')
  }
}

async function toggleLike(post: any) {
  if (!auth.isAuthenticated) { toast.warning('Войдите, чтобы ставить лайки'); return }
  try {
    const { data } = await api.post(`/api/v1/posts/${post.id}/like`)
    post.is_liked = data.liked
    post.likes_count = (post.likes_count || 0) + (data.liked ? 1 : -1)
  } catch {}
}

async function toggleComments(post: any) {
  post._showComments = !post._showComments
  if (post._showComments && !post._comments.length) {
    post._loadingComments = true
    try {
      const { data } = await api.get(`/api/v1/posts/${post.id}/comments`)
      post._comments = data || []
    } catch {
      post._comments = []
    } finally {
      post._loadingComments = false
    }
  }
}

async function submitComment(post: any) {
  const text = post._newComment?.trim()
  if (!text || post._submittingComment) return
  post._submittingComment = true
  try {
    const { data } = await api.post(`/api/v1/posts/${post.id}/comments`, { content: text })
    post._comments.push(data)
    post.comments_count = (post.comments_count || 0) + 1
    post._newComment = ''
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Ошибка отправки')
  } finally {
    post._submittingComment = false
  }
}

onMounted(load)
</script>
