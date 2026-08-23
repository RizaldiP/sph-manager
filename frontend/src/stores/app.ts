import { defineStore } from 'pinia'
import { ref } from 'vue'
import { Health } from '../../wailsjs/go/main/App'

export interface HealthInfo {
  status: string
  version: string
  platform: string
  databasePath: string
}

export const useAppStore = defineStore('app', () => {
  const health = ref<HealthInfo | null>(null)
  const loaded = ref(false)
  const error = ref('')

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

  return { health, loaded, error, load }
})
