<template>
  <AppModal :model-value="modelValue" title="Pilih Material" @update:model-value="emit('update:modelValue', $event)">
    <div class="relative mb-3">
      <svg class="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
      </svg>
      <input
        v-model="search"
        type="search"
        placeholder="Cari nama, kode, atau supplier…"
        class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
      />
    </div>

    <p v-if="store.error" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ store.error }}</p>

    <ul class="max-h-[320px] divide-y divide-slate-100 overflow-y-auto rounded-lg border border-slate-200">
      <li v-if="store.loading" class="px-4 py-6 text-center text-[13px] text-slate-400">Memuat…</li>
      <li v-for="m in list" :key="m.id" class="flex items-center gap-3 px-4 py-2.5 text-[13px]">
        <div class="min-w-0 flex-1">
          <p class="truncate font-medium text-slate-700">{{ m.name }}</p>
          <p class="truncate text-xs text-slate-400">
            <span v-if="m.code" class="font-mono">{{ m.code }}</span><span v-if="m.code && m.supplier"> · </span>{{ m.supplier || '—' }}
          </p>
        </div>
        <span class="shrink-0 text-xs text-slate-400">{{ m.unit }}</span>
        <span class="whitespace-nowrap tabular-nums text-slate-600">{{ formatRupiah(m.defaultPrice) }}</span>
        <button type="button" class="shrink-0 rounded-md bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-brand-700" @click="pick(m)">Pilih</button>
      </li>
      <li v-if="!store.loading && !list.length" class="px-4 py-6 text-center text-[13px] italic text-slate-400">
        Tidak ada material aktif yang cocok. Tambahkan lewat menu Master Data → Material.
      </li>
    </ul>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppModal from './AppModal.vue'
import { useMaterialStore } from '../stores/material'
import { formatRupiah } from '../utils/format'
import type { MaterialView } from '../types/master'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'select', material: MaterialView): void
}>()

const store = useMaterialStore()
const search = ref('')

const list = computed(() => {
  const q = search.value.trim().toLowerCase()
  return store.materials.filter(
    (m) =>
      m.isActive &&
      (!q ||
        m.name.toLowerCase().includes(q) ||
        m.code.toLowerCase().includes(q) ||
        m.supplier.toLowerCase().includes(q))
  )
})

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return
    search.value = ''
    store.search = ''
    store.includeInactive = false
    await store.load()
  }
)

function pick(m: MaterialView) {
  emit('select', m)
  emit('update:modelValue', false)
}
</script>
