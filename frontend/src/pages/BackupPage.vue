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
import { computed, onMounted, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { errorMessage } from '../utils/format'
import type { BackupInfo } from '../types/backup'
import { BackupNow, DeleteBackup, ImportBackup, ListBackups, OpenBackupFolder, QuitApp, RestoreBackup } from '../../wailsjs/go/main/App'

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