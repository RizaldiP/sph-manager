import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListMaterials,
  CreateMaterial,
  UpdateMaterial,
  SetMaterialActive,
  DeleteMaterial
} from '../../wailsjs/go/main/App'
import type { MaterialView } from '../types/master'
import { stripAudit } from '../utils/payload'

// Store domain Master Data Material (FR-M7).
export const useMaterialStore = defineStore('material', () => {
  const materials = ref<MaterialView[]>([])
  const loading = ref(false)
  const error = ref('')

  const search = ref('')
  const includeInactive = ref(false)

  async function load() {
    loading.value = true
    try {
      materials.value = (await ListMaterials(includeInactive.value, search.value)) as unknown as MaterialView[]
      error.value = ''
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  function create(payload: Partial<MaterialView>) {
    return CreateMaterial(stripAudit(payload) as never) as unknown as Promise<MaterialView>
  }

  function update(id: number, payload: Partial<MaterialView>) {
    return UpdateMaterial(id, stripAudit(payload) as never) as unknown as Promise<MaterialView>
  }

  function setActive(id: number, active: boolean) {
    return SetMaterialActive(id, active)
  }

  function remove(id: number) {
    return DeleteMaterial(id)
  }

  return {
    materials,
    loading,
    error,
    search,
    includeInactive,
    load,
    create,
    update,
    setActive,
    remove
  }
})
