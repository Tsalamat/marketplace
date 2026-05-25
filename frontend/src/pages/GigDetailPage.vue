<template>
  <div class="min-h-screen bg-white dark:bg-surface-950">
    <AppNavbar />
    <div v-if="loading" class="max-w-6xl mx-auto px-4 pt-24 pb-12 flex justify-center">
      <svg class="w-8 h-8 animate-spin text-primary-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
    </div>
    <div v-else-if="service" class="max-w-6xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 py-8">
        <!-- Main Content -->
        <div class="lg:col-span-2 space-y-6">
          <div>
            <div v-if="service.category" class="badge-primary mb-3 inline-block">{{ service.category.name }}</div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white leading-tight">{{ service.title }}</h1>
          </div>

          <!-- Seller Info -->
          <div class="flex items-center gap-3">
            <img :src="sellerAvatar" class="w-11 h-11 rounded-full object-cover" />
            <div>
              <RouterLink :to="`/profile/${service.seller?.username}`" class="font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
                {{ service.seller?.username }}
              </RouterLink>
              <div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                <StarRating :rating="service.rating" size="sm" />
                <span>{{ service.rating.toFixed(1) }} ({{ service.total_reviews }})</span>
                <span>·</span>
                <span>{{ service.orders_count }} заказов</span>
              </div>
            </div>
          </div>

          <!-- Gallery -->
          <div v-if="service.gallery?.length" class="rounded-2xl overflow-hidden">
            <img :src="service.gallery[0]" class="w-full h-72 object-cover" :alt="service.title" />
          </div>

          <!-- Description -->
          <div class="card p-6">
            <h2 class="font-bold text-gray-900 dark:text-white mb-3">О данной услуге</h2>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap">{{ service.description }}</p>
            <div v-if="service.tags?.length" class="flex flex-wrap gap-2 mt-4">
              <span v-for="tag in service.tags" :key="tag" class="badge-gray">{{ tag }}</span>
            </div>
          </div>

          <!-- Вопросы и ответыs -->
          <div v-if="service.faqs?.length" class="card p-6">
            <h2 class="font-bold text-gray-900 dark:text-white mb-4">Вопросы и ответы</h2>
            <div class="space-y-4">
              <div v-for="faq in service.faqs" :key="faq.id" class="border-b border-gray-100 dark:border-surface-800 pb-4 last:border-0 last:pb-0">
                <button @click="faq._open = !faq._open" class="flex items-center justify-between w-full text-left">
                  <span class="font-medium text-gray-900 dark:text-white">{{ faq.question }}</span>
                  <svg :class="['w-4 h-4 text-gray-400 ml-2 flex-shrink-0 transition-transform', faq._open ? 'rotate-180' : '']" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
                </button>
                <p v-if="faq._open" class="text-gray-600 dark:text-gray-400 mt-2 text-sm leading-relaxed">{{ faq.answer }}</p>
              </div>
            </div>
          </div>

          <!-- Reviews -->
          <div v-if="service.reviews?.length" class="card p-6">
            <h2 class="font-bold text-gray-900 dark:text-white mb-4">Отзывы ({{ service.total_reviews }})</h2>
            <div class="space-y-5">
              <div v-for="r in service.reviews" :key="r.id" class="border-b border-gray-100 dark:border-surface-800 pb-5 last:border-0">
                <div class="flex items-center gap-3 mb-2">
                  <img :src="reviewerAvatar(r.reviewer)" class="w-8 h-8 rounded-full object-cover" />
                  <div>
                    <p class="font-semibold text-sm text-gray-900 dark:text-white">{{ r.reviewer?.username }}</p>
                    <StarRating :rating="r.rating" size="sm" />
                  </div>
                </div>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">{{ r.content }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Order Panel -->
        <div class="lg:col-span-1">
          <div class="card p-6 sticky top-24">
            <!-- Package Tabs -->
            <div v-if="service.packages?.length" class="flex rounded-xl overflow-hidden border border-gray-200 dark:border-surface-700 mb-4">
              <button v-for="(pkg, i) in service.packages" :key="pkg.id"
                @click="selectedPkg = i"
                :class="['flex-1 py-2.5 text-xs font-semibold transition-all', selectedPkg === i ? 'bg-primary-600 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-surface-800']">
                {{ pkg.name }}
              </button>
            </div>

            <div v-if="currentPkg">
              <div class="flex items-center justify-between mb-2">
                <h3 class="font-bold text-gray-900 dark:text-white">{{ currentPkg.title }}</h3>
                <span class="text-2xl font-bold text-primary-600 dark:text-primary-400">${{ currentPkg.price }}</span>
              </div>
              <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">{{ currentPkg.description }}</p>
              <div class="space-y-2 mb-5">
                <div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  <span>{{ currentPkg.delivery_days }} дн. выполнение</span>
                </div>
                <div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 .49-3"/></svg>
                  <span>{{ currentPkg.revisions }} правок</span>
                </div>
                <div v-for="f in currentPkg.features" :key="f" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="w-4 h-4 flex-shrink-0 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                  <span>{{ f }}</span>
                </div>
              </div>
              <button @click="placeOrder" :disabled="ordering" class="btn-primary w-full py-3 mb-3">
                {{ ordering ? 'Оформление...' : `Заказать · $${currentPkg.price}` }}
              </button>
              <button class="btn-ghost w-full" @click="contactSeller">Написать продавцу</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Order Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-end sm:items-center justify-center z-50 p-4" @click.self="showModal = false">
      <div class="card w-full max-w-md p-6 animate-slide-up">
        <h3 class="font-bold text-gray-900 dark:text-white text-lg mb-4">Опишите ваши требования</h3>
        <textarea v-model="requirements" rows="5" class="input mb-4" placeholder="Опишите как можно подробнее, что вам нужно..."/>
        <div class="flex gap-3">
          <button @click="showModal = false" class="btn-secondary flex-1">Cancel</button>
          <button @click="confirmOrder" :disabled="ordering" class="btn-primary flex-1">
            {{ ordering ? 'Отправка...' : 'Подтвердить заказ' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import StarRating from '@/components/common/StarRating.vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const route  = useRoute()
const router = useRouter()
const store  = useMarketplaceStore()
const auth   = useAuthStore()
const chat   = useChatStore()

const loading      = ref(false)
const service      = ref<any>(null)
const selectedPkg  = ref(0)
const showModal    = ref(false)
const ordering     = ref(false)
const requirements = ref('')

const currentPkg = computed(() => service.value?.packages?.[selectedPkg.value])
const sellerAvatar = computed(() => {
  const url = service.value?.seller?.profile?.avatar_url
  return url || `https://ui-avatars.com/api/?name=${service.value?.seller?.username}&size=44&background=3b82f6&color=fff`
})

function reviewerAvatar(u: any) {
  return u?.profile?.avatar_url || `https://ui-avatars.com/api/?name=${u?.username}&size=32&background=3b82f6&color=fff`
}

function placeOrder() {
  if (!auth.isAuthenticated) { router.push('/login'); return }
  showModal.value = true
}

async function confirmOrder() {
  if (!currentPkg.value) return
  ordering.value = true
  try {
    await api.post('/api/v1/orders', {
      service_id: service.value.id,
      package_id: currentPkg.value.id,
      requirements: requirements.value,
    })
    showModal.value = false
    toast.success('Заказ оформлен!')
    router.push('/orders')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to place order')
  } finally {
    ordering.value = false
  }
}

async function contactSeller() {
  if (!auth.isAuthenticated) { router.push('/login'); return }
  const c = await chat.getOrCreateDirect(service.value.seller_id)
  await chat.openChat(c)
  router.push('/chat')
}

onMounted(async () => {
  loading.value = true
  try {
    service.value = await store.fetchService(route.params.slug as string)
  } finally {
    loading.value = false
  }
})
</script>
