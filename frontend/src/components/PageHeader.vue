<template>
  <div class="mb-5 flex items-end justify-between">
    <div>
      <h1 class="text-lg font-semibold text-slate-800">{{ title }}</h1>
      <p v-if="subtitle" class="mt-0.5 text-[13px] text-slate-500">{{ subtitle }}</p>
    </div>
    <div class="flex items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>
