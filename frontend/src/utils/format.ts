// Format angka & teks untuk UI (offline — memakai Intl bawaan).

export function formatRupiah(value: number | undefined | null): string {
  const n = Number(value ?? 0)
  return 'Rp' + new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(n)
}

export function formatQty(value: number | undefined | null): string {
  const n = Number(value ?? 0)
  return new Intl.NumberFormat('id-ID', { maximumFractionDigits: 3 }).format(n)
}

// Ambil pesan ramah dari hasil reject binding Wails (string atau Error).
export function errorMessage(e: unknown): string {
  if (typeof e === 'string') return e
  if (e instanceof Error) return e.message
  return String(e)
}
