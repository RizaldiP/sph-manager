<template>
  <div class="rounded-xl border border-slate-200 bg-white p-4">
    <!-- Header -->
    <div class="mb-3 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span class="flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold" :class="live ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">
          <span class="h-1.5 w-1.5 rounded-full" :class="live ? 'bg-emerald-500' : 'bg-slate-400'"></span>
          {{ live ? 'LIVE' : 'OFFLINE' }}
        </span>
        <span class="text-xs font-semibold text-slate-500">{{ store.modeLabel }}</span>
      </div>
      <button v-if="live" type="button" :disabled="leaving" class="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs font-medium text-red-700 transition-colors hover:bg-red-50 disabled:opacity-50" @click="handleLeave">
        {{ leaving ? 'Menutup...' : (store.isHost ? 'Tutup Room' : 'Keluar') }}
      </button>
    </div>

    <!-- Room info -->
    <div v-if="snap.room" class="mb-3 space-y-1 text-[13px]">
      <p class="font-semibold text-slate-800">{{ snap.room.roomName }}</p>
      <p class="text-xs text-slate-400">{{ snap.room.documentNumber }} · {{ snap.room.projectName }}</p>
    </div>

    <!-- Connection info (host only) -->
    <div v-if="store.isHost && snap.room" class="mb-3 rounded-lg border border-brand-200 bg-brand-50 p-3">
      <div class="mb-2 flex items-center justify-between">
        <h4 class="text-[13px] font-semibold text-brand-700">Untuk Bergabung</h4>
        <button type="button" class="rounded p-1 text-brand-500 transition-colors hover:bg-brand-100" :title="showConnInfo ? 'Sembunyikan' : 'Tampilkan'" @click="showConnInfo = !showConnInfo">
          <svg v-if="showConnInfo" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
          </svg>
          <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
      </div>

      <template v-if="showConnInfo">
        <div v-if="snap.room.hostIPs && snap.room.hostIPs.length" class="mb-2">
          <p class="mb-1 text-[11px] font-medium uppercase tracking-wide text-brand-600">Alamat IP</p>
          <div v-for="ip in snap.room.hostIPs" :key="ip" class="flex items-center gap-1.5">
            <code class="flex-1 rounded bg-white px-2 py-1 font-mono text-[13px] font-semibold text-slate-800 ring-1 ring-brand-200">{{ ip }}</code>
            <button type="button" class="shrink-0 rounded p-1.5 text-slate-400 transition-colors hover:bg-white hover:text-brand-600" title="Salin" @click="copyToClipboard(ip)">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            </button>
          </div>
        </div>
        <div v-else class="mb-2">
          <p class="mb-1 text-[11px] font-medium uppercase tracking-wide text-brand-600">Alamat IP</p>
          <p class="text-[12px] italic text-brand-400">Tidak terdeteksi</p>
        </div>

        <div class="mb-2">
          <p class="mb-1 text-[11px] font-medium uppercase tracking-wide text-brand-600">Port</p>
          <div class="flex items-center gap-1.5">
            <code class="flex-1 rounded bg-white px-2 py-1 font-mono text-[13px] font-semibold text-slate-800 ring-1 ring-brand-200">{{ snap.room.port }}</code>
            <button type="button" class="shrink-0 rounded p-1.5 text-slate-400 transition-colors hover:bg-white hover:text-brand-600" title="Salin" @click="copyToClipboard(String(snap.room.port))">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            </button>
          </div>
        </div>

        <div>
          <p class="mb-1 text-[11px] font-medium uppercase tracking-wide text-brand-600">Access Code</p>
          <div class="flex items-center gap-1.5">
            <code class="flex-1 rounded bg-white px-2 py-1 font-mono text-[13px] font-semibold text-slate-800 ring-1 ring-brand-200">{{ snap.room.accessCode }}</code>
            <button type="button" class="shrink-0 rounded p-1.5 text-slate-400 transition-colors hover:bg-white hover:text-brand-600" title="Salin" @click="copyToClipboard(snap.room.accessCode ?? '')">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            </button>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="space-y-1.5">
          <div class="flex items-center justify-between">
            <span class="text-[12px] text-brand-600">IP</span>
            <span class="font-mono text-[12px] text-brand-400">••••••••</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-[12px] text-brand-600">Port</span>
            <span class="font-mono text-[12px] text-brand-400">••••</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-[12px] text-brand-600">Access Code</span>
            <span class="font-mono text-[12px] text-brand-400">••••••</span>
          </div>
        </div>
      </template>

      <p v-if="copied" class="mt-2 text-center text-[11px] font-medium text-emerald-600">Tersalin!</p>
    </div>

    <!-- Connection status -->
    <div v-if="snap.connection && snap.connection !== 'CONNECTED'" class="mb-3 flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[13px] text-amber-700">
      <span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
      {{ connLabel }}
    </div>

    <!-- Firewall warning -->
    <div v-if="snap.room?.firewallWarning" class="mb-3 flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[13px] text-amber-700">
      <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126z" />
      </svg>
      <span>{{ snap.room.firewallWarning }}</span>
    </div>

    <!-- Error / notice -->
    <div v-if="snap.error" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ snap.error }}</div>
    <div v-if="snap.notice" class="mb-3 rounded-lg border border-blue-200 bg-blue-50 px-3 py-2 text-[13px] text-blue-700">{{ snap.notice }}</div>

    <!-- Turn Assignment Panel (Host only) -->
    <div v-if="store.isHost && live" class="mb-3 rounded-lg border border-slate-200 bg-slate-50 p-3">
      <h4 class="mb-2 text-[13px] font-semibold text-slate-700">Atur Hak Edit</h4>
      <div v-if="!otherParticipants.length" class="text-[12px] italic text-slate-400">
        Menunggu client terhubung...
      </div>
      <template v-else>
        <div v-for="p in otherParticipants" :key="p.id" class="mb-2 last:mb-0">
          <p class="mb-1 text-[12px] font-medium text-slate-600">{{ p.displayName }}</p>
          <div class="flex flex-wrap gap-1.5">
            <label v-for="sec in allSections" :key="sec.id" class="flex items-center gap-1 rounded border border-slate-200 bg-white px-2 py-1 text-[11px] text-slate-600 transition-colors hover:bg-slate-50">
              <input type="checkbox" :checked="isAssigned(p.id, sec.id)" class="h-3 w-3 rounded border-slate-300 text-brand-600 focus:ring-brand-500" @change="toggleAssignment(p.id, sec.id, $event)" />
              {{ sec.label }}
            </label>
          </div>
        </div>
        <button type="button" class="mt-2 w-full rounded-lg border border-brand-200 bg-brand-50 px-3 py-1.5 text-[12px] font-medium text-brand-700 transition-colors hover:bg-brand-100" @click="applyAssignments">
          Terapkan
        </button>
      </template>
    </div>

    <!-- Client: Active edit controls -->
    <div v-if="store.isClient && live" class="mb-3 space-y-1.5">
      <div v-for="sec in editableSections" :key="sec.id" class="flex items-center justify-between rounded border border-slate-200 bg-slate-50 px-3 py-1.5">
        <span class="text-[12px] font-medium text-slate-600">{{ sec.label }}</span>
        <div class="flex items-center gap-1.5">
          <span v-if="activeEditorFor(sec.id)" class="text-[11px] text-amber-600">
            {{ activeEditorFor(sec.id) === myId ? 'Anda sedang edit' : activeEditorFor(sec.id) }}
          </span>
          <button
            v-if="!activeEditorFor(sec.id)"
            type="button"
            class="rounded border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 transition-colors hover:bg-emerald-100"
            @click="store.requestEdit(sec.id)"
          >
            Edit
          </button>
          <button
            v-else-if="activeEditorFor(sec.id) === myId"
            type="button"
            class="rounded border border-amber-200 bg-amber-50 px-2 py-0.5 text-[11px] font-medium text-amber-700 transition-colors hover:bg-amber-100"
            @click="handleReleaseAndSync(sec.id)"
          >
            Selesai &amp; Sync
          </button>
        </div>
      </div>
    </div>

    <!-- Participants -->
    <div v-if="participants.length" class="mb-3">
      <h4 class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Pengguna ({{ participants.length }})</h4>
      <ul class="space-y-1">
        <li v-for="p in participants" :key="p.id" class="flex items-center gap-2 text-[13px]">
          <span class="h-1.5 w-1.5 rounded-full" :class="p.role === 'HOST' ? 'bg-brand-500' : 'bg-slate-400'"></span>
          <span class="font-medium text-slate-700">{{ p.displayName }}</span>
          <span class="text-xs text-slate-400">· {{ p.deviceName }}</span>
          <span v-if="p.role === 'HOST'" class="rounded bg-brand-50 px-1 text-[10px] font-semibold text-brand-600">HOST</span>
        </li>
      </ul>
    </div>

    <!-- Activity log -->
    <div v-if="activities.length">
      <h4 class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Aktivitas Terakhir</h4>
      <ul class="max-h-[140px] space-y-1 overflow-y-auto">
        <li v-for="(a, i) in activities" :key="i" class="text-[12px] text-slate-500">
          <span class="font-medium text-slate-600">{{ a.actor }}</span> {{ a.summary }}
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCollaborationStore } from '../../stores/collaboration'
import { ConnLabel, SectionLabel } from '../../types/collaboration'

const store = useCollaborationStore()

const emit = defineEmits<{ closed: [] }>()

const showConnInfo = ref(false)
const copied = ref(false)
const leaving = ref(false)
const pendingAssignments = ref<Record<string, string[]>>({})

const snap = computed(() => store.snapshot)
const live = computed(() => store.isLive)
const participants = computed(() => store.snapshot.participants ?? [])
const activities = computed(() => store.snapshot.activities ?? [])
const connLabel = computed(() => ConnLabel[snap.value.connection ?? ''] ?? snap.value.connection ?? '')

const otherParticipants = computed(() =>
  participants.value.filter(p => p.role !== 'HOST')
)

const allSections = [
  { id: 'header', label: 'Header' },
  { id: 'items', label: 'Items' },
  { id: 'subitems', label: 'Sub Items' }
]

const editableSections = computed(() => {
  const assigned = store.myAssignments
  return allSections.filter(s => assigned.includes(s.id))
})

const myId = computed(() => {
  const parts = store.snapshot.participants ?? []
  return parts.find(p => p.role !== 'HOST')?.id ?? ''
})

function isAssigned(participantId: string, sectionId: string): boolean {
  const a = store.turn?.assignments ?? pendingAssignments.value
  return (a[participantId] ?? []).includes(sectionId)
}

function toggleAssignment(participantId: string, sectionId: string, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  const current = pendingAssignments.value[participantId] ?? [...(store.turn?.assignments?.[participantId] ?? [])]
  if (checked) {
    if (!current.includes(sectionId)) current.push(sectionId)
  } else {
    const idx = current.indexOf(sectionId)
    if (idx >= 0) current.splice(idx, 1)
  }
  pendingAssignments.value = { ...pendingAssignments.value, [participantId]: current }
}

async function applyAssignments() {
  await store.assignTurns(pendingAssignments.value)
  pendingAssignments.value = {}
}

function activeEditorFor(sectionId: string): string {
  const editorId = store.turn?.activeEdits?.[sectionId]
  if (!editorId) return ''
  const parts = store.snapshot.participants ?? []
  return parts.find(p => p.id === editorId)?.displayName ?? ''
}

async function handleReleaseAndSync(sectionId: string) {
  await store.releaseEdit(sectionId)
}

async function handleLeave() {
  if (leaving.value) return
  leaving.value = true
  try {
    if (store.isHost) {
      await store.closeRoom()
      emit('closed')
    } else {
      await store.leaveRoom()
    }
  } finally {
    leaving.value = false
  }
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  }
}
</script>
