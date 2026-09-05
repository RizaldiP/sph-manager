<template>
  <div>
    <PageHeader title="Template Pekerjaan" subtitle="Kumpulan pekerjaan siap pakai untuk menyusun SPH lebih cepat">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="openCreate"
        >
          <span class="text-base leading-none">+</span> Template
        </button>
      </template>
    </PageHeader>

    <p v-if="store.templatesError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.templatesError }}
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
            v-model="store.tplSearch"
            type="search"
            placeholder="Cari kode atau nama template…"
            class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <label class="flex cursor-pointer items-center gap-2 text-[13px] text-slate-600">
          <input v-model="store.tplIncludeInactive" type="checkbox" class="h-4 w-4 rounded border-slate-300 accent-brand-600" />
          Tampilkan yang nonaktif
        </label>
        <span v-if="canReorder" class="ml-auto text-xs text-slate-400">Seret baris untuk mengatur urutan</span>
      </div>

      <table v-if="store.templates.length" class="w-full text-left">
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
            v-for="(tpl, idx) in store.templates"
            :key="tpl.id"
            class="border-b border-slate-50 text-[13px] transition-colors last:border-b-0 hover:bg-slate-50/70"
            :class="{ 'opacity-55': !tpl.isActive, 'cursor-grab': canReorder, 'bg-brand-50/40': draggingId === tpl.id }"
            :draggable="canReorder"
            @dragstart="startDrag(tpl.id)"
            @dragenter.prevent="enterDrag(tpl.id)"
            @dragover.prevent
            @dragend="endDrag"
          >
            <td class="px-4 py-2.5 text-slate-300">
              <span v-if="canReorder" class="select-none">&#8942;&#8942;</span>
              <span v-else>{{ tpl.sequence || idx + 1 }}</span>
            </td>
            <td class="px-3 py-2.5">
              <span v-if="tpl.code" class="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ tpl.code }}</span>
              <span v-else class="text-slate-300">—</span>
            </td>
            <td class="px-3 py-2.5 font-medium text-slate-700">{{ tpl.name }}</td>
            <td class="max-w-[280px] truncate px-3 py-2.5 text-slate-500">{{ tpl.description || '—' }}</td>
            <td class="px-3 py-2.5 tabular-nums text-slate-600">{{ tpl.itemCount }}</td>
            <td class="px-3 py-2.5">
              <span
                class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
                :class="tpl.isActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="tpl.isActive ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                {{ tpl.isActive ? 'Aktif' : 'Nonaktif' }}
              </span>
            </td>
            <td class="whitespace-nowrap px-3 py-2.5 text-right">
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openItems(tpl)">Kelola Isi</button>
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openEdit(tpl)">Edit</button>
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="duplicate(tpl)">Duplikat</button>
              <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleActive(tpl)">
                {{ tpl.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
              </button>
              <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDelete(tpl)">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!store.templatesLoading" class="p-6">
        <EmptyState
          title="Belum ada template"
          description="Buat template pertama berisi kumpulan pekerjaan yang sering dipakai bersama."
        >
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
            @click="openCreate"
          >
            + Template Pertama
          </button>
        </EmptyState>
      </div>

      <div v-if="store.templatesLoading" class="px-4 py-3 text-[13px] text-slate-400">Memuat…</div>
    </div>

    <!-- Modal data template -->
    <AppModal v-model="formOpen" :title="editing ? 'Edit Template' : 'Tambah Template'">
      <form class="space-y-3.5" @submit.prevent="submitForm" @keydown.ctrl.enter.prevent="submitForm">
        <p v-if="savedNote" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ savedNote }}</p>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
          <input
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
            maxlength="200"
            placeholder="misal Repair PLC"
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

    <!-- Modal isi template -->
    <AppModal v-model="itemsOpen" :title="'Isi Template — ' + (itemsTarget?.name ?? '')" size="lg">
      <div class="space-y-4">
        <p class="text-[13px] text-slate-500">Susun daftar pekerjaan yang termasuk dalam template ini. Urutan di sini menjadi urutan awal saat dipakai.</p>

        <div class="flex flex-wrap items-end gap-2 rounded-lg border border-slate-200 bg-slate-50/70 p-3">
          <div class="min-w-[320px] flex-1">
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Tambah pekerjaan</label>
            <select
              v-model="pickId"
              class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            >
              <option :value="0" disabled>Pilih pekerjaan…</option>
              <optgroup v-for="cat in master.categories" :key="cat.id" :label="cat.name">
                <option v-for="wi in itemsByCategory(cat.id)" :key="wi.id" :value="wi.id">{{ wi.code }} — {{ wi.name }}</option>
              </optgroup>
            </select>
          </div>
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!pickId || alreadyPicked(pickId)"
            @click="addPick"
          >
            + Tambah
          </button>
        </div>

        <div v-if="pickId && alreadyPicked(pickId)" class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[13px] text-amber-700">
          Pekerjaan itu sudah ada di dalam template.
        </div>

        <div class="overflow-hidden rounded-lg border border-slate-200">
          <table v-if="rows.length" class="w-full text-left">
            <thead>
              <tr class="border-b border-slate-100 text-xs uppercase tracking-wide text-slate-400">
                <th class="w-10 px-3 py-2 font-medium"></th>
                <th class="w-10 px-2 py-2 font-medium">#</th>
                <th class="px-3 py-2 font-medium">Pekerjaan</th>
                <th class="px-3 py-2 font-medium">Catatan</th>
                <th class="w-14 px-3 py-2 text-right font-medium"></th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in rows"
                :key="row.id"
                class="border-b border-slate-50 text-[13px] transition-colors last:border-b-0 hover:bg-slate-50/70"
                :class="{ 'cursor-grab': rows.length > 1, 'bg-brand-50/40': draggingRow === row.id }"
                :draggable="rows.length > 1"
                @dragstart="startRowDrag(row.id)"
                @dragenter.prevent="enterRowDrag(row.id)"
                @dragover.prevent
                @dragend="endRowDrag"
              >
                <td class="px-3 py-2 text-slate-300">
                  <span v-if="rows.length > 1" class="select-none">&#8942;&#8942;</span>
                </td>
                <td class="px-2 py-2 tabular-nums text-slate-400">{{ idx + 1 }}</td>
                <td class="px-3 py-2 font-medium text-slate-700">{{ rowLabel(row) }}</td>
                <td class="px-3 py-2">
                  <input
                    v-model="row.notes"
                    type="text"
                    maxlength="500"
                    placeholder="opsional…"
                    class="w-full min-w-[160px] rounded-md border border-slate-200 px-2 py-1.5 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
                  />
                </td>
                <td class="px-3 py-2 text-right">
                  <button type="button" class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="removeRow(row.id)">Hapus</button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-else class="px-4 py-6 text-center text-[13px] text-slate-400">Belum ada pekerjaan di template ini.</p>
        </div>

        <p v-if="itemsError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ itemsError }}</p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50"
            :disabled="savingItems"
            @click="itemsOpen = false"
          >
            Batal
          </button>
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60"
            :disabled="savingItems"
            @click="submitItems"
          >
            {{ savingItems ? 'Menyimpan…' : 'Simpan Isi' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Konfirmasi hapus -->
    <ConfirmDialog
      v-model="deleteOpen"
      title="Hapus Template"
      :message="`Hapus template '${deleteTarget?.name ?? ''}'?`"
      detail="Data tidak akan tampil lagi di daftar utama namun tetap tersimpan."
      confirm-label="Hapus"
      :busy="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { useTemplateStore } from '../stores/template'
import { useMasterStore } from '../stores/master'
import { errorMessage } from '../utils/format'
import type { TemplateView } from '../types/template'
import { useDragSort } from '../composables/useDragSort'

const store = useTemplateStore()
const master = useMasterStore()

const actionError = ref('')

// ===== daftar template =====
const listForDrag = computed(() => store.templates)
const canReorder = computed(() => !store.tplSearch && store.templates.length > 1)

const { draggingId, startDrag, enterDrag, endDrag } = useDragSort(listForDrag, async (ids) => {
  if (!canReorder.value) return
  try {
    await store.reorderTemplates(ids)
  } catch (e) {
    actionError.value = errorMessage(e)
    await store.loadTemplates()
  }
})

// ===== form data template =====
const formOpen = ref(false)
const editing = ref<TemplateView | null>(null)
const form = ref({ name: '', description: '', notes: '' })
const formError = ref('')
const saving = ref(false)
const savedNote = ref('')

let savedNoteTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  savedNote.value = '✓ Data template tersimpan.'
  if (savedNoteTimer) clearTimeout(savedNoteTimer)
  savedNoteTimer = setTimeout(() => (savedNote.value = ''), 3000)
}

function openCreate() {
  editing.value = null
  form.value = { name: '', description: '', notes: '' }
  formError.value = ''
  formOpen.value = true
}

function openEdit(tpl: TemplateView) {
  editing.value = tpl
  form.value = { name: tpl.name, description: tpl.description, notes: tpl.notes }
  formError.value = ''
  formOpen.value = true
}

async function submitForm() {
  formError.value = ''
  if (!form.value.name.trim()) {
    formError.value = 'Nama template wajib diisi.'
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await store.updateTemplate(editing.value.id, form.value)
      formOpen.value = false
    } else {
      await store.createTemplate(form.value)
      form.value = { name: '', description: '', notes: '' }
      formError.value = ''
      flashSaved()
      await store.loadTemplates()
    }
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

// ===== aksi baris =====
async function toggleActive(tpl: TemplateView) {
  actionError.value = ''
  try {
    await store.setTemplateActive(tpl.id, !tpl.isActive)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

async function duplicate(tpl: TemplateView) {
  actionError.value = ''
  try {
    await store.duplicateTemplate(tpl.id)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

const deleteOpen = ref(false)
const deleteTarget = ref<TemplateView | null>(null)
const deleting = ref(false)

function askDelete(tpl: TemplateView) {
  actionError.value = ''
  deleteTarget.value = tpl
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await store.deleteTemplate(deleteTarget.value.id)
    deleteOpen.value = false
  } catch (e) {
    actionError.value = errorMessage(e)
    deleteOpen.value = false
  } finally {
    deleting.value = false
  }
}

// ===== editor isi template =====
interface Row {
  id: number
  workItemId: number
  notes: string
}

let uidSeq = -1
function nextUid() {
  return uidSeq--
}

const itemsOpen = ref(false)
const itemsTarget = ref<TemplateView | null>(null)
const rows = ref<Row[]>([])
const itemsError = ref('')
const savingItems = ref(false)
const pickId = ref(0)
const fallbackLabels = ref(new Map<number, string>())

const { draggingId: draggingRow, startDrag: startRowDrag, enterDrag: enterRowDrag, endDrag: endRowDrag } = useDragSort(rows, async () => {
  /* urutan lokal saja; tersimpan saat Simpan Isi */
})

async function openItems(tpl: TemplateView) {
  itemsError.value = ''
  pickId.value = 0
  itemsTarget.value = tpl
  try {
    // Muat referensi kategori & pekerjaan aktif untuk pemilih (abaikan filter halaman lain).
    master.wiCategoryId = 0
    master.wiSearch = ''
    master.wiIncludeInactive = false
    await Promise.all([master.loadCategories(), master.loadWorkItems()])

    const detail = await store.getTemplateDetail(tpl.id)
    fallbackLabels.value = new Map()
    rows.value = (detail.items ?? []).map((it) => {
      if (it.workItem?.name) {
        const label = [it.workItem?.code, it.workItem?.name].filter(Boolean).join(' — ')
        fallbackLabels.value.set(it.workItemId, label)
      }
      return { id: it.id ?? nextUid(), workItemId: it.workItemId, notes: it.notes ?? '' }
    })
    itemsOpen.value = true
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

function itemsByCategory(categoryId: number) {
  return master.workItems.filter((wi) => wi.categoryId === categoryId)
}

function alreadyPicked(workItemId: number) {
  return rows.value.some((r) => r.workItemId === workItemId)
}

function addPick() {
  if (!pickId.value || alreadyPicked(pickId.value)) return
  rows.value.push({ id: nextUid(), workItemId: pickId.value, notes: '' })
  pickId.value = 0
}

function removeRow(id: number) {
  rows.value = rows.value.filter((r) => r.id !== id)
}

function rowLabel(row: Row): string {
  const wi = master.workItems.find((w) => w.id === row.workItemId)
  if (wi) return `${wi.code} — ${wi.name}`
  return fallbackLabels.value.get(row.workItemId) ?? `Pekerjaan #${row.workItemId}`
}

async function submitItems() {
  if (!itemsTarget.value) return
  itemsError.value = ''
  if (rows.value.length === 0) {
    itemsError.value = 'Template masih kosong. Tambahkan minimal satu pekerjaan.'
    return
  }
  savingItems.value = true
  try {
    await store.saveItems(
      itemsTarget.value.id,
      rows.value.map((r) => ({ workItemId: r.workItemId, notes: r.notes.trim() }))
    )
    itemsOpen.value = false
  } catch (e) {
    itemsError.value = errorMessage(e)
  } finally {
    savingItems.value = false
  }
}

onMounted(() => store.loadTemplates())
</script>
