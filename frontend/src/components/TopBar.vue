<template>
  <header class="flex h-14 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-6">
    <nav class="flex items-center gap-1.5 text-sm">
      <template v-for="(crumb, i) in breadcrumb" :key="i">
        <span v-if="i > 0" class="text-slate-300">/</span>
        <span :class="i === breadcrumb.length - 1 ? 'font-medium text-slate-800' : 'text-slate-400'">
          {{ crumb }}
        </span>
      </template>
    </nav>

    <div class="flex items-center gap-3">
      <div class="flex w-64 cursor-not-allowed items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-1.5 text-[13px] text-slate-400">
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.6" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
        </svg>
        <span>Cari… (Ctrl+F)</span>
      </div>
      <span
        v-if="store.health"
        class="flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
        Terhubung
      </span>
      <span
        v-else-if="store.loaded"
        class="flex items-center gap-1.5 rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-red-500"></span>
        Terputus
      </span>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const route = useRoute()
const store = useAppStore()

const breadcrumb = computed<string[]>(() => {
  const crumbs = route.meta.breadcrumb as string[] | undefined
  if (crumbs && crumbs.length > 0) {
    return crumbs
  }
  const title = route.meta.title as string | undefined
  if (!title) {
    return []
  }
  return [title]
})
</script>
