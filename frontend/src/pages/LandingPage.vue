<template>
  <div class="min-h-screen bg-white dark:bg-surface-950">

    <!-- Navbar -->
    <nav class="fixed top-0 inset-x-0 z-50 bg-white/90 dark:bg-surface-950/90 backdrop-blur-xl border-b border-gray-100 dark:border-surface-800">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 flex items-center justify-between h-16">
        <RouterLink to="/" class="flex items-center gap-2.5">
          <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-700 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
          </div>
          <span class="font-bold text-gray-900 dark:text-white tracking-tight">StudentMarket</span>
        </RouterLink>
        <div class="hidden md:flex items-center gap-6">
          <RouterLink to="/marketplace" class="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">Browse</RouterLink>
          <a href="#how-it-works" class="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">How it works</a>
          <a href="#categories" class="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">Categories</a>
        </div>
        <div class="flex items-center gap-3">
          <RouterLink to="/login"    class="btn-ghost text-sm">Login</RouterLink>
          <RouterLink to="/register" class="btn-primary text-sm">Get Started</RouterLink>
        </div>
      </div>
    </nav>

    <!-- Hero -->
    <section class="pt-32 pb-24 px-4 relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-b from-primary-50/60 via-white to-white dark:from-primary-950/20 dark:via-surface-950 dark:to-surface-950 -z-10"/>
      <div class="absolute -top-40 -right-40 w-[600px] h-[600px] bg-primary-100 dark:bg-primary-900/10 rounded-full blur-3xl opacity-60 -z-10"/>
      <div class="absolute -bottom-20 -left-20 w-96 h-96 bg-blue-100 dark:bg-blue-900/10 rounded-full blur-3xl opacity-60 -z-10"/>

      <div class="max-w-4xl mx-auto text-center">
        <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 text-sm font-medium mb-8 border border-primary-200 dark:border-primary-800">
          <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"/>
          The #1 Marketplace for Students
        </div>

        <h1 class="text-5xl sm:text-6xl lg:text-7xl font-extrabold text-gray-900 dark:text-white leading-[1.1] tracking-tight mb-6">
          Find Talented<br>
          <span class="bg-gradient-to-r from-primary-600 to-blue-500 bg-clip-text text-transparent">Student Freelancers</span>
        </h1>

        <p class="text-lg sm:text-xl text-gray-500 dark:text-gray-400 mb-10 max-w-2xl mx-auto leading-relaxed">
          Connect with skilled students for tutoring, programming, design, and 50+ more services — at student-friendly prices.
        </p>

        <!-- Search -->
        <form @submit.prevent="search" class="flex gap-2 max-w-xl mx-auto mb-8">
          <div class="relative flex-1">
            <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <input v-model="searchQuery" type="text" placeholder='Try "math tutor" or "logo design"' class="input pl-10 py-3.5 shadow-sm" />
          </div>
          <button type="submit" class="btn-primary px-6 py-3.5 whitespace-nowrap shadow-glow">Search</button>
        </form>

        <div class="flex flex-wrap justify-center gap-x-2 gap-y-1 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-gray-400">Popular:</span>
          <button v-for="tag in popularSearches" :key="tag" @click="searchQuery = tag; search()" class="text-primary-600 dark:text-primary-400 hover:underline">{{ tag }}</button>
        </div>
      </div>

      <!-- Stats -->
      <div class="max-w-3xl mx-auto mt-20 grid grid-cols-3 gap-8">
        <div v-for="s in stats" :key="s.label" class="text-center">
          <div class="text-3xl font-extrabold text-gray-900 dark:text-white">{{ s.value }}</div>
          <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ s.label }}</div>
        </div>
      </div>
    </section>

    <!-- Categories -->
    <section id="categories" class="py-20 px-4 bg-gray-50 dark:bg-surface-900">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold text-gray-900 dark:text-white mb-3">Explore Categories</h2>
          <p class="text-gray-500 dark:text-gray-400">Find the perfect service for your needs</p>
        </div>
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          <RouterLink v-for="cat in categories" :key="cat.slug" :to="`/marketplace?category=${cat.slug}`"
            class="card p-5 text-center hover:shadow-card-hover hover:-translate-y-1 transition-all duration-200 group cursor-pointer">
            <div class="w-12 h-12 mx-auto mb-3 rounded-xl bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center group-hover:bg-primary-100 dark:group-hover:bg-primary-900/40 transition-colors">
              <component :is="CatIcon" :path="cat.iconPath" class="w-6 h-6 text-primary-600 dark:text-primary-400"/>
            </div>
            <div class="font-semibold text-gray-900 dark:text-white text-sm group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">{{ cat.name }}</div>
            <div class="text-xs text-gray-400 mt-1">{{ cat.count }}+ services</div>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- Featured Gigs -->
    <section class="py-20 px-4">
      <div class="max-w-7xl mx-auto">
        <div class="flex items-center justify-between mb-10">
          <div>
            <h2 class="text-3xl font-bold text-gray-900 dark:text-white">Featured Services</h2>
            <p class="text-gray-500 dark:text-gray-400 mt-1">Top-rated student freelancers</p>
          </div>
          <RouterLink to="/marketplace" class="text-sm font-medium text-primary-600 dark:text-primary-400 hover:underline flex items-center gap-1">
            View all
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </RouterLink>
        </div>
        <div v-if="featuredServices.length === 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <div v-for="i in 4" :key="i" class="card animate-pulse overflow-hidden">
            <div class="h-44 bg-gray-100 dark:bg-surface-800"/>
            <div class="p-4 space-y-3"><div class="h-3 bg-gray-100 dark:bg-surface-800 rounded w-1/3"/><div class="h-4 bg-gray-100 dark:bg-surface-800 rounded"/><div class="h-4 bg-gray-100 dark:bg-surface-800 rounded w-3/4"/></div>
          </div>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <ServiceCard v-for="s in featuredServices" :key="s.id" :service="s"/>
        </div>
      </div>
    </section>

    <!-- How It Works -->
    <section id="how-it-works" class="py-20 px-4 bg-gray-50 dark:bg-surface-900">
      <div class="max-w-5xl mx-auto">
        <div class="text-center mb-14">
          <h2 class="text-3xl font-bold text-gray-900 dark:text-white mb-3">How It Works</h2>
          <p class="text-gray-500 dark:text-gray-400">Simple. Secure. Student-powered.</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div v-for="(step, i) in steps" :key="i" class="text-center group">
            <div class="w-16 h-16 rounded-2xl bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center mx-auto mb-5 group-hover:scale-110 transition-transform">
              <svg class="w-7 h-7 text-primary-600 dark:text-primary-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75">
                <path :d="step.iconPath"/>
              </svg>
            </div>
            <div class="text-xs font-bold text-primary-600 dark:text-primary-400 uppercase tracking-widest mb-2">Step {{ i + 1 }}</div>
            <h3 class="font-bold text-gray-900 dark:text-white text-lg mb-2">{{ step.title }}</h3>
            <p class="text-gray-500 dark:text-gray-400 text-sm leading-relaxed">{{ step.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="py-24 px-4">
      <div class="max-w-3xl mx-auto text-center bg-gradient-to-r from-primary-600 to-blue-600 rounded-3xl p-12 shadow-2xl">
        <h2 class="text-4xl font-extrabold text-white mb-4">Ready to get started?</h2>
        <p class="text-primary-100 mb-10 text-lg">Join 10,000+ students already using StudentMarket</p>
        <div class="flex gap-4 justify-center flex-wrap">
          <RouterLink to="/register" class="px-8 py-3.5 bg-white text-primary-700 rounded-xl font-semibold hover:bg-gray-50 transition-colors shadow-lg">
            Sign Up Free
          </RouterLink>
          <RouterLink to="/marketplace" class="px-8 py-3.5 bg-primary-700/50 text-white border border-white/20 rounded-xl font-semibold hover:bg-primary-700/70 transition-colors">
            Browse Services
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="bg-gray-900 text-gray-400 py-12 px-4">
      <div class="max-w-7xl mx-auto grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
        <div>
          <div class="flex items-center gap-2 mb-4">
            <div class="w-7 h-7 bg-primary-600 rounded-lg flex items-center justify-center">
              <svg class="w-3.5 h-3.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </div>
            <span class="text-white font-bold">StudentMarket</span>
          </div>
          <p class="text-sm">The marketplace built by students, for students.</p>
        </div>
        <div v-for="col in footerLinks" :key="col.title">
          <div class="text-white font-semibold mb-4 text-sm">{{ col.title }}</div>
          <ul class="space-y-2">
            <li v-for="link in col.links" :key="link"><a href="#" class="text-sm hover:text-white transition-colors">{{ link }}</a></li>
          </ul>
        </div>
      </div>
      <div class="max-w-7xl mx-auto pt-8 border-t border-gray-800 text-sm text-center">
        © 2025 StudentMarket. All rights reserved.
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, defineComponent, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ServiceCard from '@/components/marketplace/ServiceCard.vue'
import { useMarketplaceStore } from '@/stores/marketplace'

const router = useRouter()
const store  = useMarketplaceStore()
const searchQuery    = ref('')
const featuredServices = ref<any[]>([])

const CatIcon = defineComponent({
  props: { path: String, class: String },
  setup(props) {
    return () => h('svg', { class: props.class, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [
      h('path', { d: props.path })
    ])
  }
})

const popularSearches = ['Math tutor', 'Logo design', 'Website', 'Video editing', 'Resume']

const stats = [
  { value: '10,000+', label: 'Students' },
  { value: '5,000+',  label: 'Services' },
  { value: '$500K+',  label: 'Earned by students' },
]

const categories = [
  { name: 'Programming', slug: 'programming', count: 320, iconPath: 'M16 18l6-6-6-6M8 6l-6 6 6 6' },
  { name: 'Design',      slug: 'design',      count: 280, iconPath: 'M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z' },
  { name: 'Tutoring',    slug: 'tutoring',    count: 450, iconPath: 'M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2zM22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z' },
  { name: 'Writing',     slug: 'writing',     count: 190, iconPath: 'M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z' },
  { name: 'Video',       slug: 'video',       count: 120, iconPath: 'M23 7l-7 5 7 5V7zM1 5h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H1a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2z' },
  { name: 'Photography', slug: 'photography', count: 85,  iconPath: 'M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2zM12 17a4 4 0 1 0 0-8 4 4 0 0 0 0 8z' },
  { name: 'Delivery',    slug: 'delivery',    count: 60,  iconPath: 'M5 17H3a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11a2 2 0 0 1 2 2v3m4 0h2l3 5v3h-6m0 0a2 2 0 1 0 4 0 2 2 0 0 0-4 0m-6 0a2 2 0 1 0 4 0 2 2 0 0 0-4 0' },
  { name: 'Fitness',     slug: 'fitness',     count: 75,  iconPath: 'M18 20V10M12 20V4M6 20v-6' },
  { name: 'Music',       slug: 'music',       count: 95,  iconPath: 'M9 18V5l12-2v13M9 18a3 3 0 1 0-3 3 3 3 0 0 0 3-3zm12 0a3 3 0 1 0-3 3 3 3 0 0 0 3-3z' },
  { name: 'Business',    slug: 'business',    count: 110, iconPath: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12' },
]

const steps = [
  { iconPath: 'M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z', title: 'Find a service', desc: 'Browse hundreds of student-powered services across 10+ categories.' },
  { iconPath: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 0 2 2z',  title: 'Chat & Order',   desc: 'Message sellers, discuss requirements, and place your order securely.' },
  { iconPath: 'M22 11.08V12a10 10 0 1 1-5.93-9.14M22 4L12 14.01l-3-3',           title: 'Get it done',   desc: 'Receive your work, request revisions if needed, and release payment.' },
]

const footerLinks = [
  { title: 'For Buyers',  links: ['How to buy', 'Categories', 'Pricing', 'Trust & Safety'] },
  { title: 'For Sellers', links: ['Become a seller', 'Seller FAQ', 'Success tips', 'Pro badge'] },
  { title: 'Company',     links: ['About', 'Blog', 'Careers', 'Contact'] },
]

function search() {
  router.push({ name: 'marketplace', query: { q: searchQuery.value } })
}

onMounted(async () => {
  const data = await store.fetchFeatured()
  featuredServices.value = data.slice(0, 8)
})
</script>
