<template>
  <div>
    <PageHeader title="Pengaturan" subtitle="Profil perusahaan, penomoran SPH, dan preferensi dokumen" />

    <p v-if="store.error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.error }}
    </p>

    <div v-if="store.loading && !loadedOnce" class="rounded-xl border border-slate-200 bg-white px-4 py-6 text-center text-[13px] text-slate-400">Memuat…</div>

    <form v-else class="mx-auto max-w-3xl space-y-5" @submit.prevent="submit">
      <p v-if="saved" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-[13px] text-emerald-700">Pengaturan berhasil disimpan.</p>
      <p v-if="formError" class="rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ formError }}</p>

      <!-- Profil perusahaan -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Profil Perusahaan</h2>
        <div class="space-y-3.5">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Perusahaan <span class="text-red-500">*</span></label>
            <input v-model="form.companyName" type="text" maxlength="200" required class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div class="grid grid-cols-2 gap-x-4">
            <div>
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Kota</label>
              <input v-model="form.companyCity" type="text" maxlength="100" placeholder="Surabaya" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            </div>
            <div>
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Alamat</label>
              <input v-model="form.companyAddress" type="text" maxlength="500" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            </div>
          </div>
          <p class="text-xs text-slate-400">Tampil pada kop dokumen SPH hasil export.</p>
        </div>
      </section>

      <!-- Logo -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Logo Perusahaan</h2>
        <div class="flex items-start gap-4">
          <div class="flex h-28 w-28 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-dashed border-slate-300 bg-slate-50">
            <img v-if="logoDataUrl" :src="logoDataUrl" alt="Logo" class="max-h-full max-w-full object-contain" />
            <span v-else class="px-2 text-center text-xs text-slate-400">Belum ada logo</span>
          </div>
          <div class="min-w-0 flex-1 space-y-2">
            <p class="text-xs text-slate-500">PNG atau JPG, maksimal 5 MB. Disalin ke folder data aplikasi sehingga aman jika file asal dipindah.</p>
            <div class="flex flex-wrap gap-2">
              <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60" :disabled="logoBusy || saving" @click="pickLogo">
                {{ logoBusy ? 'Memproses…' : 'Pilih Logo…' }}
              </button>
              <button v-if="logoDataUrl" type="button" class="rounded-lg px-3.5 py-2 text-[13px] font-medium text-red-600 transition-colors hover:bg-red-50 disabled:opacity-60" :disabled="logoBusy || saving" @click="removeLogo">
                Hapus
              </button>
            </div>
            <p v-if="logoError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ logoError }}</p>
          </div>
        </div>
      </section>

      <!-- Penomoran SPH -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Penomoran SPH</h2>
        <div class="space-y-3.5">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Format Nomor <span class="text-red-500">*</span></label>
            <input v-model="form.sphNumberFormat" type="text" maxlength="100" required class="w-full rounded-lg border border-slate-200 px-3 py-2 font-mono text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <code v-for="t in placeholders" :key="t" class="cursor-pointer rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600 hover:bg-brand-50 hover:text-brand-700" title="Klik untuk menyisipkan ke format" @click="insertPlaceholder(t)">{{ t }}</code>
            <span class="ml-1 text-xs text-slate-400">— klik untuk menyisipkan</span>
          </div>
          <div class="rounded-lg bg-slate-50 px-3 py-2.5 text-[13px]">
            <span class="text-slate-500">Contoh nomor berikutnya:</span>
            <span v-if="numberError" class="text-red-600">{{ numberError }}</span>
            <span v-else-if="numberPreview" class="font-mono font-medium text-slate-700">{{ numberPreview }}</span>
            <span v-else class="italic text-slate-400">…</span>
          </div>
        </div>
      </section>

      <!-- Penandatangan -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Penandatangan Dokumen</h2>
        <div class="grid grid-cols-2 gap-x-4">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama</label>
            <input v-model="form.signerName" type="text" maxlength="100" placeholder="Matawai" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Jabatan</label>
            <input v-model="form.signerPosition" type="text" maxlength="100" placeholder="Direktur" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
        </div>
      </section>

      <!-- Stempel & tanda tangan -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Stempel & Tanda Tangan</h2>
        <p v-if="assetError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ assetError }}</p>
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div class="rounded-lg border border-slate-200 bg-slate-50/50 p-4">
            <h3 class="mb-1 text-[13px] font-semibold text-slate-700">Stempel Perusahaan</h3>
            <p class="mb-3 text-xs text-slate-400">PNG dengan latar transparan, dirender pada blok tanda tangan dokumen export.</p>
            <div class="flex items-start gap-3">
              <div class="flex h-24 w-24 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-dashed border-slate-300 bg-white">
                <img v-if="stampDataUrl" :src="stampDataUrl" alt="Stempel" class="max-h-full max-w-full object-contain" />
                <span v-else class="px-2 text-center text-xs text-slate-400">Belum ada stempel</span>
              </div>
              <div class="min-w-0 flex-1 space-y-2">
                <div class="flex flex-wrap gap-2">
                  <button type="button" class="rounded-lg border border-slate-200 px-3 py-1.5 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60" :disabled="assetBusy || saving" @click="pickStamp">
                    {{ assetBusy ? 'Memproses…' : 'Pilih Stempel…' }}
                  </button>
                  <button v-if="stampDataUrl" type="button" class="rounded-lg px-3 py-1.5 text-[13px] font-medium text-red-600 transition-colors hover:bg-red-50 disabled:opacity-60" :disabled="assetBusy || saving" @click="clearStamp">
                    Hapus
                  </button>
                </div>
              </div>
            </div>
            <div class="mt-4 space-y-3">
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Posisi Horizontal</span><span>{{ pct(form.stampPosX) }}</span></label>
                <input v-model.number="form.stampPosX" type="range" min="0" max="1" step="0.01" class="w-full accent-brand-600" @change="saveStampPosition" />
              </div>
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Posisi Vertikal</span><span>{{ pct(form.stampPosY) }}</span></label>
                <input v-model.number="form.stampPosY" type="range" min="0" max="1" step="0.01" class="w-full accent-brand-600" @change="saveStampPosition" />
              </div>
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Ukuran</span><span>{{ pct(form.stampSize) }}</span></label>
                <input v-model.number="form.stampSize" type="range" min="0.1" max="1" step="0.01" class="w-full accent-brand-600" @change="saveStampPosition" />
              </div>
            </div>
          </div>
          <div class="rounded-lg border border-slate-200 bg-slate-50/50 p-4">
            <h3 class="mb-1 text-[13px] font-semibold text-slate-700">Tanda Tangan</h3>
            <p class="mb-3 text-xs text-slate-400">PNG transparan hasil digitalisasi ttd, dirender tepat di atas baris nama penandatangan.</p>
            <div class="flex items-start gap-3">
              <div class="flex h-24 w-24 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-dashed border-slate-300 bg-white">
                <img v-if="signatureDataUrl" :src="signatureDataUrl" alt="Tanda tangan" class="max-h-full max-w-full object-contain" />
                <span v-else class="px-2 text-center text-xs text-slate-400">Belum ada ttd</span>
              </div>
              <div class="min-w-0 flex-1 space-y-2">
                <div class="flex flex-wrap gap-2">
                  <button type="button" class="rounded-lg border border-slate-200 px-3 py-1.5 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-60" :disabled="assetBusy || saving" @click="pickSignature">
                    {{ assetBusy ? 'Memproses…' : 'Pilih Tanda Tangan…' }}
                  </button>
                  <button v-if="signatureDataUrl" type="button" class="rounded-lg px-3 py-1.5 text-[13px] font-medium text-red-600 transition-colors hover:bg-red-50 disabled:opacity-60" :disabled="assetBusy || saving" @click="clearSignature">
                    Hapus
                  </button>
                </div>
              </div>
            </div>
            <div class="mt-4 space-y-3">
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Posisi Horizontal</span><span>{{ pct(form.signaturePosX) }}</span></label>
                <input v-model.number="form.signaturePosX" type="range" min="0" max="1" step="0.01" class="w-full accent-brand-600" @change="saveSignaturePosition" />
              </div>
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Posisi Vertikal</span><span>{{ pct(form.signaturePosY) }}</span></label>
                <input v-model.number="form.signaturePosY" type="range" min="0" max="1" step="0.01" class="w-full accent-brand-600" @change="saveSignaturePosition" />
              </div>
              <div>
                <label class="mb-1 flex justify-between text-xs font-medium text-slate-500"><span>Ukuran</span><span>{{ pct(form.signatureSize) }}</span></label>
                <input v-model.number="form.signatureSize" type="range" min="0.1" max="1" step="0.01" class="w-full accent-brand-600" @change="saveSignaturePosition" />
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Catatan default -->
      <section class="rounded-xl border border-slate-200 bg-white p-5">
        <h2 class="mb-4 text-sm font-semibold text-slate-800">Catatan Default Dokumen</h2>
        <textarea v-model="form.defaultNotes" rows="3" maxlength="1000" placeholder="Contoh: Harga penawaran berlaku 30 hari kerja…" class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"></textarea>
      </section>

      <div class="flex justify-end pb-2">
        <button type="submit" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="saving || logoBusy || assetBusy">
          {{ saving ? 'Menyimpan…' : 'Simpan Pengaturan' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import { useSettingsStore } from '../stores/settings'
import { errorMessage } from '../utils/format'
import { emptySettings, type SettingsView } from '../types/settings'
import {
  PreviewSphNumber,
  PickLogo,
  ClearLogo,
  LogoDataUrl,
  PickStamp,
  ClearStamp,
  StampDataUrl,
  PickSignature,
  ClearSignature,
  SignatureDataUrl,
  SetStampPosition,
  SetSignaturePosition
} from '../../wailsjs/go/main/App'

const store = useSettingsStore()

const loadedOnce = ref(false)
const saved = ref(false)
const saving = ref(false)
const formError = ref('')

const form = reactive<SettingsView>(emptySettings())

const logoDataUrl = ref('')
const logoBusy = ref(false)
const logoError = ref('')

const stampDataUrl = ref('')
const signatureDataUrl = ref('')
const assetBusy = ref(false)
const assetError = ref('')

const numberPreview = ref('')
const numberError = ref('')
let previewTimer: ReturnType<typeof setTimeout> | null = null

const placeholders = ['{YYYY}', '{MM}', '{ROMAN}', '{SEQ}']

onMounted(async () => {
  await store.load()
  Object.assign(form, store.settings)
  loadedOnce.value = true
  await Promise.all([loadLogo(), loadStamp(), loadSignature(), refreshPreview()])
})

function pct(v: number) {
  return `${Math.round((v || 0) * 100)}%`
}

async function loadLogo() {
  try {
    logoDataUrl.value = await LogoDataUrl()
  } catch {
    logoDataUrl.value = ''
  }
}

async function loadStamp() {
  try {
    stampDataUrl.value = await StampDataUrl()
  } catch {
    stampDataUrl.value = ''
  }
}

async function loadSignature() {
  try {
    signatureDataUrl.value = await SignatureDataUrl()
  } catch {
    signatureDataUrl.value = ''
  }
}

function schedulePreview() {
  if (previewTimer) clearTimeout(previewTimer)
  previewTimer = setTimeout(refreshPreview, 300)
}

watch(() => form.sphNumberFormat, schedulePreview)

async function refreshPreview() {
  try {
    numberPreview.value = await PreviewSphNumber(form.sphNumberFormat)
    numberError.value = ''
  } catch (e) {
    numberPreview.value = ''
    numberError.value = errorMessage(e)
  }
}

function insertPlaceholder(t: string) {
  form.sphNumberFormat = (form.sphNumberFormat || '') + t
}

async function submit() {
  formError.value = ''
  saved.value = false
  if (!form.companyName.trim()) {
    formError.value = 'Nama perusahaan wajib diisi.'
    return
  }
  if (!form.sphNumberFormat.includes('{SEQ}')) {
    formError.value = 'Format nomor SPH harus memuat placeholder {SEQ}.'
    return
  }
  saving.value = true
  try {
    const view = (await store.save({
      companyName: form.companyName,
      companyCity: form.companyCity,
      companyAddress: form.companyAddress,
      sphNumberFormat: form.sphNumberFormat,
      signerName: form.signerName,
      signerPosition: form.signerPosition,
      defaultNotes: form.defaultNotes,
      collabPort: form.collabPort ?? 48765,
      collabDisplayName: form.collabDisplayName ?? ''
    })) as unknown as SettingsView
    Object.assign(form, view)
    saved.value = true
    await refreshPreview()
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function pickLogo() {
  logoError.value = ''
  logoBusy.value = true
  try {
    const view = (await PickLogo()) as unknown as SettingsView | null
    if (view) {
      form.logoPath = view.logoPath
      await loadLogo()
    }
  } catch (e) {
    logoError.value = errorMessage(e)
  } finally {
    logoBusy.value = false
  }
}

async function removeLogo() {
  logoError.value = ''
  logoBusy.value = true
  try {
    const view = (await ClearLogo()) as unknown as SettingsView
    form.logoPath = view.logoPath
    logoDataUrl.value = ''
  } catch (e) {
    logoError.value = errorMessage(e)
  } finally {
    logoBusy.value = false
  }
}

async function pickStamp() {
  assetError.value = ''
  assetBusy.value = true
  try {
    const view = (await PickStamp()) as unknown as SettingsView | null
    if (view) {
      Object.assign(form, view)
      await loadStamp()
    }
  } catch (e) {
    assetError.value = errorMessage(e)
  } finally {
    assetBusy.value = false
  }
}

async function clearStamp() {
  assetError.value = ''
  assetBusy.value = true
  try {
    const view = (await ClearStamp()) as unknown as SettingsView
    Object.assign(form, view)
    stampDataUrl.value = ''
  } catch (e) {
    assetError.value = errorMessage(e)
  } finally {
    assetBusy.value = false
  }
}

async function pickSignature() {
  assetError.value = ''
  assetBusy.value = true
  try {
    const view = (await PickSignature()) as unknown as SettingsView | null
    if (view) {
      Object.assign(form, view)
      await loadSignature()
    }
  } catch (e) {
    assetError.value = errorMessage(e)
  } finally {
    assetBusy.value = false
  }
}

async function clearSignature() {
  assetError.value = ''
  assetBusy.value = true
  try {
    const view = (await ClearSignature()) as unknown as SettingsView
    Object.assign(form, view)
    signatureDataUrl.value = ''
  } catch (e) {
    assetError.value = errorMessage(e)
  } finally {
    assetBusy.value = false
  }
}

async function saveStampPosition() {
  if (assetBusy.value || saving.value) return
  try {
    const view = (await SetStampPosition(form.stampPosX, form.stampPosY, form.stampSize)) as unknown as SettingsView
    form.stampPosX = view.stampPosX
    form.stampPosY = view.stampPosY
    form.stampSize = view.stampSize
  } catch (e) {
    assetError.value = errorMessage(e)
  }
}

async function saveSignaturePosition() {
  if (assetBusy.value || saving.value) return
  try {
    const view = (await SetSignaturePosition(form.signaturePosX, form.signaturePosY, form.signatureSize)) as unknown as SettingsView
    form.signaturePosX = view.signaturePosX
    form.signaturePosY = view.signaturePosY
    form.signatureSize = view.signatureSize
  } catch (e) {
    assetError.value = errorMessage(e)
  }
}
</script>