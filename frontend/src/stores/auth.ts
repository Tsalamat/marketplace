import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/axios'

export interface User {
  id: string
  email: string
  username: string
  role: 'admin' | 'seller' | 'buyer'
  is_verified: boolean
  email_verified: boolean
  last_active: string
  created_at: string
  profile?: Profile
}

export interface Profile {
  first_name: string
  last_name: string
  avatar_url: string
  cover_url: string
  bio: string
  tagline: string
  skills: string[]
  languages: string[]
  location: string
  university: string
  department: string
  year_of_study: number
  rating: number
  total_reviews: number
  completed_jobs: number
  is_online: boolean
  currency_pref: 'USD' | 'KZT' | 'EUR'
  github_url?: string
  linkedin_url?: string
  portfolio_url?: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)

  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isSeller = computed(() => user.value?.role === 'seller' || user.value?.role === 'admin')
  const fullName = computed(() => {
    const p = user.value?.profile
    if (!p) return user.value?.username || ''
    return `${p.first_name} ${p.last_name}`.trim() || user.value?.username || ''
  })

  async function initAuth() {
    const token = localStorage.getItem('access_token')
    if (!token) {
      initialized.value = true
      return
    }
    try {
      const { data } = await api.get('/api/v1/auth/me')
      user.value = data
    } catch {
      clearSession()
    } finally {
      initialized.value = true
    }
  }

  async function register(payload: {
    email: string
    username: string
    password: string
    first_name: string
    last_name: string
    university?: string
  }) {
    const { data } = await api.post('/api/v1/auth/register', payload)
    setSession(data)
    return data
  }

  async function login(email: string, password: string) {
    const { data } = await api.post('/api/v1/auth/login', { email, password })
    setSession(data)
    return data
  }

  async function logout() {
    try { await api.post('/api/v1/auth/logout') } catch {}
    clearSession()
  }

  async function updateProfile(updates: Partial<Profile>) {
    const { data } = await api.put('/api/v1/users/profile', updates)
    if (user.value) user.value.profile = { ...user.value.profile, ...data }
    return data
  }

  function setSession(data: { user: User; access_token: string; refresh_token: string }) {
    user.value = data.user
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
  }

  function clearSession() {
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return {
    user, initialized, isAuthenticated, isAdmin, isSeller, fullName,
    initAuth, register, login, logout, updateProfile, clearSession,
  }
})
