<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="text-center">
      <div class="w-16 h-16 mx-auto mb-6 rounded-2xl flex items-center justify-center"
        :class="status === 'ok' ? 'bg-green-100 dark:bg-green-900/30' : status === 'err' ? 'bg-red-100 dark:bg-red-900/30' : 'bg-primary-100 dark:bg-primary-900/30'">
        <svg v-if="status === 'ok'" class="w-8 h-8 text-green-600 dark:text-green-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
        <svg v-else-if="status === 'err'" class="w-8 h-8 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        <svg v-else class="w-8 h-8 text-primary-600 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
        {{ status === 'ok' ? 'Email Verified!' : status === 'err' ? 'Invalid Token' : 'Verifying...' }}
      </h1>
      <RouterLink v-if="status !== 'loading'" to="/login" class="btn-primary mt-6 inline-flex">Go to Login</RouterLink>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/axios'
const route = useRoute()
const status = ref<'loading' | 'ok' | 'err'>('loading')
onMounted(async () => {
  try { await api.get(`/api/v1/auth/verify-email?token=${route.query.token}`); status.value = 'ok' }
  catch { status.value = 'err' }
})
</script>
