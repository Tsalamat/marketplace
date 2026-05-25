<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Reports</h1>
    <div v-if="reports.length === 0" class="card p-8 text-center text-gray-500 dark:text-gray-400">
      <svg class="w-10 h-10 text-green-400 mx-auto mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="20 6 9 17 4 12"/></svg>
      <p>No pending reports</p>
    </div>
    <div v-else class="space-y-4">
      <div v-for="r in reports" :key="r.id" class="card p-5">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="font-semibold text-gray-900 dark:text-white">{{ r.reason }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ r.description }}</p>
            <p class="text-xs text-gray-400 mt-2">By: {{ r.reporter?.username }}</p>
          </div>
          <div class="flex gap-2">
            <button @click="resolve(r.id, 'resolved')" class="btn-primary text-sm">Resolve</button>
            <button @click="resolve(r.id, 'dismissed')" class="btn-secondary text-sm">Dismiss</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
const reports = ref<any[]>([])
async function load() { try { const { data } = await api.get('/api/v1/admin/reports'); reports.value = data } catch {} }
async function resolve(id: string, status: string) {
  await api.post(`/api/v1/admin/reports/${id}/resolve`, { status })
  reports.value = reports.value.filter(r => r.id !== id)
}
onMounted(load)
</script>
