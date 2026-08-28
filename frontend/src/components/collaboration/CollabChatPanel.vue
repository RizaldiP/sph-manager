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
        <div
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

    <div class="flex items-center gap-2 border-t border-slate-100 px-4 py-2.5">
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
import type { ChatMessage } from '../../types/collaboration'

const store = useCollaborationStore()
const draft = ref('')
const sending = ref(false)
const scrollRef = ref<HTMLElement | null>(null)

const messages = computed(() => store.messages)

const myName = computed(() => {
  const parts = store.snapshot.participants ?? []
  const me = parts.find(p => p.role !== 'HOST')
  return me?.displayName ?? store.snapshot.room?.hostName ?? ''
})

function isOwn(m: ChatMessage): boolean {
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
