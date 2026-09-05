<template>
  <div>
    <PageHeader title="Material" subtitle="Master material & suku cadang dengan harga default (FR-M7)">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="openForm()"
        >
          <span class="text-base leading-none">+</span> Material
        </button>
      </template>
    </PageHeader>

    <p v-if="store.error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.error }}
    </p>
    <p v-if="actionError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ actionError }}
    </p>

    <div class="rounded-xl border border-slate-200 bg-white">
      <div class="flex flex-wrap items-center gap-3 border-b border-slate-100 px-4 py-3">
        <div class="relative w-72">
          <svg class="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
          </svg>
          <input
            v-model="store.search"
            type="search"
            placeholder="Cari nama, kode, atau supplier…"
            class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <label class="flex cursor-pointer items-center gap-2 text-[13px] text-slate-600">
          <input v-model="store.includeInactive" type="checkbox" class="h-4 w-4 rounded border-slate-300 accent-brand-600" @change="store.load()" />
          Tampilkan yang nonaktif
        </label>
      </div>

      <table v-if="store.materials.length" class="w-full text-left">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50/60 text-[11px] uppercase tracking-wide text-slate-400">
            <th class="px-4 py-2 font-medium">Kode</th>
            <th class="px-3 py-2 font-medium">Nama</th>
            <th class="px-3 py-2 font-medium">Satuan</th>
            <th class="px-3 py-2 text-right font-medium">Harga Default</th>
            <th class="px-3 py-2 font-medium">Supplier</th>
            <th class="px-3 py-2 font-medium">Status</th>
            <th class="px-4 py-2 text-right font-medium">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in store.materials" :key="m.id" class="border-b border-slate-50 text-[13px] last:border-b-0" :class="{ 'opacity-55': !m.isActive }">
            <td class="px-4 py-2">
              <span v-if="m.code" class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ m.code }}</span>
              <span v-else class="text-slate-300">—</span>
            </td>
            <td class="max-w-[280px] px-3 py-2">
              <p class="truncate font-medium text-slate-700">{{ m.name }}</p>
              <p v-if="m.description" class="truncate text-xs text-slate-400">{{ m.description }}</p>
            </td>
            <td class="px-3 py-2 text-slate-500">{{ m.unit || '—' }}</td>
            <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(m.defaultPrice) }}</td>
            <td class="max-w-[180px] truncate px-3 py-2 text-slate-500">{{ m.supplier || '—' }}</td>
            <td class="px-3 py-2">
              <span
                class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                :class="m.isActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'"
              >
                {{ m.isActive ? 'Aktif' : 'Nonaktif' }}
              </span>
            </td>
            <td class="whitespace-nowrap px-4 py-2 text-right">
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openForm(m)">Edit</button>
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleActive(m)">
                {{ m.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
              </button>
              <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDelete(m)">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!store.loading" class="px-4 py-8 text-center text-[13px] text-slate-400">
        Belum ada material. Tambahkan material pertama untuk mempercepat pengisian harga.
      </div>

      <div v-if="store.loading" class="px-4 py-3 text-[13px] text-slate-400">Memuat…</div>
    </div>

    <!-- Form material -->
    <AppModal v-model="formOpen" :title="editing ? 'Edit Material' : 'Tambah Material'">
      <form class="space-y-3.5" @submit.prevent="submit" @keydown.ctrl.enter.prevent="submit">
        <p v-if="savedNote" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ savedNote }}</p>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama <span class="text-red-500">*</span></label>
          <input
            v-model="form.name"
            type="text"
            required
            maxlength="300"
            placeholder="Oli Sistem 15W40"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Satuan</label>
            <input
              v-model="form.unit"
              type="text"
              maxlength="30"
              placeholder="pcs / liter / drum"
              class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Harga Default (Rp)</label>
            <input
              v-model.number="form.defaultPrice"
              type="number"
              min="0"
              step="any"
              class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Supplier</label>
          <input
            v-model="form.supplier"
            type="text"
            maxlength="200"
            placeholder="PT Sumber Teknik"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Deskripsi</label>
          <textarea
            v-model="form.description"
            rows="2"
            maxlength="1000"
            class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          ></textarea>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
          <textarea
            v-model="form.notes"
            rows="2"
            maxlength="1000"
            class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          ></textarea>
        </div>
        <p class="text-xs italic text-slate-400">Kode material dibuat otomatis oleh sistem (MAT-…).</p>
        <p v-if="formError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ formError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="formOpen = false">Batal</button>
          <button type="button" :disabled="busy" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" @click="submit">
            {{ busy ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <ConfirmDialog
      v-model="confirmOpen"
      title="Hapus Material"
      :message="confirmMessage"
      confirm-label="Hapus"
      danger
      @confirm="runConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { useMaterialStore } from '../stores/material'
import { errorMessage, formatRupiah } from '../utils/format'
import { emptyMaterial, type MaterialView } from '../types/master'

const store = useMaterialStore()

// debounce pencarian server-side
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => store.search,
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => void store.load(), 250)
  }
)

// ===== form =====
const formOpen = ref(false)
const editing = ref<MaterialView | null>(null)
const form = reactive(emptyMaterial())
const formError = ref('')
const busy = ref(false)
const savedNote = ref('')

let savedNoteTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  savedNote.value = '✓ Data material tersimpan.'
  if (savedNoteTimer) clearTimeout(savedNoteTimer)
  savedNoteTimer = setTimeout(() => (savedNote.value = ''), 3000)
}

function openForm(m?: MaterialView) {
  formError.value = ''
  editing.value = m ?? null
  Object.assign(form, emptyMaterial(), m ? { ...m } : {})
  formOpen.value = true
}

async function submit() {
  busy.value = true
  formError.value = ''
  try {
    if (editing.value) {
      await store.update(editing.value.id, form)
      formOpen.value = false
    } else {
      await store.create(form)
      Object.assign(form, emptyMaterial())
      formError.value = ''
      flashSaved()
    }
    await store.load()
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function toggleActive(m: MaterialView) {
  actionError.value = ''
  try {
    await store.setActive(m.id, !m.isActive)
    await store.load()
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

// ===== konfirmasi hapus =====
const confirmOpen = ref(false)
const confirmMessage = ref('')
const confirmAction = ref<(() => Promise<void>) | null>(null)
const actionError = ref('')

function askDelete(m: MaterialView) {
  confirmMessage.value = `Material "${m.name}" akan dihapus dari daftar. Lanjutkan?`
  confirmAction.value = async () => {
    await store.remove(m.id)
    await store.load()
  }
  confirmOpen.value = true
}

async function runConfirm() {
  if (!confirmAction.value) return
  actionError.value = ''
  try {
    await confirmAction.value()
    confirmOpen.value = false
  } catch (e) {
    confirmOpen.value = false
    actionError.value = errorMessage(e)
  }
}

onMounted(() => void store.load())
</script>
