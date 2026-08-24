import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListCustomers,
  CreateCustomer,
  UpdateCustomer,
  SetCustomerActive,
  DeleteCustomer,
  CreateVessel,
  UpdateVessel,
  SetVesselActive,
  DeleteVessel
} from '../../wailsjs/go/main/App'
import type { CustomerView, VesselView } from '../types/partner'
import { stripAudit } from '../utils/payload'

// Store domain Master Data Customer & Kapal (FR-M5, FR-M6).
export const usePartnerStore = defineStore('partner', () => {
  const customers = ref<CustomerView[]>([])
  const loading = ref(false)
  const error = ref('')

  const search = ref('')
  const includeInactive = ref(false)

  async function load() {
    loading.value = true
    try {
      customers.value = (await ListCustomers(includeInactive.value, search.value)) as unknown as CustomerView[]
      error.value = ''
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  function createCustomer(payload: Partial<CustomerView>) {
    return CreateCustomer(stripAudit(payload) as never) as unknown as Promise<CustomerView>
  }

  function updateCustomer(id: number, payload: Partial<CustomerView>) {
    return UpdateCustomer(id, stripAudit(payload) as never) as unknown as Promise<CustomerView>
  }

  function createVessel(payload: Partial<VesselView>) {
    return CreateVessel(stripAudit(payload) as never) as unknown as Promise<CustomerView>
  }

  function updateVessel(id: number, payload: Partial<VesselView>) {
    return UpdateVessel(id, stripAudit(payload) as never) as unknown as Promise<CustomerView>
  }

  function setCustomerActive(id: number, active: boolean) {
    return SetCustomerActive(id, active)
  }

  function deleteCustomer(id: number) {
    return DeleteCustomer(id)
  }

  function setVesselActive(id: number, active: boolean) {
    return SetVesselActive(id, active)
  }

  function deleteVessel(id: number) {
    return DeleteVessel(id)
  }

  return {
    customers,
    loading,
    error,
    search,
    includeInactive,
    load,
    createCustomer,
    updateCustomer,
    setCustomerActive,
    deleteCustomer,
    createVessel,
    updateVessel,
    setVesselActive,
    deleteVessel
  }
})
