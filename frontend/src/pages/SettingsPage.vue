<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-2xl mx-auto px-4 sm:px-6 pt-20 pb-16">
      <div class="py-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Settings</h1>
      </div>

      <!-- Success banner -->
      <div v-if="saved" class="mb-5 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-xl flex items-center gap-3">
        <svg class="w-5 h-5 text-green-600 dark:text-green-400 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20 6 9 17 4 12"/>
        </svg>
        <span class="text-sm text-green-700 dark:text-green-300 font-medium">Changes saved successfully!</span>
      </div>

      <form @submit.prevent="save" class="space-y-6">

        <!-- Avatar -->
        <div class="card p-6">
          <h2 class="font-bold text-gray-900 dark:text-white mb-5">Profile Photo</h2>
          <div class="flex items-center gap-5">
            <div class="relative flex-shrink-0">
              <img :src="avatarPreview" class="w-20 h-20 rounded-2xl object-cover ring-2 ring-gray-200 dark:ring-surface-700"/>
              <div v-if="uploadingAvatar" class="absolute inset-0 bg-black/50 rounded-2xl flex items-center justify-center">
                <svg class="w-6 h-6 text-white animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
              </div>
            </div>
            <div>
              <label class="btn-secondary text-sm cursor-pointer inline-flex items-center gap-2">
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
                </svg>
                Upload Photo
                <input type="file" class="hidden" accept="image/*" @change="onAvatarChange"/>
              </label>
              <p class="text-xs text-gray-400 mt-1.5">JPG, PNG, WebP · Max 5MB</p>
            </div>
          </div>
        </div>

        <!-- Profile info -->
        <div class="card p-6 space-y-4">
          <h2 class="font-bold text-gray-900 dark:text-white">Profile Information</h2>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">First Name</label>
              <input v-model="form.first_name" class="input"/>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Last Name</label>
              <input v-model="form.last_name" class="input"/>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tagline</label>
            <input v-model="form.tagline" class="input" placeholder="Full-stack dev & Math tutor"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Bio</label>
            <textarea v-model="form.bio" rows="3" class="input" placeholder="Tell buyers about yourself…"/>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">University</label>
              <input v-model="form.university" class="input"/>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Department</label>
              <input v-model="form.department" class="input"/>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Location (city)</label>
            <input v-model="form.location" class="input" placeholder="Almaty, Kazakhstan"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Skills <span class="text-gray-400 font-normal">(comma-separated)</span></label>
            <input v-model="skillsInput" class="input" placeholder="React, Python, Figma…"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Languages</label>
            <input v-model="langInput" class="input" placeholder="English, Russian, Kazakh…"/>
          </div>
        </div>

        <!-- Links -->
        <div class="card p-6 space-y-4">
          <h2 class="font-bold text-gray-900 dark:text-white">Links</h2>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">GitHub</label>
            <input v-model="form.github_url" class="input" placeholder="https://github.com/you"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">LinkedIn</label>
            <input v-model="form.linkedin_url" class="input" placeholder="https://linkedin.com/in/you"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Portfolio</label>
            <input v-model="form.portfolio_url" class="input" placeholder="https://yoursite.com"/>
          </div>
        </div>

        <!-- Password -->
        <div class="card p-6 space-y-4">
          <h2 class="font-bold text-gray-900 dark:text-white">Change Password</h2>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Current Password</label>
            <input v-model="pass.current" type="password" class="input"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">New Password</label>
            <input v-model="pass.new" type="password" minlength="8" class="input"/>
          </div>
          <button type="button" @click="changePassword"
            :disabled="!pass.current || !pass.new || changingPass"
            class="btn-secondary text-sm">
            <svg v-if="changingPass" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
          <span>{{ changingPass ? 'Updating…' : 'Update Password' }}</span>
          </button>
        </div>

        <!-- Danger zone -->
        <div class="card p-6 border border-red-200 dark:border-red-900/50">
          <h2 class="font-bold text-red-600 dark:text-red-400 mb-2">Account</h2>
          <button type="button" @click="logout"
            class="flex items-center gap-2 text-sm text-red-600 dark:text-red-400 hover:underline">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
            </svg>
            Sign Out
          </button>
        </div>

        <button type="submit" :disabled="loading" class="btn-primary w-full py-3.5 text-base">
          <svg v-if="loading" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
          <span>{{ loading ? 'Saving…' : 'Save Changes' }}</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const auth   = useAuthStore()
const chat   = useChatStore()
const router = useRouter()
const loading        = ref(false)
const saved          = ref(false)
const changingPass   = ref(false)
const uploadingAvatar = ref(false)
const skillsInput    = ref('')
const langInput      = ref('')

const form = ref({
  first_name: '', last_name: '', tagline: '', bio: '',
  university: '', department: '', location: '',
  github_url: '', linkedin_url: '', portfolio_url: '',
  avatar_url: '',
})
const pass = ref({ current: '', new: '' })

const avatarPreview = computed(() =>
  form.value.avatar_url ||
  auth.user?.profile?.avatar_url ||
  `https://ui-avatars.com/api/?name=${auth.user?.username}&size=80&background=2563eb&color=fff`
)

async function onAvatarChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploadingAvatar.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    const { data } = await api.post('/api/v1/upload/avatar', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    form.value.avatar_url = data.url
    toast.success('Avatar uploaded!')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Upload failed')
  } finally {
    uploadingAvatar.value = false
  }
}

async function save() {
  loading.value = true
  saved.value   = false
  try {
    await auth.updateProfile({
      ...form.value,
      skills:    skillsInput.value.split(',').map(s => s.trim()).filter(Boolean),
      languages: langInput.value.split(',').map(s => s.trim()).filter(Boolean),
    } as any)
    saved.value = true
    setTimeout(() => saved.value = false, 4000)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Save failed')
  } finally {
    loading.value = false
  }
}

async function changePassword() {
  changingPass.value = true
  try {
    await api.post('/api/v1/auth/change-password', {
      current_password: pass.value.current,
      new_password:     pass.value.new,
    })
    toast.success('Password updated!')
    pass.value = { current: '', new: '' }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed')
  } finally {
    changingPass.value = false
  }
}

async function logout() {
  await auth.logout()
  chat.disconnectWS()
  toast.info('Signed out')
  router.push('/')
}

onMounted(() => {
  const p = auth.user?.profile
  if (p) {
    form.value.first_name    = p.first_name    || ''
    form.value.last_name     = p.last_name     || ''
    form.value.tagline       = p.tagline       || ''
    form.value.bio           = p.bio           || ''
    form.value.university    = p.university    || ''
    form.value.department    = p.department    || ''
    form.value.location      = p.location      || ''
    form.value.github_url    = p.github_url    || ''
    form.value.linkedin_url  = p.linkedin_url  || ''
    form.value.portfolio_url = p.portfolio_url || ''
    form.value.avatar_url    = p.avatar_url    || ''
    skillsInput.value = p.skills?.join(', ')    || ''
    langInput.value   = p.languages?.join(', ') || ''
  }
})
</script>
