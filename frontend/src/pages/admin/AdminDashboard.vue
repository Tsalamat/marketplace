<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Dashboard</h1>
    <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 8" :key="i" class="card p-5 animate-pulse h-24"/>
    </div>
    <div v-else class="space-y-6">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="s in statCards" :key="s.label" class="card p-5">
          <div class="w-8 h-8 rounded-lg bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center mb-3">
            <svg class="w-4 h-4 text-primary-600 dark:text-primary-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path :d="s.icon"/>
            </svg>
          </div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ s.value }}</div>
          <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ s.label }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
const loading = ref(false)
const statCards = ref<any[]>([])
onMounted(async () => {
  loading.value = true
  try {
    const { data } = await api.get('/api/v1/admin/dashboard')
    statCards.value = [
      { icon: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z', label: 'Total Users', value: data.total_users?.toLocaleString() },
      { icon: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z', label: 'Total Services', value: data.total_services?.toLocaleString() },
      { icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z', label: 'Total Orders', value: data.total_orders?.toLocaleString() },
      { icon: 'M22 11.08V12a10 10 0 1 1-5.93-9.14M22 4L12 14.01l-3-3', label: 'Completed Orders', value: data.completed_orders?.toLocaleString() },
      { icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6', label: 'Total Revenue', value: `$${Number(data.total_revenue || 0).toLocaleString()}` },
      { icon: 'M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z', label: 'Platform Fees', value: `$${Number(data.platform_fees || 0).toLocaleString()}` },
      { icon: 'M5 12.55a11 11 0 0 1 14.08 0M1.42 9a16 16 0 0 1 21.16 0M8.53 16.11a6 6 0 0 1 6.95 0M12 20h.01', label: 'Online Users', value: data.active_users?.toLocaleString() },
      { icon: 'M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7', label: 'Total Posts', value: data.total_posts?.toLocaleString() },
    ]
  } finally { loading.value = false }
})
</script>
