<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      @mousedown.self="close"
    >
      <div
        class="flex max-h-[90vh] w-full flex-col rounded-xl border border-slate-200 bg-white shadow-xl"
        :class="size === 'lg' ? 'max-w-2xl' : 'max-w-lg'"
        role="dialog"
        aria-modal="true"
      >
        <div class="flex items-center justify-between border-b border-slate-100 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-800">{{ title }}</h2>
          <button
            v-if="showClose"
            class="rounded-md p-1 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            aria-label="Tutup"
            @click="close"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
          <slot />
        </div>

        <div v-if="$slots.footer" class="border-t border-slate-100 px-5 py-3">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    size?: 'md' | 'lg'
    dismissible?: boolean
    showClose?: boolean
  }>(),
  {
    size: 'md',
    dismissible: true,
    showClose: true
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function close() {
  if (props.dismissible) emit('update:modelValue', false)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue && props.dismissible) {
    emit('update:modelValue', false)
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) window.addEventListener('keydown', onKeydown)
    else window.removeEventListener('keydown', onKeydown)
  }
)
</script>
