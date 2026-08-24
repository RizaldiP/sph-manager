import type { WorkItemView } from './master'

export interface TemplateView {
  id: number
  code: string
  name: string
  description: string
  notes: string
  sequence: number
  isActive: boolean
  itemCount: number
  createdAt: string
  updatedAt: string
}

// Satu baris isi template pada editor maupun detail dari backend.
export interface TemplateItemRow {
  id?: number
  workItemId: number
  notes: string
  // Relasi tidak dideklarasikan di binding generated — dikirim runtime oleh Go.
  workItem?: Partial<WorkItemView>
}

export interface TemplateDetail {
  id: number
  code: string
  name: string
  description: string
  notes: string
  isActive: boolean
  createdAt?: string
  updatedAt?: string
  items?: TemplateItemRow[]
}
