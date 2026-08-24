import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListCategories,
  ListWorkItems,
  CreateCategory,
  UpdateCategory,
  SetCategoryActive,
  DeleteCategory,
  ReorderCategories,
  GetWorkItemDetail,
  CreateWorkItem,
  UpdateWorkItem,
  SetWorkItemActive,
  DeleteWorkItem,
  DeleteWorkItems,
  ReorderWorkItems,
  CreateSubItem,
  UpdateSubItem,
  SetSubItemActive,
  DeleteSubItem,
  DeleteSubItems,
  ReorderSubItems
} from '../../wailsjs/go/main/App'
import type { CategoryView, WorkItemView, WorkItemDetail, DeleteResult } from '../types/master'
import { stripAudit } from '../utils/payload'

// Store domain Master Pekerjaan (kategori + pekerjaan + sub-pekerjaan).
// Mutasi melempar error apa adanya agar halaman bisa menampilkannya di tempat yang sesuai;
// pemuatan daftar menangkap error sendiri ke state banner.
export const useMasterStore = defineStore('master', () => {
  // ===== state =====
  const categories = ref<CategoryView[]>([])
  const categoriesLoading = ref(false)
  const categoriesError = ref('')

  const workItems = ref<WorkItemView[]>([])
  const workItemsLoading = ref(false)
  const workItemsError = ref('')

  // filter kategori
  const catSearch = ref('')
  const catIncludeInactive = ref(false)

  // filter pekerjaan
  const wiCategoryId = ref(0) // 0 = semua kategori
  const wiSearch = ref('')
  const wiIncludeInactive = ref(false)

  // ===== kategori =====
  async function loadCategories() {
    categoriesLoading.value = true
    try {
      categories.value = await ListCategories(catIncludeInactive.value, catSearch.value)
      categoriesError.value = ''
    } catch (e) {
      categoriesError.value = String(e)
    } finally {
      categoriesLoading.value = false
    }
  }

  function createCategory(payload: Partial<CategoryView>) {
    return CreateCategory(stripAudit(payload) as never) as Promise<CategoryView>
  }

  function updateCategory(id: number, payload: Partial<CategoryView>) {
    return UpdateCategory(id, stripAudit(payload) as never) as Promise<CategoryView>
  }

  async function setCategoryActive(id: number, active: boolean) {
    await SetCategoryActive(id, active)
    await loadCategories()
  }

  async function deleteCategory(id: number) {
    await DeleteCategory(id)
    await loadCategories()
  }

  async function reorderCategories(ids: number[]) {
    await ReorderCategories(ids)
    await loadCategories()
  }

  // ===== pekerjaan =====
  async function loadWorkItems() {
    workItemsLoading.value = true
    try {
      workItems.value = await ListWorkItems(wiCategoryId.value, wiIncludeInactive.value, wiSearch.value)
      workItemsError.value = ''
    } catch (e) {
      workItemsError.value = String(e)
    } finally {
      workItemsLoading.value = false
    }
  }

  function getWorkItemDetail(id: number) {
    return GetWorkItemDetail(id) as unknown as Promise<WorkItemDetail>
  }

  function createWorkItem(payload: Partial<WorkItemView>) {
    return CreateWorkItem(stripAudit(payload) as never) as unknown as Promise<WorkItemDetail>
  }

  function updateWorkItem(id: number, payload: Partial<WorkItemView>) {
    return UpdateWorkItem(id, stripAudit(payload) as never) as unknown as Promise<WorkItemDetail>
  }

  async function setWorkItemActive(id: number, active: boolean) {
    await SetWorkItemActive(id, active)
    await loadWorkItems()
  }

  async function deleteWorkItem(id: number) {
    await DeleteWorkItem(id)
    await loadWorkItems()
  }

  // Hapus massal pekerjaan (kaskade sub-pekerjaannya) dalam satu transaksi.
  function deleteWorkItems(ids: number[]) {
    return DeleteWorkItems(ids) as unknown as Promise<DeleteResult>
  }

  async function reorderWorkItems(ids: number[]) {
    if (wiCategoryId.value === 0) return
    await ReorderWorkItems(wiCategoryId.value, ids)
    await loadWorkItems()
  }

  // ===== sub-pekerjaan =====
  function createSubItem(payload: object) {
    return CreateSubItem(stripAudit(payload) as never) as unknown as Promise<WorkItemDetail>
  }

  function updateSubItem(id: number, payload: object) {
    return UpdateSubItem(id, stripAudit(payload) as never) as unknown as Promise<WorkItemDetail>
  }

  function setSubItemActive(id: number, active: boolean) {
    return SetSubItemActive(id, active)
  }

  function deleteSubItem(id: number) {
    return DeleteSubItem(id)
  }

  // Hapus massal sub-pekerjaan dalam satu transaksi.
  function deleteSubItems(ids: number[]) {
    return DeleteSubItems(ids) as unknown as Promise<DeleteResult>
  }

  function reorderSubItems(workItemId: number, ids: number[]) {
    return ReorderSubItems(workItemId, ids)
  }

  return {
    categories,
    categoriesLoading,
    categoriesError,
    workItems,
    workItemsLoading,
    workItemsError,
    catSearch,
    catIncludeInactive,
    wiCategoryId,
    wiSearch,
    wiIncludeInactive,
    loadCategories,
    createCategory,
    updateCategory,
    setCategoryActive,
    deleteCategory,
    reorderCategories,
    loadWorkItems,
    getWorkItemDetail,
    createWorkItem,
    updateWorkItem,
    setWorkItemActive,
    deleteWorkItem,
    deleteWorkItems,
    reorderWorkItems,
    createSubItem,
    updateSubItem,
    setSubItemActive,
    deleteSubItem,
    deleteSubItems,
    reorderSubItems
  }
})
