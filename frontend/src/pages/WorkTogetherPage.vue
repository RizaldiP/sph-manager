<template>
  <div>
    <PageHeader title="Work Together" subtitle="Kolaborasi real-time dokumen SPH di jaringan lokal">
      <template #actions>
        <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="openCreate">
          + Mulai Room Baru
        </button>
        <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="manualOpen = true">
          Gabung via IP…
        </button>
      </template>
    </PageHeader>

    <!-- Status bar -->
    <div v-if="collabStore.isLive" class="mb-4 flex items-center gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3">
      <span class="flex items-center gap-1.5 text-[13px] font-semibold text-emerald-700">
        <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
        LIVE
      </span>
      <span class="text-[13px] text-emerald-600">{{ collabStore.roomName }} · {{ collabStore.participantCount }} pengguna</span>
      <button type="button" class="ml-auto rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-emerald-700" @click="goToDoc">
        Buka Dokumen →
      </button>
      <button type="button" class="rounded-lg border border-red-200 px-3 py-1.5 text-xs font-medium text-red-700 transition-colors hover:bg-red-50" @click="handleLeave">
        {{ collabStore.isHost ? 'Tutup Room' : 'Keluar' }}
      </button>
    </div>

    <!-- Error -->
    <p v-if="collabStore.error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ collabStore.error }}</p>

    <!-- Discovered rooms -->
    <div class="rounded-xl border border-slate-200 bg-white p-5">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-[13px] font-semibold text-slate-700">Room di Jaringan Lokal</h2>
        <button type="button" class="rounded-lg border border-slate-200 px-2.5 py-1.5 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-50" @click="refreshList">
          Refresh
        </button>
      </div>

      <div v-if="!discovered.length" class="py-10 text-center">
        <svg class="mx-auto mb-3 h-10 w-10 text-slate-300" fill="none" viewBox="0 0 24 24" stroke-width="1.2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z" />
        </svg>
        <p class="text-[13px] text-slate-400">Belum ada room ditemukan di jaringan ini.</p>
        <p class="mt-1 text-xs text-slate-400">Mulai room baru atau gabung via IP manual.</p>
      </div>

      <ul v-else class="divide-y divide-slate-100">
        <li v-for="r in discovered" :key="r.roomId" class="flex items-center gap-4 px-2 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-[13px] font-semibold text-slate-800">{{ r.roomName }}</p>
            <p class="text-xs text-slate-400">{{ r.documentNumber }} · {{ r.projectName }}</p>
            <p class="text-xs text-slate-400">Host: {{ r.hostName }} ({{ r.hostIP }}) · {{ r.users }} pengguna</p>
          </div>
          <button type="button" class="shrink-0 rounded-lg bg-brand-600 px-3 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="openJoin(r)">
            Join
          </button>
        </li>
      </ul>
    </div>

    <!-- Dialogs -->
    <CreateRoomDialog v-model="createOpen" @started="afterCreate" />
    <JoinDialog v-model="joinOpen" :room="joinTarget" @joined="afterJoin" />
    <JoinManualDialog v-model="manualOpen" @joined="afterJoin" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import CreateRoomDialog from '../components/collaboration/CreateRoomDialog.vue'
import JoinDialog from '../components/collaboration/JoinDialog.vue'
import JoinManualDialog from '../components/collaboration/JoinManualDialog.vue'
import { useCollaborationStore } from '../stores/collaboration'
import type { DiscoveredRoom } from '../types/collaboration'

const router = useRouter()
const collabStore = useCollaborationStore()

const createOpen = ref(false)
const joinOpen = ref(false)
const joinTarget = ref<DiscoveredRoom | null>(null)
const manualOpen = ref(false)

const discovered = computed(() => collabStore.discovered)

function openCreate() {
  createOpen.value = true
}

function openJoin(room: DiscoveredRoom) {
  joinTarget.value = room
  joinOpen.value = true
}

function afterCreate() {
  const docId = collabStore.sphDocumentId
  if (docId) {
    router.push(`/sph/${docId}/edit`)
  }
}

function afterJoin() {
  const docId = collabStore.sphDocumentId
  if (docId) {
    router.push(`/sph/${docId}/edit`)
  }
}

function goToDoc() {
  const docId = collabStore.sphDocumentId
  if (docId) {
    router.push(`/sph/${docId}/edit`)
  }
}

async function handleLeave() {
  if (collabStore.isHost) {
    await collabStore.closeRoom()
  } else {
    await collabStore.leaveRoom()
  }
}

let refreshTimer: ReturnType<typeof setInterval> | null = null

async function refreshList() {
  await collabStore.refreshDiscovered()
}

onMounted(async () => {
  await collabStore.loadDefaults()
  await collabStore.startDiscovery()
  await refreshList()
  refreshTimer = setInterval(refreshList, 5000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  collabStore.stopDiscovery()
})
</script>
