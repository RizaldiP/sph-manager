import { createRouter, createWebHashHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import MasterPekerjaanPage from '../pages/MasterPekerjaanPage.vue'
import KategoriPage from '../pages/KategoriPage.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      component: MainLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: DashboardPage,
          meta: { title: 'Dashboard', breadcrumb: ['Dashboard'] }
        },
        { path: 'sph', name: 'sph-list', component: PlaceholderPage, meta: { title: 'Semua SPH', breadcrumb: ['SPH', 'Semua SPH'], phase: 'Phase 5' } },
        { path: 'sph/draft', name: 'sph-draft', component: PlaceholderPage, meta: { title: 'Draft SPH', breadcrumb: ['SPH', 'Draft'], phase: 'Phase 5' } },
        { path: 'sph/final', name: 'sph-final', component: PlaceholderPage, meta: { title: 'SPH Final', breadcrumb: ['SPH', 'Final'], phase: 'Phase 5' } },
        { path: 'sph/baru', name: 'sph-baru', component: PlaceholderPage, meta: { title: 'Buat SPH', breadcrumb: ['SPH', 'Buat SPH'], phase: 'Phase 5' } },
        { path: 'pekerjaan', name: 'pekerjaan', component: MasterPekerjaanPage, meta: { title: 'Master Pekerjaan', breadcrumb: ['Pekerjaan', 'Master Pekerjaan'] } },
        { path: 'pekerjaan/kategori', name: 'kategori', component: KategoriPage, meta: { title: 'Kategori', breadcrumb: ['Pekerjaan', 'Kategori'] } },
        { path: 'pekerjaan/template', name: 'template', component: PlaceholderPage, meta: { title: 'Template', breadcrumb: ['Pekerjaan', 'Template'], phase: 'Phase 4' } },
        { path: 'data/customer', name: 'customer', component: PlaceholderPage, meta: { title: 'Customer', breadcrumb: ['Master Data', 'Customer'], phase: 'Phase 3' } },
        { path: 'data/kapal', name: 'kapal', component: PlaceholderPage, meta: { title: 'Kapal', breadcrumb: ['Master Data', 'Kapal'], phase: 'Phase 3' } },
        { path: 'data/material', name: 'material', component: PlaceholderPage, meta: { title: 'Material', breadcrumb: ['Master Data', 'Material'], phase: 'Phase 3' } },
        { path: 'impor-ekspor', name: 'impor-ekspor', component: PlaceholderPage, meta: { title: 'Import / Export', breadcrumb: ['Import / Export'], phase: 'Phase 8 & 9' } },
        { path: 'backup', name: 'backup', component: PlaceholderPage, meta: { title: 'Backup', breadcrumb: ['Backup'], phase: 'Phase 10' } },
        { path: 'pengaturan', name: 'pengaturan', component: PlaceholderPage, meta: { title: 'Pengaturan', breadcrumb: ['Pengaturan'], phase: 'Phase 7' } }
      ]
    }
  ]
})

export default router
