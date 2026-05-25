<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-surface-950">
    <div class="text-center">
      <div v-if="error" class="space-y-4">
        <div class="w-16 h-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto">
          <svg class="w-8 h-8 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Auth Failed</h2>
        <p class="text-gray-500 text-sm">{{ errorMsg }}</p>
        <RouterLink to="/login" class="btn-primary inline-flex">Try Again</RouterLink>
      </div>
      <div v-else class="space-y-4">
        <div class="w-16 h-16 bg-primary-100 dark:bg-primary-900/30 rounded-full flex items-center justify-center mx-auto">
          <svg class="w-8 h-8 text-primary-600 dark:text-primary-400 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
        </div>
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Signing you in…</h2>
        <p class="text-sm text-gray-500">Please wait</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { toast } from 'vue3-toastify'
import api from '@/api/axios'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()
const chat   = useChatStore()
const error    = ref(false)
const errorMsg = ref('')

onMounted(async () => {
  const errParam = route.query.error as string
  if (errParam === 'account_banned') {
    error.value = true; errorMsg.value = 'Your account has been suspended.'; return
  }

  // New secure flow: tokens stored as httpOnly cookies, fetch them
  const success = route.query.success === 'true'

  // Try query params first (legacy), then cookie-based endpoint
  const accessToken  = route.query.access_token as string
  const refreshToken = route.query.refresh_token as string

  if (accessToken && refreshToken) {
    localStorage.setItem('access_token',  accessToken)
    localStorage.setItem('refresh_token', refreshToken)
  } else if (success) {
    // Fetch tokens from secure cookie endpoint
    try {
      const { data } = await api.get('/api/v1/auth/oauth-tokens', { withCredentials: true })
      localStorage.setItem('access_token',  data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
    } catch {
      error.value = true; errorMsg.value = 'Could not retrieve tokens. Please try again.'; return
    }
  } else {
    error.value = true; errorMsg.value = 'No authentication data received.'; return
  }

  try {
    await auth.initAuth()
    chat.connectWS()
    toast.success(`Welcome, ${auth.fullName || auth.user?.username}!`)
    router.replace('/marketplace')
  } catch {
    error.value = true; errorMsg.value = 'Failed to load profile.'
  }
})
</script>
