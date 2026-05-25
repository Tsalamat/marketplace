<template>
  <nav class="fixed top-0 inset-x-0 z-50 bg-white/95 dark:bg-surface-900/95 backdrop-blur-xl border-b border-gray-100 dark:border-surface-800">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 flex items-center gap-3 h-16">

      <!-- Logo -->
      <RouterLink to="/marketplace" class="flex items-center gap-2.5 mr-4 flex-shrink-0">
        <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-700 rounded-lg flex items-center justify-center shadow-sm">
          <svg class="w-4.5 h-4.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
            <polyline points="9 22 9 12 15 12 15 22"/>
          </svg>
        </div>
        <span class="font-bold text-gray-900 dark:text-white hidden sm:block tracking-tight">StudentMarket</span>
      </RouterLink>

      <!-- Search -->
      <div class="flex-1 max-w-md hidden md:block">
        <div class="relative">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            v-model="searchQuery"
            @keyup.enter="goSearch"
            type="text"
            placeholder="Search services..."
            class="input py-2 pl-9 pr-4 text-sm"
          />
        </div>
      </div>

      <div class="flex items-center gap-1 ml-auto">
        <RouterLink to="/marketplace" class="btn-ghost hidden md:inline-flex text-sm">Browse</RouterLink>
        <RouterLink to="/community"   class="btn-ghost hidden md:inline-flex text-sm">Community</RouterLink>
        <RouterLink v-if="auth.isAuthenticated" to="/friends" class="btn-ghost hidden md:inline-flex text-sm">Friends</RouterLink>
        <RouterLink v-if="auth.isAuthenticated" to="/map"     class="btn-ghost hidden md:inline-flex text-sm">
          <svg class="w-4 h-4 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
          </svg>
          Map
        </RouterLink>

        <template v-if="auth.isAuthenticated">
          <RouterLink v-if="auth.isSeller" to="/create-gig" class="hidden sm:inline-flex items-center gap-1.5 btn-secondary text-sm px-3 py-2">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/>
            </svg>
            New Gig
          </RouterLink>

          <!-- Notifications -->
          <RouterLink to="/notifications" class="relative p-2 btn-ghost rounded-lg">
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
              <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
            </svg>
            <span v-if="unreadNotifs > 0"
              class="absolute top-1.5 right-1.5 w-4 h-4 bg-red-500 rounded-full text-white text-[10px] font-bold flex items-center justify-center">
              {{ unreadNotifs > 9 ? '9+' : unreadNotifs }}
            </span>
          </RouterLink>

          <!-- Chat -->
          <RouterLink to="/chat" class="relative p-2 btn-ghost rounded-lg">
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            <span v-if="chatStore.totalUnread > 0"
              class="absolute top-1.5 right-1.5 w-4 h-4 bg-primary-600 rounded-full text-white text-[10px] font-bold flex items-center justify-center">
              {{ chatStore.totalUnread > 9 ? '9+' : chatStore.totalUnread }}
            </span>
          </RouterLink>

          <!-- Avatar Menu -->
          <div class="relative" ref="avatarMenuRef">
            <button @click="menuOpen = !menuOpen" class="flex items-center gap-2 ml-1 rounded-full p-0.5 hover:ring-2 hover:ring-primary-500 transition-all">
              <img
                :src="auth.user?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${auth.user?.username}&background=2563eb&color=fff&size=36`"
                class="w-8 h-8 rounded-full object-cover"
              />
            </button>

            <Transition name="dropdown">
              <div v-if="menuOpen"
                class="absolute right-0 top-11 w-56 bg-white dark:bg-surface-900 rounded-2xl shadow-xl border border-gray-100 dark:border-surface-800 overflow-hidden z-50">
                <!-- User info -->
                <div class="px-4 py-3 border-b border-gray-100 dark:border-surface-800">
                  <p class="font-semibold text-gray-900 dark:text-white text-sm truncate">{{ auth.fullName || auth.user?.username }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ auth.user?.email }}</p>
                  <span class="inline-block mt-1 badge-primary text-xs capitalize">{{ auth.user?.role }}</span>
                </div>

                <div class="py-1">
                  <NavItem to="/dashboard"                        icon="grid" label="Dashboard"  @click="menuOpen=false"/>
                  <NavItem :to="`/profile/${auth.user?.username}`" icon="user" label="My Profile"  @click="menuOpen=false"/>
                  <NavItem to="/orders"                           icon="package" label="My Orders"  @click="menuOpen=false"/>
                  <NavItem to="/create-gig"                       icon="plus-circle" label="Create Gig"  @click="menuOpen=false"/>
                  <NavItem to="/friends"                          icon="users"    label="Friends"   @click="menuOpen=false"/>
                  <NavItem to="/map"                              icon="map-pin"  label="Map"       @click="menuOpen=false"/>
                  <NavItem to="/settings"                         icon="settings" label="Settings"  @click="menuOpen=false"/>
                  <NavItem v-if="auth.isAdmin" to="/admin"        icon="shield" label="Admin Panel" @click="menuOpen=false" highlight/>
                  <div class="border-t border-gray-100 dark:border-surface-800 my-1"/>
                  <button @click="logout" class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/10 transition-colors">
                    <Icon name="log-out" class="w-4 h-4"/>
                    Sign Out
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </template>

        <template v-else>
          <RouterLink to="/login"    class="btn-ghost text-sm">Login</RouterLink>
          <RouterLink to="/register" class="btn-primary text-sm px-4 py-2">Get Started</RouterLink>
        </template>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { onClickOutside } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { toast } from 'vue3-toastify'
import api from '@/api/axios'

// Inline icon component
const Icon = defineComponent({
  props: { name: String, class: String },
  setup(props) {
    const icons: Record<string, string> = {
      'grid':        'M3 3h7v7H3zm11 0h7v7h-7zM3 14h7v7H3zm11 0h7v7h-7z',
      'user':        'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z',
      'package':     'M16.5 9.4l-9-5.19M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12',
      'plus-circle': 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10zM12 8v8M8 12h8',
      'settings':    'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z',
      'shield':      'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
      'log-out':     'M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9',
    }
    return () => h('svg', {
      class: props.class || 'w-4 h-4',
      viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2',
    }, [h('path', { d: icons[props.name || ''] || '' })])
  }
})

const NavItem = defineComponent({
  props: { to: String, icon: String, label: String, highlight: Boolean },
  emits: ['click'],
  setup(props, { emit }) {
    return () => h(resolveComponent('RouterLink'), {
      to: props.to,
      onClick: () => emit('click'),
      class: `flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${props.highlight ? 'text-primary-600 dark:text-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/10' : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-surface-800'}`,
    }, () => [
      h(Icon, { name: props.icon, class: 'w-4 h-4 flex-shrink-0' }),
      props.label,
    ])
  }
})

import { defineComponent, h, resolveComponent } from 'vue'

const auth = useAuthStore()
const chatStore = useChatStore()
const router = useRouter()
const searchQuery = ref('')
const menuOpen = ref(false)
const avatarMenuRef = ref()
const unreadNotifs = ref(0)

onClickOutside(avatarMenuRef, () => { menuOpen.value = false })

function goSearch() {
  router.push({ name: 'marketplace', query: { q: searchQuery.value } })
  searchQuery.value = ''
}

async function logout() {
  menuOpen.value = false
  await auth.logout()
  chatStore.disconnectWS()
  toast.success('Signed out')
  router.push({ name: 'landing' })
}

onMounted(async () => {
  if (auth.isAuthenticated) {
    try {
      const { data } = await api.get('/api/v1/notifications?limit=1')
      unreadNotifs.value = data.unread_count || 0
    } catch {}
  }
})
</script>

<style scoped>
.dropdown-enter-active { transition: opacity .15s, transform .15s; }
.dropdown-leave-active { transition: opacity .1s, transform .1s; }
.dropdown-enter-from  { opacity: 0; transform: translateY(-6px) scale(.98); }
.dropdown-leave-to    { opacity: 0; transform: translateY(-6px) scale(.98); }
</style>
