<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 to-blue-50 dark:from-surface-950 dark:to-surface-900 flex items-center justify-center px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <RouterLink to="/" class="inline-flex items-center gap-2">
          <div class="w-10 h-10 bg-primary-600 rounded-xl flex items-center justify-center text-white font-bold">SM</div>
        </RouterLink>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white mt-6">Forgot Password</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Enter your email to receive a reset link</p>
      </div>
      <div class="card p-8">
        <div v-if="sent" class="text-center py-4">
          <svg class="w-12 h-12 text-primary-600 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>
          </svg>
          <p class="text-gray-700 dark:text-gray-300">If that email exists, a reset link has been sent.</p>
          <RouterLink to="/login" class="btn-primary mt-6 inline-flex">Back to Login</RouterLink>
        </div>
        <form v-else @submit.prevent="submit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Email</label>
            <input v-model="email" type="email" required class="input" placeholder="you@university.edu" />
          </div>
          <button type="submit" :disabled="loading" class="btn-primary w-full py-3">
            <svg v-if="loading" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
            <span>{{ loading ? 'Sending...' : 'Send Reset Link' }}</span>
          </button>
          <RouterLink to="/login" class="block text-center text-sm text-gray-500 hover:underline">Back to Login</RouterLink>
        </form>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api/axios'
const email = ref(''); const loading = ref(false); const sent = ref(false)
async function submit() {
  loading.value = true
  try { await api.post('/api/v1/auth/forgot-password', { email: email.value }); sent.value = true }
  finally { loading.value = false }
}
</script>
