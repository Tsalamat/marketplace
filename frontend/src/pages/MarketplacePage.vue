<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />

    <div class="max-w-7xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <!-- Header -->
      <div class="py-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-1">Marketplace</h1>
        <p class="text-gray-500 dark:text-gray-400">{{ store.total }} services available</p>
      </div>

      <div class="flex gap-6">
        <!-- Sidebar Filters -->
        <aside class="hidden lg:block w-64 flex-shrink-0">
          <div class="card p-5 sticky top-24">
            <h3 class="font-semibold text-gray-900 dark:text-white mb-4">Filters</h3>

            <!-- Category -->
            <div class="mb-6">
              <label class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3 block">Category</label>
              <div class="space-y-2">
                <label v-for="cat in store.categories" :key="cat.slug" class="flex items-center gap-2 cursor-pointer group">
                  <input type="radio" :value="cat.slug" v-model="selectedCategory" class="text-primary-600" />
                  <span class="text-sm text-gray-700 dark:text-gray-300 group-hover:text-primary-600 dark:group-hover:text-primary-400">
                    {{ cat.name }}
                  </span>
                </label>
              </div>
            </div>

            <!-- Price Range -->
            <div class="mb-6">
              <label class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3 block">Price Range</label>
              <div class="grid grid-cols-2 gap-2">
                <input v-model.number="minPrice" type="number" placeholder="Min" class="input text-sm py-2 px-3" />
                <input v-model.number="maxPrice" type="number" placeholder="Max" class="input text-sm py-2 px-3" />
              </div>
            </div>

            <!-- Min Rating -->
            <div class="mb-6">
              <label class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3 block">Minimum Rating</label>
              <div class="space-y-1">
                <label v-for="r in [4.5, 4, 3.5, 0]" :key="r" class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" :value="r || undefined" v-model="minRating" class="text-primary-600" />
                  <StarRating :rating="r || 1" size="sm" />
                  <span class="text-xs text-gray-600 dark:text-gray-400">{{ r ? `${r}+` : 'Any' }}</span>
                </label>
              </div>
            </div>

            <button @click="applyFilters" class="btn-primary w-full">Apply Filters</button>
            <button @click="resetFilters" class="btn-ghost w-full mt-2 text-sm">Reset</button>
          </div>
        </aside>

        <!-- Main Content -->
        <div class="flex-1 min-w-0">
          <!-- Toolbar -->
          <div class="flex items-center gap-4 mb-6 flex-wrap">
            <!-- Search -->
            <div class="flex-1 min-w-64">
              <input
                v-model="searchQuery"
                @keyup.enter="applyFilters"
                type="text"
                placeholder="Search services..."
                class="input"
              />
            </div>

            <!-- Sort -->
            <select v-model="sortBy" @change="applyFilters" class="input w-auto pr-8">
              <option value="trending">Trending</option>
              <option value="newest">Newest</option>
              <option value="rating">Top Rated</option>
              <option value="price_asc">Price: Low to High</option>
              <option value="price_desc">Price: High to Low</option>
            </select>
          </div>

          <!-- Loading -->
          <div v-if="store.loading" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-5">
            <div v-for="i in 9" :key="i" class="card overflow-hidden animate-pulse">
              <div class="h-44 bg-gray-200 dark:bg-surface-800"/>
              <div class="p-4 space-y-3">
                <div class="h-3 bg-gray-200 dark:bg-surface-800 rounded w-1/3"/>
                <div class="h-4 bg-gray-200 dark:bg-surface-800 rounded"/>
                <div class="h-4 bg-gray-200 dark:bg-surface-800 rounded w-3/4"/>
              </div>
            </div>
          </div>

          <!-- Grid -->
          <div v-else-if="store.services.length" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-5">
            <ServiceCard v-for="s in store.services" :key="s.id" :service="s" />
          </div>

          <!-- Empty -->
          <div v-else class="text-center py-24">
            <svg class="w-14 h-14 text-gray-300 dark:text-surface-700 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">No services found</h3>
            <p class="text-gray-500 dark:text-gray-400 mb-6">Try adjusting your filters or search terms</p>
            <button @click="resetFilters" class="btn-primary">Clear Filters</button>
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 mt-10">
            <button
              :disabled="currentPage === 1"
              @click="goToPage(currentPage - 1)"
              class="btn-secondary px-4"
            >← Prev</button>
            <span class="text-sm text-gray-600 dark:text-gray-400 px-4">
              Page {{ currentPage }} of {{ totalPages }}
            </span>
            <button
              :disabled="currentPage === totalPages"
              @click="goToPage(currentPage + 1)"
              class="btn-secondary px-4"
            >Next →</button>
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
import ServiceCard from '@/components/marketplace/ServiceCard.vue'
import StarRating from '@/components/common/StarRating.vue'
import { useMarketplaceStore } from '@/stores/marketplace'

const route = useRoute()
const store = useMarketplaceStore()

const searchQuery = ref('')
const selectedCategory = ref('')
const minPrice = ref<number | undefined>()
const maxPrice = ref<number | undefined>()
const minRating = ref<number | undefined>()
const sortBy = ref('trending')
const currentPage = ref(1)

const totalPages = computed(() => Math.ceil(store.total / 20))

function applyFilters() {
  currentPage.value = 1
  store.fetchServices({
    q: searchQuery.value || undefined,
    category: selectedCategory.value || undefined,
    min_price: minPrice.value,
    max_price: maxPrice.value,
    min_rating: minRating.value,
    sort: sortBy.value,
    page: 1,
  })
}

function resetFilters() {
  searchQuery.value = ''
  selectedCategory.value = ''
  minPrice.value = undefined
  maxPrice.value = undefined
  minRating.value = undefined
  sortBy.value = 'trending'
  applyFilters()
}

function goToPage(page: number) {
  currentPage.value = page
  store.fetchServices({ page })
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(async () => {
  await store.fetchCategories()

  if (route.query.q) searchQuery.value = String(route.query.q)
  if (route.query.category) selectedCategory.value = String(route.query.category)

  await store.fetchServices({
    q: searchQuery.value || undefined,
    category: selectedCategory.value || undefined,
  })
})
</script>
