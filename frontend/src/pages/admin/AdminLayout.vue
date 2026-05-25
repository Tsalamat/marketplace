<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950 flex">
    <!-- Sidebar -->
    <aside class="w-56 bg-white dark:bg-surface-900 border-r border-gray-200 dark:border-surface-800 flex flex-col pt-16">
      <div class="p-4 border-b border-gray-100 dark:border-surface-800">
        <div class="text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-widest">Admin Panel</div>
      </div>
      <nav class="flex-1 p-3 space-y-1">
        <RouterLink v-for="item in nav" :key="item.to" :to="item.to"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors"
          :class="$route.path === item.to ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-400' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-surface-800'">
          <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path :d="item.icon"/>
          </svg>
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="p-4 border-t border-gray-100 dark:border-surface-800">
        <RouterLink to="/marketplace" class="text-xs text-gray-500 dark:text-gray-400 hover:underline">← Back to site</RouterLink>
      </div>
    </aside>
    <!-- Content -->
    <div class="flex-1 overflow-auto">
      <AppNavbar />
      <div class="pt-16 p-6">
        <RouterView />
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import AppNavbar from '@/components/common/AppNavbar.vue'
const nav = [
  { icon: 'M3 3h7v7H3zm11 0h7v7h-7zM3 14h7v7H3zm11 0h7v7h-7z', label: 'Dashboard', to: '/admin/dashboard' },
  { icon: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8zM23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75', label: 'Users', to: '/admin/users' },
  { icon: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z', label: 'Services', to: '/admin/services' },
  { icon: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z', label: 'Orders', to: '/admin/orders' },
  { icon: 'M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4M12 17h.01', label: 'Reports', to: '/admin/reports' },
  { icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6', label: 'Revenue', to: '/admin/revenue' },
]
</script>
