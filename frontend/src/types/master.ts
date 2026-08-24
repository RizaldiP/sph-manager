// Tipe domain Master Pekerjaan — mencerminkan JSON dari backend Go (Phase 3).

export interface CategoryView {
  id: number
  code: string
  name: string
  description: string
  sequence: number
  isActive: boolean
  workItemCount: number
  createdAt: string
  updatedAt: string
}

export interface WorkItemView {
  id: number
  categoryId: number
  code: string
  name: string
  description: string
  defaultUnit: string
  defaultQuantity: number
  defaultServicePrice: number
  defaultMaterialPrice: number
  notes: string
  sequence: number
  isActive: boolean
  subItemCount: number
  createdAt: string
  updatedAt: string
}

export interface WorkSubItem {
  id: number
  workItemId: number
  code: string
  sequence: number
  name: string
  description: string
  difficultyWeight: number
  defaultUnit: string
  defaultQuantity: number
  defaultServicePrice: number
  defaultMaterialPrice: number
  notes: string
  isActive: boolean
}

// Hasil GetWorkItemDetail / CreateSubItem / UpdateSubItem:
// model WorkItem + relasi subItems (relasi tidak dideklarasikan di wailsjs, jadi kita definisikan sendiri).
export interface WorkItemDetail {
  id: number
  categoryId: number
  code: string
  name: string
  description: string
  defaultUnit: string
  defaultQuantity: number
  defaultServicePrice: number
  defaultMaterialPrice: number
  notes: string
  sequence: number
  isActive: boolean
  subItems?: WorkSubItem[]
}

export function emptyCategory(): CategoryView & { id: number } {
  return {
    id: 0,
    code: '',
    name: '',
    description: '',
    sequence: 0,
    isActive: true,
    workItemCount: 0,
    createdAt: '',
    updatedAt: ''
  }
}

export function emptyWorkItem(categoryId = 0): WorkItemView {
  return {
    id: 0,
    categoryId,
    code: '',
    name: '',
    description: '',
    defaultUnit: 'giat',
    defaultQuantity: 1,
    defaultServicePrice: 0,
    defaultMaterialPrice: 0,
    notes: '',
    sequence: 0,
    isActive: true,
    subItemCount: 0,
    createdAt: '',
    updatedAt: ''
  }
}

export function emptySubItem(workItemId = 0): WorkSubItem {
  return {
    id: 0,
    workItemId,
    code: '',
    sequence: 0,
    name: '',
    description: '',
    difficultyWeight: 0,
    defaultUnit: 'giat',
    defaultQuantity: 1,
    defaultServicePrice: 0,
    defaultMaterialPrice: 0,
    notes: '',
    isActive: true
  }
}
