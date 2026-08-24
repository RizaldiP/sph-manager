export type VesselView = {
  id: number
  customerId: number
  code: string
  name: string
  vesselNumber: string
  vesselType: string
  notes: string
  isActive: boolean
}

export type CustomerView = {
  id: number
  code: string
  name: string
  address: string
  phone: string
  email: string
  picName: string
  picPosition: string
  notes: string
  isActive: boolean
  vessels: VesselView[]
  createdAt: string
  updatedAt: string
}

export function emptyCustomer(): CustomerView {
  return {
    id: 0,
    code: '',
    name: '',
    address: '',
    phone: '',
    email: '',
    picName: '',
    picPosition: '',
    notes: '',
    isActive: true,
    vessels: [],
    createdAt: '',
    updatedAt: ''
  }
}

export function emptyVessel(customerId = 0): VesselView {
  return {
    id: 0,
    customerId,
    code: '',
    name: '',
    vesselNumber: '',
    vesselType: '',
    notes: '',
    isActive: true
  }
}
