<template>
  <div class="rounded-xl border border-slate-200 bg-white shadow-sm">
    <div class="flex items-center justify-between border-b border-slate-100 px-4 py-2.5">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-brand-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
        </svg>
        <h4 class="text-[13px] font-semibold text-slate-700">Master Data</h4>
        <span v-if="pendingInbox.length" class="rounded-full bg-rose-500 px-1.5 py-0.5 text-[10px] font-bold text-white">{{ pendingInbox.length }}</span>
      </div>
      <button
        type="button"
        :disabled="store.working"
        class="rounded-lg bg-brand-600 px-3 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-40"
        @click="openSend"
      >
        + Kirim Master Data
      </button>
    </div>

    <div v-if="store.error" class="mx-4 mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ store.error }}</div>

    <!-- Send dialog -->
    <div v-if="showSend" class="border-b border-slate-100 bg-slate-50/60 px-4 py-4">
      <div class="mb-2">
        <p class="text-[13px] font-semibold text-slate-700">Rakit &amp; Kirim Package</p>
        <p class="text-[12px] text-slate-400">Package berisi seluruh Master Data aktif di perangkat ini.</p>
      </div>

      <template v-if="built">
        <div class="mb-3 grid grid-cols-4 gap-2 text-center">
          <div class="rounded-lg border border-slate-200 bg-white p-2">
            <p class="text-lg font-bold text-slate-800">{{ built.data.categories?.length ?? 0 }}</p>
            <p class="text-[11px] text-slate-400">Kategori</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-2">
            <p class="text-lg font-bold text-slate-800">{{ built.data.workItems?.length ?? 0 }}</p>
            <p class="text-[11px] text-slate-400">Pekerjaan</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-2">
            <p class="text-lg font-bold text-slate-800">{{ built.data.workSubItems?.length ?? 0 }}</p>
            <p class="text-[11px] text-slate-400">Sub Pekerjaan</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-2">
            <p class="text-lg font-bold text-slate-800">{{ built.data.materials?.length ?? 0 }}</p>
            <p class="text-[11px] text-slate-400">Material</p>
          </div>
        </div>

        <p class="mb-1.5 text-[12px] font-medium text-slate-600">Kirim ke:</p>
        <div v-if="clientTargets.length" class="mb-3 space-y-1.5">
          <label v-for="t in clientTargets" :key="t.id" class="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-[13px] text-slate-700">
            <input type="checkbox" v-model="selectedTargets" :value="t.id" class="h-3.5 w-3.5 rounded border-slate-300 text-brand-600 focus:ring-brand-500" />
            <span class="font-medium">{{ t.displayName }}</span>
            <span class="ml-auto text-[11px] text-slate-400">{{ t.deviceName }}</span>
          </label>
        </div>
        <p v-else class="mb-3 text-[12px] italic text-slate-400">Tidak ada client yang terhubung.</p>

        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3 py-1.5 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-100" @click="showSend = false">
            Batal
          </button>
          <button
            type="button"
            :disabled="!selectedTargets.length || store.working"
            class="rounded-lg bg-emerald-600 px-3 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-emerald-700 disabled:opacity-40"
            @click="doSend"
          >
            {{ store.working ? 'Mengirim...' : 'Kirim' }}
          </button>
        </div>
      </template>

      <button
        v-else
        type="button"
        :disabled="store.working"
        class="w-full rounded-lg border border-brand-200 bg-white px-3 py-2 text-[13px] font-medium text-brand-700 transition-colors hover:bg-brand-50 disabled:opacity-40"
        @click="doBuild"
      >
        {{ store.working ? 'Merakit...' : 'Rakit Package' }}
      </button>
    </div>

    <!-- Inbox -->
    <div v-if="store.masterInbox.length" class="px-4 py-3">
      <h4 class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">Inbox ({{ store.masterInbox.length }})</h4>
      <div class="space-y-2">
        <div
          v-for="item in store.masterInbox"
          :key="item.packageId"
          class="rounded-lg border p-3 transition-colors"
          :class="item.status === 'INSTALLED' || item.status === 'REJECTED' ? 'border-slate-200 bg-slate-50' : 'border-brand-200 bg-brand-50/40'"
        >
          <div class="mb-1 flex items-start justify-between gap-2">
            <div class="min-w-0">
              <p class="truncate text-[13px] font-semibold text-slate-800">
                <span v-if="item.status === 'PENDING'" class="mr-1.5 inline-block h-2 w-2 rounded-full bg-rose-500 align-middle"></span>
                {{ item.title || 'Master Data' }}
              </p>
              <p class="text-[12px] text-slate-500">Dari {{ item.senderName }} · {{ item.itemCount }} item</p>
            </div>
            <span class="shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold" :class="statusBadge(item.status)">
              {{ MasterStatusLabel[item.status] ?? item.status }}
            </span>
          </div>
          <p v-if="item.summary" class="mb-2 line-clamp-2 text-[12px] text-slate-500">{{ item.summary }}</p>
          <p class="mb-2 text-[11px] text-slate-400">{{ formatTime(item.receivedAt) }}</p>

          <template v-if="previewOpen === item.packageId">
            <div v-if="store.masterPreview.length" class="mb-2 max-h-52 space-y-1 overflow-y-auto rounded-lg border border-slate-200 bg-white p-2">
              <div v-for="(d, i) in store.masterPreview" :key="i" class="flex items-center gap-2 text-[12px]">
                <span class="shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold" :class="diffBadge(d.kind)">{{ MasterDiffLabel[d.kind] ?? d.kind }}</span>
                <span class="truncate font-medium text-slate-700">{{ d.name || d.code || '-' }}</span>
                <span class="ml-auto shrink-0 text-[11px] text-slate-400">{{ d.summary }}</span>
              </div>
            </div>
            <p v-else class="mb-2 text-[12px] italic text-slate-400">Tidak ada perubahan yang terdeteksi.</p>

            <template v-if="item.status === 'PENDING' || item.status === 'VIEWED'">
              <div class="mb-2">
                <p class="mb-1 text-[11px] font-medium text-slate-500">Strategi konflik:</p>
                <div class="flex flex-wrap gap-1.5">
                  <label v-for="opt in strategyOptions" :key="opt.value" class="flex items-center gap-1 rounded border border-slate-200 bg-white px-2 py-1 text-[11px] text-slate-600">
                    <input type="radio" v-model="strategy" :value="opt.value" class="h-3 w-3 border-slate-300 text-brand-600 focus:ring-brand-500" />
                    {{ opt.label }}
                  </label>
                </div>
              </div>
              <div class="flex gap-2">
                <button
                  type="button"
                  :disabled="store.working"
                  class="flex-1 rounded-lg bg-emerald-600 px-3 py-1.5 text-[12px] font-medium text-white transition-colors hover:bg-emerald-700 disabled:opacity-40"
                  @click="doInstall(item.packageId)"
                >
                  {{ store.working ? 'Memasang...' : 'Pasang' }}
                </button>
                <button
                  type="button"
                  :disabled="store.working"
                  class="rounded-lg border border-red-200 px-3 py-1.5 text-[12px] font-medium text-red-700 transition-colors hover:bg-red-50 disabled:opacity-40"
                  @click="doReject(item.packageId)"
                >
                  Tolak
                </button>
              </div>
            </template>
          </template>

          <div v-else class="flex gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-medium text-slate-600 transition-colors hover:bg-slate-50"
              :disabled="store.working"
              @click="doPreview(item)"
            >
              {{ store.working && previewOpen === item.packageId ? 'Memuat...' : 'Pratinjau' }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="px-4 py-5 text-center">
      <p class="text-[12px] italic text-slate-400">Belum ada Master Data masuk.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCollaborationStore } from '../../stores/collaboration'
import {
  MasterDiffLabel,
  MasterStatusLabel,
  MasterStrategy,
  type MasterDataPackage
} from '../../types/collaboration'

const store = useCollaborationStore()

const showSend = ref(false)
const built = ref<MasterDataPackage | null>(null)
const selectedTargets = ref<string[]>([])
const previewOpen = ref<string>('')
const strategy = ref<string>(MasterStrategy.USE_INCOMING)

const pendingInbox = computed(() =>
  store.masterInbox.filter(i => i.status === 'PENDING' || i.status === 'VIEWED')
)

const clientTargets = computed(() =>
  (store.snapshot.participants ?? []).filter(p => p.role !== 'HOST')
)

const strategyOptions = [
  { value: MasterStrategy.USE_INCOMING, label: 'Pakai data masuk' },
  { value: MasterStrategy.USE_LOCAL, label: 'Pertahankan data lokal' },
  { value: MasterStrategy.SKIP, label: 'Lewati konflik' },
  { value: MasterStrategy.PROMPT, label: 'Tanya dulu' }
]

function statusBadge(status: string): string {
  switch (status) {
    case 'INSTALLED': return 'bg-emerald-100 text-emerald-700'
    case 'REJECTED': return 'bg-slate-200 text-slate-600'
    case 'FAILED': return 'bg-red-100 text-red-700'
    case 'VIEWED': return 'bg-amber-100 text-amber-700'
    default: return 'bg-rose-100 text-rose-700'
  }
}

function diffBadge(kind: string): string {
  switch (kind) {
    case 'NEW': return 'bg-emerald-100 text-emerald-700'
    case 'UPDATED': return 'bg-blue-100 text-blue-700'
    case 'CONFLICT': return 'bg-amber-100 text-amber-700'
    default: return 'bg-slate-200 text-slate-500'
  }
}

function formatTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
}

function openSend() {
  showSend.value = !showSend.value
}

async function doBuild() {
  try {
    built.value = await store.buildMasterDataPackage()
  } catch (e) {
    store.error = String(e)
  }
}

async function doSend() {
  if (!built.value || !selectedTargets.value.length) return
  try {
    await store.sendMasterData(built.value, selectedTargets.value)
    showSend.value = false
    built.value = null
    selectedTargets.value = []
    await store.refreshMasterInbox()
  } catch (e) {
    store.error = String(e)
  }
}

async function doPreview(item: { packageId: string; status: string }) {
  previewOpen.value = previewOpen.value === item.packageId ? '' : item.packageId
  if (previewOpen.value) {
    await store.previewMasterData(item.packageId)
    if (item.status === 'PENDING') {
      await store.markMasterInboxViewed(item.packageId)
    }
  }
}

async function doInstall(packageId: string) {
  try {
    await store.installMasterData(packageId, strategy.value, {})
    previewOpen.value = ''
  } catch (e) {
    store.error = String(e)
  }
}

async function doReject(packageId: string) {
  try {
    await store.rejectMasterData(packageId)
    previewOpen.value = ''
  } catch (e) {
    store.error = String(e)
  }
}
</script>
