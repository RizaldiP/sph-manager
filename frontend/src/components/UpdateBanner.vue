<template>
  <div>
    <!-- Banner update tersedia -->
    <div
      v-if="store.updateStatus && store.updateStatus.hasUpdate && !dismissed"
      class="shrink-0 border-b border-brand-200 bg-brand-50 px-6 py-2.5"
    >
      <div class="flex items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-2.5 text-[13px] text-brand-800">
          <svg class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
          </svg>
          <span class="font-medium">Versi baru v{{ store.updateStatus.latestVersion }} tersedia</span>
          <span class="hidden text-brand-600 sm:inline">(versi saat ini v{{ store.updateStatus.currentVersion }})</span>
          <span v-if="store.updateStatus.notes" class="hidden truncate text-brand-600 md:inline">· {{ store.updateStatus.notes }}</span>
        </div>
        <div class="flex shrink-0 items-center gap-1.5">
          <button class="rounded-lg bg-brand-600 px-3 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="store.openUpdater()">
            Pasang Sekarang
          </button>
          <button
            class="rounded-lg p-1.5 text-brand-500 transition-colors hover:bg-brand-100 hover:text-brand-700"
            aria-label="Tutup notifikasi"
            @click="dismissed = true"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Dialog pembaruan -->
    <AppModal
      :model-value="open"
      title="Pembaruan Aplikasi"
      :dismissible="!busy"
      :show-close="!busy"
      @update:model-value="onModalChange"
    >
      <div class="space-y-4 py-1">
        <!-- memeriksa -->
        <div v-if="store.updateBusy && !store.updateStatus" class="flex items-center gap-2.5 py-3 text-[13px] text-slate-500">
          <svg class="h-4 w-4 animate-spin text-brand-600" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 0 1 8-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Memeriksa pembaruan…
        </div>

        <!-- gagal memeriksa -->
        <div v-else-if="store.updateError && !store.updateStatus" class="space-y-3">
          <p class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[13px] text-red-700">{{ store.updateError }}</p>
          <p class="text-xs text-slate-500">
            Pastikan link file aplikasi sudah terisi di halaman Pengaturan (atau kosongkan untuk memakai link bawaan) dan koneksi internet tersedia.
          </p>
          <button class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="store.checkUpdate()">
            Coba Lagi
          </button>
        </div>

        <!-- periksa selesai -->
        <template v-else-if="store.updateStatus">
          <!-- sudah terbaru -->
          <div v-if="!store.updateStatus.hasUpdate" class="py-1">
            <p class="text-[13px] text-slate-600">
              Aplikasi sudah dalam versi terbaru
              <span class="font-semibold text-slate-800">v{{ store.updateStatus.currentVersion }}</span>.
            </p>
          </div>

          <!-- ada pembaruan -->
          <div v-else class="space-y-4">
            <div class="flex items-center gap-5 rounded-lg border border-slate-200 bg-slate-50/60 px-4 py-3">
              <div>
                <p class="text-[11px] uppercase tracking-wide text-slate-400">Versi terpasang</p>
                <p class="text-lg font-semibold text-slate-500">v{{ store.updateStatus.currentVersion }}</p>
              </div>
              <svg class="pb-1 text-slate-300" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
              </svg>
              <div>
                <p class="text-[11px] uppercase tracking-wide text-slate-400">Versi baru</p>
                <p class="text-lg font-semibold text-brand-700">v{{ store.updateStatus.latestVersion }}</p>
              </div>
            </div>

            <div v-if="store.updateStatus.notes" class="rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-[13px] text-slate-600">
              {{ store.updateStatus.notes }}
            </div>

            <!-- Langkah 1: backup (wajib) -->
            <div class="rounded-lg border border-slate-200 p-3.5" :class="backupResult ? 'border-emerald-200 bg-emerald-50/50' : ''">
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-[13px] font-semibold text-slate-700">1. Backup data (wajib)</p>
                  <p v-if="backupResult" class="mt-0.5 truncate text-xs text-emerald-700">Selesai: {{ backupResult.name }} · {{ backupResult.modified }}</p>
                  <p v-else class="mt-0.5 text-xs text-slate-400">Cadangkan database sebelum mengganti aplikasi demi keamanan data.</p>
                </div>
                <button
                  class="shrink-0 rounded-lg px-3.5 py-2 text-[13px] font-medium transition-colors disabled:opacity-60"
                  :class="backupResult ? 'border border-emerald-300 bg-emerald-100/60 text-emerald-800' : 'border border-slate-200 bg-white text-slate-600 hover:bg-slate-50'"
                  :disabled="busy"
                  @click="startBackup"
                >
                  <span v-if="backupBusy">Membuat…</span>
                  <span v-else-if="backupResult">Backup selesai</span>
                  <span v-else>Buat Backup</span>
                </button>
              </div>
            </div>

            <!-- Langkah 2: unduh & pasang -->
            <div class="rounded-lg border border-slate-200 p-3.5">
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-[13px] font-semibold text-slate-700">2. Unduh &amp; pasang</p>
                  <p class="mt-0.5 text-xs text-slate-400">Aplikasi akan ditutup lalu dijalankan ulang secara otomatis.</p>
                </div>
                <button
                  class="shrink-0 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
                  :disabled="busy || !backupResult"
                  @click="startDownload"
                >
                  <span v-if="downloadBusy">Mengunduh…</span>
                  <span v-else>Unduh &amp; Pasang</span>
                </button>
              </div>

              <div v-if="!downloadFinalizing" class="mt-3">
                <div class="flex items-center justify-between text-xs text-slate-500">
                  <span>{{ progressLabel }}</span>
                  <span v-if="progressText">{{ progressText }}</span>
                </div>
                <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-100">
                  <div
                    class="h-full rounded-full bg-brand-600 transition-all"
                    :class="progressIndeterminate ? 'indeterminate-bar' : ''"
                    :style="progressIndeterminate ? undefined : { width: progressPct + '%' }"
                  ></div>
                </div>
              </div>
              <p v-else class="mt-2 rounded-md border border-brand-200 bg-brand-50 px-2.5 py-1.5 text-xs text-brand-700">
                Unduhan selesai. Aplikasi akan segera ditutup dan dijalankan ulang dengan versi baru…
              </p>

              <p v-if="stepError" class="mt-2 rounded-md border border-red-200 bg-red-50 px-2.5 py-1.5 text-xs text-red-700">{{ stepError }}</p>
            </div>
          </div>
        </template>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <button
            class="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60"
            :disabled="busy || store.updateBusy"
            @click="store.closeUpdater()"
          >
            {{ busy ? 'Tunggu…' : 'Tutup' }}
          </button>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AppModal from './AppModal.vue'
import { useAppStore } from '../stores/app'
import { errorMessage } from '../utils/format'
import { BackupForUpdate, DownloadAndApplyUpdate } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface BackupResult {
  name: string
  path: string
  size: number
  modified: string
}

const store = useAppStore()

const dismissed = ref(false)

const backupBusy = ref(false)
const backupResult = ref<BackupResult | null>(null)
const downloadBusy = ref(false)
const downloadFinalizing = ref(false)
const downloadProgress = ref<{ done: number; total: number } | null>(null)
const stepError = ref('')

const busy = computed(() => backupBusy.value || downloadBusy.value)

const open = computed({
  get: () => store.updateDialogOpen,
  set: (v: boolean) => {
    if (v) store.openUpdater()
    else store.closeUpdater()
  }
})

function onModalChange(v: boolean) {
  open.value = v
}

const progressPct = computed(() => {
  const p = downloadProgress.value
  if (!p || p.total <= 0) return 0
  return Math.min(100, Math.round((p.done / p.total) * 100))
})

const progressIndeterminate = computed(() => {
  const p = downloadProgress.value
  return !!p && p.total <= 0
})

const progressLabel = computed(() => {
  if (!downloadProgress.value) return downloadBusy.value ? 'Menghitung ukuran…' : 'Belum dimulai'
  if (progressPct.value > 0) return `${progressPct.value}%`
  if (progressIndeterminate.value) return 'Mengunduh…'
  return 'Mengunduh…'
})

const progressText = computed(() => {
  const p = downloadProgress.value
  if (!p || p.done <= 0) return ''
  const mb = (v: number) => (v / 1024 / 1024).toFixed(1)
  return p.total > 0 ? `${mb(p.done)} / ${mb(p.total)} MB` : `${mb(p.done)} MB`
})

watch(
  () => store.updateDialogOpen,
  (openNow) => {
    if (openNow) {
      backupBusy.value = false
      backupResult.value = null
      downloadBusy.value = false
      downloadFinalizing.value = false
      downloadProgress.value = null
      stepError.value = ''
    }
  }
)

onMounted(() => {
  if (!store.updateChecked) {
    void store.checkUpdate()
  }
})

onUnmounted(() => EventsOff('update:progress'))

async function startBackup() {
  stepError.value = ''
  backupBusy.value = true
  try {
    backupResult.value = (await BackupForUpdate()) as unknown as BackupResult
  } catch (e) {
    stepError.value = errorMessage(e)
  } finally {
    backupBusy.value = false
  }
}

async function startDownload() {
  stepError.value = ''
  downloadBusy.value = true
  downloadFinalizing.value = false
  downloadProgress.value = null
  EventsOn('update:progress', (data: { done: number; total: number }) => {
    downloadProgress.value = data
    if (data.done === -1 || data.total === -1) {
      downloadFinalizing.value = true
    }
  })
  try {
    await DownloadAndApplyUpdate()
    downloadFinalizing.value = true
  } catch (e) {
    stepError.value = errorMessage(e)
    downloadFinalizing.value = false
  } finally {
    downloadBusy.value = false
    EventsOff('update:progress')
  }
}
</script>

<style scoped>
.indeterminate-bar {
  width: 40%;
  animation: spm-indeterminate 1.2s ease-in-out infinite;
}

@keyframes spm-indeterminate {
  0% {
    transform: translateX(-110%);
  }
  100% {
    transform: translateX(275%);
  }
}
</style>