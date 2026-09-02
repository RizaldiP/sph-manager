export type CollabMode = '' | 'HOST' | 'CLIENT'
export type CollabConnection = '' | 'CONNECTED' | 'RECONNECTING' | 'DISCONNECTED'

export interface CollabSnapshot {
  mode: CollabMode
  connection?: CollabConnection
  room?: RoomInfo
  doc?: SphSaveInput
  participants?: Participant[]
  activities?: CollabActivity[]
  turn?: TurnState
  chat?: ChatMessage[]
  unread?: number
  version?: number
  error?: string
  notice?: string
  masterStatus?: MasterStatusEntry[]
}

export type ChatMessageType = 'text' | 'system' | 'master_data'

export interface ChatMessage {
  messageId: string
  roomId?: string
  senderId?: string
  senderName?: string
  messageType: ChatMessageType
  content: string
  status?: string
  refPackage?: string
  refMeta?: string
  createdAt: string
}

export interface RoomInfo {
  roomId: string
  sphDocumentId: number
  documentNumber: string
  projectName: string
  roomCode: string
  roomName: string
  accessCode?: string
  hostName: string
  hostDevice: string
  hostIPs?: string[]
  port: number
  status: string
  version: number
  participants?: Participant[]
  firewallWarning?: string
  createdAt: string
}

export interface Participant {
  id: string
  displayName: string
  deviceName: string
  role: string
  joinedAt: string
  lastSeen: string
}

export interface DiscoveredRoom {
  roomId: string
  roomName: string
  documentNumber: string
  projectName: string
  hostIP: string
  hostName: string
  port: number
  users: number
  lastSeen: string
}

export interface CollabActivity {
  actor: string
  action: string
  summary: string
}

export interface OpPayload {
  type: string
  itemId?: number
  subItemId?: number
  toIndex?: number
  header?: HeaderPatch
  item?: ItemFields
  subItem?: SubItemFields
}

export interface HeaderPatch {
  date: string
  sequence?: string
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

export interface ItemFields {
  workItemId?: number
  name: string
  description: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  pricingMode: string
  notes: string
}

export interface SubItemFields {
  name: string
  description: string
  quantity: number
  unit: string
  weight: number
  serviceUnitPrice: number
  materialUnitPrice: number
  notes: string
}

export interface CollabDefaults {
  deviceName: string
  port: number
  displayName: string
}

export interface SphHeaderInput {
  date: string
  sequence?: string
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

export interface SphSubItemInput {
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

export interface SphItemInput {
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

export interface SphSaveInput {
  header: SphHeaderInput
  items: SphItemInput[]
}

export interface TurnState {
  assignments: Record<string, string[]>  // participantID → []sectionID
  activeEdits: Record<string, string>    // sectionID → participantID
}

export const SectionLabel: Record<string, string> = {
  header: 'Header',
  items: 'Items',
  subitems: 'Sub Items'
}

export const OpType = {
  HEADER_UPDATED: 'HEADER_UPDATED',
  ITEM_ADDED: 'ITEM_ADDED',
  ITEM_UPDATED: 'ITEM_UPDATED',
  ITEM_DELETED: 'ITEM_DELETED',
  ITEM_MOVED: 'ITEM_MOVED',
  SUB_ITEM_ADDED: 'SUB_ITEM_ADDED',
  SUB_ITEM_UPDATED: 'SUB_ITEM_UPDATED',
  SUB_ITEM_DELETED: 'SUB_ITEM_DELETED',
  SUB_ITEM_MOVED: 'SUB_ITEM_MOVED'
} as const

export const ModeLabel: Record<string, string> = {
  '': 'Offline',
  HOST: 'Host',
  CLIENT: 'Client'
}

export const ConnLabel: Record<string, string> = {
  '': 'Terputus',
  CONNECTED: 'Terhubung',
  RECONNECTING: 'Menyambung ulang…',
  DISCONNECTED: 'Terputus'
}

// ===== Master Data transfer =====
export interface MasterDataPackage {
  metadata: MasterPackageMetadata
  data: MasterPackageData
  checksum: string
}
export interface MasterPackageMetadata {
  packageId: string
  senderId?: string
  senderName?: string
  roomId?: string
  createdAt: string
  schemaVersion: string
  packageVersion?: string
  sourceVersion?: number
}
export interface MasterPackageData {
  categories?: PackageCategory[]
  workItems?: PackageWorkItem[]
  workSubItems?: PackageWorkSubItem[]
  materials?: PackageMaterial[]
}
export interface PackageCategory {
  code: string
  name: string
  description?: string
  sequence: number
  isActive: boolean
}
export interface PackageWorkItem {
  code: string
  name: string
  description?: string
  defaultUnit?: string
  defaultQuantity: number
  defaultServicePrice: number
  defaultMaterialPrice: number
  notes?: string
  sequence: number
  isActive: boolean
  categoryCode?: string
}
export interface PackageWorkSubItem {
  code?: string
  sequence: number
  name: string
  description?: string
  difficultyWeight: number
  defaultUnit?: string
  defaultQuantity: number
  defaultServicePrice: number
  defaultMaterialPrice: number
  notes?: string
  isActive: boolean
  workItemCode?: string
}
export interface PackageMaterial {
  code?: string
  name: string
  description?: string
  unit?: string
  defaultPrice: number
  supplier?: string
  notes?: string
  isActive: boolean
}
export interface MasterDataList {
  categories: PackageCategory[]
  workItems: PackageWorkItem[]
  workSubItems: PackageWorkSubItem[]
  materials: PackageMaterial[]
}
export interface MasterDataSelection {
  categoryCodes: string[]
  workItemCodes: string[]
  subItemKeys: string[]
  materialCodes: string[]
  sendAll: boolean
}
export interface MasterInboxItem {
  packageId: string
  senderId: string
  senderName: string
  roomId: string
  sourceVersion: number
  title: string
  summary: string
  itemCount: number
  status: string
  receivedAt: string
  installedAt?: string
  rejectedAt?: string
}
export type MasterDiffKind = 'NEW' | 'UPDATED' | 'UNCHANGED' | 'CONFLICT'
export interface MasterDiffItem {
  kind: MasterDiffKind
  entity: string
  code: string
  name: string
  summary: string
}
export interface MasterInstallSummary {
  categoriesCreated: number
  categoriesUpdated: number
  workItemsCreated: number
  workItemsUpdated: number
  subItemsCreated: number
  subItemsUpdated: number
  materialsCreated: number
  materialsUpdated: number
  skipped: number
  conflicts: number
}
export interface MasterSentItem {
  packageId: string
  roomId: string
  sourceVersion: number
  title: string
  itemCount: number
  status: string
  recipients: string
  sentAt: string
}
export interface MasterStatusEntry {
  packageId: string
  targetId?: string
  targetName?: string
  status: string
  at: string
}
export const MasterStrategy = {
  PROMPT: 'PROMPT',
  USE_LOCAL: 'USE_LOCAL',
  USE_INCOMING: 'USE_INCOMING',
  SKIP: 'SKIP'
} as const
export const MasterDiffLabel: Record<MasterDiffKind, string> = {
  NEW: 'Baru',
  UPDATED: 'Diperbarui',
  UNCHANGED: 'Tidak berubah',
  CONFLICT: 'Konflik'
}
export const MasterStatusLabel: Record<string, string> = {
  PENDING: 'Menunggu',
  VIEWED: 'Dibuka',
  INSTALLED: 'Terpasang',
  REJECTED: 'Ditolak',
  FAILED: 'Gagal'
}
