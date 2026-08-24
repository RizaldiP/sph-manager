import { ref, type Ref } from 'vue'

// useDragSort: drag-and-drop reorder sederhana berbasis HTML5 DnD tanpa dependensi.
// Saat item diseret melewati target, array sumber diubah seketika (optimistik).
// Saat dilepas, daftar ID hasil urutan dikirim ke `persist` (panggilan backend).
export function useDragSort<T extends { id: number }>(
  items: Ref<T[]>,
  persist: (orderedIds: number[]) => Promise<void>
) {
  const draggingId = ref<number | null>(null)

  function startDrag(id: number) {
    draggingId.value = id
  }

  function enterDrag(targetId: number) {
    const from = draggingId.value
    if (from === null || from === targetId) return
    const arr = items.value
    const fromIdx = arr.findIndex((x) => x.id === from)
    const toIdx = arr.findIndex((x) => x.id === targetId)
    if (fromIdx < 0 || toIdx < 0) return
    arr.splice(toIdx, 0, arr.splice(fromIdx, 1)[0])
  }

  async function endDrag() {
    const from = draggingId.value
    draggingId.value = null
    if (from === null) return
    await persist(items.value.map((x) => x.id))
  }

  return { draggingId, startDrag, enterDrag, endDrag }
}
