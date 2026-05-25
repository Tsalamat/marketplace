<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-3xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div v-if="loading" class="py-24 flex justify-center">
        <svg class="w-8 h-8 animate-spin text-primary-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
      </div>
      <div v-else-if="order" class="py-8 space-y-6">
        <!-- Header -->
        <div class="flex items-start justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ order.service?.title }}</h1>
            <p class="text-gray-500 dark:text-gray-400 mt-1">Order #{{ order.id.slice(0, 8).toUpperCase() }}</p>
          </div>
          <span :class="['badge text-sm px-3 py-1', statusClass(order.status)]">{{ order.status.replace('_', ' ') }}</span>
        </div>

        <!-- Parties -->
        <div class="card p-6 grid grid-cols-2 gap-6">
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wide">Buyer</p>
            <div class="flex items-center gap-3">
              <img :src="avatar(order.buyer)" class="w-10 h-10 rounded-full object-cover" />
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ order.buyer?.username }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ order.buyer?.profile?.university }}</p>
              </div>
            </div>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wide">Seller</p>
            <div class="flex items-center gap-3">
              <img :src="avatar(order.seller)" class="w-10 h-10 rounded-full object-cover" />
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ order.seller?.username }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ order.seller?.profile?.university }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Package Details -->
        <div class="card p-6">
          <h2 class="font-bold text-gray-900 dark:text-white mb-4">Package Details</h2>
          <div class="grid grid-cols-3 gap-4 text-center">
            <div class="bg-gray-50 dark:bg-surface-800 rounded-xl p-4">
              <div class="text-xl font-bold text-primary-600 dark:text-primary-400">${{ order.amount }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Order Total</div>
            </div>
            <div class="bg-gray-50 dark:bg-surface-800 rounded-xl p-4">
              <div class="text-xl font-bold text-gray-900 dark:text-white">{{ order.package?.delivery_days }}d</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Delivery Time</div>
            </div>
            <div class="bg-gray-50 dark:bg-surface-800 rounded-xl p-4">
              <div class="text-xl font-bold text-gray-900 dark:text-white">{{ order.revision_count }}/{{ order.max_revisions }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">Revisions Used</div>
            </div>
          </div>
        </div>

        <!-- Requirements -->
        <div v-if="order.requirements" class="card p-6">
          <h2 class="font-bold text-gray-900 dark:text-white mb-3">Requirements</h2>
          <p class="text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap">{{ order.requirements }}</p>
        </div>

        <!-- Actions -->
        <div class="card p-6">
          <h2 class="font-bold text-gray-900 dark:text-white mb-4">Actions</h2>
          <div class="flex flex-wrap gap-3">
            <button v-if="isSeller && order.status === 'pending'" @click="updateStatus('in_progress')" class="btn-primary">Accept Order</button>
            <button v-if="isSeller && order.status === 'in_progress'" @click="updateStatus('delivered')" class="btn-primary">Mark Delivered</button>
            <button v-if="isBuyer && order.status === 'delivered'" @click="updateStatus('completed')" class="btn-primary">Accept Delivery</button>
            <button v-if="isBuyer && order.status === 'delivered' && order.revision_count < order.max_revisions" @click="updateStatus('revision')" class="btn-secondary">Request Revision</button>
            <button v-if="order.status === 'pending'" @click="updateStatus('cancelled')" class="btn-secondary text-red-600 dark:text-red-400">Cancel</button>
            <RouterLink to="/chat" class="btn-ghost flex items-center gap-2">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
              Open Chat
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const route = useRoute()
const auth = useAuthStore()
const order = ref<any>(null)
const loading = ref(false)

const isBuyer  = computed(() => order.value?.buyer_id === auth.user?.id)
const isSeller = computed(() => order.value?.seller_id === auth.user?.id)

function avatar(user: any) {
  return user?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${user?.username}&size=40&background=3b82f6&color=fff`
}

function statusClass(s: string) {
  return { pending: 'badge-warning', in_progress: 'badge-primary', delivered: 'badge', completed: 'badge-success', cancelled: 'badge-danger' }[s] || 'badge-gray'
}

async function updateStatus(status: string) {
  try {
    await api.patch(`/api/v1/orders/${order.value.id}/status`, { status })
    order.value.status = status
    toast.success(`Order ${status.replace('_', ' ')}`)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to update status')
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await api.get(`/api/v1/orders/${route.params.id}`)
    order.value = data
  } finally {
    loading.value = false
  }
})
</script>
