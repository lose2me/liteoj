import type { SelectOption } from 'naive-ui'
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

export const verdictLabel: Record<string, string> = {
  AC: 'Accepted',
  WA: 'Wrong Answer',
  TLE: 'Time Limit Exceeded',
  MLE: 'Memory Limit Exceeded',
  OLE: 'Output Limit Exceeded',
  RE: 'Runtime Error',
  CE: 'Compile Error',
  PE: 'Presentation Error',
  SE: 'System Error',
  UKE: 'Unknown Error',
  PENDING: 'Pending',
}

export function verdictType(v: string | undefined): NaiveType {
  if (!v) return 'default'
  return verdictColor[v] ?? 'default'
}

export function allOption(extra: SelectOption[]): SelectOption[] {
  return [{ label: t.common.all, value: '' }, ...extra]
}

// Four-state problem status driven by the backend `my_status` field:
// ""          : not attempted
// "attempted" : submissions exist but none AC
// "AC"        : has AC, and no non-AC afterwards (持续 AC)
// "AC_FADED"  : has AC, but at least one non-AC submission came later (回归错)
export type ProblemStatus = '' | 'attempted' | 'AC' | 'AC_FADED'

export function statusLabel(s: string | undefined): string {
  switch (s) {
    case 'AC': return t.verdict.statusAc
    case 'AC_FADED': return t.verdict.statusAcFaded
    case 'attempted': return t.verdict.statusAttempted
    default: return t.verdict.statusEmpty
  }
}

export function statusColor(s: string | undefined): string {
  switch (s) {
    case 'AC': return '#18a058'
    case 'AC_FADED': return '#7bc96f'
    case 'attempted': return '#f0a020'
    default: return '#c4c4c4'
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
