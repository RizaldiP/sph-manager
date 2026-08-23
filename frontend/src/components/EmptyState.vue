<template>
  <div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-slate-300 bg-white px-6 py-14">
    <svg class="h-10 w-10 text-slate-300" fill="none" viewBox="0 0 24 24" stroke-width="1.4" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" :d="icon" />
    </svg>
    <p class="mt-3 text-sm font-medium text-slate-600">{{ title }}</p>
    <p v-if="description" class="mt-1 max-w-sm text-center text-[13px] text-slate-400">{{ description }}</p>
    <div v-if="$slots.default" class="mt-4">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    description?: string
    icon?: string
  }>(),
  {
    icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z'
  }
)
</script>
