import { t } from '../i18n'

export type NaiveType = 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info'

export const verdictColor: Record<string, NaiveType> = {
  AC: 'success',
  WA: 'error',
  TLE: 'warning',
  MLE: 'warning',
  OLE: 'warning',
  RE: 'error',
  CE: 'error',
  PE: 'warning',
  SE: 'error',
  UKE: 'default',
  PENDING: 'default',
}

export const verdictKeys = Object.keys(verdictColor)

export function verdictType(v: string | undefined): NaiveType {
  if (!v) return 'default'
  return verdictColor[v] ?? 'default'
}

export function statusLabel(s: string | undefined): string {
  switch (s) {
    case 'AC': return t.verdict.statusAc
    case 'AC_FADED': return t.verdict.statusAcFaded
    case 'attempted': return t.verdict.statusAttempted
    default: return t.verdict.statusEmpty
  }
}

export function statusTagType(s: string | undefined): NaiveType {
  switch (s) {
    case 'AC':
    case 'AC_FADED': return 'success'
    case 'attempted': return 'warning'
    default: return 'default'
  }
}
