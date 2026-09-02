<template>
  <div class="rounded-xl border border-slate-200 bg-white shadow-sm">
    <div class="flex items-center justify-between border-b border-slate-100 px-4 py-2.5">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
        </svg>
        <h4 class="text-[13px] font-semibold text-slate-700">Chat Room</h4>
        <span v-if="store.unreadCount" class="rounded-full bg-rose-500 px-1.5 py-0.5 text-[10px] font-bold text-white">{{ store.unreadCount }}</span>
      </div>
    </div>

    <div ref="scrollRef" class="h-64 space-y-2.5 overflow-y-auto px-4 py-3">
      <div v-if="!messages.length" class="flex h-full items-center justify-center">
        <p class="text-[12px] italic text-slate-400">Belum ada pesan. Mulai percakapan!</p>
      </div>
      <div
        v-for="m in messages"
        :key="m.messageId"
        class="flex"
        :class="isOwn(m) ? 'justify-end' : m.messageType === 'system' ? 'justify-center' : 'justify-start'"
      >
        <div v-if="m.messageType === 'master_data'" class="max-w-[90%]">
          <p v-if="!isOwn(m)" class="mb-0.5 text-[11px] font-medium text-slate-400">{{ m.senderName || 'System' }}</p>
          <div class="overflow-hidden rounded-xl border border-brand-200 bg-brand-50/50">
            <div class="border-b border-brand-100 bg-white px-3 py-2">
              <p class="text-[12px] font-semibold text-brand-700">Master Data Dikirim</p>
              <p class="text-[11px] text-slate-500">{{ m.content || 'Dikirim melalui kolaborasi' }}</p>
            </div>
            <div v-if="cardOpen === m.messageId" class="px-3 py-2.5">
              <div v-if="store.masterPreview.length" class="mb-2 max-h-40 space-y-1 overflow-y-auto rounded-lg border border-slate-200 bg-white p-2">
                <div v-for="(d, i) in store.masterPreview" :key="i" class="flex items-center gap-2 text-[11px]">
                  <span class="shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold" :class="diffBadge(d.kind)">{{ MasterDiffLabel[d.kind] ?? d.kind }}</span>
                  <span class="truncate font-medium text-slate-700">{{ d.name || d.code || '-' }}</span>
                  <span class="ml-auto shrink-0 text-[10px] text-slate-400">{{ d.summary }}</span>
                </div>
              </div>
              <p v-else class="mb-2 text-[12px] italic text-slate-400">{{ store.working ? 'Memuat…' : 'Tidak ada perubahan terdeteksi.' }}</p>
              <div class="flex gap-2">
                <button
                  type="button"
                  :disabled="store.working"
                  class="flex-1 rounded-lg bg-emerald-600 px-2.5 py-1.5 text-[11px] font-medium text-white transition-colors hover:bg-emerald-700 disabled:opacity-40"
                  @click="installFromChat(m.refPackage)"
                >
                  {{ store.working ? 'Memasang…' : 'Pasang' }}
                </button>
                <button
                  type="button"
                  :disabled="store.working"
                  class="rounded-lg border border-red-200 px-2.5 py-1.5 text-[11px] font-medium text-red-700 transition-colors hover:bg-red-50 disabled:opacity-40"
                  @click="rejectFromChat(m.refPackage)"
                >
                  Tolak
                </button>
              </div>
            </div>
            <button
              v-else
              type="button"
              class="w-full px-3 py-2 text-left text-[12px] font-medium text-brand-700 transition-colors hover:bg-brand-100"
              @click="openCard(m)"
            >
              {{ store.working && cardLoading === m.refPackage ? 'Memuat…' : 'Pratinjau &amp; Pasang →' }}
            </button>
          </div>
          <p class="mt-1 text-right text-[10px] text-slate-300">{{ formatTime(m.createdAt) }}</p>
        </div>
        <div
          v-else
          class="max-w-[85%] rounded-xl px-3 py-1.5 text-[13px]"
          :class="bubbleClass(m)"
        >
          <p v-if="m.messageType !== 'system'" class="mb-0.5 text-[11px] font-medium" :class="isOwn(m) ? 'text-emerald-100' : 'text-slate-400'">
            {{ m.senderName || 'System' }}
          </p>
          <p class="leading-snug" :class="isOwn(m) && m.messageType !== 'system' ? 'text-white' : 'text-slate-700'">{{ m.content }}</p>
          <p class="mt-0.5 text-right text-[10px]" :class="isOwn(m) && m.messageType !== 'system' ? 'text-emerald-200' : 'text-slate-300'">
            {{ formatTime(m.createdAt) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Send Master Data dialog (dipicu dari icon peniti) -->
    <div v-if="sendOpen" class="border-t border-slate-100 bg-slate-50/60 px-4 py-4">
      <div class="mb-2 flex items-start justify-between gap-2">
        <div>
          <p class="text-[13px] font-semibold text-slate-700">Rakit &amp; Kirim Master Data</p>
          <p class="text-[12px] text-slate-400">Package berisi seluruh Master Data aktif di perangkat ini.</p>
        </div>
        <button type="button" class="rounded p-1 text-slate-400 transition-colors hover:bg-slate-200" title="Tutup" @click="closeSend">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
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

        <p class="mb-2 flex items-center gap-1.5 text-[12px] text-slate-500">
          <svg class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125" /></svg>
          Dikirim ke semua anggota room
        </p>

        <p v-if="!allTargets.length" class="mb-2 flex items-start gap-1.5 rounded-lg border border-amber-200 bg-amber-50 px-2.5 py-1.5 text-[12px] text-amber-700">
          <svg class="mt-0.5 h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126z" /></svg>
          <span>Room belum memiliki member lain. Master Data tetap terkirim sebagai pesan chat, tetapi tidak ada yang dapat memasangnya.</span>
        </p>

        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3 py-1.5 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-100" @click="closeSend">
            Batal
          </button>
          <button
            type="button"
            :disabled="store.working"
            class="rounded-lg bg-emerald-600 px-3 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-emerald-700 disabled:opacity-40"
            @click="doSend"
          >
            {{ store.working ? 'Mengirim...' : 'Kirim ke Room' }}
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

    <!-- Error -->
    <div v-if="store.error" class="border-t border-slate-100 px-4 py-2 text-[12px] text-red-700">{{ store.error }}</div>

    <div class="relative flex items-center gap-2 border-t border-slate-100 px-4 py-2.5">
      <!-- Paperclip -->
      <div class="relative">
        <button
          type="button"
          class="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-brand-600"
          :title="attachOpen ? 'Tutup' : 'Lampirkan'"
          @click="attachOpen = !attachOpen"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M18.375 12.739l-7.693 7.693a4.5 4.5 0 01-6.364-6.364l10.94-10.94A3 3 0 1119.5 7.372L8.552 18.32m.009-.01l-.01.01m5.699-9.941l-7.81 7.81a1.5 1.5 0 002.112 2.13" />
          </svg>
        </button>
        <div v-if="attachOpen" class="fixed inset-0 z-10" @click="attachOpen = false"></div>
        <div v-if="attachOpen" class="absolute bottom-full left-0 z-20 mb-2 w-52 overflow-hidden rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
          <button
            type="button"
            class="flex w-full items-center gap-2.5 px-3.5 py-2 text-left text-[13px] text-slate-700 transition-colors hover:bg-brand-50"
            @click="openSend"
          >
            <svg class="h-4 w-4 text-brand-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
            </svg>
            <span class="font-medium">Kirim Master Data</span>
          </button>
        </div>
      </div>

      <input
        v-model="draft"
        class="min-w-0 flex-1 rounded-lg border border-slate-200 bg-slate-50 px-3 py-1.5 text-[13px] text-slate-700 placeholder:text-slate-400 focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
        placeholder="Tulis pesan…"
        @keydown.enter="send"
      />
      <button
        type="button"
        :disabled="!draft.trim() || sending"
        class="rounded-lg bg-brand-600 px-3 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-40"
        @click="send"
      >
        Kirim
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useCollaborationStore } from '../../stores/collaboration'
import {
  MasterDiffLabel,
  MasterStrategy,
  type ChatMessage,
  type MasterDataPackage
} from '../../types/collaboration'

const store = useCollaborationStore()
const draft = ref('')
const sending = ref(false)
const scrollRef = ref<HTMLElement | null>(null)
const cardOpen = ref<string>('')
const cardLoading = ref<string>('')
const attachOpen = ref(false)
const sendOpen = ref(false)
const built = ref<MasterDataPackage | null>(null)

const messages = computed(() => store.messages)

const ownId = computed(() => {
  const parts = store.snapshot.participants ?? []
  if (store.isHost) return parts.find(p => p.role === 'HOST')?.id ?? ''
  return parts.find(p => p.role !== 'HOST')?.id ?? ''
})

const myName = computed(() => {
  if (store.isHost) return store.snapshot.room?.hostName ?? ''
  const parts = store.snapshot.participants ?? []
  return parts.find(p => p.role !== 'HOST')?.displayName ?? ''
})

const allTargets = computed(() => {
  const parts = store.snapshot.participants ?? []
  return parts.filter(p => p.id !== ownId.value).map(p => p.id)
})

function isOwn(m: ChatMessage): boolean {
  if (m.senderId) return m.senderId === ownId.value
  return m.senderName === myName.value && m.senderName !== ''
}

function bubbleClass(m: ChatMessage): string {
  if (m.messageType === 'system' || m.messageType === 'master_data') {
    return 'border border-slate-200 bg-slate-50'
  }
  return isOwn(m) ? 'bg-emerald-600' : 'border border-slate-200 bg-white'
}

function formatTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

function diffBadge(kind: string): string {
  switch (kind) {
    case 'NEW': return 'bg-emerald-100 text-emerald-700'
    case 'UPDATED': return 'bg-blue-100 text-blue-700'
    case 'CONFLICT': return 'bg-amber-100 text-amber-700'
    default: return 'bg-slate-200 text-slate-500'
  }
}

function openSend() {
  attachOpen.value = false
  sendOpen.value = true
}

function closeSend() {
  sendOpen.value = false
  built.value = null
}

async function doBuild() {
  try {
    built.value = await store.buildMasterDataPackage()
  } catch (e) {
    store.error = String(e)
  }
}

async function doSend() {
  if (!built.value) return
  try {
    await store.sendMasterData(built.value, allTargets.value)
    closeSend()
    await scrollToBottom()
  } catch (e) {
    store.error = String(e)
  }
}

async function openCard(m: ChatMessage) {
  if (cardOpen.value === m.messageId) {
    cardOpen.value = ''
    return
  }
  cardOpen.value = m.messageId
  if (!m.refPackage) return
  cardLoading.value = m.refPackage
  try {
    await store.previewMasterData(m.refPackage)
  } finally {
    cardLoading.value = ''
  }
}

async function installFromChat(packageId?: string) {
  if (!packageId) return
  try {
    await store.installMasterData(packageId, MasterStrategy.USE_INCOMING, {})
    cardOpen.value = ''
    await scrollToBottom()
  } catch (e) {
    store.error = String(e)
  }
}

async function rejectFromChat(packageId?: string) {
  if (!packageId) return
  try {
    await store.rejectMasterData(packageId)
    cardOpen.value = ''
  } catch (e) {
    store.error = String(e)
  }
}

async function scrollToBottom() {
  await nextTick()
  if (scrollRef.value) {
    scrollRef.value.scrollTop = scrollRef.value.scrollHeight
  }
}

async function send() {
  const text = draft.value.trim()
  if (!text || sending.value) return
  sending.value = true
  try {
    await store.sendChat(text, 'text')
    draft.value = ''
    await scrollToBottom()
  } finally {
    sending.value = false
  }
}

watch(
  () => store.messages.length,
  () => scrollToBottom()
)

watch(
  () => store.unreadCount,
  (n) => {
    if (n && n > 0) store.clearChatUnread()
  },
  { immediate: true }
)

scrollToBottom()
</script>