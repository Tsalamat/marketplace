import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/api/axios'

export interface ServicePackage {
  id: string
  name: 'basic' | 'standard' | 'premium'
  title: string
  description: string
  price: number
  currency: string
  delivery_days: number
  revisions: number
  features: string[]
}

export interface Service {
  id: string
  seller_id: string
  category_id: string
  title: string
  slug: string
  description: string
  tags: string[]
  gallery: string[]
  is_active: boolean
  is_featured: boolean
  views: number
  orders_count: number
  rating: number
  total_reviews: number
  created_at: string
  seller?: any
  category?: any
  packages?: ServicePackage[]
  faqs?: any[]
  reviews?: any[]
}

export interface ServiceFilters {
  q?: string
  category?: string
  min_price?: number
  max_price?: number
  min_rating?: number
  sort?: string
  page?: number
  limit?: number
}

export const useMarketplaceStore = defineStore('marketplace', () => {
  const services = ref<Service[]>([])
  const featured = ref<Service[]>([])
  const categories = ref<any[]>([])
  const currentService = ref<Service | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const filters = ref<ServiceFilters>({ sort: 'trending', page: 1, limit: 20 })

  async function fetchServices(f?: ServiceFilters) {
    if (f) filters.value = { ...filters.value, ...f }
    loading.value = true
    try {
      const params = Object.fromEntries(
        Object.entries(filters.value).filter(([, v]) => v !== undefined && v !== '')
      )
      const { data } = await api.get('/api/v1/services', { params })
      services.value = data.data
      total.value = data.total
      return data
    } finally {
      loading.value = false
    }
  }

  async function fetchFeatured() {
    const { data } = await api.get('/api/v1/services/featured')
    featured.value = data
    return data
  }

  async function fetchCategories() {
    const { data } = await api.get('/api/v1/services/categories')
    categories.value = data
    return data
  }

  async function fetchService(slug: string) {
    loading.value = true
    try {
      const { data } = await api.get(`/api/v1/services/${slug}`)
      currentService.value = data
      return data
    } finally {
      loading.value = false
    }
  }

  async function createService(payload: any) {
    const { data } = await api.post('/api/v1/services', payload)
    return data
  }

  function setFilter(key: keyof ServiceFilters, value: any) {
    filters.value = { ...filters.value, [key]: value, page: 1 }
  }

  function nextPage() {
    filters.value.page = (filters.value.page || 1) + 1
  }

  return {
    services, featured, categories, currentService, total, loading, filters,
    fetchServices, fetchFeatured, fetchCategories, fetchService, createService, setFilter, nextPage,
  }
})
