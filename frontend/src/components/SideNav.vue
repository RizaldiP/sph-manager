<template>
  <aside class="flex w-64 shrink-0 flex-col border-r border-slate-200 bg-white">
    <div class="flex items-center gap-3 px-5 py-4">
      <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-brand-600 text-sm font-bold text-white">
        SM
      </div>
      <div>
        <p class="text-sm font-semibold text-slate-800">SPH Manager</p>
        <p class="text-xs text-slate-500">Ganesha Energi Indonesia</p>
      </div>
    </div>

    <nav class="min-h-0 flex-1 overflow-y-auto px-3 pb-4">
      <div v-for="section in sections" :key="section.title" class="mb-4">
        <p class="px-2 pb-1.5 pt-2 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
          {{ section.title }}
        </p>
        <router-link
          v-for="item in section.items"
          :key="item.to"
          :to="item.to"
          class="mt-0.5 flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] transition-colors"
          :class="isActive(item.to) ? 'bg-brand-600 font-medium text-white' : 'text-slate-600 hover:bg-slate-100'"
        >
          <svg
            class="h-[18px] w-[18px] shrink-0"
            :class="isActive(item.to) ? 'text-white' : 'text-slate-400'"
            fill="none"
            viewBox="0 0 24 24"
            stroke-width="1.6"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" />
          </svg>
          <span>{{ item.label }}</span>
        </router-link>
      </div>
    </nav>

    <div class="border-t border-slate-200 px-5 py-3">
      <p class="text-xs text-slate-400">
        v{{ store.health?.version ?? '—' }} · Offline
      </p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const route = useRoute()
const store = useAppStore()

interface NavItem {
  label: string
  to: string
  icon: string
}

interface NavSection {
  title: string
  items: NavItem[]
}

const icons = {
  home: 'M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75',
  document:
    'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z',
  wrench:
    'M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z',
  database:
    'M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 5.625c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125',
  transfer:
    'M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5',
  shield:
    'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z',
  cog: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
  users: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z'
}

const sections: NavSection[] = [
  {
    title: 'Menu Utama',
    items: [{ label: 'Dashboard', to: '/', icon: icons.home }]
  },
  {
    title: 'SPH',
    items: [
      { label: 'Semua SPH', to: '/sph', icon: icons.document },
      { label: 'Buat SPH', to: '/sph/baru', icon: icons.document }
    ]
  },
  {
    title: 'Pekerjaan',
    items: [
      { label: 'Master Pekerjaan', to: '/pekerjaan', icon: icons.wrench },
      { label: 'Kategori', to: '/pekerjaan/kategori', icon: icons.wrench },
      { label: 'Template', to: '/pekerjaan/template', icon: icons.wrench }
    ]
  },
  {
    title: 'Master Data',
    items: [
      { label: 'Customer', to: '/data/customer', icon: icons.database },
      { label: 'Material', to: '/data/material', icon: icons.database }
    ]
  },
  {
    title: 'Lainnya',
    items: [
      { label: 'Import', to: '/import', icon: icons.transfer },
      { label: 'Work Together', to: '/work-together', icon: icons.users },
      { label: 'Backup', to: '/backup', icon: icons.shield },
      { label: 'Pengaturan', to: '/pengaturan', icon: icons.cog }
    ]
  }
]

function isActive(to: string): boolean {
  const matchLen = (p: string): number =>
    route.path === p || route.path.startsWith(p + '/') ? p.length : 0
  const myLen = matchLen(to)
  if (myLen === 0) return false
  return !sections.some((s) => s.items.some((i) => i.to !== to && matchLen(i.to) > myLen))
}
</script>
