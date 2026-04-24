import { defineStore } from 'pinia'

const LEGACY_KEY = 'liteoj.editor_zoom_step'
const CODE_KEY = 'liteoj.code_editor_zoom_step'
const MARKDOWN_KEY = 'liteoj.markdown_editor_zoom_step'
const MIN_STEP = -3
const MAX_STEP = 10

export type EditorZoomAction = 'in' | 'out' | 'reset'

function clampStep(step: number) {
  return Math.min(MAX_STEP, Math.max(MIN_STEP, step))
}

function readStoredStep(key: string, fallbackKey?: string) {
  const current = Number(localStorage.getItem(key))
  if (Number.isFinite(current)) return clampStep(current)
  if (fallbackKey) {
    const legacy = Number(localStorage.getItem(fallbackKey))
    if (Number.isFinite(legacy)) return clampStep(legacy)
  }
  return 0
}

export function matchEditorZoomShortcut(
  event: Pick<KeyboardEvent, 'ctrlKey' | 'metaKey' | 'altKey' | 'key' | 'code'>,
): EditorZoomAction | null {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null

  if (event.key === '0' || event.code === 'Digit0' || event.code === 'Numpad0') {
    return 'reset'
  }
  if (event.key === '+' || event.key === '=' || event.code === 'NumpadAdd') {
    return 'in'
  }
  if (event.key === '-' || event.key === '_' || event.code === 'Minus' || event.code === 'NumpadSubtract') {
    return 'out'
  }
  return null
}

export function handleEditorZoomShortcut(
  event: KeyboardEvent,
  apply: (action: EditorZoomAction) => void,
) {
  const action = matchEditorZoomShortcut(event)
  if (!action) return false
  event.preventDefault()
  event.stopPropagation()
  apply(action)
  return true
}

export function matchEditorZoomWheel(
  event: Pick<WheelEvent, 'ctrlKey' | 'metaKey' | 'altKey' | 'deltaY'>,
): EditorZoomAction | null {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null
  if (event.deltaY < 0) return 'in'
  if (event.deltaY > 0) return 'out'
  return null
}

export function handleEditorZoomWheel(
  event: WheelEvent,
  apply: (action: EditorZoomAction) => void,
) {
  const action = matchEditorZoomWheel(event)
  if (!action) return false
  event.preventDefault()
  event.stopPropagation()
  apply(action)
  return true
}

export const useCodeEditorZoomStore = defineStore('codeEditorZoom', {
  state: () => ({
    step: 0,
  }),
  getters: {
    monacoFontSize: (s) => 14 + s.step,
  },
  actions: {
    hydrate() {
      this.step = readStoredStep(CODE_KEY, LEGACY_KEY)
    },
    persist() {
      localStorage.setItem(CODE_KEY, String(this.step))
    },
    zoomIn() {
      this.step = clampStep(this.step + 1)
      this.persist()
    },
    zoomOut() {
      this.step = clampStep(this.step - 1)
      this.persist()
    },
    reset() {
      this.step = 0
      this.persist()
    },
    apply(action: EditorZoomAction) {
      if (action === 'in') this.zoomIn()
      else if (action === 'out') this.zoomOut()
      else this.reset()
    },
  },
})

export const useMarkdownEditorZoomStore = defineStore('markdownEditorZoom', {
  state: () => ({
    step: 0,
  }),
  getters: {
    markdownBodyFontSize: (s) => 16 + s.step,
    markdownCodeFontSize: (s) => 14 + s.step,
    markdownSmallFontSize: (s) => Math.max(12, 12 + s.step),
    markdownLineHeight: (s) => 20 + s.step,
  },
  actions: {
    hydrate() {
      this.step = readStoredStep(MARKDOWN_KEY)
    },
    persist() {
      localStorage.setItem(MARKDOWN_KEY, String(this.step))
    },
    zoomIn() {
      this.step = clampStep(this.step + 1)
      this.persist()
    },
    zoomOut() {
      this.step = clampStep(this.step - 1)
      this.persist()
    },
    reset() {
      this.step = 0
      this.persist()
    },
    apply(action: EditorZoomAction) {
      if (action === 'in') this.zoomIn()
      else if (action === 'out') this.zoomOut()
      else this.reset()
    },
  },
})
