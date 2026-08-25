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
  HEADER_UPDATED: 'header_updated',
  ITEM_ADDED: 'item_added',
  ITEM_UPDATED: 'item_updated',
  ITEM_DELETED: 'item_deleted',
  ITEM_MOVED: 'item_moved',
  SUB_ITEM_ADDED: 'sub_item_added',
  SUB_ITEM_UPDATED: 'sub_item_updated',
  SUB_ITEM_DELETED: 'sub_item_deleted',
  SUB_ITEM_MOVED: 'sub_item_moved'
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
