<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-7xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          Welcome back, {{ auth.fullName || auth.user?.username }}
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Here's what's happening with your account.</p>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
        <div v-for="s in stats" :key="s.label" class="card p-5">
          <div class="w-9 h-9 rounded-xl bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center mb-3">
            <svg class="w-5 h-5 text-primary-600 dark:text-primary-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path :d="s.icon"/>
            </svg>
          </div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ s.value }}</div>
          <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ s.label }}</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Recent Orders -->
        <div class="card p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="font-bold text-gray-900 dark:text-white">Recent Orders</h2>
            <RouterLink to="/orders" class="text-sm text-primary-600 dark:text-primary-400 hover:underline">View all</RouterLink>
          </div>
          <div v-if="orders.length === 0" class="text-center py-8">
            <svg class="w-10 h-10 text-gray-300 dark:text-surface-700 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M16.5 9.4l-9-5.19M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12"/>
            </svg>
            <p class="text-sm text-gray-400">No orders yet</p>
          </div>
          <div v-else class="space-y-3">
            <div v-for="o in orders.slice(0, 5)" :key="o.id"
              class="flex items-center justify-between p-3 bg-gray-50 dark:bg-surface-800 rounded-xl">
              <div class="flex-1 min-w-0 mr-3">
                <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ o.service?.title }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">${{ o.amount }}</p>
              </div>
              <span :class="statusBadge(o.status)">{{ o.status.replace('_', ' ') }}</span>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="card p-6">
          <h2 class="font-bold text-gray-900 dark:text-white mb-4">Quick Actions</h2>
          <div class="grid grid-cols-2 gap-3">
            <RouterLink v-for="a in actions" :key="a.label" :to="a.to"
              class="flex flex-col items-center gap-2 p-4 bg-gray-50 dark:bg-surface-800 rounded-xl hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-colors group">
              <svg class="w-6 h-6 text-gray-500 dark:text-gray-400 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path :d="a.icon"/>
              </svg>
              <span class="text-xs font-medium text-gray-700 dark:text-gray-300 group-hover:text-primary-600 dark:group-hover:text-primary-400 text-center">{{ a.label }}</span>
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'

const auth = useAuthStore()
const orders = ref<any[]>([])

const stats = ref([
  { icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z', label: 'Active Orders', value: '0' },
  { icon: 'M22 11.08V12a10 10 0 1 1-5.93-9.14M22 4L12 14.01l-3-3', label: 'Completed', value: '0' },
  { icon: 'M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z', label: 'My Rating', value: '—' },
  { icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6', label: 'Earned', value: '$0' },
])

const actions = [
  { icon: 'M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z', label: 'Browse Services', to: '/marketplace' },
  { icon: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10zM12 8v8M8 12h8', label: 'Create Gig', to: '/create-gig' },
  { icon: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z', label: 'Messages', to: '/chat' },
  { icon: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75', label: 'Community', to: '/community' },
  { icon: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z', label: 'My Profile', to: `/profile/${auth.user?.username}` },
  { icon: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z', label: 'Settings', to: '/settings' },
]

function statusBadge(status: string) {
  const map: Record<string, string> = {
    pending: 'badge-warning', in_progress: 'badge-primary',
    delivered: 'badge', completed: 'badge-success', cancelled: 'badge-danger',
  }
  return map[status] || 'badge-gray'
}

onMounted(async () => {
  try {
    const { data } = await api.get('/api/v1/orders?as=buyer&limit=5')
    orders.value = data
    const active = data.filter((o: any) => ['pending', 'in_progress', 'delivered'].includes(o.status)).length
    const done   = data.filter((o: any) => o.status === 'completed').length
    stats.value[0].value = String(active)
    stats.value[1].value = String(done)
    if (auth.user?.profile?.rating) {
      stats.value[2].value = Number(auth.user.profile.rating).toFixed(1)
    }
  } catch {}
})
</script>
