import { http } from './http'

// Module-level cache: fetched once per page load. /api/meta is cheap but
// there's no point re-hitting it for every Submissions page mount.
let cached: string[] | null = null
let inflight: Promise<string[]> | null = null

export async function getLanguages(): Promise<string[]> {
  if (cached) return cached
  if (inflight) return inflight
  inflight = http
    .get('/meta')
    .then((res) => {
      cached = (res.data?.languages as string[]) || []
      return cached
    })
    .finally(() => {
      inflight = null
    })
  return inflight
}
