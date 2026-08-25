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
      <button v-if="live" type="button" class="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs font-medium text-red-700 transition-colors hover:bg-red-50" @click="handleLeave">
        {{ store.isHost ? 'Tutup Room' : 'Keluar' }}
      </button>
    </div>

    <!-- Room info -->
    <div v-if="snap.room" class="mb-3 space-y-1 text-[13px]">
      <p class="font-semibold text-slate-800">{{ snap.room.roomName }}</p>
      <p class="text-xs text-slate-400">{{ snap.room.documentNumber }} · {{ snap.room.projectName }}</p>
      <p v-if="snap.room.accessCode && store.isHost" class="text-xs text-slate-400">Code: <span class="font-mono font-semibold text-slate-600">{{ snap.room.accessCode }}</span></p>
    </div>

    <!-- Connection status -->
    <div v-if="snap.connection && snap.connection !== 'CONNECTED'" class="mb-3 flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[13px] text-amber-700">
      <span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
      {{ connLabel }}
    </div>

    <!-- Error / notice -->
    <div v-if="snap.error" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ snap.error }}</div>
    <div v-if="snap.notice" class="mb-3 rounded-lg border border-blue-200 bg-blue-50 px-3 py-2 text-[13px] text-blue-700">{{ snap.notice }}</div>

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
import { computed } from 'vue'
import { useCollaborationStore } from '../../stores/collaboration'
import { ConnLabel } from '../../types/collaboration'

const store = useCollaborationStore()

const snap = computed(() => store.snapshot)
const live = computed(() => store.isLive)
const participants = computed(() => store.snapshot.participants ?? [])
const activities = computed(() => store.snapshot.activities ?? [])
const connLabel = computed(() => ConnLabel[snap.value.connection ?? ''] ?? snap.value.connection ?? '')

function handleLeave() {
  if (store.isHost) {
    store.closeRoom()
  } else {
    store.leaveRoom()
  }
}
</script>
