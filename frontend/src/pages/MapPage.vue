<template>
  <div class="min-h-screen bg-gray-50 dark:bg-surface-950">
    <AppNavbar />
    <div class="pt-16 flex flex-col" style="height:100vh">
      <!-- Header -->
      <div class="bg-white dark:bg-surface-900 border-b border-gray-100 dark:border-surface-800 px-4 py-3 flex items-center justify-between flex-shrink-0">
        <div>
          <h1 class="font-bold text-gray-900 dark:text-white">Friends Map</h1>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ mapFriends.length }} online · sharing location</p>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2">
            <div :class="['w-2 h-2 rounded-full', sharing ? 'bg-green-500 animate-pulse' : 'bg-gray-300']"/>
            <span class="text-xs text-gray-600 dark:text-gray-400">{{ sharing ? 'Sharing' : 'Location off' }}</span>
          </div>
          <button @click="toggleSharing"
            :class="['text-xs px-3 py-1.5 rounded-lg font-medium transition-colors', sharing ? 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400' : 'btn-primary']">
            {{ sharing ? 'Stop Sharing' : 'Share Location' }}
          </button>
        </div>
      </div>

      <!-- Map + sidebar -->
      <div class="flex-1 relative">
        <div ref="mapEl" class="w-full h-full"/>

        <!-- Friend list -->
        <div class="absolute top-4 right-4 w-56 bg-white/95 dark:bg-surface-900/95 backdrop-blur-sm rounded-2xl shadow-card border border-gray-100 dark:border-surface-800 overflow-hidden z-[1000]">
          <div class="px-4 py-3 border-b border-gray-100 dark:border-surface-800">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">On Map ({{ mapFriends.length }})</p>
          </div>
          <div class="max-h-64 overflow-y-auto">
            <div v-if="!mapFriends.length" class="px-4 py-6 text-center text-xs text-gray-400">
              No friends sharing location
            </div>
            <button v-for="f in mapFriends" :key="f.user_id"
              @click="flyTo(f)"
              class="w-full flex items-center gap-3 px-4 py-3 hover:bg-gray-50 dark:hover:bg-surface-800 transition-colors text-left">
              <img :src="f.avatar || `https://ui-avatars.com/api/?name=${f.username}&size=32&background=2563eb&color=fff`"
                class="w-8 h-8 rounded-full object-cover flex-shrink-0"/>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ f.username }}</p>
                <p class="text-xs text-gray-400">{{ f.lat.toFixed(4) }}, {{ f.lng.toFixed(4) }}</p>
              </div>
              <div class="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"/>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import AppNavbar from '@/components/common/AppNavbar.vue'
import { useChatStore } from '@/stores/chat'
import api from '@/api/axios'

// Fix Leaflet default marker icon paths broken by Vite bundling
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'
delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({ iconUrl: markerIcon, iconRetinaUrl: markerIcon2x, shadowUrl: markerShadow })

const chat       = useChatStore()
const mapEl      = ref<HTMLDivElement>()
const sharing    = ref(false)
const mapFriends = ref<any[]>([])

let map: L.Map
const markers = new Map<string, L.Marker>()
let myMarker: L.CircleMarker | null = null
let watchId: number | null = null
let pollTimer: ReturnType<typeof setInterval>

function initMap() {
  if (!mapEl.value) return
  map = L.map(mapEl.value, { zoomControl: true }).setView([51.505, -0.09], 5)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 19,
  }).addTo(map)
}

async function loadFriendLocations() {
  try {
    const { data } = await api.get('/api/v1/friends/locations')
    mapFriends.value = data
    data.forEach((f: any) => updateMarker(f))
  } catch {}
}

function updateMarker(f: any) {
  if (!map || !f.lat || !f.lng) return
  const pos: L.LatLngTuple = [f.lat, f.lng]
  const icon = L.divIcon({
    html: `<img src="${f.avatar || `https://ui-avatars.com/api/?name=${f.username}&size=40&background=2563eb&color=fff`}"
      style="width:36px;height:36px;border-radius:50%;border:2px solid #2563eb;object-fit:cover"/>`,
    iconSize: [36, 36],
    iconAnchor: [18, 18],
    className: '',
  })
  if (markers.has(f.user_id)) {
    markers.get(f.user_id)!.setLatLng(pos)
  } else {
    const m = L.marker(pos, { icon })
      .addTo(map)
      .bindPopup(`<strong>${f.username}</strong><br><span style="color:#6b7280;font-size:12px">Sharing live location</span>`)
    markers.set(f.user_id, m)
  }
}

function flyTo(f: any) {
  if (!map || !f.lat || !f.lng) return
  map.flyTo([f.lat, f.lng], 14)
}

function toggleSharing() {
  sharing.value ? stopSharing() : startSharing()
}

function startSharing() {
  if (!navigator.geolocation) { alert('Geolocation not supported'); return }
  sharing.value = true
  watchId = navigator.geolocation.watchPosition(
    async ({ coords: { latitude: lat, longitude: lng } }) => {
      const pos: L.LatLngTuple = [lat, lng]
      if (!myMarker) {
        myMarker = L.circleMarker(pos, {
          radius: 10, color: '#2563EB', fillColor: '#2563EB',
          fillOpacity: 1, weight: 3,
        }).addTo(map).bindPopup('You')
        map.flyTo(pos, 13)
      } else {
        myMarker.setLatLng(pos)
      }
      try { await api.patch('/api/v1/users/location', { lat, lng }) } catch {}
      chat.sendLocation(lat, lng)
    },
    (err) => { console.warn('Geolocation error', err); stopSharing() },
    { enableHighAccuracy: true, maximumAge: 5000, timeout: 10000 }
  )
}

function stopSharing() {
  sharing.value = false
  if (watchId !== null) { navigator.geolocation.clearWatch(watchId); watchId = null }
  if (myMarker) { myMarker.remove(); myMarker = null }
  api.patch('/api/v1/users/location', { lat: 0, lng: 0 }).catch(() => {})
}

function handleWSLocation(event: any) {
  if (event.type !== 'location') return
  const { user_id, lat, lng, username, avatar } = event.payload
  const existing = mapFriends.value.find(f => f.user_id === user_id)
  if (existing) { existing.lat = lat; existing.lng = lng }
  else mapFriends.value.push({ user_id, username, avatar, lat, lng })
  updateMarker({ user_id, username, avatar, lat, lng })
}

onMounted(async () => {
  initMap()
  await loadFriendLocations()
  pollTimer = setInterval(loadFriendLocations, 10_000)
  chat.onWSMessage(handleWSLocation)
})

onUnmounted(() => {
  stopSharing()
  clearInterval(pollTimer)
  chat.offWSMessage(handleWSLocation)
  map?.remove()
})
</script>
