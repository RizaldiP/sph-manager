<template>
  <div>
    <PageHeader title="Dashboard" subtitle="Ringkasan aktivitas SPH perusahaan" />

    <p v-if="store.error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      Gagal terhubung ke backend: {{ store.error }}
    </p>

    <div class="grid grid-cols-4 gap-4">
      <div v-for="stat in stats" :key="stat.label" class="rounded-xl border border-slate-200 bg-white p-4">
        <p class="text-[13px] text-slate-500">{{ stat.label }}</p>
        <p class="mt-1 text-xl font-semibold text-slate-300">—</p>
        <p class="mt-0.5 text-xs text-slate-400">{{ stat.note }}</p>
      </div>
    </div>

    <div class="mt-5 grid grid-cols-3 gap-4">
      <div class="col-span-2 rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-800">Aksi Cepat</h2>
        <div class="mt-3 grid grid-cols-2 gap-3">
          <button
            v-for="action in actions"
            :key="action.label"
            disabled
            class="flex items-center gap-2.5 rounded-lg border px-3.5 py-3 text-left text-[13px] transition-colors disabled:cursor-not-allowed"
            :class="action.primary ? 'border-accent-200 bg-accent-50 text-accent-700' : 'border-slate-200 bg-white text-slate-600'"
          >
            <span
              class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-white"
              :class="action.primary ? 'bg-accent-500' : 'bg-brand-600'"
            >
              <span class="text-base font-semibold leading-none">+</span>
            </span>
            <span>
              {{ action.label }}
              <span class="block text-[11px] text-slate-400">Phase {{ action.phase }}</span>
            </span>
          </button>
        </div>
      </div>

      <div class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-800">Status Sistem</h2>
        <dl class="mt-3 space-y-2.5 text-[13px]">
          <div class="flex justify-between gap-3">
            <dt class="text-slate-500">Status</dt>
            <dd class="font-medium" :class="store.health?.status === 'ok' ? 'text-emerald-600' : 'text-red-600'">
              {{ store.loaded ? (store.health?.status === 'ok' ? 'Berjalan normal' : 'Gangguan') : 'Memeriksa…' }}
            </dd>
          </div>
          <div class="flex justify-between gap-3">
            <dt class="text-slate-500">Versi</dt>
            <dd class="font-medium text-slate-700">{{ store.health?.version ?? '—' }}</dd>
          </div>
          <div class="flex justify-between gap-3">
            <dt class="text-slate-500">Platform</dt>
            <dd class="font-medium text-slate-700">{{ store.health?.platform ?? '—' }}</dd>
          </div>
          <div>
            <dt class="text-slate-500">Lokasi Database</dt>
            <dd class="mt-0.5 break-all rounded-md bg-slate-50 px-2 py-1.5 font-mono text-[11px] leading-relaxed text-slate-600">
              {{ store.health?.databasePath ?? '—' }}
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <div class="mt-5">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-sm font-semibold text-slate-800">SPH Terbaru</h2>
        <span class="text-xs text-slate-400">Phase 5</span>
      </div>
      <EmptyState
        title="Belum ada dokumen SPH"
        description="Dokumen yang dibuat akan tampil di sini. Fitur SPH tersedia pada Phase 5."
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import { useAppStore } from '../stores/app'

const store = useAppStore()

interface Stat {
  label: string
  note: string
}

const stats: Stat[] = [
  { label: 'Total SPH', note: 'Phase 5' },
  { label: 'SPH Draft', note: 'Phase 5' },
  { label: 'SPH Final', note: 'Phase 5' },
  { label: 'Nilai Bulan Ini', note: 'Phase 5' }
]

interface QuickAction {
  label: string
  phase: string
  primary?: boolean
}

const actions: QuickAction[] = [
  { label: 'Buat SPH', phase: '5', primary: true },
  { label: 'Tambah Pekerjaan', phase: '3' },
  { label: 'Tambah Template', phase: '4' },
  { label: 'Import Excel', phase: '8' }
]
</script>
