import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  GetCollabDefaults,
  CreateCollabRoom,
  CloseCollabRoom,
  StartDiscoveryListener,
  StopDiscoveryListener,
  ListDiscoveredRooms,
  JoinCollabRoom,
  LeaveCollabRoom,
  SendCollabOp,
  GetCollabSession,
  AssignTurns,
  RequestEdit,
  ReleaseEdit,
  SyncPush,
  SendChatMessage,
  ClearChatUnread,
  BuildMasterDataPackage,
  ListMasterDataForSelection,
  BuildMasterDataPackageFiltered,
  SendMasterData,
  ListMasterInbox,
  PreviewMasterData,
  InstallMasterData,
  RejectMasterData,
  MarkMasterInboxViewed
} from '../../wailsjs/go/main/App'
import type {
  CollabSnapshot,
  CollabDefaults,
  CollabMode,
  CollabConnection,
  DiscoveredRoom,
  OpPayload,
  TurnState,
  ChatMessageType,
  SphSaveInput,
  MasterDataPackage,
  MasterDataList,
  MasterDataSelection,
  MasterInboxItem,
  MasterDiffItem,
  MasterInstallSummary
} from '../types/collaboration'

export const useCollaborationStore = defineStore('collaboration', () => {
  const snapshot = ref<CollabSnapshot>({ mode: '' })
  const defaults = ref<CollabDefaults | null>(null)
  const discovered = ref<DiscoveredRoom[]>([])
  const loading = ref(false)
  const error = ref('')

  const isLive = computed(() => snapshot.value.mode !== '' && snapshot.value.mode !== undefined)
  const isHost = computed(() => snapshot.value.mode === 'HOST')
  const isClient = computed(() => snapshot.value.mode === 'CLIENT')
  const modeLabel = computed(() => {
    const m = snapshot.value.mode
    if (m === 'HOST') return 'Host'
    if (m === 'CLIENT') return 'Client'
    return ''
  })
  const roomName = computed(() => snapshot.value.room?.roomName ?? '')
  const participantCount = computed(() => snapshot.value.participants?.length ?? 0)
  const documentNumber = computed(() => snapshot.value.room?.documentNumber ?? '')
  const sphDocumentId = computed(() => snapshot.value.room?.sphDocumentId ?? 0)
  const turn = computed(() => snapshot.value.turn)
  const myAssignments = computed(() => {
    // Host has all sections assigned
    if (isHost.value) return ['header', 'items', 'subitems']
    const parts = snapshot.value.participants ?? []
    const myPart = parts.find(p => p.role !== 'HOST')
    if (!myPart) return []
    return turn.value?.assignments?.[myPart.id] ?? []
  })
  const myActiveEdits = computed(() => {
    if (isHost.value) {
      // Host can see all active edits
      return Object.keys(turn.value?.activeEdits ?? {})
    }
    const parts = snapshot.value.participants ?? []
    const myPart = parts.find(p => p.role !== 'HOST')
    if (!myPart) return []
    const edits = turn.value?.activeEdits ?? {}
    return Object.entries(edits).filter(([, editorId]) => editorId === myPart.id).map(([section]) => section)
  })

  function applySnapshot(raw: unknown) {
    const s = raw as Record<string, unknown>
    snapshot.value = {
      mode: (s.mode as CollabMode) ?? '',
      connection: (s.connection as CollabConnection) ?? '',
      room: s.room as CollabSnapshot['room'],
      doc: s.doc as CollabSnapshot['doc'],
      participants: s.participants as CollabSnapshot['participants'],
      activities: s.activities as CollabSnapshot['activities'],
      turn: s.turn as CollabSnapshot['turn'],
      chat: s.chat as CollabSnapshot['chat'],
      unread: s.unread as number | undefined,
      version: s.version as number | undefined,
      error: s.error as string | undefined,
      notice: s.notice as string | undefined,
      masterStatus: (s.masterStatus as CollabSnapshot['masterStatus']) ?? []
    }
  }

  async function loadDefaults() {
    try {
      defaults.value = (await GetCollabDefaults()) as unknown as CollabDefaults
    } catch (e) {
      error.value = String(e)
    }
  }

  async function createRoom(sphDocumentId: number, roomName: string, displayName: string) {
    loading.value = true
    error.value = ''
    try {
      await CreateCollabRoom(sphDocumentId, roomName, displayName)
      await refreshSession()
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      loading.value = false
    }
  }

  async function closeRoom() {
    loading.value = true
    error.value = ''
    try {
      await CloseCollabRoom()
      snapshot.value = { mode: '' }
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function joinRoom(hostIP: string, port: number, accessCode: string, roomCode: string, displayName: string) {
    loading.value = true
    error.value = ''
    try {
      await JoinCollabRoom(hostIP, port, accessCode, roomCode, displayName)
      await refreshSession()
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      loading.value = false
    }
  }

  async function leaveRoom() {
    loading.value = true
    error.value = ''
    try {
      await LeaveCollabRoom()
      snapshot.value = { mode: '' }
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function sendOp(op: OpPayload) {
    try {
      await SendCollabOp(op as never)
    } catch (e) {
      error.value = String(e)
      throw e
    }
  }

  async function startDiscovery() {
    try {
      await StartDiscoveryListener()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function stopDiscovery() {
    try {
      await StopDiscoveryListener()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function refreshDiscovered() {
    try {
      discovered.value = (await ListDiscoveredRooms()) as unknown as DiscoveredRoom[]
    } catch (e) {
      error.value = String(e)
    }
  }

  async function refreshSession() {
    try {
      const raw = (await GetCollabSession()) as unknown as Record<string, unknown>
      applySnapshot(raw)
    } catch (e) {
      error.value = String(e)
    }
  }

  function reset() {
    snapshot.value = { mode: '' }
    error.value = ''
  }

  async function assignTurns(assignments: Record<string, string[]>) {
    try {
      await AssignTurns(assignments)
      await refreshSession()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function requestEdit(sectionID: string) {
    try {
      await RequestEdit(sectionID)
      await refreshSession()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function releaseEdit(sectionID: string) {
    try {
      await ReleaseEdit(sectionID)
      await refreshSession()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function syncPush(input: SphSaveInput) {
    try {
      await SyncPush(input as never)
    } catch (e) {
      error.value = String(e)
    }
  }

  const messages = computed(() => snapshot.value.chat ?? [])
  const unreadCount = computed(() => snapshot.value.unread ?? 0)
  const masterStatuses = computed(() => snapshot.value.masterStatus ?? [])

  async function sendChat(content: string, messageType: ChatMessageType = 'text') {
    try {
      await SendChatMessage(messageType, content, '', '')
      await refreshSession()
    } catch (e) {
      error.value = String(e)
      throw e
    }
  }

  async function clearChatUnread() {
    await ClearChatUnread()
    await refreshSession()
  }

  const masterInbox = ref<MasterInboxItem[]>([])
  const masterPreview = ref<MasterDiffItem[]>([])
  const working = ref(false)

  async function buildMasterDataPackage(): Promise<MasterDataPackage> {
    return (await BuildMasterDataPackage()) as unknown as MasterDataPackage
  }

  async function listMasterDataForSelection(): Promise<MasterDataList> {
    return (await ListMasterDataForSelection()) as unknown as MasterDataList
  }

  async function buildMasterDataPackageFiltered(sel: MasterDataSelection): Promise<MasterDataPackage> {
    return (await BuildMasterDataPackageFiltered(sel as never)) as unknown as MasterDataPackage
  }

  async function sendMasterData(pkg: MasterDataPackage, targets: string[]) {
    working.value = true
    error.value = ''
    try {
      await SendMasterData(pkg as never, targets)
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      working.value = false
    }
  }

  async function refreshMasterInbox() {
    try {
      masterInbox.value = (await ListMasterInbox()) as unknown as MasterInboxItem[]
    } catch (e) {
      error.value = String(e)
    }
  }

  async function previewMasterData(packageId: string) {
    working.value = true
    error.value = ''
    try {
      masterPreview.value = ((await PreviewMasterData(packageId)) ?? []) as unknown as MasterDiffItem[]
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      working.value = false
    }
  }

  async function installMasterData(packageId: string, strategy: string, decisions: Record<string, string> = {}): Promise<MasterInstallSummary> {
    working.value = true
    error.value = ''
    try {
      const sum = (await InstallMasterData(packageId, strategy, decisions)) as unknown as MasterInstallSummary
      await refreshMasterInbox()
      return sum
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      working.value = false
    }
  }

  async function rejectMasterData(packageId: string) {
    working.value = true
    error.value = ''
    try {
      await RejectMasterData(packageId)
      await refreshMasterInbox()
    } catch (e) {
      error.value = String(e)
      throw e
    } finally {
      working.value = false
    }
  }

  async function markMasterInboxViewed(packageId: string) {
    try {
      await MarkMasterInboxViewed(packageId)
    } catch {
      // non-fatal
    }
  }

  return {
    snapshot,
    defaults,
    discovered,
    loading,
    error,
    isLive,
    isHost,
    isClient,
    modeLabel,
    roomName,
    participantCount,
    documentNumber,
    sphDocumentId,
    turn,
    myAssignments,
    myActiveEdits,
    messages,
    unreadCount,
    masterStatuses,
    applySnapshot,
    loadDefaults,
    createRoom,
    closeRoom,
    joinRoom,
    leaveRoom,
    sendOp,
    assignTurns,
    requestEdit,
    releaseEdit,
    syncPush,
    sendChat,
    clearChatUnread,
    masterInbox,
    masterPreview,
    working,
    buildMasterDataPackage,
    listMasterDataForSelection,
    buildMasterDataPackageFiltered,
    sendMasterData,
    refreshMasterInbox,
    previewMasterData,
    installMasterData,
    rejectMasterData,
    markMasterInboxViewed,
    startDiscovery,
    stopDiscovery,
    refreshDiscovered,
    refreshSession,
    reset
  }
})
