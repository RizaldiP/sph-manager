import { onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useCollaborationStore } from '../stores/collaboration'
import { useSphStore } from '../stores/sph'
import { OpType } from '../types/collaboration'
import type { OpPayload, HeaderPatch, ItemFields, SubItemFields, SphSaveInput } from '../types/collaboration'

export function useCollabSync() {
  const collabStore = useCollaborationStore()
  const sphStore = useSphStore()

  function onSync(raw: unknown) {
    collabStore.applySnapshot(raw)

    const snap = collabStore.snapshot
    if (snap.doc && collabStore.isLive) {
      const doc = snap.doc as unknown as SphSaveInput
      const docId = collabStore.sphDocumentId
      if (docId && doc) {
        sphStore.list = sphStore.list
      }
    }
  }

  function sendHeaderOp(patch: HeaderPatch) {
    const op: OpPayload = { type: OpType.HEADER_UPDATED, header: patch }
    return collabStore.sendOp(op)
  }

  function sendItemOp(type: string, itemId: number, itemFields: ItemFields, toIndex?: number) {
    const op: OpPayload = { type, itemId, item: itemFields }
    if (toIndex !== undefined) op.toIndex = toIndex
    return collabStore.sendOp(op)
  }

  function sendItemAdded(itemFields: ItemFields, toIndex?: number) {
    return sendItemOp(OpType.ITEM_ADDED, 0, itemFields, toIndex)
  }

  function sendItemUpdated(itemId: number, itemFields: ItemFields) {
    return sendItemOp(OpType.ITEM_UPDATED, itemId, itemFields)
  }

  function sendItemDeleted(itemId: number) {
    return collabStore.sendOp({ type: OpType.ITEM_DELETED, itemId })
  }

  function sendItemMoved(itemId: number, toIndex: number) {
    return collabStore.sendOp({ type: OpType.ITEM_MOVED, itemId, toIndex })
  }

  function sendSubItemOp(type: string, itemId: number, subItemId: number, subItemFields: SubItemFields, toIndex?: number) {
    const op: OpPayload = { type, itemId, subItemId, subItem: subItemFields }
    if (toIndex !== undefined) op.toIndex = toIndex
    return collabStore.sendOp(op)
  }

  function sendSubItemAdded(itemId: number, subItemFields: SubItemFields, toIndex?: number) {
    return sendSubItemOp(OpType.SUB_ITEM_ADDED, itemId, 0, subItemFields, toIndex)
  }

  function sendSubItemUpdated(itemId: number, subItemId: number, subItemFields: SubItemFields) {
    return sendSubItemOp(OpType.SUB_ITEM_UPDATED, itemId, subItemId, subItemFields)
  }

  function sendSubItemDeleted(itemId: number, subItemId: number) {
    return collabStore.sendOp({ type: OpType.SUB_ITEM_DELETED, itemId, subItemId })
  }

  function sendSubItemMoved(itemId: number, subItemId: number, toIndex: number) {
    return collabStore.sendOp({ type: OpType.SUB_ITEM_MOVED, itemId, subItemId, toIndex })
  }

  onMounted(() => {
    EventsOn('collab:sync', onSync)
  })

  onUnmounted(() => {
    EventsOff('collab:sync')
  })

  return {
    sendHeaderOp,
    sendItemAdded,
    sendItemUpdated,
    sendItemDeleted,
    sendItemMoved,
    sendSubItemAdded,
    sendSubItemUpdated,
    sendSubItemDeleted,
    sendSubItemMoved
  }
}
