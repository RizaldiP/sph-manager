<template>
  <div>
    <div v-if="loadError" class="rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ loadError }}
    </div>

    <template v-if="doc">
      <PageHeader :title="doc.documentNumber" :subtitle="headerSubtitle">
        <template #actions>
          <button v-if="doc.status === 'DRAFT'" type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="router.push(`/sph/${doc.id}/edit`)">
            Edit Draft
          </button>
          <button type="button" :disabled="busyAction !== ''" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60" @click="doDuplicate">
            Duplikat
          </button>
          <button
            v-if="['FINAL', 'SENT', 'REJECTED'].includes(doc.status)"
            type="button"
            :disabled="busyAction !== ''"
            class="rounded-lg border border-brand-200 px-3.5 py-2 text-[13px] font-medium text-brand-700 transition-colors hover:bg-brand-50 disabled:opacity-60"
            @click="doRevision"
          >
            Buat Revisi
          </button>
        </template>
      </PageHeader>

      <p v-if="actionError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ actionError }}</p>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <!-- Dokumen -->
        <article class="lg:col-span-2 rounded-xl border border-slate-200 bg-white p-6 text-[13px] leading-relaxed">
          <header class="mb-5 flex items-start justify-between border-b border-slate-100 pb-4">
            <div>
              <p class="font-mono text-base font-bold text-slate-800">{{ doc.documentNumber }}</p>
              <p class="mt-0.5 text-slate-500">Tanggal: {{ fmtDate(doc.date) }}</p>
              <p v-if="doc.validUntil" class="text-slate-500">Berlaku s.d. {{ fmtDate(doc.validUntil) }}</p>
            </div>
            <span class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold" :class="statusToneOf(doc.status)">
              {{ statusLabelOf(doc.status) }}
              <span v-if="doc.revision > 0" class="ml-1">· Rev {{ doc.revision }}</span>
            </span>
          </header>

          <div class="mb-5 grid grid-cols-2 gap-x-6 gap-y-1.5 md:grid-cols-3">
            <Info label="Customer" :value="doc.customer?.name || '-'" />
            <Info label="Kapal" :value="doc.vessel?.name || '-'" />
            <Info label="PIC Customer" :value="doc.picName || '-'" />
            <Info label="Proyek" :value="doc.projectName || '-'" />
            <Info label="Subjek" :value="doc.subject || '-'" />
            <Info label="Referensi" :value="doc.reference || '-'" />
            <Info label="Lokasi" :value="doc.location || '-'" />
            <Info label="Dibuat" :value="fmtDateTime(doc.createdAt)" />
            <Info label="Finalisasi" :value="doc.finalizedAt ? fmtDateTime(doc.finalizedAt) : '-'" />
          </div>

          <table class="w-full text-left">
            <thead>
              <tr class="border-y border-slate-200 text-xs uppercase tracking-wide text-slate-500">
                <th class="py-2 pr-2 font-medium">No</th>
                <th class="px-2 py-2 font-medium">Uraian</th>
                <th class="px-2 py-2 text-right font-medium">Qty</th>
                <th class="px-2 py-2 text-center font-medium">Sat</th>
                <th class="px-2 py-2 text-right font-medium">Jasa</th>
                <th class="px-2 py-2 text-right font-medium">Mat.</th>
                <th class="pl-2 py-2 text-right font-medium">Jumlah</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(it, idx) in doc.items ?? []" :key="it.id">
                <tr class="border-b border-slate-50 align-top">
                  <td class="py-2 pr-2 text-slate-400">{{ idx + 1 }}</td>
                  <td class="max-w-[320px] px-2 py-2">
                    <p class="font-medium text-slate-800">{{ it.nameSnapshot }}</p>
                    <p v-if="it.descriptionSnapshot" class="text-slate-500">{{ it.descriptionSnapshot }}</p>
                    <p v-if="it.notes" class="mt-0.5 text-xs italic text-slate-400">{{ it.notes }}</p>
                  </td>
                  <td class="px-2 py-2 text-right tabular-nums">{{ formatQty(it.quantity) }}</td>
                  <td class="px-2 py-2 text-center text-slate-600">{{ it.unit }}</td>
                  <td class="whitespace-nowrap px-2 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(it.serviceUnitPrice) }}</td>
                  <td class="whitespace-nowrap px-2 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(it.materialUnitPrice) }}</td>
                  <td class="whitespace-nowrap pl-2 py-2 text-right font-medium tabular-nums text-slate-800">{{ formatRupiah(it.total) }}</td>
                </tr>
                <tr v-for="sub in it.subItems ?? []" :key="`${it.id}-${sub.id}`" class="border-b border-slate-50 align-top text-[12px] text-slate-500">
                  <td class="py-1.5 pr-2"></td>
                  <td class="max-w-[320px] py-1.5 pl-8 pr-2">
                    <p>
                      — {{ sub.nameSnapshot }}
                      <span v-if="it.pricingMode === 'PEMBOBOTAN'" class="ml-0.5 rounded bg-brand-50 px-1 text-[10px] font-semibold text-brand-700">{{ sub.weight }}%</span>
                    </p>
                    <p v-if="sub.descriptionSnapshot" class="text-slate-400">{{ sub.descriptionSnapshot }}</p>
                    <p v-if="it.pricingMode === 'PEMBOBOTAN'" class="mt-0.5 text-[11px] italic text-slate-400">alokasi dari bobot {{ sub.weight }}%</p>
                  </td>
                  <td class="px-2 py-1.5 text-right tabular-nums">{{ formatQty(sub.quantity) }}</td>
                  <td class="px-2 py-1.5 text-center">{{ sub.unit }}</td>
                  <td class="whitespace-nowrap px-2 py-1.5 text-right tabular-nums">{{ formatRupiah(sub.serviceUnitPrice) }}</td>
                  <td class="whitespace-nowrap px-2 py-1.5 text-right tabular-nums">{{ formatRupiah(sub.materialUnitPrice) }}</td>
                  <td class="whitespace-nowrap py-1.5 pl-2 text-right tabular-nums">{{ formatRupiah(sub.total) }}</td>
                </tr>
              </template>
            </tbody>
            <tfoot>
              <tr class="border-t border-slate-200 text-slate-600">
                <td colspan="5"></td>
                <td class="py-1.5 pl-2 text-right">Subtotal Jasa</td>
                <td class="py-1.5 text-right tabular-nums">{{ formatRupiah(doc.subtotalService) }}</td>
              </tr>
              <tr class="text-slate-600">
                <td colspan="5"></td>
                <td class="py-1.5 pl-2 text-right">Subtotal Material</td>
                <td class="py-1.5 text-right tabular-nums">{{ formatRupiah(doc.subtotalMaterial) }}</td>
              </tr>
              <tr class="font-bold text-brand-700">
                <td colspan="5"></td>
                <td class="py-2 pl-2 text-right">Grand Total</td>
                <td class="py-2 text-right text-[14px] tabular-nums">{{ formatRupiah(doc.grandTotal) }}</td>
              </tr>
            </tfoot>
          </table>
          <p class="mt-2 italic text-slate-500">Terbilang: <strong>{{ doc.terbilang }}</strong></p>
        </article>

        <!-- Panel aksi & riwayat -->
        <aside class="space-y-4">
          <div class="rounded-xl border border-slate-200 bg-white p-4">
            <h3 class="mb-3 text-[13px] font-semibold uppercase tracking-wide text-slate-400">Aksi Status</h3>
            <div class="space-y-2">
              <button
                v-for="act in availableActions"
                :key="act.target"
                type="button"
                :disabled="busyAction !== ''"
                class="flex w-full items-center justify-between rounded-lg border px-3.5 py-2.5 text-left text-[13px] font-medium transition-colors disabled:opacity-50"
                :class="act.danger ? 'border-red-200 text-red-700 hover:bg-red-50' : 'border-brand-200 text-brand-700 hover:bg-brand-50'"
                @click="askStatus(act)"
              >
                <span>{{ act.label }}</span>
                <span class="text-xs text-slate-400">→ {{ statusLabelOf(act.target) }}</span>
              </button>
              <button
                v-if="doc.status === 'DRAFT'"
                type="button"
                :disabled="busyAction !== ''"
                class="flex w-full items-center justify-between rounded-lg border border-red-200 px-3.5 py-2.5 text-left text-[13px] font-medium text-red-700 transition-colors hover:bg-red-50 disabled:opacity-50"
                @click="deleteOpen = true"
              >
                Hapus Draft
              </button>
              <p v-if="!availableActions.length && doc.status !== 'DRAFT'" class="text-[13px] italic text-slate-400">Tidak ada transisi lagi untuk status ini.</p>
            </div>
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-4">
            <h3 class="mb-3 text-[13px] font-semibold uppercase tracking-wide text-slate-400">Riwayat Revisi</h3>
            <p v-if="!(doc.revisions ?? []).length" class="text-[13px] italic text-slate-400">Belum ada revisi.</p>
            <ol v-else class="space-y-2.5">
              <li v-for="rev in [...(doc.revisions ?? [])].reverse()" :key="rev.id" class="flex gap-2.5 text-[13px]">
                <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-brand-400"></span>
                <div>
                  <p class="font-medium text-slate-700">Rev {{ rev.revisionNumber }}</p>
                  <p class="text-xs text-slate-400">{{ rev.note }} · {{ fmtDateTime(rev.createdAt) }}</p>
                </div>
              </li>
            </ol>
          </div>

          <div v-if="doc.notes" class="rounded-xl border border-slate-200 bg-white p-4">
            <h3 class="mb-2 text-[13px] font-semibold uppercase tracking-wide text-slate-400">Catatan</h3>
            <p class="whitespace-pre-wrap text-[13px] text-slate-600">{{ doc.notes }}</p>
          </div>
        </aside>
      </div>

      <ConfirmDialog v-model="statusOpen" :title="statusAction?.label ?? ''" :message="statusMessage" :confirm-label="statusAction?.label ?? 'Ya'" :danger="statusAction?.danger" :busy="busyAction !== ''" @confirm="confirmStatus" />
      <ConfirmDialog v-model="deleteOpen" title="Hapus Draft SPH" :message="`Draft ${doc.documentNumber} akan dihapus dan tidak bisa dikembalikan. Lanjutkan?`" confirm-label="Hapus" danger :busy="busyAction !== ''" @confirm="confirmDelete" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import { useSphStore } from '../stores/sph'
import { errorMessage, formatQty, formatRupiah } from '../utils/format'
import { statusLabelOf, statusToneOf, type SphDetail } from '../types/sph'
import { h, type FunctionalComponent } from 'vue'

// Kartu kecil label-nilai untuk ringkasan header dokumen.
const Info: FunctionalComponent<{ label: string; value: string }> = (props) =>
  h('div', { class: 'min-w-0' }, [
    h('p', { class: 'text-xs text-slate-400' }, props.label),
    h('p', { class: 'truncate font-medium text-slate-700' }, props.value || '-')
  ])
Info.props = ['label', 'value']

const route = useRoute()
const router = useRouter()
const store = useSphStore()

const docId = computed(() => Number(route.params.id || 0))
const doc = ref<SphDetail | null>(null)
const loadError = ref('')
const actionError = ref('')
const busyAction = ref('')

async function reload() {
  try {
    doc.value = await store.getDetail(docId.value)
  } catch (e) {
    loadError.value = errorMessage(e)
  }
}

const headerSubtitle = computed(() => {
  if (!doc.value) return ''
  const parts = [doc.value.customer?.name ?? '', doc.value.projectName].filter(Boolean)
  return parts.join(' · ')
})

// ===== transisi status (BR-08) =====
interface StatusAction {
  target: string
  label: string
  danger?: boolean
}

const availableActions = computed<StatusAction[]>(() => {
  switch (doc.value?.status) {
    case 'DRAFT':
      return [
        { target: 'REVIEW', label: 'Ajukan Review' },
        { target: 'CANCELLED', label: 'Batalkan Dokumen', danger: true }
      ]
    case 'REVIEW':
      return [{ target: 'FINAL', label: 'Finalisasi' }]
    case 'FINAL':
      return [{ target: 'SENT', label: 'Tandai Terkirim' }]
    case 'SENT':
      return [
        { target: 'ACCEPTED', label: 'Disetujui Customer' },
        { target: 'REJECTED', label: 'Ditolak Customer', danger: true }
      ]
    default:
      return []
  }
})

function fmtDate(v?: unknown): string {
  const s = String(v ?? '')
  return s ? s.slice(0, 10) : '-'
}
function fmtDateTime(v?: unknown): string {
  const s = String(v ?? '').replace('T', ' ')
  return s ? s.slice(0, 16) : '-'
}

// ===== aksi =====
const statusOpen = ref(false)
const statusAction = ref<StatusAction | null>(null)
const deleteOpen = ref(false)

const statusMessage = computed(() => {
  if (!doc.value || !statusAction.value) return ''
  const extra =
    statusAction.value.target === 'FINAL'
      ? ' Setelah final, isi dokumen tidak dapat diedit lagi.'
      : statusAction.value.target === 'CANCELLED'
        ? ' Dokumen yang dibatalkan tidak dapat dipakai lagi.'
        : ''
  return `Ubah status ${doc.value.documentNumber} menjadi "${statusLabelOf(statusAction.value.target)}"?${extra}`
})

function askStatus(act: StatusAction) {
  statusAction.value = act
  statusOpen.value = true
}

async function confirmStatus() {
  if (!statusAction.value) return
  busyAction.value = statusAction.value.target
  actionError.value = ''
  try {
    await store.setStatus(docId.value, statusAction.value.target)
    statusOpen.value = false
    await reload()
  } catch (e) {
    actionError.value = errorMessage(e)
    statusOpen.value = false
  } finally {
    busyAction.value = ''
    statusAction.value = null
  }
}

async function confirmDelete() {
  busyAction.value = 'DELETE'
  actionError.value = ''
  try {
    await store.remove(docId.value)
    await router.push('/sph')
  } catch (e) {
    actionError.value = errorMessage(e)
    deleteOpen.value = false
  } finally {
    busyAction.value = ''
  }
}

async function doDuplicate() {
  busyAction.value = 'DUPLICATE'
  actionError.value = ''
  try {
    const clone = await store.duplicate(docId.value)
    await router.push(`/sph/${clone.id}/edit`)
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    busyAction.value = ''
  }
}

async function doRevision() {
  busyAction.value = 'REVISION'
  actionError.value = ''
  try {
    const rev = await store.createRevision(docId.value)
    await router.push(`/sph/${rev.id}/edit`)
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    busyAction.value = ''
  }
}

onMounted(reload)
</script>
