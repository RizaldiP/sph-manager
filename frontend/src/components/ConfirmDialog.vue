<template>
  <AppModal :model-value="modelValue" :title="title" size="md" @update:model-value="$emit('update:modelValue', $event)">
    <div class="flex items-start gap-3.5">
      <span
        class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
        :class="danger ? 'bg-red-50 text-red-600' : 'bg-brand-50 text-brand-600'"
      >
        <svg v-if="danger" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
          />
        </svg>
        <svg v-else class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z"
          />
        </svg>
      </span>
      <div>
        <p class="text-[13px] leading-relaxed text-slate-600">{{ message }}</p>
        <p v-if="detail" class="mt-1 text-[13px] text-slate-400">{{ detail }}</p>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50"
          :disabled="busy"
          @click="$emit('update:modelValue', false)"
        >
          Batal
        </button>
        <button
          type="button"
          class="rounded-lg px-3.5 py-2 text-[13px] font-medium text-white transition-colors disabled:opacity-60"
          :class="danger ? 'bg-red-600 hover:bg-red-700' : 'bg-brand-600 hover:bg-brand-700'"
          :disabled="busy"
          @click="$emit('confirm')"
        >
          {{ busy ? 'Memproses…' : confirmLabel }}
        </button>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import AppModal from './AppModal.vue'

withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    message: string
    detail?: string
    confirmLabel?: string
    danger?: boolean
    busy?: boolean
  }>(),
  {
    confirmLabel: 'Ya, Lanjutkan',
    danger: true,
    busy: false,
    detail: ''
  }
)

defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
}>()
</script>
