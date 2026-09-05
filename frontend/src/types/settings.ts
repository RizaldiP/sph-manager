export interface SettingsInput {
  companyName: string
  companyCity: string
  companyAddress: string
  sphNumberFormat: string
  signerName: string
  signerPosition: string
  defaultNotes: string
  collabPort: number
  collabDisplayName: string
}

export interface SettingsView {
  companyName: string
  companyCity: string
  companyAddress: string
  logoPath: string
  stampPath: string
  signaturePath: string
  stampPosX: number
  stampPosY: number
  stampSize: number
  signaturePosX: number
  signaturePosY: number
  signatureSize: number
  sphNumberFormat: string
  signerName: string
  signerPosition: string
  defaultNotes: string
  collabPort: number
  collabDisplayName: string
}

export const DEFAULT_SPH_FORMAT = 'SPH/GEI/{ROMAN}/{YYYY}/{SEQ}'

export function emptySettings(): SettingsView {
  return {
    companyName: '',
    companyCity: '',
    companyAddress: '',
    logoPath: '',
    stampPath: '',
    signaturePath: '',
    stampPosX: 0.5,
    stampPosY: 0.4,
    stampSize: 0.45,
    signaturePosX: 0.5,
    signaturePosY: 0.18,
    signatureSize: 0.3,
    sphNumberFormat: DEFAULT_SPH_FORMAT,
    signerName: '',
    signerPosition: '',
    defaultNotes: '',
    collabPort: 48765,
    collabDisplayName: ''
  }
}
