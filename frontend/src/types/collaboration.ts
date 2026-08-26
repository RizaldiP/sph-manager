export type CollabMode = '' | 'HOST' | 'CLIENT'
export type CollabConnection = '' | 'CONNECTED' | 'RECONNECTING' | 'DISCONNECTED'

export interface CollabSnapshot {
  mode: CollabMode
  connection?: CollabConnection
  room?: RoomInfo
  doc?: SphSaveInput
  participants?: Participant[]
  activities?: CollabActivity[]
  version?: number
  error?: string
  notice?: string
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
