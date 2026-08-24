// Tipe domain Import Excel (FR-IE1..IE4) — cermin struct Go di internal/importers
// dan internal/services.

export interface ColumnMapping {
  nameCol: number
  nameSpan: number
  qtyCol: number
  unitCol: number
  serviceCol: number
  materialCol: number
  // Kolom "Harga Satuan" umum — dipakai hanya bila JASA & MATERIAL kosong.
  // Selalu kirim -1 bila tidak dipetakan (0 berarti kolom A).
  unitPriceCol: number
  unitPriceAs: 'service' | 'material'
  serviceTotal: boolean
  materialTotal: boolean
  headerRows: number
}

export interface PreviewRow {
  rowIndex: number
  suggested: string // main | sub | unknown
  level: string
  marker: string
  name: string
  qty: number
  unit: string
  servicePrice: number
  materialPrice: number
  raw: string
  errors?: string[]
}

export interface SheetPreview {
  grid: string[][]
  totalRows: number
  totalCols: number
  suggestedMapping: ColumnMapping
  notes?: string[]
  mainCount: number
  subCount: number
  unknownCount: number
}

export interface ConfirmRow {
  rowIndex: number
  level: 'main' | 'sub' | 'skip'
}

export interface ImportResult {
  itemsCreated: number
  subsCreated: number
  skipped: number
}

// Klasifikasi yang dipilih pengguna per baris unknown.
export type RowDecision = '' | 'main' | 'sub' | 'skip'
