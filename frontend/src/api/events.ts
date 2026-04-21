// Real-time event stream shared by all subscribing components. One
// EventSource connection is opened on first subscription and closed when the
// last subscriber unregisters. Reconnection is handled by the browser.
//
// The backend broadcasts a small set of typed events (see
// backend/internal/handlers/events.go). Consumers register a handler via
// `onEvent` and typically respond by re-fetching their local slice of state.

export type LiteEvent = {
  type:
    | 'submission:new'
    | 'submission:done'
    | 'ai:tasks:changed'
    | 'ai:task:done'
    // admin 改题目（含测试用例）——订阅者：/problems 列表、/problems/:id
    | 'problem:changed'
    // admin 改题单本体（规则/可见性/题目列表）——订阅者：题单页、排名、/problems 列表的 chip
    | 'problemset:changed'
    // 踢人 / unban / 学生 Join——订阅者：排名、题单提交 tab、/problems 列表 chip
    | 'problemset:members:changed'
    // admin 在后台编辑首页 markdown——订阅者：/ 首页
    | 'home:changed'
  data: any
}

type Handler = (ev: LiteEvent) => void

const handlers = new Set<Handler>()
let es: EventSource | null = null
let refCount = 0

const TYPES: LiteEvent['type'][] = [
  'submission:new',
  'submission:done',
  'ai:tasks:changed',
  'ai:task:done',
  'problem:changed',
  'problemset:changed',
  'problemset:members:changed',
  'home:changed',
]

function ensureConnected() {
  if (es) return
  es = new EventSource('/api/events/stream')
  for (const type of TYPES) {
    es.addEventListener(type, (e: MessageEvent) => {
      let data: any = null
      try { data = JSON.parse(e.data) } catch { /* keep null */ }
      const ev: LiteEvent = { type, data }
      for (const h of handlers) h(ev)
    })
  }
}

function maybeClose() {
  if (refCount === 0 && es) {
    es.close()
    es = null
  }
}

// Subscribe to server events. Returns an unsubscribe function; call it in
// onUnmounted. The underlying connection is reference-counted.
export function onEvent(h: Handler): () => void {
  handlers.add(h)
  refCount++
  ensureConnected()
  return () => {
    if (handlers.delete(h)) {
      refCount--
      maybeClose()
    }
  }
}
