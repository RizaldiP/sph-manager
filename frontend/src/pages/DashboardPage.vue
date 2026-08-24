<template>
  <div>
    <PageHeader title="Dashboard" subtitle="Ringkasan aktivitas SPH perusahaan" />

    <p v-if="store.error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      Gagal terhubung ke backend: {{ store.error }}
    </p>
    <p v-if="sphStore.statsError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ sphStore.statsError }}
    </p>

    <div class="grid grid-cols-2 gap-4 md:grid-cols-5">
      <div class="rounded-xl border border-slate-200 bg-white p-4">
        <p class="text-[13px] text-slate-500">Total SPH</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-slate-800">{{ sphStore.stats?.totalSph ?? '—' }}</p>
        <p class="mt-0.5 text-xs text-slate-400">seluruh dokumen</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4">
        <p class="text-[13px] text-slate-500">SPH Draft</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-slate-800">{{ sphStore.stats?.draftCount ?? '—' }}</p>
        <p class="mt-0.5 text-xs text-slate-400">belum final</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4">
        <p class="text-[13px] text-slate-500">SPH Final</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-slate-800">{{ sphStore.stats?.finalCount ?? '—' }}</p>
        <p class="mt-0.5 text-xs text-slate-400">final & terkirim</p>
      </div>
      <div class="rounded-xl border border-emerald-200 bg-emerald-50/50 p-4">
        <p class="text-[13px] text-emerald-700">Disetujui</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-emerald-800">{{ sphStore.stats?.acceptedCount ?? '—' }}</p>
        <p class="mt-0.5 text-xs text-emerald-600/70">menang tender</p>
      </div>
      <div class="rounded-xl border border-brand-200 bg-brand-50/60 p-4">
        <p class="text-[13px] text-brand-700">Nilai Bulan Ini</p>
        <p class="mt-1 whitespace-nowrap text-lg font-semibold tabular-nums text-brand-800">{{ formatRupiah(sphStore.stats?.monthValue) }}</p>
        <p class="mt-0.5 text-xs text-brand-600/70">{{ monthLabel }}</p>
      </div>
    </div>

    <div class="mt-5 grid grid-cols-3 gap-4">
      <div class="col-span-2 rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-800">Aksi Cepat</h2>
        <div class="mt-3 grid grid-cols-2 gap-3">
          <button
            v-for="action in actions"
            :key="action.label"
            :disabled="!action.to"
            class="flex items-center gap-2.5 rounded-lg border px-3.5 py-3 text-left text-[13px] transition-colors disabled:cursor-not-allowed"
            :class="action.primary ? 'border-accent-200 bg-accent-50 text-accent-700' : 'border-slate-200 bg-white text-slate-600'"
            @click="action.to && router.push(action.to)"
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
        <button v-if="(sphStore.stats?.recent.length ?? 0) > 0" type="button" class="text-xs font-medium text-brand-600 hover:text-brand-700" @click="router.push('/sph')">
          Lihat semua →
        </button>
      </div>

      <div v-if="(sphStore.stats?.recent.length ?? 0) > 0" class="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <table class="w-full text-left text-[13px]">
          <tbody>
            <tr
              v-for="d in sphStore.stats!.recent"
              :key="d.id"
              class="cursor-pointer border-b border-slate-50 transition-colors last:border-b-0 hover:bg-slate-50/70"
              @click="router.push(`/sph/${d.id}`)"
            >
              <td class="whitespace-nowrap px-4 py-2.5 font-mono text-xs font-semibold text-slate-700">
                {{ d.documentNumber }}
                <span v-if="d.revision > 0" class="ml-1 rounded bg-brand-50 px-1 py-0.5 text-[10px] text-brand-600">R{{ d.revision }}</span>
              </td>
              <td class="max-w-[220px] truncate px-3 py-2.5 text-slate-600">{{ d.customerName }}</td>
              <td class="whitespace-nowrap px-3 py-2.5 text-right font-medium tabular-nums text-slate-800">{{ formatRupiah(d.grandTotal) }}</td>
              <td class="whitespace-nowrap px-3 py-2.5">
                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :class="statusToneOf(d.status)">{{ statusLabelOf(d.status) }}</span>
              </td>
              <td class="whitespace-nowrap pl-3 pr-4 py-2.5 text-right text-xs tabular-nums text-slate-400">{{ d.date }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <EmptyState
        v-else
        title="Belum ada dokumen SPH"
        description="Dokumen yang dibuat akan tampil di sini. Mulai dengan tombol Buat SPH."
      >
        <button
          type="button"
          class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="router.push('/sph/baru')"
        >
          + Buat SPH
        </button>
      </EmptyState>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import { useAppStore } from '../stores/app'
import { useSphStore } from '../stores/sph'
import { formatRupiah } from '../utils/format'
import { statusLabelOf, statusToneOf } from '../types/sph'

const store = useAppStore()
const sphStore = useSphStore()
const router = useRouter()

const monthLabel = computed(() =>
  new Date().toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
)

interface QuickAction {
  label: string
  phase: string
  primary?: boolean
  to?: string
}

const actions: QuickAction[] = [
  { label: 'Buat SPH', phase: '5', primary: true, to: '/sph/baru' },
  { label: 'Tambah Pekerjaan', phase: '3', to: '/pekerjaan' },
  { label: 'Tambah Template', phase: '4', to: '/pekerjaan/template' },
  { label: 'Import Excel', phase: '8' }
]

onMounted(() => {
  void sphStore.loadStats()
})
</script>
