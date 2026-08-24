export interface SettingsView {
  companyName: string
  companyCity: string
  companyAddress: string
  logoPath: string
  sphNumberFormat: string
  signerName: string
  signerPosition: string
  defaultNotes: string
}

export const DEFAULT_SPH_FORMAT = 'SPH/GEI/{ROMAN}/{YYYY}/{SEQ}'

export function emptySettings(): SettingsView {
  return {
    companyName: '',
    companyCity: '',
    companyAddress: '',
    logoPath: '',
    sphNumberFormat: DEFAULT_SPH_FORMAT,
    signerName: '',
    signerPosition: '',
    defaultNotes: ''
  }
}
