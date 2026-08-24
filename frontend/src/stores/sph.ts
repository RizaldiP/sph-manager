import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListSph,
  GetSph,
  DashboardStats as DashboardStatsBinding,
  CreateSph,
  UpdateDraftSph,
  DeleteSph,
  SetSphStatus,
  DuplicateSph,
  CreateSphRevision
} from '../../wailsjs/go/main/App'
import type {
  DashboardStats,
  SphDetail,
  SphDocumentView,
  SphSaveInput,
  SphScope
} from '../types/sph'

// Store domain SPH: daftar, detail, wizard, dan aksi lifecycle.
// Mutasi melempar error apa adanya; daftar/statistik menangkap error sendiri.
export const useSphStore = defineStore('sph', () => {
  // ===== daftar =====
  const list = ref<SphDocumentView[]>([])
  const listLoading = ref(false)
  const listError = ref('')
  const search = ref('')
  const scope = ref<SphScope>('')

  async function loadList() {
    listLoading.value = true
    try {
      list.value = (await ListSph(scope.value, search.value, 0)) as unknown as SphDocumentView[]
      listError.value = ''
    } catch (e) {
      listError.value = String(e)
    } finally {
      listLoading.value = false
    }
  }

  // ===== detail =====
  function getDetail(id: number) {
    return GetSph(id) as unknown as Promise<SphDetail>
  }

  // ===== wizard =====
  function create(payload: SphSaveInput) {
    return CreateSph(payload as never) as unknown as Promise<SphDocumentView>
  }

  function updateDraft(id: number, payload: SphSaveInput) {
    return UpdateDraftSph(id, payload as never) as unknown as Promise<SphDetail>
  }

  // ===== aksi dokumen =====
  async function remove(id: number) {
    await DeleteSph(id)
    await loadList()
  }

  async function setStatus(id: number, status: string) {
    await SetSphStatus(id, status)
    await loadList()
  }

  async function duplicate(id: number) {
    const doc = (await DuplicateSph(id)) as unknown as SphDocumentView
    await loadList()
    return doc
  }

  async function createRevision(id: number) {
    const doc = (await CreateSphRevision(id)) as unknown as SphDocumentView
    await loadList()
    return doc
  }

  // ===== dashboard =====
  const stats = ref<DashboardStats | null>(null)
  const statsLoading = ref(false)
  const statsError = ref('')

  async function loadStats() {
    statsLoading.value = true
    try {
      stats.value = (await DashboardStatsBinding()) as unknown as DashboardStats
    } catch (e) {
      statsError.value = String(e)
    } finally {
      statsLoading.value = false
    }
  }

  return {
    list,
    listLoading,
    listError,
    search,
    scope,
    loadList,
    getDetail,
    create,
    updateDraft,
    remove,
    setStatus,
    duplicate,
    createRevision,
    stats,
    statsLoading,
    statsError,
    loadStats
  }
})
