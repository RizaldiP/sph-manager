<template>
  <div>
    <PageHeader :title="meta.title" :subtitle="meta.subtitle">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="router.push('/sph/baru')"
        >
          <span class="text-base leading-none">+</span> SPH Baru
        </button>
      </template>
    </PageHeader>

    <p v-if="store.listError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.listError }}
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
            placeholder="Cari nomor, customer, atau proyek…"
            class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
        </div>
        <span class="text-xs text-slate-400">{{ store.list.length }} dokumen</span>
      </div>

      <table v-if="store.list.length" class="w-full text-left">
        <thead>
          <tr class="border-b border-slate-100 text-xs uppercase tracking-wide text-slate-400">
            <th class="px-4 py-2.5 font-medium">Nomor</th>
            <th class="px-3 py-2.5 font-medium">Tanggal</th>
            <th class="px-3 py-2.5 font-medium">Customer</th>
            <th class="px-3 py-2.5 font-medium">Proyek</th>
            <th class="px-3 py-2.5 text-center font-medium">Item</th>
            <th class="px-3 py-2.5 text-right font-medium">Grand Total</th>
            <th class="px-3 py-2.5 font-medium">Status</th>
            <th class="px-3 py-2.5 text-right font-medium">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="doc in store.list"
            :key="doc.id"
            class="cursor-pointer border-b border-slate-50 text-[13px] transition-colors last:border-b-0 hover:bg-slate-50/70"
            @click="router.push(`/sph/${doc.id}`)"
          >
            <td class="whitespace-nowrap px-4 py-2.5">
              <span class="font-mono text-xs font-semibold text-slate-700">{{ doc.documentNumber }}</span>
              <span v-if="doc.revision > 0" class="ml-1 rounded bg-brand-50 px-1 py-0.5 text-[11px] font-medium text-brand-600">Rev {{ doc.revision }}</span>
            </td>
            <td class="whitespace-nowrap px-3 py-2.5 tabular-nums text-slate-600">{{ doc.date }}</td>
            <td class="max-w-[180px] truncate px-3 py-2.5 text-slate-700">{{ doc.customerName }}</td>
            <td class="max-w-[220px] truncate px-3 py-2.5 text-slate-500">{{ doc.projectName || '—' }}</td>
            <td class="px-3 py-2.5 text-center tabular-nums text-slate-600">{{ doc.itemCount }}</td>
            <td class="whitespace-nowrap px-3 py-2.5 text-right font-medium tabular-nums text-slate-800">{{ formatRupiah(doc.grandTotal) }}</td>
            <td class="px-3 py-2.5">
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :class="statusToneOf(doc.status)">
                {{ statusLabelOf(doc.status) }}
              </span>
            </td>
            <td class="whitespace-nowrap px-3 py-2.5 text-right" @click.stop>
              <button v-if="doc.status === 'DRAFT'" class="mr-1 rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="router.push(`/sph/${doc.id}/edit`)">Edit</button>
              <button v-if="doc.status === 'DRAFT'" class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDelete(doc)">Hapus</button>
              <button v-if="doc.status !== 'DRAFT' && doc.status !== 'CANCELLED'" class="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="doDuplicate(doc)">Duplikat</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!store.listLoading" class="p-6">
        <EmptyState title="Belum ada SPH" :description="emptyText">
          <button
            type="button"
            class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
            @click="router.push('/sph/baru')"
          >
            + Buat SPH Pertama
          </button>
        </EmptyState>
      </div>

      <div v-if="store.listLoading" class="px-4 py-3 text-[13px] text-slate-400">Memuat…</div>
    </div>

    <ConfirmDialog
      v-model="deleteOpen"
      title="Hapus Draft SPH"
      :message="`Draft ${deleteTarget?.documentNumber ?? ''} akan dihapus dan tidak bisa dikembalikan. Lanjutkan?`"
      confirm-label="Hapus"
      danger
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { useSphStore } from '../stores/sph'
import { formatRupiah, errorMessage } from '../utils/format'
import { statusLabelOf, statusToneOf, type SphDocumentView, type SphScope } from '../types/sph'

// Satu halaman untuk tiga menu: Semua / Draft / Final (beda scope + copy).
const props = defineProps<{ mode?: 'all' | 'draft' | 'final' }>()

const route = useRoute()
const router = useRouter()
const store = useSphStore()

const mode = computed(() => props.mode ?? ((route.name === 'sph-draft' && 'draft') || (route.name === 'sph-final' && 'final') || 'all'))

const meta = computed(() => {
  switch (mode.value) {
    case 'draft':
      return { title: 'Draft SPH', subtitle: 'Dokumen yang masih disusun — belum final dan belum terkirim' }
    case 'final':
      return { title: 'SPH Final', subtitle: 'Dokumen final, terkirim, dan hasil keputusannya' }
    default:
      return { title: 'Semua SPH', subtitle: 'Seluruh dokumen penawaran di satu tempat' }
  }
})

const emptyText = computed(() =>
  mode.value === 'draft'
    ? 'Tidak ada draft saat ini. Mulai susun penawaran baru.'
    : mode.value === 'final'
      ? 'Belum ada dokumen yang difinalisasi.'
      : 'Buat SPH pertama dari master pekerjaan, template, atau salinan dokumen lama.'
)

const actionError = ref('')
const deleteOpen = ref(false)
const deleteTarget = ref<SphDocumentView | null>(null)

function askDelete(doc: SphDocumentView) {
  deleteTarget.value = doc
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  try {
    await store.remove(deleteTarget.value.id)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
  deleteTarget.value = null
}

async function doDuplicate(doc: SphDocumentView) {
  actionError.value = ''
  try {
    const clone = await store.duplicate(doc.id)
    await router.push(`/sph/${clone.id}/edit`)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

watch(
  () => [store.search, mode.value] as const,
  ([, m], [, prevMode]) => {
    if (m !== prevMode) return
    loadDebounced()
  }
)

let debounceTimer: ReturnType<typeof setTimeout> | undefined
function loadDebounced() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => void store.loadList(), 250)
}

onMounted(() => {
  const scopes: Record<string, SphScope> = { all: '', draft: 'open', final: 'final' }
  store.scope = scopes[mode.value]
  void store.loadList()
})

// Ganti tab sidebar antar list → muat ulang dengan scope baru.
watch(
  () => route.name,
  (name, old) => {
    if (name === old) return
    if (!['sph-list', 'sph-draft', 'sph-final'].includes(String(name))) return
    const scopes: Record<string, SphScope> = { 'sph-list': '', 'sph-draft': 'open', 'sph-final': 'final' }
    store.scope = scopes[String(name)]
    store.search = ''
    void store.loadList()
  }
)
</script>
