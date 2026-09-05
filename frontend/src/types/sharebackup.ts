// Tipe domain Backup yang Dapat Dibagikan (".sphbak") — cermin struct Go di
// internal/sharebackup dan binding app_sharebackup.go.

export interface ShareBackupCreateResult {
  path: string
  items: number
}

export interface SectionCounts {
  sph: number
  workItems: number
  categories: number
  templates: number
  customers: number
  materials: number
}

export interface ShareBackupPreview {
  path: string
  deviceName: string
  createdAt: string
  counts: SectionCounts
}

export interface SectionInstallResult {
  added: number
  skipped: number
  codeGenerated: number
}

export interface InstallSummary {
  categories: SectionInstallResult
  workItems: SectionInstallResult
  subItems: SectionInstallResult
  templates: SectionInstallResult
  customers: SectionInstallResult
  vessels: SectionInstallResult
  materials: SectionInstallResult
  sph: SectionInstallResult
  templateItemsAdded: number
  templateItemsMissed: number
  sphItemsUnlinked: number
}

export type ShareSectionKey = 'sph' | 'workItems' | 'categories' | 'templates' | 'customers' | 'materials'

export interface ShareSectionOption {
  key: ShareSectionKey
  label: string
  description: string
}