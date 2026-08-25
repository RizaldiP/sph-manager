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
  GetCollabSession
} from '../../wailsjs/go/main/App'
import type {
  CollabSnapshot,
  CollabDefaults,
  CollabMode,
  CollabConnection,
  DiscoveredRoom,
  OpPayload
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

  function applySnapshot(raw: unknown) {
    const s = raw as Record<string, unknown>
    snapshot.value = {
      mode: (s.mode as CollabMode) ?? '',
      connection: (s.connection as CollabConnection) ?? '',
      room: s.room as CollabSnapshot['room'],
      doc: s.doc as CollabSnapshot['doc'],
      participants: s.participants as CollabSnapshot['participants'],
      activities: s.activities as CollabSnapshot['activities'],
      version: s.version as number | undefined,
      error: s.error as string | undefined,
      notice: s.notice as string | undefined
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
    applySnapshot,
    loadDefaults,
    createRoom,
    closeRoom,
    joinRoom,
    leaveRoom,
    sendOp,
    startDiscovery,
    stopDiscovery,
    refreshDiscovered,
    refreshSession,
    reset
  }
})
