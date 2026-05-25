<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Users</h1>
      <input v-model="search" @input="load" type="text" placeholder="Search..." class="input w-64" />
    </div>
    <div class="card overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 dark:bg-surface-800 text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
          <tr><th class="px-4 py-3 text-left">User</th><th class="px-4 py-3 text-left">Role</th><th class="px-4 py-3 text-left">Status</th><th class="px-4 py-3 text-left">Joined</th><th class="px-4 py-3"></th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-surface-800">
          <tr v-for="u in users" :key="u.id" class="hover:bg-gray-50 dark:hover:bg-surface-800/50">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <img :src="`https://ui-avatars.com/api/?name=${u.username}&size=32&background=3b82f6&color=fff`" class="w-8 h-8 rounded-full" />
                <div><p class="font-medium text-gray-900 dark:text-white">{{ u.username }}</p><p class="text-xs text-gray-500 dark:text-gray-400">{{ u.email }}</p></div>
              </div>
            </td>
            <td class="px-4 py-3"><span class="badge-primary capitalize">{{ u.role }}</span></td>
            <td class="px-4 py-3"><span :class="u.is_active ? 'badge-success' : 'badge-danger'">{{ u.is_active ? 'Active' : 'Banned' }}</span></td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ formatDate(u.created_at) }}</td>
            <td class="px-4 py-3">
              <button v-if="u.is_active" @click="ban(u)" class="text-xs text-red-600 dark:text-red-400 hover:underline">Ban</button>
              <button v-else @click="unban(u)" class="text-xs text-green-600 dark:text-green-400 hover:underline">Unban</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="loading" class="p-8 text-center text-gray-400">Loading...</div>
      <div v-else-if="users.length === 0" class="p-8 text-center text-gray-400">No users found</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { format } from 'date-fns'
import api from '@/api/axios'
const users = ref<any[]>([]); const loading = ref(false); const search = ref('')
async function load() {
  loading.value = true
  try { const { data } = await api.get(`/api/v1/admin/users?q=${search.value}`); users.value = data.data || [] }
  finally { loading.value = false }
}
async function ban(u: any) { await api.post(`/api/v1/admin/users/${u.id}/ban`, { reason: 'Admin ban' }); u.is_active = false }
async function unban(u: any) { await api.post(`/api/v1/admin/users/${u.id}/unban`); u.is_active = true }
function formatDate(d: string) { try { return format(new Date(d), 'MMM d, yyyy') } catch { return d } }
onMounted(load)
</script>
