<template>
  <RouterLink :to="`/gig/${service.slug}`"
    class="card hover:shadow-card-hover hover:-translate-y-1 transition-all duration-200 overflow-hidden group block">

    <!-- Thumbnail -->
    <div class="relative h-44 bg-gray-100 dark:bg-surface-800 overflow-hidden">
      <img v-if="service.gallery?.[0]" :src="service.gallery[0]" :alt="service.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
      <div v-else class="w-full h-full flex items-center justify-center">
        <svg class="w-12 h-12 text-gray-300 dark:text-surface-700" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path :d="categoryIcon"/>
        </svg>
      </div>
      <div v-if="service.is_featured"
        class="absolute top-2 left-2 flex items-center gap-1 px-2 py-1 bg-amber-400 rounded-lg text-amber-900 text-xs font-bold shadow">
        <svg class="w-3 h-3" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
        Featured
      </div>
    </div>

    <!-- Content -->
    <div class="p-4">
      <!-- Seller row -->
      <div class="flex items-center gap-2 mb-2.5">
        <img :src="sellerAvatar" :alt="service.seller?.username"
          class="w-6 h-6 rounded-full object-cover flex-shrink-0 ring-1 ring-gray-200 dark:ring-surface-700"/>
        <span class="text-xs font-medium text-gray-600 dark:text-gray-400 truncate">{{ service.seller?.username }}</span>
        <span v-if="service.seller?.profile?.is_online"
          class="ml-auto w-2 h-2 rounded-full bg-green-500 flex-shrink-0"/>
      </div>

      <!-- Title -->
      <h3 class="font-semibold text-gray-900 dark:text-white text-sm leading-snug line-clamp-2 mb-3 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
        {{ service.title }}
      </h3>

      <!-- Rating -->
      <div class="flex items-center gap-1.5 mb-3">
        <div class="flex">
          <svg v-for="i in 5" :key="i" class="w-3.5 h-3.5"
            :class="i <= Math.round(service.rating) ? 'text-amber-400' : 'text-gray-200 dark:text-surface-700'"
            viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
        </div>
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ service.rating > 0 ? service.rating.toFixed(1) : 'New' }}
          <span v-if="service.total_reviews > 0">({{ service.total_reviews }})</span>
        </span>
      </div>

      <!-- Price -->
      <div class="flex items-center justify-between pt-3 border-t border-gray-100 dark:border-surface-800">
        <span class="text-xs text-gray-400">Starting at</span>
        <span class="font-bold text-gray-900 dark:text-white text-sm">${{ minPrice }}</span>
      </div>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Service } from '@/stores/marketplace'

const props = defineProps<{ service: Service }>()

const catIcons: Record<string, string> = {
  programming: 'M16 18l6-6-6-6M8 6l-6 6 6 6',
  design:      'M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z',
  tutoring:    'M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2zM22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z',
  writing:     'M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7',
  video:       'M23 7l-7 5 7 5V7z',
  photography: 'M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2zM12 17a4 4 0 1 0 0-8 4 4 0 0 0 0 8z',
  delivery:    'M5 17H3a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11a2 2 0 0 1 2 2v3m4 0h2l3 5v3h-6',
  fitness:     'M18 20V10M12 20V4M6 20v-6',
  music:       'M9 18V5l12-2v13',
  business:    'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
}

const categoryIcon = computed(() =>
  catIcons[props.service.category?.slug || ''] ||
  'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z'
)

const sellerAvatar = computed(() => {
  const url = props.service.seller?.profile?.avatar_url
  return url || `https://ui-avatars.com/api/?name=${props.service.seller?.username || 'U'}&background=2563eb&color=fff&size=32`
})

const minPrice = computed(() => {
  if (!props.service.packages?.length) return 0
  return Math.min(...props.service.packages.map(p => p.price))
})
</script>
