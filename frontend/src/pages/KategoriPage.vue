<template>
  <div>
    <PageHeader title="Kategori Pekerjaan" subtitle="Kelompok pekerjaan: Electrical, Mechanical, dan lainnya">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="openCreate"
        >
          <span class="text-base leading-none">+</span> Kategori
        </button>
      </template>
    </PageHeader>

    <p v-if="store.categoriesError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.categoriesError }}
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
            v-model="store.catSearch"
            type="search"
            placeholder="Cari kode atau nama kategori…"
            class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <label class="flex cursor-pointer items-center gap-2 text-[13px] text-slate-600">
          <input v-model="store.catIncludeInactive" type="checkbox" class="h-4 w-4 rounded border-slate-300 accent-brand-600" />
          Tampilkan yang nonaktif
        </label>
        <span v-if="canReorder" class="ml-auto text-xs text-slate-400">Seret baris untuk mengatur urutan</span>
      </div>

      <table v-if="store.categories.length" class="w-full text-left">
        <thead>
          <tr class="border-b border-slate-100 text-xs uppercase tracking-wide text-slate-400">
            <th class="w-10 px-4 py-2.5 font-medium"></th>
            <th class="px-3 py-2.5 font-medium">Kode</th>
            <th class="px-3 py-2.5 font-medium">Nama</th>
            <th class="px-3 py-2.5 font-medium">Deskripsi</th>
            <th class="px-3 py-2.5 font-medium">Pekerjaan</th>
            <th class="px-3 py-2.5 font-medium">Status</th>
            <th class="px-3 py-2.5 text-right font-medium">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(cat, idx) in store.categories"
            :key="cat.id"
            class="border-b border-slate-50 text-[13px] transition-colors last:border-b-0 hover:bg-slate-50/70"
            :class="{ 'opacity-55': !cat.isActive, 'cursor-grab': canReorder, 'bg-brand-50/40': draggingId === cat.id }"
            :draggable="canReorder"
            @dragstart="startDrag(cat.id)"
            @dragenter.prevent="enterDrag(cat.id)"
            @dragover.prevent
            @dragend="endDrag"
          >
            <td class="px-4 py-2.5 text-slate-300">
              <span v-if="canReorder" class="select-none">&#8942;&#8942;</span>
              <span v-else>{{ cat.sequence || idx + 1 }}</span>
            </td>
            <td class="px-3 py-2.5">
              <span v-if="cat.code" class="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ cat.code }}</span>
              <span v-else class="text-slate-300">—</span>
            </td>
            <td class="px-3 py-2.5 font-medium text-slate-700">{{ cat.name }}</td>
            <td class="max-w-[280px] truncate px-3 py-2.5 text-slate-500">{{ cat.description || '—' }}</td>
            <td class="px-3 py-2.5 tabular-nums text-slate-600">{{ cat.workItemCount }}</td>
            <td class="px-3 py-2.5">
              <span
                class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
                :class="cat.isActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="cat.isActive ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                {{ cat.isActive ? 'Aktif' : 'Nonaktif' }}
              </span>
            </td>
            <td class="whitespace-nowrap px-3 py-2.5 text-right">
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openEdit(cat)">Edit</button>
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleActive(cat)">
                {{ cat.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
              </button>
              <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDelete(cat)">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!store.categoriesLoading" class="p-6">
        <EmptyState
          title="Belum ada kategori"
          description="Buat kategori pertama untuk mulai menyusun master pekerjaan."
        >
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
            @click="openCreate"
          >
            + Kategori Pertama
          </button>
        </EmptyState>
      </div>

      <div v-if="store.categoriesLoading" class="px-4 py-3 text-[13px] text-slate-400">Memuat…</div>
    </div>

    <!-- Modal form kategori -->
    <AppModal v-model="formOpen" :title="editing ? 'Edit Kategori' : 'Tambah Kategori'">
      <form class="space-y-3.5" @submit.prevent="submitForm" @keydown.ctrl.enter.prevent="submitForm">
        <p v-if="savedNote" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ savedNote }}</p>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
          <input
            v-model="form.code"
            type="text"
            disabled
            placeholder="(otomatis)"
            class="w-full cursor-not-allowed rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-[13px] text-slate-400 outline-none"
          />
          <p class="mt-1 text-xs text-slate-400">Kode dibuat otomatis oleh sistem.</p>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama <span class="text-red-500">*</span></label>
          <input
            v-model="form.name"
            type="text"
            required
            maxlength="150"
            placeholder="misal Electrical"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Deskripsi</label>
          <textarea
            v-model="form.description"
            rows="3"
            maxlength="500"
            class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          ></textarea>
        </div>
        <p v-if="formError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ formError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50"
            :disabled="saving"
            @click="formOpen = false"
          >
            Batal
          </button>
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
            :disabled="saving"
            @click="submitForm"
          >
            {{ saving ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Konfirmasi hapus -->
    <ConfirmDialog
      v-model="deleteOpen"
      title="Hapus Kategori"
      :message="`Hapus kategori '${deleteTarget?.name ?? ''}'?`"
      detail="Data tidak akan tampil lagi di daftar utama namun tetap tersimpan."
      confirm-label="Hapus"
      :busy="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { useMasterStore } from '../stores/master'
import { errorMessage } from '../utils/format'
import type { CategoryView } from '../types/master'
import { useDragSort } from '../composables/useDragSort'

const store = useMasterStore()

const actionError = ref('')
const formOpen = ref(false)
const editing = ref<CategoryView | null>(null)
const form = ref({ code: '', name: '', description: '' })
const formError = ref('')
const saving = ref(false)
const savedNote = ref('')

let savedNoteTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  savedNote.value = '✓ Data kategori tersimpan.'
  if (savedNoteTimer) clearTimeout(savedNoteTimer)
  savedNoteTimer = setTimeout(() => (savedNote.value = ''), 3000)
}

const deleteOpen = ref(false)
const deleteTarget = ref<CategoryView | null>(null)
const deleting = ref(false)

// Drag urutan hanya aktif saat daftar tidak difilter (backend butuh urutan lengkap).
const listForDrag = computed(() => store.categories)
const canReorder = computed(() => !store.catSearch && store.categories.length > 1)

const { draggingId, startDrag, enterDrag, endDrag } = useDragSort(listForDrag, async (ids) => {
  if (!canReorder.value) return
  try {
    await store.reorderCategories(ids)
  } catch (e) {
    actionError.value = errorMessage(e)
    await store.loadCategories()
  }
})

function openCreate() {
  editing.value = null
  form.value = { code: '', name: '', description: '' }
  formError.value = ''
  formOpen.value = true
}

function openEdit(cat: CategoryView) {
  editing.value = cat
  form.value = { code: cat.code, name: cat.name, description: cat.description }
  formError.value = ''
  formOpen.value = true
}

async function submitForm() {
  formError.value = ''
  if (!form.value.name.trim()) {
    formError.value = 'Nama kategori wajib diisi.'
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await store.updateCategory(editing.value.id, form.value)
      formOpen.value = false
    } else {
      await store.createCategory(form.value)
      form.value = { code: '', name: '', description: '' }
      formError.value = ''
      flashSaved()
      await store.loadCategories()
    }
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function toggleActive(cat: CategoryView) {
  actionError.value = ''
  try {
    await store.setCategoryActive(cat.id, !cat.isActive)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

function askDelete(cat: CategoryView) {
  actionError.value = ''
  deleteTarget.value = cat
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await store.deleteCategory(deleteTarget.value.id)
    deleteOpen.value = false
  } catch (e) {
    actionError.value = errorMessage(e)
    deleteOpen.value = false
  } finally {
    deleting.value = false
  }
}

// debounce pencarian
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => store.catSearch,
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => store.loadCategories(), 250)
  }
)
watch(
  () => store.catIncludeInactive,
  () => store.loadCategories()
)

onMounted(() => store.loadCategories())
</script>
