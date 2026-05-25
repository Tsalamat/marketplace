<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-4xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-8 flex items-center justify-between">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">My Orders</h1>
        <div class="flex gap-2">
          <button @click="role = 'buyer'" :class="role === 'buyer' ? 'btn-primary' : 'btn-secondary'" class="text-sm">As Buyer</button>
          <button @click="role = 'seller'" :class="role === 'seller' ? 'btn-primary' : 'btn-secondary'" class="text-sm">As Seller</button>
        </div>
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="i in 4" :key="i" class="card p-5 animate-pulse">
          <div class="h-4 bg-gray-200 dark:bg-surface-800 rounded w-1/2 mb-3"/>
          <div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/4"/>
        </div>
      </div>

      <div v-else-if="orders.length === 0" class="text-center py-24">
        <svg class="w-14 h-14 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12"/>
        </svg>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">No orders yet</h3>
        <RouterLink to="/marketplace" class="btn-primary">Browse Services</RouterLink>
      </div>

      <div v-else class="space-y-4">
        <RouterLink v-for="o in orders" :key="o.id" :to="`/orders/${o.id}`"
          class="card p-5 flex items-center gap-4 hover:shadow-card-hover transition-shadow block">
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-gray-900 dark:text-white mb-1 truncate">{{ o.service?.title }}</h3>
            <div class="flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
              <span>${{ o.amount }}</span>
              <span>·</span>
              <span>{{ o.package?.name }}</span>
              <span>·</span>
              <span>{{ formatDate(o.created_at) }}</span>
            </div>
          </div>
          <span :class="['badge', statusClass(o.status)]">{{ o.status.replace('_', ' ') }}</span>
          <span class="text-gray-400">›</span>
        </RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { format } from 'date-fns'
import AppNavbar from '@/components/common/AppNavbar.vue'
import api from '@/api/axios'

const role = ref<'buyer' | 'seller'>('buyer')
const orders = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get(`/api/v1/orders?as=${role.value}`)
    orders.value = data
  } finally {
    loading.value = false
  }
}

function statusClass(s: string) {
  return { pending: 'badge-warning', in_progress: 'badge-primary', delivered: 'badge', completed: 'badge-success', cancelled: 'badge-danger', revision: 'badge-warning' }[s] || 'badge-gray'
}

function formatDate(d: string) {
  try { return format(new Date(d), 'MMM d, yyyy') } catch { return d }
}

watch(role, load)
onMounted(load)
</script>
