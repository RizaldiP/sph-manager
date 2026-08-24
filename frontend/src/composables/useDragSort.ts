import { ref, type Ref } from 'vue'

// useDragSort: drag-and-drop reorder sederhana berbasis HTML5 DnD tanpa dependensi.
// Saat item diseret melewati target, array sumber diubah seketika (optimistik).
// Saat dilepas, daftar kunci hasil urutan dikirim ke `persist` (panggilan backend).
// `keyOf` opsional untuk array dengan kunci selain numeric id (mis. uid string wizard).
export function useDragSort<T, K extends string | number = number>(
  items: Ref<T[]>,
  persist: (orderedKeys: K[]) => Promise<void> | void,
  keyOf: (item: T) => K = ((x) => (x as { id: number }).id) as (item: T) => K
) {
  const draggingId = ref<K | null>(null)

  function startDrag(key: K) {
    draggingId.value = key
  }

  function enterDrag(targetKey: K) {
    const from = draggingId.value
    if (from === null || from === targetKey) return
    const arr = items.value
    const fromIdx = arr.findIndex((x) => keyOf(x) === from)
    const toIdx = arr.findIndex((x) => keyOf(x) === targetKey)
    if (fromIdx < 0 || toIdx < 0) return
    arr.splice(toIdx, 0, arr.splice(fromIdx, 1)[0])
  }

  function endDrag() {
    const from = draggingId.value
    draggingId.value = null
    if (from === null) return
    void persist(items.value.map(keyOf))
  }

  return { draggingId, startDrag, enterDrag, endDrag }
}
