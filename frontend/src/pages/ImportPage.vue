<template>
  <div>
    <PageHeader title="Import Excel" subtitle="Impor daftar pekerjaan dari file Excel ke Master Pekerjaan" />

    <!-- Langkah -->
    <ol class="mb-5 flex flex-wrap items-center gap-2 text-[13px]">
      <li v-for="(s, i) in stepLabels" :key="s" class="flex items-center gap-2">
        <span
          class="flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold"
          :class="step > i ? 'bg-brand-600 text-white' : step === i + 1 ? 'bg-brand-100 text-brand-700 ring-1 ring-brand-300' : 'bg-slate-100 text-slate-400'"
        >{{ i + 1 }}</span>
        <span :class="step === i + 1 ? 'font-medium text-slate-800' : 'text-slate-400'">{{ s }}</span>
        <span v-if="i < stepLabels.length - 1" class="mx-1 text-slate-300">—</span>
      </li>
    </ol>

    <p v-if="pageError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ pageError }}</p>

    <!-- ===== Langkah 1: pilih file ===== -->
    <section v-if="step === 1" class="rounded-xl border border-slate-200 bg-white p-5">
      <p class="mb-3 text-[13px] text-slate-500">Format yang didukung: <code class="rounded bg-slate-100 px-1">.xls</code> dan <code class="rounded bg-slate-100 px-1">.xlsx</code>. Seluruh baris akan dipratinjau lebih dulu — belum ada data yang tersimpan sampai langkah terakhir.</p>
      <button type="button" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="busy" @click="pickFile">
        {{ busy ? 'Membuka…' : 'Pilih File Excel…' }}
      </button>
      <p v-if="filePath" class="mt-3 truncate rounded-lg bg-slate-50 px-3 py-2 font-mono text-xs text-slate-600">{{ filePath }}</p>
    </section>

    <!-- ===== Langkah 2: sheet & mapping ===== -->
    <section v-else-if="step === 2" class="space-y-4">
      <div class="rounded-xl border border-slate-200 bg-white p-5">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Sheet</label>
            <select v-model="sheet" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" @change="loadPreview">
              <option v-for="s in sheets" :key="s" :value="s">{{ s }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Kategori Tujuan <span class="text-red-500">*</span></label>
            <select v-model.number="categoryId" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100">
              <option value="" disabled>Pilih kategori…</option>
              <option v-for="c in master.categories" :key="c.id" :value="c.id">{{ c.code }} — {{ c.name }}</option>
            </select>
          </div>
        </div>
        <p class="mt-2 text-xs text-slate-400">Seluruh hasil import ditambahkan ke kategori ini sebagai data baru.</p>
      </div>

      <div v-if="previewLoading" class="rounded-xl border border-slate-200 bg-white px-4 py-6 text-center text-[13px] text-slate-400">Menganalisis sheet…</div>

      <template v-if="preview && !previewLoading">
        <p v-for="(n, i) in preview.notes" :key="i" class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-[13px] text-amber-800">{{ n }}</p>

        <div class="rounded-xl border border-slate-200 bg-white p-5">
          <h2 class="mb-1 text-sm font-semibold text-slate-800">Mapping Kolom</h2>
          <p class="mb-4 text-xs text-slate-400">Petakan kolom Excel ke field aplikasi. Baris header pertama data diisi sesuai posisi baris judul berakhir.</p>
          <div class="grid grid-cols-2 gap-x-4 gap-y-3 md:grid-cols-3">
            <div v-for="f in mappingFields" :key="f.key">
              <label class="mb-1 block text-[13px] font-medium text-slate-600">{{ f.label }}</label>
              <select v-model.number="mapping[f.key]" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100">
                <option value="-1">— tidak dipetakan —</option>
                <option v-for="c in preview.totalCols" :key="c - 1" :value="c - 1">{{ colLetter(c - 1) }}</option>
              </select>
            </div>
            <div>
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Gabung Kolom Uraian</label>
              <input v-model.number="mapping.nameSpan" type="number" min="1" max="8" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            </div>
            <div>
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Baris Data Pertama</label>
              <input v-model.number="mapping.headerRows" type="number" min="0" :max="preview.totalRows - 1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            </div>
          </div>
          <div class="mt-3 flex flex-wrap gap-x-6 gap-y-2">
            <label class="flex items-center gap-2 text-[13px] text-slate-600">
              <input v-model="mapping.serviceTotal" type="checkbox" class="h-4 w-4 accent-brand-600" />
              Kolom JASA berisi nilai total (harga satuan = total ÷ qty)
            </label>
            <label class="flex items-center gap-2 text-[13px] text-slate-600">
              <input v-model="mapping.materialTotal" type="checkbox" class="h-4 w-4 accent-brand-600" />
              Kolom MATERIAL berisi nilai total
            </label>
          </div>
          <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[13px] text-slate-600">
            <span>Harga Satuan dipakai bila JASA &amp; MATERIAL kosong, dihitung sebagai:</span>
            <label class="flex cursor-pointer items-center gap-1.5">
              <input type="radio" :checked="mapping.unitPriceAs === 'service'" class="h-4 w-4 accent-brand-600" @change="setUnitPriceAs('service')" />
              Harga Jasa
            </label>
            <label class="flex cursor-pointer items-center gap-1.5">
              <input type="radio" :checked="mapping.unitPriceAs === 'material'" class="h-4 w-4 accent-brand-600" @change="setUnitPriceAs('material')" />
              Harga Material
            </label>
          </div>
        </div>

        <div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <div class="max-h-72 overflow-auto">
            <table class="w-full text-left text-xs">
              <thead class="sticky top-0 bg-slate-50 text-slate-500">
                <tr>
                  <th class="px-3 py-2 font-medium">&nbsp;</th>
                  <th v-for="c in preview.totalCols" :key="c" class="whitespace-nowrap px-3 py-2 font-mono font-medium">{{ colLetter(c - 1) }}{{ isMapped(c - 1) ? ' •' : '' }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, ri) in preview.grid" :key="ri" :class="ri === mapping.headerRows ? 'border-y-2 border-dashed border-brand-300 bg-brand-50/40' : ''">
                  <td class="whitespace-nowrap px-3 py-1.5 text-right font-mono text-slate-400">{{ ri + 1 }}</td>
                  <td v-for="c in preview.totalCols" :key="c" class="max-w-[220px] truncate whitespace-nowrap px-3 py-1.5 text-slate-600">{{ row[c - 1] }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="border-t border-slate-100 px-3 py-2 text-xs text-slate-400">Garis putus-putus menandai awal baris data. Tanda • menandakan kolom terpetakan.</p>
        </div>

        <div class="flex justify-between pb-2">
          <button type="button" class="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="resetAll">Batal</button>
          <button type="button" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="!canAnalyze || busy" @click="analyze">
            {{ busy ? 'Memproses…' : 'Analisis Baris →' }}
          </button>
        </div>
      </template>
    </section>

    <!-- ===== Langkah 3: pratinjau & klasifikasi ===== -->
    <section v-else-if="step === 3" class="space-y-4">
      <div class="flex flex-wrap gap-2 text-[13px]">
        <span class="rounded-lg bg-slate-100 px-3 py-1.5 text-slate-600">Induk: <b>{{ countBy('main') }}</b></span>
        <span class="rounded-lg bg-slate-100 px-3 py-1.5 text-slate-600">Sub: <b>{{ countBy('sub') }}</b></span>
        <span :class="['rounded-lg px-3 py-1.5', pendingUnknown() ? 'bg-red-100 text-red-700' : 'bg-slate-100 text-slate-600']">Perlu diputuskan: <b>{{ pendingUnknown() }}</b></span>
        <span class="rounded-lg bg-slate-100 px-3 py-1.5 text-slate-600">Bermasalah: <b>{{ errorCount() }}</b></span>
      </div>

      <p v-if="blockers.length" class="max-h-32 overflow-auto rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
        <span v-for="(b, i) in blockers" :key="i" class="block">{{ b }}</span>
      </p>

      <div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <div class="max-h-[480px] overflow-auto">
          <table class="w-full text-left text-[13px]">
            <thead class="sticky top-0 bg-slate-50 text-xs text-slate-500">
              <tr>
                <th class="px-3 py-2 font-medium">Baris</th>
                <th class="px-3 py-2 font-medium">Uraian</th>
                <th class="px-3 py-2 text-right font-medium">Qty</th>
                <th class="px-3 py-2 font-medium">Sat</th>
                <th class="px-3 py-2 text-right font-medium">Jasa (sat.)</th>
                <th class="px-3 py-2 text-right font-medium">Material (sat.)</th>
                <th class="px-3 py-2 font-medium">Klasifikasi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in rows" :key="r.rowIndex" :class="decisionOf(r) === '' ? 'bg-red-50/60' : r.errors?.length ? 'bg-amber-50/50' : ''" class="border-t border-slate-100 align-top">
                <td class="whitespace-nowrap px-3 py-2 font-mono text-xs text-slate-400">{{ r.rowIndex + 1 }}</td>
                <td class="min-w-[240px] px-3 py-2 text-slate-700">
                  <span v-if="r.marker" class="mr-1.5 inline-block rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-500">{{ r.marker }}</span>
                  {{ r.name || r.raw }}
                  <span v-for="(e, ei) in r.errors" :key="ei" class="block text-xs text-red-600">{{ e }}</span>
                  <span v-if="!r.name" class="block text-xs italic text-slate-400">(teks uraian kosong)</span>
                </td>
                <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-600">{{ formatQty(r.qty) }}</td>
                <td class="whitespace-nowrap px-3 py-2 text-slate-600">{{ r.unit || '—' }}</td>
                <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-600">{{ r.servicePrice ? formatRupiah(r.servicePrice) : '—' }}</td>
                <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-600">{{ r.materialPrice ? formatRupiah(r.materialPrice) : '—' }}</td>
                <td class="whitespace-nowrap px-3 py-2">
                  <span v-if="r.suggested !== 'unknown'" class="mr-2 inline-block rounded px-1.5 py-0.5 text-xs" :class="r.suggested === 'main' ? 'bg-brand-50 text-brand-700' : 'bg-emerald-50 text-emerald-700'">{{ r.suggested === 'main' ? 'Induk' : 'Sub' }}</span>
                  <select v-model="decisions[r.rowIndex]" class="rounded-lg border px-2 py-1 text-xs outline-none focus:ring-2 focus:ring-brand-100" :class="decisions[r.rowIndex] ? 'border-slate-200' : 'border-red-300 bg-red-50'">
                    <option value="">Pilih…</option>
                    <option value="main">Induk</option>
                    <option value="sub">Sub-Pekerjaan</option>
                    <option value="skip">Lewati</option>
                  </select>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="flex justify-between pb-2">
        <button type="button" class="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="step = 2">← Kembali</button>
        <button type="button" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="!canImport || importing" @click="runImport">
          {{ importing ? `Mengimport… ${progress.done}/${progress.total}` : `Import ${countIncluded()} Baris` }}
        </button>
      </div>

      <div v-if="importing" class="rounded-xl border border-slate-200 bg-white p-4">
        <div class="mb-2 flex justify-between text-xs text-slate-500"><span>Menyimpan ke database…</span><span>{{ progress.done }}/{{ progress.total }}</span></div>
        <div class="h-2 overflow-hidden rounded-full bg-slate-100">
          <div class="h-full rounded-full bg-brand-500 transition-all" :style="{ width: progressPct + '%' }"></div>
        </div>
      </div>
    </section>

    <!-- ===== Langkah 4: hasil ===== -->
    <section v-else class="rounded-xl border border-slate-200 bg-white p-8 text-center">
      <div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-emerald-100 text-2xl text-emerald-600">✓</div>
      <h2 class="text-base font-semibold text-slate-800">Import Selesai</h2>
      <p class="mt-1 mb-6 text-[13px] text-slate-500">Data berhasil disimpan ke kategori terpilih.</p>
      <div class="mx-auto grid max-w-md grid-cols-3 gap-3 text-center">
        <div class="rounded-xl bg-slate-50 px-3 py-4">
          <div class="text-2xl font-semibold text-slate-800">{{ result?.itemsCreated ?? 0 }}</div>
          <div class="text-xs text-slate-400">Pekerjaan</div>
        </div>
        <div class="rounded-xl bg-slate-50 px-3 py-4">
          <div class="text-2xl font-semibold text-slate-800">{{ result?.subsCreated ?? 0 }}</div>
          <div class="text-xs text-slate-400">Sub-Pekerjaan</div>
        </div>
        <div class="rounded-xl bg-slate-50 px-3 py-4">
          <div class="text-2xl font-semibold text-slate-800">{{ result?.skipped ?? 0 }}</div>
          <div class="text-xs text-slate-400">Dilewati</div>
        </div>
      </div>
      <div class="mt-6 flex justify-center gap-2">
        <button type="button" class="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="resetAll">Import File Lain</button>
        <RouterLink to="/pekerjaan" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700">Buka Master Pekerjaan</RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import { useMasterStore } from '../stores/master'
import { errorMessage, formatQty, formatRupiah } from '../utils/format'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import {
  PickImportFile,
  ListImportSheets,
  PreviewImportSheet,
  ParseImportRows,
  ValidateImportRows,
  RunWorkItemImport
} from '../../wailsjs/go/main/App'
import type { ColumnMapping, ConfirmRow, ImportResult, PreviewRow, RowDecision, SheetPreview } from '../types/import'

const master = useMasterStore()

const stepLabels = ['Pilih File', 'Sheet & Mapping', 'Pratinjau', 'Selesai']
const step = ref(1)
const busy = ref(false)
const pageError = ref('')

const filePath = ref('')
const sheets = ref<string[]>([])
const sheet = ref('')
const categoryId = ref<number | ''>('')

const previewLoading = ref(false)
const preview = ref<SheetPreview | null>(null)

const rows = ref<PreviewRow[]>([])
const decisions = reactive<Record<number, RowDecision>>({})

const importing = ref(false)
const progress = reactive({ done: 0, total: 0 })
const result = ref<ImportResult | null>(null)
const blockers = ref<string[]>([])

const mappingFields = [
  { key: 'nameCol', label: 'Uraian Kegiatan' },
  { key: 'qtyCol', label: 'JML / Qty' },
  { key: 'unitCol', label: 'Satuan' },
  { key: 'serviceCol', label: 'Harga JASA' },
  { key: 'materialCol', label: 'Harga MATERIAL' },
  { key: 'unitPriceCol', label: 'Harga Satuan (umum)' }
] as const

const mapping = reactive<ColumnMapping>({
  nameCol: -1,
  nameSpan: 3,
  qtyCol: -1,
  unitCol: -1,
  serviceCol: -1,
  materialCol: -1,
  unitPriceCol: -1,
  unitPriceAs: 'service',
  serviceTotal: true,
  materialTotal: true,
  headerRows: 1
})

function setUnitPriceAs(v: string) {
  mapping.unitPriceAs = v === 'material' ? 'material' : 'service'
}

onMounted(async () => {
  if (!master.categories.length) await master.loadCategories()
})

function colLetter(i: number): string {
  let s = ''
  let n = i
  do {
    s = String.fromCharCode(65 + (n % 26)) + s
    n = Math.floor(n / 26) - 1
  } while (n >= 0)
  return s
}

function currentMapping(): ColumnMapping {
  return { ...mapping }
}

// ===== Langkah 1 =====
async function pickFile() {
  pageError.value = ''
  busy.value = true
  try {
    const p = await PickImportFile()
    if (!p) return // dialog dibatalkan
    filePath.value = p
    sheets.value = (await ListImportSheets(p)) as unknown as string[]
    if (!sheets.value.length) {
      pageError.value = 'Tidak ada sheet yang dapat dibaca dari file ini.'
      filePath.value = ''
      return
    }
    sheet.value = sheets.value[0]
    resetAnalysis()
    step.value = 2
    await loadPreview()
  } catch (e) {
    pageError.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

// ===== Langkah 2 =====
async function loadPreview() {
  if (!filePath.value || !sheet.value) return
  previewLoading.value = true
  pageError.value = ''
  try {
    const pv = (await PreviewImportSheet(filePath.value, sheet.value)) as unknown as SheetPreview
    preview.value = pv
    Object.assign(mapping, pv.suggestedMapping)
  } catch (e) {
    pageError.value = errorMessage(e)
    preview.value = null
  } finally {
    previewLoading.value = false
  }
}

const canAnalyze = computed(() =>
  !!preview.value && mapping.nameCol >= 0 && mapping.headerRows >= 0 && categoryId.value !== ''
)

function isMapped(col: number): boolean {
  return (
    col === mapping.nameCol ||
    col === mapping.qtyCol ||
    col === mapping.unitCol ||
    col === mapping.serviceCol ||
    col === mapping.materialCol
  )
}

async function analyze() {
  if (!canAnalyze.value || !preview.value) return
  busy.value = true
  pageError.value = ''
  try {
    const parsed = (await ParseImportRows(filePath.value, sheet.value, currentMapping())) as unknown as PreviewRow[]
    rows.value = parsed
    for (const k of Object.keys(decisions)) delete decisions[Number(k)]
    for (const r of parsed) {
      decisions[r.rowIndex] =
        r.suggested === 'main' ? 'main'
        : r.suggested === 'sub' ? 'sub'
        : '' // unknown wajib diputuskan pengguna (FR-IE3)
    }
    step.value = 3
  } catch (e) {
    pageError.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

// ===== Langkah 3 =====
function decisionOf(r: PreviewRow): RowDecision {
  return decisions[r.rowIndex] ?? ''
}

function countBy(level: string): number {
  return rows.value.filter((r) => decisionOf(r) === level).length
}

function pendingUnknown(): number {
  return rows.value.filter((r) => r.suggested === 'unknown' && !decisionOf(r)).length
}

function errorCount(): number {
  return rows.value.filter((r) => decisionOf(r) !== 'skip' && decisionOf(r) !== '' && r.errors?.length).length
}

function countIncluded(): number {
  return rows.value.filter((r) => decisionOf(r) === 'main' || decisionOf(r) === 'sub').length
}

const canImport = computed(() => {
  if (!rows.value.length || categoryId.value === '') return false
  if (!countIncluded()) return false
  if (pendingUnknown()) return false
  return !errorCount()
})

function confirmList(): ConfirmRow[] {
  const out: ConfirmRow[] = []
  for (const r of rows.value) {
    const d = decisionOf(r)
    if (d === 'main' || d === 'sub' || d === 'skip') {
      out.push({ rowIndex: r.rowIndex, level: d })
    }
  }
  return out
}

async function runImport() {
  if (!canImport.value || categoryId.value === '') return
  importing.value = true
  pageError.value = ''
  blockers.value = []
  progress.done = 0
  progress.total = countIncluded()
  try {
    EventsOn('import:progress', (data: { done: number; total: number }) => {
      progress.done = data.done
      progress.total = data.total
    })
    const res = (await RunWorkItemImport(
      categoryId.value as number,
      filePath.value,
      sheet.value,
      currentMapping(),
      confirmList()
    )) as unknown as ImportResult
    result.value = res
    step.value = 4
  } catch (e) {
    pageError.value = errorMessage(e)
    try {
      blockers.value = (await ValidateImportRows(
        filePath.value,
        sheet.value,
        currentMapping(),
        confirmList()
      )) as unknown as string[]
    } catch {
      /* abaikan */
    }
  } finally {
    EventsOff('import:progress')
    importing.value = false
  }
}

const progressPct = computed(() => (progress.total ? Math.round((progress.done / progress.total) * 100) : 0))

// ===== util =====
function resetAnalysis() {
  preview.value = null
  rows.value = []
  result.value = null
  blockers.value = []
  for (const k of Object.keys(decisions)) delete decisions[Number(k)]
}

function resetAll() {
  filePath.value = ''
  sheets.value = []
  sheet.value = ''
  categoryId.value = ''
  resetAnalysis()
  step.value = 1
  pageError.value = ''
}

onUnmounted(() => EventsOff('import:progress'))
</script>
