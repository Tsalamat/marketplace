<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Services</h1>
    <div class="card overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 dark:bg-surface-800 text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
          <tr><th class="px-4 py-3 text-left">Title</th><th class="px-4 py-3 text-left">Seller</th><th class="px-4 py-3">Orders</th><th class="px-4 py-3">Rating</th><th class="px-4 py-3">Featured</th><th class="px-4 py-3">Active</th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-surface-800">
          <tr v-for="s in services" :key="s.id" class="hover:bg-gray-50 dark:hover:bg-surface-800/50">
            <td class="px-4 py-3 font-medium text-gray-900 dark:text-white max-w-xs truncate">{{ s.title }}</td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ s.seller?.username }}</td>
            <td class="px-4 py-3 text-center">{{ s.orders_count }}</td>
            <td class="px-4 py-3 text-center">{{ s.rating?.toFixed(1) ?? '—' }}</td>
            <td class="px-4 py-3 text-center">
              <button @click="toggle(s, 'is_featured')" :class="s.is_featured ? 'text-yellow-500' : 'text-gray-400'">★</button>
            </td>
            <td class="px-4 py-3 text-center">
              <button @click="toggle(s, 'is_active')" :class="s.is_active ? 'badge-success' : 'badge-danger'" class="badge">{{ s.is_active ? 'Active' : 'Hidden' }}</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="loading" class="p-8 text-center text-gray-400">Loading...</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
const services = ref<any[]>([]); const loading = ref(false)
async function load() {
  loading.value = true
  try { const { data } = await api.get('/api/v1/services?limit=50'); services.value = data.data || [] }
  finally { loading.value = false }
}
async function toggle(s: any, field: string) {
  const newVal = !s[field]
  await api.patch(`/api/v1/admin/services/${s.id}`, { [field]: newVal })
  s[field] = newVal
}
onMounted(load)
</script>
