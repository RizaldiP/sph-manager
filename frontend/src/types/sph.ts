export type SphStatus =
  | 'DRAFT'
  | 'REVIEW'
  | 'FINAL'
  | 'SENT'
  | 'ACCEPTED'
  | 'REJECTED'
  | 'CANCELLED'

export type SphScope = '' | 'open' | 'final'
export type SphDocumentView = {
  id: number
  documentNumber: string
  revision: number
  date: string
  customerId: number
  customerName: string
  vesselId?: number
  vesselName: string
  projectName: string
  subject: string
  status: SphStatus
  itemCount: number
  grandTotal: number
  terbilang: string
  finalizedAt?: string
  createdAt: string
  updatedAt: string
}

export type DashboardStats = {
  totalSph: number
  draftCount: number
  finalCount: number
  acceptedCount: number
  monthValue: number
  recent: SphDocumentView[]
}

// ===== hasil GetSph (dokumen lengkap). Tanggal datang sebagai string RFC3339 dari Go.
export interface SphSubItemRow {
  id: number
  sequence: number
  nameSnapshot: string
  descriptionSnapshot: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  weight: number
  allocatedValue: number
  serviceTotal: number
  materialTotal: number
  total: number
  notes: string
}

export interface SphItemRow {
  id: number
  sequence: number
  workItemId?: number
  nameSnapshot: string
  descriptionSnapshot: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  serviceTotal: number
  materialTotal: number
  total: number
  pricingMode: string
  notes: string
  subItems?: SphSubItemRow[]
}

export interface SphRevisionRow {
  id: number
  revisionNumber: number
  fromDocumentId?: number
  note: string
  createdAt: string
}

export interface CustomerBrief {
  id: number
  name: string
}

export interface SphDetail {
  id: number
  documentNumber: string
  revision: number
  date: string
  customerId: number
  customer?: CustomerBrief
  vesselId?: number
  vessel?: CustomerBrief & { vesselNumber?: string }
  projectName: string
  subject: string
  reference: string
  location: string
  validUntil?: string
  picName: string
  status: string
  subtotalService: number
  subtotalMaterial: number
  grandTotal: number
  terbilang: string
  notes: string
  finalizedAt?: string
  createdAt: string
  updatedAt: string
  items?: SphItemRow[]
  revisions?: SphRevisionRow[]
}

// ===== payload simpan (tanpa field waktu agar aman lewat binding Wails) =====

export type SphSubItemInput = {
  id?: number
  name: string
  description: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  weight: number
  notes: string
}

export type SphItemInput = {
  id?: number
  workItemId?: number
  name: string
  description: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  pricingMode: string
  notes: string
  subItems: SphSubItemInput[]
}

export type SphHeaderInput = {
  date: string // "YYYY-MM-DD"
  customerId: number
  vesselId?: number
  projectName: string
  subject: string
  reference: string
  location: string
  validUntil: string
  picName: string
  notes: string
}

export type SphSaveInput = {
  header: SphHeaderInput
  items: SphItemInput[]
}

export const sphStatusLabel: Record<SphStatus, string> = {
  DRAFT: 'Draft',
  REVIEW: 'Review',
  FINAL: 'Final',
  SENT: 'Terkirim',
  ACCEPTED: 'Disetujui',
  REJECTED: 'Ditolak',
  CANCELLED: 'Dibatalkan'
}

export const sphStatusTone: Record<SphStatus, string> = {
  DRAFT: 'bg-slate-100 text-slate-700',
  REVIEW: 'bg-amber-100 text-amber-700',
  FINAL: 'bg-blue-100 text-blue-700',
  SENT: 'bg-indigo-100 text-indigo-700',
  ACCEPTED: 'bg-emerald-100 text-emerald-700',
  REJECTED: 'bg-red-100 text-red-700',
  CANCELLED: 'bg-zinc-100 text-zinc-500'
}

// Versi aman untuk nilai status bertipe string umum (mis. model generated).
export function statusLabelOf(status: string): string {
  return sphStatusLabel[status as SphStatus] ?? status
}

export function statusToneOf(status: string): string {
  return sphStatusTone[status as SphStatus] ?? ''
}
