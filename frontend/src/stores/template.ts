import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListTemplates,
  GetTemplateDetail,
  CreateTemplate,
  UpdateTemplate,
  SetTemplateItems,
  DuplicateTemplate,
  SetTemplateActive,
  DeleteTemplate,
  ReorderTemplates
} from '../../wailsjs/go/main/App'
import type { TemplateView, TemplateDetail } from '../types/template'
import { stripAudit } from '../utils/payload'

// Store domain Template Pekerjaan (Phase 4).
// Mutasi melempar error apa adanya agar halaman bisa menampilkannya di tempat yang sesuai;
// pemuatan daftar menangkap error sendiri ke state banner.
export const useTemplateStore = defineStore('template', () => {
  // ===== state =====
  const templates = ref<TemplateView[]>([])
  const templatesLoading = ref(false)
  const templatesError = ref('')

  const tplSearch = ref('')
  const tplIncludeInactive = ref(false)

  // ===== aksi =====
  async function loadTemplates() {
    templatesLoading.value = true
    try {
      templates.value = await ListTemplates(tplIncludeInactive.value, tplSearch.value)
      templatesError.value = ''
    } catch (e) {
      templatesError.value = String(e)
    } finally {
      templatesLoading.value = false
    }
  }

  function createTemplate(payload: Partial<TemplateView>) {
    return CreateTemplate(stripAudit(payload) as never) as Promise<TemplateView>
  }

  function updateTemplate(id: number, payload: Partial<TemplateView>) {
    return UpdateTemplate(id, stripAudit(payload) as never) as Promise<TemplateView>
  }

  async function getTemplateDetail(id: number): Promise<TemplateDetail> {
    return (await GetTemplateDetail(id)) as unknown as TemplateDetail
  }

  async function saveItems(id: number, items: Array<{ workItemId: number; notes: string }>) {
    const detail = (await SetTemplateItems(id, items as never)) as unknown as TemplateDetail
    await loadTemplates()
    return detail
  }

  async function duplicateTemplate(id: number) {
    const view = (await DuplicateTemplate(id)) as unknown as TemplateView
    await loadTemplates()
    return view
  }

  async function setTemplateActive(id: number, active: boolean) {
    await SetTemplateActive(id, active)
    await loadTemplates()
  }

  async function deleteTemplate(id: number) {
    await DeleteTemplate(id)
    await loadTemplates()
  }

  async function reorderTemplates(ids: number[]) {
    await ReorderTemplates(ids)
    await loadTemplates()
  }

  return {
    templates,
    templatesLoading,
    templatesError,
    tplSearch,
    tplIncludeInactive,
    loadTemplates,
    createTemplate,
    updateTemplate,
    getTemplateDetail,
    saveItems,
    duplicateTemplate,
    setTemplateActive,
    deleteTemplate,
    reorderTemplates
  }
})
