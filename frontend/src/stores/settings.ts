import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetSettings, UpdateSettings } from '../../wailsjs/go/main/App'
import type { SettingsView } from '../types/settings'
import { emptySettings } from '../types/settings'

// Store domain Pengaturan aplikasi (FR-U4).
export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<SettingsView>(emptySettings())
  const loading = ref(false)
  const error = ref('')

  async function load() {
    loading.value = true
    try {
      settings.value = (await GetSettings()) as unknown as SettingsView
      error.value = ''
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  function save(payload: SettingsView) {
    return UpdateSettings(payload) as unknown as Promise<SettingsView>
  }

  return {
    settings,
    loading,
    error,
    load,
    save
  }
})
