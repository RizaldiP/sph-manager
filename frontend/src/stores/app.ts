import { defineStore } from 'pinia'
import { ref } from 'vue'
import { Health, CheckForUpdate } from '../../wailsjs/go/main/App'

export interface HealthInfo {
  status: string
  version: string
  platform: string
  databasePath: string
}

export interface UpdateStatusInfo {
  currentVersion: string
  latestVersion: string
  hasUpdate: boolean
  notes?: string
}

export const useAppStore = defineStore('app', () => {
  const health = ref<HealthInfo | null>(null)
  const loaded = ref(false)
  const error = ref('')

  const updateStatus = ref<UpdateStatusInfo | null>(null)
  const updateChecked = ref(false)
  const updateBusy = ref(false)
  const updateDialogOpen = ref(false)
  const updateError = ref('')

  async function load() {
    try {
      health.value = await Health()
      error.value = ''
    } catch (e) {
      error.value = String(e)
    } finally {
      loaded.value = true
    }
  }

  async function checkUpdate() {
    updateBusy.value = true
    updateError.value = ''
    try {
      updateStatus.value = (await CheckForUpdate()) as unknown as UpdateStatusInfo
      updateChecked.value = true
    } catch (e) {
      updateError.value = String(e)
    } finally {
      updateBusy.value = false
    }
  }

  function openUpdater() {
    updateDialogOpen.value = true
    void checkUpdate()
  }

  function closeUpdater() {
    updateDialogOpen.value = false
  }

  return {
    health,
    loaded,
    error,
    load,
    updateStatus,
    updateChecked,
    updateBusy,
    updateDialogOpen,
    updateError,
    checkUpdate,
    openUpdater,
    closeUpdater
  }
})
