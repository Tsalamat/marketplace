<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="max-w-3xl mx-auto px-4 sm:px-6 pt-20 pb-12">
      <div class="py-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Create a New Gig</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Fill in the details to list your service</p>
      </div>

      <div v-if="error" class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl text-sm text-red-600 dark:text-red-400">
        {{ error }}
      </div>

      <form @submit.prevent="submit" class="space-y-6">
        <!-- Basic Info -->
        <div class="card p-6 space-y-5">
          <h2 class="font-bold text-gray-900 dark:text-white">Basic Information</h2>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Service Title</label>
            <input v-model="form.title" required minlength="10" maxlength="200" class="input" placeholder="I will create a professional logo for your brand" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Category</label>
            <select v-model="form.category_id" required class="input">
              <option value="">Select category...</option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Description</label>
            <textarea v-model="form.description" required minlength="50" rows="6" class="input" placeholder="Describe your service in detail. What will the buyer get? What makes you qualified?" />
            <p class="text-xs text-gray-400 mt-1">{{ form.description.length }}/5000 characters (min. 50)</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tags (comma-separated)</label>
            <input v-model="tagsInput" class="input" placeholder="logo, design, branding" />
          </div>
        </div>

        <!-- Gallery -->
        <div class="card p-6 space-y-4">
          <h2 class="font-bold text-gray-900 dark:text-white">Photos / Gallery</h2>
          <p class="text-sm text-gray-500 dark:text-gray-400">Add up to 5 photos showcasing your work. First image becomes the cover.</p>

          <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <!-- Uploaded images -->
            <div
              v-for="(url, i) in form.gallery"
              :key="url"
              class="relative aspect-video rounded-xl overflow-hidden border border-gray-200 dark:border-surface-700 group"
            >
              <img :src="url" class="w-full h-full object-cover" />
              <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                <span v-if="i === 0" class="absolute top-2 left-2 text-xs bg-primary-600 text-white rounded px-2 py-0.5 font-medium">Cover</span>
                <button type="button" @click="removeImage(i)" class="w-8 h-8 rounded-full bg-red-500 text-white flex items-center justify-center hover:bg-red-600">
                  <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Upload button -->
            <label
              v-if="form.gallery.length < 5"
              :class="['aspect-video rounded-xl border-2 border-dashed border-gray-300 dark:border-surface-700 flex flex-col items-center justify-center gap-2 cursor-pointer transition-colors', uploadingImage ? 'opacity-50 cursor-not-allowed' : 'hover:border-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/10']"
            >
              <svg v-if="!uploadingImage" class="w-8 h-8 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
              </svg>
              <svg v-else class="w-6 h-6 text-primary-500 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
              <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">{{ uploadingImage ? 'Uploading...' : 'Add Photo' }}</span>
              <input type="file" class="hidden" accept="image/*" :disabled="uploadingImage" @change="uploadImage" />
            </label>
          </div>
        </div>

        <!-- Packages -->
        <div class="card p-6 space-y-5">
          <h2 class="font-bold text-gray-900 dark:text-white">Pricing Packages</h2>
          <div v-for="(pkg, i) in form.packages" :key="i" class="border border-gray-200 dark:border-surface-700 rounded-xl p-4 space-y-4">
            <div class="flex items-center gap-2">
              <span class="badge-primary capitalize">{{ pkg.name }}</span>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Package Title</label>
                <input v-model="pkg.title" required class="input" :placeholder="`${pkg.name.charAt(0).toUpperCase() + pkg.name.slice(1)} Package`" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Price (USD)</label>
                <input v-model.number="pkg.price" type="number" min="1" required class="input" placeholder="25" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Delivery Days</label>
                <input v-model.number="pkg.delivery_days" type="number" min="1" required class="input" placeholder="3" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Revisions</label>
                <input v-model.number="pkg.revisions" type="number" min="0" required class="input" placeholder="2" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Description</label>
              <textarea v-model="pkg.description" rows="2" class="input" placeholder="What's included in this package?" />
            </div>
          </div>
        </div>

        <!-- FAQs -->
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="font-bold text-gray-900 dark:text-white">FAQ (optional)</h2>
            <button type="button" @click="addFaq" class="btn-ghost text-sm">+ Add FAQ</button>
          </div>
          <div v-for="(faq, i) in form.faqs" :key="i" class="space-y-3 border border-gray-200 dark:border-surface-700 rounded-xl p-4">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-600 dark:text-gray-400">FAQ {{ i + 1 }}</span>
              <button type="button" @click="form.faqs.splice(i, 1)" class="text-red-500 text-sm hover:underline">Remove</button>
            </div>
            <input v-model="faq.question" class="input" placeholder="What question do buyers ask?" />
            <textarea v-model="faq.answer" rows="2" class="input" placeholder="Your answer..." />
          </div>
        </div>

        <button type="submit" :disabled="loading" class="btn-primary w-full py-4 text-base font-semibold">
          {{ loading ? 'Publishing...' : 'Publish Gig' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import api from '@/api/axios'
import { toast } from 'vue3-toastify'

const router = useRouter()
const store  = useMarketplaceStore()

const loading        = ref(false)
const uploadingImage = ref(false)
const error          = ref('')
const categories     = ref<any[]>([])
const tagsInput      = ref('')

const form = ref({
  title: '', category_id: '', description: '',
  packages: [
    { name: 'basic',    title: '', price: 10, delivery_days: 3, revisions: 1, currency: 'USD', description: '' },
    { name: 'standard', title: '', price: 25, delivery_days: 5, revisions: 2, currency: 'USD', description: '' },
    { name: 'premium',  title: '', price: 50, delivery_days: 7, revisions: 3, currency: 'USD', description: '' },
  ],
  faqs:    [] as { question: string; answer: string }[],
  gallery: [] as string[],
})

function addFaq() {
  form.value.faqs.push({ question: '', answer: '' })
}

function removeImage(i: number) {
  form.value.gallery.splice(i, 1)
}

async function uploadImage(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (form.value.gallery.length >= 5) { toast.warning('Max 5 photos'); return }

  uploadingImage.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    const { data } = await api.post('/api/v1/upload/image', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    form.value.gallery.push(data.url)
    toast.success('Photo added!')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Upload failed')
  } finally {
    uploadingImage.value = false
    ;(e.target as HTMLInputElement).value = ''
  }
}

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const payload = {
      ...form.value,
      tags: tagsInput.value.split(',').map(t => t.trim()).filter(Boolean),
    }
    const data = await store.createService(payload)
    toast.success('Gig published!')
    router.push(`/gig/${data.slug}`)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to create gig'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  categories.value = await store.fetchCategories()
})
</script>
