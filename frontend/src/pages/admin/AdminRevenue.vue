<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Revenue Analytics</h1>
    <div v-if="loading" class="grid grid-cols-3 gap-4">
      <div v-for="i in 3" :key="i" class="card p-5 animate-pulse h-24"/>
    </div>
    <div v-else class="space-y-6">
      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">30-day Revenue</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">${{ totalRevenue.toLocaleString() }}</p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">Platform Fees</p>
          <p class="text-2xl font-bold text-primary-600 dark:text-primary-400 mt-1">${{ totalFees.toLocaleString() }}</p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">Orders Completed</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ totalOrders }}</p>
        </div>
      </div>

      <!-- Chart -->
      <div class="card p-6">
        <h2 class="font-bold text-gray-900 dark:text-white mb-4">Daily Revenue (last 30 days)</h2>
        <div v-if="!chartData.length" class="text-center py-12 text-gray-400">No completed orders yet</div>
        <div v-else class="overflow-x-auto">
          <div class="flex items-end gap-1 h-48 min-w-full">
            <div v-for="d in chartData" :key="d.date" class="flex-1 flex flex-col items-center gap-1 group">
              <div class="relative flex-1 flex items-end w-full">
                <div
                  :style="{ height: barHeight(d.revenue) + '%' }"
                  class="w-full bg-primary-500 dark:bg-primary-600 rounded-t-sm group-hover:bg-primary-600 dark:group-hover:bg-primary-500 transition-colors min-h-[2px]"
                  :title="`$${d.revenue} on ${d.date}`"
                />
              </div>
              <span class="text-[8px] text-gray-400 -rotate-45 origin-top-left">{{ d.date.slice(5) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="card overflow-hidden">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-surface-800 text-xs uppercase text-gray-500 dark:text-gray-400">
            <tr>
              <th class="px-4 py-3 text-left">Date</th>
              <th class="px-4 py-3 text-right">Orders</th>
              <th class="px-4 py-3 text-right">Revenue</th>
              <th class="px-4 py-3 text-right">Platform Fee</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-surface-800">
            <tr v-for="d in chartData" :key="d.date" class="hover:bg-gray-50 dark:hover:bg-surface-800/50">
              <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ d.date }}</td>
              <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ d.orders }}</td>
              <td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white">${{ d.revenue.toFixed(2) }}</td>
              <td class="px-4 py-3 text-right text-primary-600 dark:text-primary-400">${{ d.platform.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api/axios'

const loading   = ref(false)
const chartData = ref<any[]>([])

const totalRevenue = computed(() => chartData.value.reduce((s, d) => s + d.revenue, 0))
const totalFees    = computed(() => chartData.value.reduce((s, d) => s + d.platform, 0))
const totalOrders  = computed(() => chartData.value.reduce((s, d) => s + d.orders, 0))

const maxRevenue = computed(() => Math.max(...chartData.value.map(d => d.revenue), 1))
function barHeight(r: number) { return Math.max(2, (r / maxRevenue.value) * 100) }

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await api.get('/api/v1/admin/revenue')
    chartData.value = data
  } finally { loading.value = false }
})
</script>
