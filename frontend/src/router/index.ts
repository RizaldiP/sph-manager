import { createRouter, createWebHashHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import MasterPekerjaanPage from '../pages/MasterPekerjaanPage.vue'
import KategoriPage from '../pages/KategoriPage.vue'
import TemplatePage from '../pages/TemplatePage.vue'
import DataPartnerPage from '../pages/DataPartnerPage.vue'
import MaterialPage from '../pages/MaterialPage.vue'
import SphListPage from '../pages/SphListPage.vue'
import BuatSphPage from '../pages/BuatSphPage.vue'
import SphDetailPage from '../pages/SphDetailPage.vue'
import PengaturanPage from '../pages/PengaturanPage.vue'
import ImportPage from '../pages/ImportPage.vue'
import WorkTogetherPage from '../pages/WorkTogetherPage.vue'

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
        { path: 'sph', name: 'sph-list', component: SphListPage, meta: { title: 'Semua SPH', breadcrumb: ['SPH', 'Semua SPH'] } },
        { path: 'sph/draft', name: 'sph-draft', component: SphListPage, meta: { title: 'Draft SPH', breadcrumb: ['SPH', 'Draft'] } },
        { path: 'sph/final', name: 'sph-final', component: SphListPage, meta: { title: 'SPH Final', breadcrumb: ['SPH', 'Final'] } },
        { path: 'sph/baru', name: 'sph-baru', component: BuatSphPage, meta: { title: 'Buat SPH', breadcrumb: ['SPH', 'Buat SPH'] } },
        { path: 'sph/:id(\\d+)', name: 'sph-detail', component: SphDetailPage, meta: { title: 'Detail SPH', breadcrumb: ['SPH', 'Detail'] } },
        { path: 'sph/:id(\\d+)/edit', name: 'sph-edit', component: BuatSphPage, meta: { title: 'Edit Draft SPH', breadcrumb: ['SPH', 'Edit Draft'] } },
        { path: 'pekerjaan', name: 'pekerjaan', component: MasterPekerjaanPage, meta: { title: 'Master Pekerjaan', breadcrumb: ['Pekerjaan', 'Master Pekerjaan'] } },
        { path: 'pekerjaan/kategori', name: 'kategori', component: KategoriPage, meta: { title: 'Kategori', breadcrumb: ['Pekerjaan', 'Kategori'] } },
        { path: 'pekerjaan/template', name: 'template', component: TemplatePage, meta: { title: 'Template', breadcrumb: ['Pekerjaan', 'Template'] } },
        { path: 'data/customer', name: 'customer', component: DataPartnerPage, meta: { title: 'Customer & Kapal', breadcrumb: ['Master Data', 'Customer'] } },
        { path: 'data/material', name: 'material', component: MaterialPage, meta: { title: 'Material', breadcrumb: ['Master Data', 'Material'] } },
        { path: 'impor-ekspor', name: 'impor-ekspor', component: ImportPage, meta: { title: 'Import / Export', breadcrumb: ['Import / Export'] } },
        { path: 'backup', name: 'backup', component: PlaceholderPage, meta: { title: 'Backup', breadcrumb: ['Backup'], phase: 'Phase 11' } },
        { path: 'work-together', name: 'work-together', component: WorkTogetherPage, meta: { title: 'Work Together', breadcrumb: ['Work Together'] } },
        { path: 'pengaturan', name: 'pengaturan', component: PengaturanPage, meta: { title: 'Pengaturan', breadcrumb: ['Pengaturan'] } }
      ]
    }
  ]
})

export default router
