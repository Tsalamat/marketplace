<template>
  <div class="flex items-center gap-0.5">
    <svg
      v-for="i in 5" :key="i"
      :class="[
        sizeClass,
        i <= fullStars ? 'text-yellow-400' :
        i === halfStar ? 'text-yellow-300' : 'text-gray-200 dark:text-gray-700'
      ]"
      viewBox="0 0 20 20" fill="currentColor"
    >
      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  rating: number
  size?: 'sm' | 'md' | 'lg'
}>()

const sizeClass = computed(() => ({
  sm: 'w-3.5 h-3.5',
  md: 'w-4 h-4',
  lg: 'w-5 h-5',
}[props.size || 'md']))

const fullStars = computed(() => Math.floor(props.rating))
const halfStar = computed(() => props.rating % 1 >= 0.5 ? fullStars.value + 1 : 0)
</script>
