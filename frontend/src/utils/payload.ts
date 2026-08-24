// Buang field audit (timestamp) dari payload sebelum dikirim ke Go.
// Binding Wails mem-parse argumen ke struct Go yang punya time.Time;
// string kosong/format tampilan akan gagal parse — timestamp biarlah server (GORM) yang mengisi.
export function stripAudit<T>(o: T): T {
  const { createdAt, updatedAt, deletedAt, ...rest } = o as Record<string, unknown>
  return rest as T
}
