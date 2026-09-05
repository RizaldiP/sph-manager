<template>
  <div>
    <PageHeader title="Customer & Kapal" subtitle="Data pelanggan dan armada kapal untuk header SPH">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="openCustomerForm()"
        >
          <span class="text-base leading-none">+</span> Customer
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
            placeholder="Cari kode atau nama customer…"
            class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <label class="flex cursor-pointer items-center gap-2 text-[13px] text-slate-600">
          <input v-model="store.includeInactive" type="checkbox" class="h-4 w-4 rounded border-slate-300 accent-brand-600" @change="store.load()" />
          Tampilkan yang nonaktif
        </label>
      </div>

      <p v-if="!store.customers.length && !store.loading" class="px-4 py-8 text-center text-[13px] text-slate-400">
        Belum ada customer. Tambahkan data pelanggan pertama untuk mulai menyusun SPH.
      </p>

      <!-- Kartu per customer berisi tabel kapalnya -->
      <div v-for="c in store.customers" :key="c.id" class="border-b border-slate-100 last:border-b-0">
        <div class="flex flex-wrap items-center gap-2 px-4 py-3">
          <button
            type="button"
            class="flex min-w-0 flex-1 items-center gap-2 text-left"
            @click="toggleExpand(c.id)"
          >
            <svg
              class="h-4 w-4 shrink-0 text-slate-400 transition-transform"
              :class="{ 'rotate-90': expanded.has(c.id) }"
              fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
            <span v-if="c.code" class="shrink-0 rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ c.code }}</span>
            <span class="truncate font-medium text-slate-700">{{ c.name }}</span>
            <span v-if="!c.isActive" class="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500">Nonaktif</span>
            <span class="ml-auto shrink-0 text-xs text-slate-400">{{ c.vessels.length }} kapal</span>
          </button>
          <button class="rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openVesselForm(c)">+ Kapal</button>
          <button class="rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openCustomerForm(c)">Edit</button>
          <button class="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleActive(c)">
            {{ c.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
          </button>
          <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDeleteCustomer(c)">Hapus</button>
        </div>

        <table v-if="expanded.has(c.id) && c.vessels.length" class="mb-1 w-full text-left">
          <thead>
            <tr class="border-y border-slate-100 bg-slate-50/60 text-[11px] uppercase tracking-wide text-slate-400">
              <th class="py-2 pl-12 pr-3 font-medium">Kode</th>
              <th class="px-3 py-2 font-medium">Nama Kapal</th>
              <th class="px-3 py-2 font-medium">Nomor</th>
              <th class="px-3 py-2 font-medium">Jenis</th>
              <th class="px-3 py-2 font-medium">Status</th>
              <th class="px-3 py-2 text-right font-medium">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="v in c.vessels" :key="v.id" class="border-b border-slate-50 text-[13px] last:border-b-0" :class="{ 'opacity-55': !v.isActive }">
              <td class="py-2 pl-12 pr-3">
                <span v-if="v.code" class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ v.code }}</span>
                <span v-else class="text-slate-300">—</span>
              </td>
              <td class="px-3 py-2 font-medium text-slate-700">{{ v.name }}</td>
              <td class="px-3 py-2 tabular-nums text-slate-500">{{ v.vesselNumber || '—' }}</td>
              <td class="px-3 py-2 text-slate-500">{{ v.vesselType || '—' }}</td>
              <td class="px-3 py-2">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="v.isActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'"
                >
                  {{ v.isActive ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="whitespace-nowrap px-3 py-2 text-right">
                <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openVesselForm(c, v)">Edit</button>
                <button class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleVessel(v)">
                  {{ v.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
                </button>
                <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDeleteVessel(v)">Hapus</button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else-if="expanded.has(c.id)" class="py-2 pl-12 pr-4 text-xs italic text-slate-400">Belum ada kapal terdaftar.</p>
      </div>

      <div v-if="store.loading" class="px-4 py-3 text-[13px] text-slate-400">Memuat…</div>
    </div>

    <!-- Form customer -->
    <AppModal v-model="customerOpen" :title="editingCustomer ? 'Edit Customer' : 'Tambah Customer'" size="lg">
      <form class="grid grid-cols-1 gap-3.5 md:grid-cols-2" @submit.prevent="submitCustomer" @keydown.ctrl.enter.prevent="submitCustomer">
        <p v-if="customerSavedNote" class="md:col-span-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ customerSavedNote }}</p>
        <div class="md:col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama <span class="text-red-500">*</span></label>
          <input
            v-model="customerForm.name"
            type="text"
            required
            maxlength="200"
            placeholder="PT Laut Biru"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
          <input
            v-model="customerForm.code"
            type="text"
            maxlength="50"
            placeholder="PTL"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Telepon</label>
          <input
            v-model="customerForm.phone"
            type="text"
            maxlength="50"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div class="md:col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Alamat</label>
          <input
            v-model="customerForm.address"
            type="text"
            maxlength="500"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Email</label>
          <input
            v-model="customerForm.email"
            type="email"
            maxlength="150"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">PIC</label>
          <input
            v-model="customerForm.picName"
            type="text"
            maxlength="150"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div class="md:col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Posisi PIC</label>
          <input
            v-model="customerForm.picPosition"
            type="text"
            maxlength="150"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div class="md:col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
          <textarea
            v-model="customerForm.notes"
            rows="2"
            maxlength="1000"
            class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          ></textarea>
        </div>
        <p v-if="formError" class="md:col-span-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ formError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="customerOpen = false">Batal</button>
          <button type="button" :disabled="busy" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" @click="submitCustomer">
            {{ busy ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Form kapal -->
    <AppModal v-model="vesselOpen" :title="editingVessel ? 'Edit Kapal' : `Tambah Kapal — ${vesselCustomer?.name ?? ''}`">
      <form class="space-y-3.5" @submit.prevent="submitVessel" @keydown.ctrl.enter.prevent="submitVessel">
        <p v-if="vesselSavedNote" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ vesselSavedNote }}</p>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Kapal <span class="text-red-500">*</span></label>
          <input
            v-model="vesselForm.name"
            type="text"
            required
            maxlength="200"
            placeholder="KM Bahari"
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
            <input
              v-model="vesselForm.code"
              type="text"
              maxlength="50"
              class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Nomor Kapal</label>
            <input
              v-model="vesselForm.vesselNumber"
              type="text"
              maxlength="100"
              class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Jenis</label>
          <input
            v-model="vesselForm.vesselType"
            type="text"
            maxlength="150"
            placeholder="Tugboat / Barge / dsb."
            class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
          <textarea
            v-model="vesselForm.notes"
            rows="2"
            maxlength="1000"
            class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          ></textarea>
        </div>
        <p v-if="formError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ formError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="vesselOpen = false">Batal</button>
          <button type="button" :disabled="busy" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" @click="submitVessel">
            {{ busy ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <ConfirmDialog
      v-model="confirmOpen"
      :title="confirmTitle"
      :message="confirmMessage"
      confirm-label="Hapus"
      danger
      @confirm="runConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { usePartnerStore } from '../stores/partner'
import { errorMessage } from '../utils/format'
import { emptyCustomer, emptyVessel, type CustomerView, type VesselView } from '../types/partner'

const store = usePartnerStore()

const expanded = ref(new Set<number>())
function toggleExpand(id: number) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}

// ===== form customer =====
const customerOpen = ref(false)
const editingCustomer = ref<CustomerView | null>(null)
const customerForm = reactive(emptyCustomer())
const formError = ref('')
const busy = ref(false)
const customerSavedNote = ref('')
const vesselSavedNote = ref('')

let savedNoteTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved(target: 'customer' | 'vessel', text: string) {
  const note = target === 'customer' ? customerSavedNote : vesselSavedNote
  note.value = text
  if (savedNoteTimer) clearTimeout(savedNoteTimer)
  savedNoteTimer = setTimeout(() => (note.value = ''), 3000)
}

function openCustomerForm(c?: CustomerView) {
  formError.value = ''
  editingCustomer.value = c ?? null
  Object.assign(customerForm, emptyCustomer(), c ? { ...c } : {})
  customerOpen.value = true
}

async function submitCustomer() {
  busy.value = true
  formError.value = ''
  try {
    if (editingCustomer.value) {
      await store.updateCustomer(editingCustomer.value.id, customerForm)
      customerOpen.value = false
    } else {
      await store.createCustomer(customerForm)
      Object.assign(customerForm, emptyCustomer())
      formError.value = ''
      flashSaved('customer', '✓ Data customer tersimpan.')
    }
    await store.load()
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function toggleActive(c: CustomerView) {
  actionError.value = ''
  try {
    await store.setCustomerActive(c.id, !c.isActive)
    await store.load()
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

// ===== form kapal =====
const vesselOpen = ref(false)
const editingVessel = ref<VesselView | null>(null)
const vesselCustomer = ref<CustomerView | null>(null)
const vesselForm = reactive(emptyVessel())

function openVesselForm(c: CustomerView, v?: VesselView) {
  formError.value = ''
  expanded.value = new Set(expanded.value).add(c.id)
  vesselCustomer.value = c
  editingVessel.value = v ?? null
  Object.assign(vesselForm, emptyVessel(c.id), v ? { ...v, customerId: c.id } : {})
  vesselOpen.value = true
}

async function submitVessel() {
  busy.value = true
  formError.value = ''
  try {
    if (editingVessel.value) {
      await store.updateVessel(editingVessel.value.id, vesselForm)
      vesselOpen.value = false
    } else {
      await store.createVessel(vesselForm)
      Object.assign(vesselForm, emptyVessel(vesselCustomer.value?.id ?? 0))
      formError.value = ''
      flashSaved('vessel', '✓ Data kapal tersimpan.')
    }
    await store.load()
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function toggleVessel(v: VesselView) {
  actionError.value = ''
  try {
    await store.setVesselActive(v.id, !v.isActive)
    await store.load()
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

// ===== konfirmasi hapus =====
const confirmOpen = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
const confirmAction = ref<(() => Promise<void>) | null>(null)
const actionError = ref('')

function askDeleteCustomer(c: CustomerView) {
  confirmTitle.value = 'Hapus Customer'
  confirmMessage.value = `Customer "${c.name}" beserta datanya akan dihapus dari daftar. Lanjutkan?`
  confirmAction.value = async () => {
    await store.deleteCustomer(c.id)
    await store.load()
  }
  confirmOpen.value = true
}

function askDeleteVessel(v: VesselView) {
  confirmTitle.value = 'Hapus Kapal'
  confirmMessage.value = `Kapal "${v.name}" akan dihapus dari daftar. Lanjutkan?`
  confirmAction.value = async () => {
    await store.deleteVessel(v.id)
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
