<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 to-blue-50 dark:from-surface-950 dark:to-surface-900 flex items-center justify-center px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Set New Password</h1>
      </div>
      <div class="card p-8">
        <div v-if="done" class="text-center">
          <div class="w-14 h-14 rounded-2xl bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto mb-4">
            <svg class="w-7 h-7 text-green-600 dark:text-green-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          </div>
          <p class="text-gray-700 dark:text-gray-300 mb-4">Password reset successfully!</p>
          <RouterLink to="/login" class="btn-primary">Login</RouterLink>
        </div>
        <form v-else @submit.prevent="submit" class="space-y-4">
          <div v-if="error" class="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl text-sm text-red-600 dark:text-red-400">{{ error }}</div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">New Password</label>
            <input v-model="password" type="password" required minlength="8" class="input" placeholder="Min. 8 characters" />
          </div>
          <button type="submit" :disabled="loading" class="btn-primary w-full py-3">
            <svg v-if="loading" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
            <span>{{ loading ? 'Resetting...' : 'Reset Password' }}</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/axios'
const route = useRoute()
const password = ref(''); const loading = ref(false); const done = ref(false); const error = ref('')
async function submit() {
  loading.value = true; error.value = ''
  try { await api.post('/api/v1/auth/reset-password', { token: route.query.token, password: password.value }); done.value = true }
  catch (e: any) { error.value = e.response?.data?.error || 'Failed' }
  finally { loading.value = false }
}
</script>
