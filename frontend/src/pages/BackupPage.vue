<template>
  <div>
    <PageHeader title="Backup & Restore" subtitle="Lindungi data SPH Anda: backup manual, otomatis harian, dan pemulihan (FR-B1..B3)">
      <template #actions>
        <button
          type="button"
          class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-brand-600 transition-colors hover:bg-brand-50 disabled:opacity-60"
          :disabled="busy || importing"
          @click="importBackup"
        >
          {{ importing ? 'Memvalidasi…' : 'Import Backup…' }}
        </button>
        <button
          type="button"
          class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60"
          :disabled="busy || importing"
          @click="openFolder"
        >
          Buka Folder Backup
        </button>
        <button
          type="button"
          class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
          :disabled="busy"
          @click="createNow"
        >
          {{ backingUp ? 'Membuat…' : 'Buat Backup Sekarang' }}
        </button>
      </template>
    </PageHeader>

    <p v-if="actionError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ actionError }}
    </p>
    <p v-if="actionSuccess" class="mb-4 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-[13px] text-emerald-700">
      {{ actionSuccess }}
    </p>
    <p v-if="shareResult" class="mb-4 whitespace-pre-line rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-[13px] leading-relaxed text-emerald-700">
      {{ shareResult }}
    </p>

    <!-- Info auto backup -->
    <section class="mb-5 rounded-xl border border-slate-200 bg-white p-5">
      <h2 class="mb-3 text-sm font-semibold text-slate-800">Backup Otomatis</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="rounded-lg border border-slate-100 bg-slate-50/60 p-4">
          <p class="text-[13px] font-medium text-slate-700">Jadwal</p>
          <p class="mt-1 text-[13px] leading-relaxed text-slate-500">
            Snapshots database dibuat otomatis setiap kali aplikasi ditutup dan sekali per hari bila aplikasi
            dibiarkan menyala.
          </p>
        </div>
        <div class="rounded-lg border border-slate-100 bg-slate-50/60 p-4">
          <p class="text-[13px] font-medium text-slate-700">Penyimpanan &amp; Retention</p>
          <p class="mt-1 text-[13px] leading-relaxed text-slate-500">
            Disimpan di <code class="font-mono text-[12px] text-brand-700">{{ backupDirHint }}</code>. Hanya
            10 backup terbaru yang dipertahankan (FR-B3).
          </p>
        </div>
      </div>
      <div class="mt-3 flex items-center gap-2 rounded-lg border border-slate-100 bg-slate-50/60 px-4 py-2.5 text-[13px]">
        <span class="relative flex h-2.5 w-2.5">
          <span
            v-if="autoActive"
            class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
          ></span>
          <span class="relative inline-flex h-2.5 w-2.5 rounded-full" :class="autoActive ? 'bg-emerald-500' : 'bg-slate-300'"></span>
        </span>
        <span class="text-slate-600">
          {{ autoActive ? 'Backup hari ini sudah tersedia.' : 'Belum ada backup hari ini — otomatis saat aplikasi ditutup.' }}
        </span>
      </div>
    </section>

    <!-- Backup yang dapat dibagikan -->
    <section class="mb-5 rounded-xl border border-slate-200 bg-white p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="max-w-2xl">
          <h2 class="text-sm font-semibold text-slate-800">Backup yang Dapat Dibagikan</h2>
          <p class="mt-1 text-[13px] leading-relaxed text-slate-500">
            Kirim data ke perangkat lain lewat satu file
            <code class="font-mono text-[12px] text-brand-700">.sphbak</code>. Penerima memilih seksi mana yang ingin
            dipulihkan. Pemulihan hanya <strong>menambah</strong> data baru — data yang sudah ada
            tidak dihapus maupun ditimpa.
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-brand-600 transition-colors hover:bg-brand-50 disabled:opacity-60"
            :disabled="busy || exporting || scanning"
            @click="exportShare"
          >
            {{ exporting ? 'Menyusun…' : 'Ekspor Backup (Kirim)' }}
          </button>
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
            :disabled="busy || exporting || scanning"
            @click="importShare"
          >
            {{ scanning ? 'Membaca…' : 'Impor Backup (Pulihkan)' }}
          </button>
        </div>
      </div>
    </section>

    <!-- Daftar backup -->
    <section class="overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div class="border-b border-slate-100 px-4 py-3">
        <h2 class="text-sm font-semibold text-slate-800">Daftar Backup</h2>
      </div>

      <div v-if="loading" class="px-4 py-8 text-center text-[13px] text-slate-400">Memuat…</div>

      <table v-else-if="backups.length" class="w-full text-left">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50/60 text-[11px] uppercase tracking-wide text-slate-400">
            <th class="px-4 py-2 font-medium">Nama File</th>
            <th class="px-3 py-2 font-medium">Dibuat</th>
            <th class="px-3 py-2 text-right font-medium">Ukuran</th>
            <th class="px-4 py-2 text-right font-medium">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in backups" :key="b.name" class="border-b border-slate-50 text-[13px] last:border-b-0">
            <td class="max-w-[300px] truncate px-4 py-2 font-mono text-[12px] text-slate-700" :title="b.name">{{ b.name }}</td>
            <td class="whitespace-nowrap px-3 py-2 text-slate-500">{{ b.modified }}</td>
            <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-500">{{ formatBytes(b.size) }}</td>
            <td class="whitespace-nowrap px-4 py-2 text-right">
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="askRestore(b)">Pulihkan</button>
              <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDelete(b)">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else class="px-4 py-8 text-center text-[13px] text-slate-400">
        Belum ada backup. Buat backup pertama Anda untuk melindungi data.
      </div>
    </section>

    <!-- Konfirmasi hapus -->
    <ConfirmDialog
      v-model="deleteOpen"
      title="Hapus Backup"
      :message="deleteMessage"
      confirm-label="Hapus"
      danger
      @confirm="runDelete"
    />

    <!-- Konfirmasi pulihkan -->
    <ConfirmDialog
      v-model="restoreOpen"
      title="Pulihkan dari Backup"
      :message="restoreMessage"
      :detail="restoreDetail"
      confirm-label="Pulihkan Sekarang"
      :busy="restoring"
      @confirm="runRestore"
    />

    <!-- Pulihkan backup yang dibagikan -->
    <AppModal v-model="shareOpen" title="Pulihkan dari Backup yang Dibagikan" size="lg" :dismissible="!restoring">
      <div v-if="sharePreview">
        <div class="rounded-lg border border-brand-100 bg-brand-50/60 px-3.5 py-3">
          <p class="text-[13px] font-medium text-brand-800">File siap dipulihkan</p>
          <p class="mt-1 break-all font-mono text-[12px] leading-relaxed text-brand-700/90">{{ sharePreview.path }}</p>
          <p v-if="sharePreview.deviceName" class="mt-1 text-[12px] text-brand-700/80">
            Dibuat di <strong>{{ sharePreview.deviceName }}</strong>, {{ sharePreview.createdAt }}.
          </p>
        </div>

        <p class="mb-2 mt-4 text-[13px] font-medium text-slate-700">Pilih seksi yang ingin dipulihkan:</p>
        <div class="grid gap-2">
          <label
            v-for="opt in shareSections"
            :key="opt.key"
            class="flex cursor-pointer items-center gap-3 rounded-lg border px-3.5 py-2.5 transition-colors"
            :class="shareSelected[opt.key] ? 'border-brand-200 bg-brand-50/50' : 'border-slate-200 bg-slate-50/40'"
          >
            <input
              v-model="shareSelected[opt.key]"
              type="checkbox"
              class="h-4 w-4 rounded border-slate-300 text-brand-600 focus:ring-brand-500"
            />
            <span class="min-w-0 flex-1">
              <span class="block text-[13px] font-medium text-slate-700">{{ opt.label }}</span>
              <span class="block text-[12px] text-slate-500">{{ opt.description }}</span>
            </span>
            <span class="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-[12px] font-medium tabular-nums text-slate-600">
              {{ shareCount(opt.key) }} item
            </span>
          </label>
        </div>

        <p class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-[12px] leading-relaxed text-amber-700">
          Pemulihan hanya <strong>menambah</strong> data baru. Data yang sudah ada di perangkat ini (dibandingkan
          berdasarkan nama atau nomor dokumen) akan dilewati dan tidak dihapus maupun ditimpa.
        </p>
      </div>

      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60"
            :disabled="restoring"
            @click="shareOpen = false"
          >
            Batal
          </button>
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
            :disabled="restoring || !sharePreview || noSectionSelected"
            @click="runShareRestore"
          >
            {{ restoring ? 'Memulihkan…' : 'Pulihkan Sekarang' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Layar penutup saat aplikasi akan restart -->
    <AppModal v-model="restarting" title="Restore Berhasil" :dismissible="false" :show-close="false" size="md">
      <div class="flex items-start gap-3.5">
        <span class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-600">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
        </span>
        <div>
          <p class="text-[13px] font-medium text-slate-700">{{ restartMessage }}</p>
          <p class="mt-1 text-[13px] text-slate-500">Aplikasi akan ditutup dan dibuka kembali sebentar lagi…</p>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { errorMessage } from '../utils/format'
import type { BackupInfo } from '../types/backup'
import type { InstallSummary, SectionInstallResult, ShareBackupPreview, ShareSectionKey, ShareSectionOption } from '../types/sharebackup'
import { useMasterStore } from '../stores/master'
import { useTemplateStore } from '../stores/template'
import { usePartnerStore } from '../stores/partner'
import { useMaterialStore } from '../stores/material'
import { useSphStore } from '../stores/sph'
import {
  BackupNow,
  CreateShareableBackup,
  DeleteBackup,
  ImportBackup,
  ListBackups,
  OpenBackupFolder,
  OpenShareableBackup,
  QuitApp,
  RestoreBackup,
  RestoreShareableBackup
} from '../../wailsjs/go/main/App'

const backups = ref<BackupInfo[]>([])
const loading = ref(false)
const busy = ref(false)
const backingUp = ref(false)
const importing = ref(false)
const restoring = ref(false)
const actionError = ref('')
const actionSuccess = ref('')

const deleteOpen = ref(false)
const deleteMessage = ref('')
const deleteTarget = ref<BackupInfo | null>(null)

const restoreOpen = ref(false)
const restoreMessage = ref('')
const restoreDetail = ref('')
const restoreTarget = ref<BackupInfo | null>(null)

const restarting = ref(false)
const restartMessage = ref('')

const exporting = ref(false)
const scanning = ref(false)
const shareOpen = ref(false)
const sharePreview = ref<ShareBackupPreview | null>(null)
const shareResult = ref('')
const shareSelected = reactive<Record<ShareSectionKey, boolean>>({
  sph: true,
  workItems: true,
  categories: true,
  templates: true,
  customers: true,
  materials: true
})

const shareSections: ShareSectionOption[] = [
  { key: 'sph', label: 'Semua Data SPH', description: 'Seluruh dokumen penawaran beserta item dan riwayat revisi.' },
  { key: 'workItems', label: 'Master Pekerjaan', description: 'Daftar pekerjaan beserta sub-pekerjaannya.' },
  { key: 'categories', label: 'Kategori', description: 'Kategori pekerjaan.' },
  { key: 'templates', label: 'Template', description: 'Template isi dokumen.' },
  { key: 'customers', label: 'Customer', description: 'Daftar customer beserta kapalnya.' },
  { key: 'materials', label: 'Material', description: 'Daftar material pengeluaran.' }
]

function shareCount(key: ShareSectionKey): number {
  const c = sharePreview.value?.counts
  if (!c) return 0
  return c[key]
}

const noSectionSelected = computed(() => !shareSections.some((o) => shareSelected[o.key]))

const autoActive = computed(() => {
  const today = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  const prefix = `SPH_Backup_${today.getFullYear()}-${pad(today.getMonth() + 1)}-${pad(today.getDate())}`
  return backups.value.some((b) => b.name.startsWith(prefix))
})

const backupDirHint = computed(() => {
  const b = backups.value[0]
  if (!b) return 'folder data aplikasi'
  const i = b.path.lastIndexOf('\\')
  return i >= 0 ? b.path.slice(0, i) : b.path
})

function formatBytes(n: number): string {
  if (!n || n <= 0) return '—'
  for (const unit of ['B', 'KB', 'MB', 'GB']) {
    if (n < 1024) return `${n.toFixed(n < 10 && unit !== 'B' ? 1 : 0)} ${unit}`
    n = n / 1024
  }
  return `${n.toFixed(1)} TB`
}

async function load() {
  loading.value = true
  actionError.value = ''
  try {
    backups.value = await ListBackups()
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function createNow() {
  backingUp.value = true
  actionError.value = ''
  actionSuccess.value = ''
  try {
    const info = await BackupNow()
    actionSuccess.value = `Backup berhasil dibuat: ${info.name}`
    await load()
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    backingUp.value = false
  }
}

async function openFolder() {
  actionError.value = ''
  actionSuccess.value = ''
  try {
    await OpenBackupFolder()
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

async function importBackup() {
  importing.value = true
  actionError.value = ''
  actionSuccess.value = ''
  try {
    const info = await ImportBackup()
    if (info) {
      actionSuccess.value = `Backup berhasil diimpor: ${info.name}`
      await load()
    }
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    importing.value = false
  }
}

async function exportShare() {
  exporting.value = true
  actionError.value = ''
  actionSuccess.value = ''
  shareResult.value = ''
  try {
    const res = await CreateShareableBackup()
    if (res) {
      actionSuccess.value = `Backup yang dapat dibagikan berhasil dibuat (${res.items} item). File: ${res.path}`
    }
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    exporting.value = false
  }
}

async function importShare() {
  scanning.value = true
  actionError.value = ''
  actionSuccess.value = ''
  shareResult.value = ''
  try {
    const preview = await OpenShareableBackup()
    if (!preview) return // pengguna membatalkan
    sharePreview.value = preview
    shareSelected.sph = true
    shareSelected.workItems = true
    shareSelected.categories = true
    shareSelected.templates = true
    shareSelected.customers = true
    shareSelected.materials = true
    shareOpen.value = true
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    scanning.value = false
  }
}

async function runShareRestore() {
  if (!sharePreview.value) return
  restoring.value = true
  actionError.value = ''
  actionSuccess.value = ''
  shareResult.value = ''
  try {
    const keys = shareSections.filter((o) => shareSelected[o.key]).map((o) => o.key)
    const sum = await RestoreShareableBackup(keys)
    shareOpen.value = false
    buildShareSummary(sum)
    await reloadAfterShare()
  } catch (e) {
    shareOpen.value = false
    actionError.value = errorMessage(e)
  } finally {
    restoring.value = false
  }
}

function buildShareSummary(sum: InstallSummary) {
  const lines: string[] = []
  const add = (label: string, r: SectionInstallResult, extra?: string) => {
    const extraText = extra ? ` (${extra})` : ''
    lines.push(`• ${label}: ${r.added} baru, ${r.skipped} sudah ada${extraText}`)
  }
  add('Semua Data SPH', sum.sph)
  const subText =
    sum.subItems.added + sum.subItems.skipped > 0 ? `sub: ${sum.subItems.added} baru, ${sum.subItems.skipped} sudah ada` : ''
  add('Master Pekerjaan', sum.workItems, subText)
  add('Kategori', sum.categories)
  add('Template', sum.templates)
  const vesText =
    sum.vessels.added + sum.vessels.skipped > 0 ? `kapal: ${sum.vessels.added} baru, ${sum.vessels.skipped} sudah ada` : ''
  add('Customer', sum.customers, vesText)
  add('Material', sum.materials)
  if (sum.templateItemsMissed > 0) {
    lines.push(`• ${sum.templateItemsMissed} item template dilewati (pekerjaan induk tidak ditemukan)`)
  }
  if (sum.sphItemsUnlinked > 0) {
    lines.push(`• ${sum.sphItemsUnlinked} item SPH tidak tertaut ke pekerjaan master yang sama`)
  }
  const generated =
    sum.categories.codeGenerated +
    sum.workItems.codeGenerated +
    sum.subItems.codeGenerated +
    sum.templates.codeGenerated +
    sum.customers.codeGenerated +
    sum.vessels.codeGenerated +
    sum.materials.codeGenerated
  if (generated > 0) {
    lines.push(`Catatan: ${generated} kode baru dibuat otomatis karena kode pengirim sudah terpakai di perangkat ini.`)
  }
  shareResult.value = ['Pemulihan selesai. Data baru berhasil ditambahkan:', ...lines].join('\n')
}

async function reloadAfterShare() {
  const master = useMasterStore()
  const template = useTemplateStore()
  const partner = usePartnerStore()
  const material = useMaterialStore()
  const sph = useSphStore()
  await Promise.allSettled([
    master.loadCategories(),
    master.loadWorkItems(),
    template.loadTemplates(),
    partner.load(),
    material.load(),
    sph.loadList(),
    sph.loadStats()
  ])
}

function askDelete(b: BackupInfo) {
  deleteTarget.value = b
  deleteMessage.value = `Backup "${b.name}" akan dihapus permanen. Lanjutkan?`
  deleteOpen.value = true
}

async function runDelete() {
  if (!deleteTarget.value) return
  actionError.value = ''
  actionSuccess.value = ''
  try {
    await DeleteBackup(deleteTarget.value.name)
    deleteOpen.value = false
    actionSuccess.value = `Backup "${deleteTarget.value.name}" dihapus.`
    await load()
  } catch (e) {
    deleteOpen.value = false
    actionError.value = errorMessage(e)
  }
}

function askRestore(b: BackupInfo) {
  restoreTarget.value = b
  restoreMessage.value = `Seluruh data saat ini akan diganti dengan isi backup "${b.name}".`
  restoreDetail.value =
    'Backup kondisi sekarang dibuat otomatis terlebih dahulu, lalu aplikasi ditutup dan dibuka kembali. Yakin lanjut?'
  restoreOpen.value = true
}

async function runRestore() {
  if (!restoreTarget.value) return
  restoring.value = true
  actionError.value = ''
  actionSuccess.value = ''
  try {
    const res = await RestoreBackup(restoreTarget.value.name)
    restoreOpen.value = false
    busy.value = true
    if (res.restarting) {
      restartMessage.value = res.message
      restarting.value = true
      setTimeout(() => void QuitApp(), 2500)
    }
  } catch (e) {
    restoreOpen.value = false
    actionError.value = errorMessage(e)
  } finally {
    restoring.value = false
  }
}

onMounted(load)
</script>