// Tipe domain Backup (Phase 11 / FR-B1..B3) — cermin struct Go di internal/backup
// dan internal/services.

export interface BackupInfo {
  name: string
  path: string
  size: number
  modified: string
}

export interface RestoreResult {
  restarting: boolean
  backup: string
  message: string
}