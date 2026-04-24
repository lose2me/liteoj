<script setup lang="ts">
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { handleEditorZoomShortcut, handleEditorZoomWheel, useCodeEditorZoomStore } from '../stores/editorZoom'
import { useThemeStore } from '../stores/theme'

const props = defineProps<{ modelValue: string; language: string; height?: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()
const theme = useThemeStore()
const editorZoom = useCodeEditorZoomStore()
const shellRef = ref<HTMLElement | null>(null)

const monacoLang = computed(() => {
  switch (props.language) {
    case 'cpp': return 'cpp'
    case 'c': return 'c'
    case 'java': return 'java'
    case 'python': return 'python'
    default: return 'plaintext'
  }
})

const options = computed(() => ({
  fontSize: editorZoom.monacoFontSize,
  minimap: { enabled: false },
  automaticLayout: true,
  scrollBeyondLastLine: false,
  glyphMargin: false,
  lineDecorationsWidth: 0,
  lineNumbersMinChars: 2,
  overviewRulerBorder: false,
  overviewRulerLanes: 0,
  tabSize: 4,
}))

const editorTheme = computed(() => (theme.isDark ? 'vs-dark' : 'vs'))

const handleZoomKeydown = (event: KeyboardEvent) => {
  handleEditorZoomShortcut(event, (action) => editorZoom.apply(action))
}

const handleZoomWheel = (event: WheelEvent) => {
  handleEditorZoomWheel(event, (action) => editorZoom.apply(action))
}

onMounted(() => {
  shellRef.value?.addEventListener('wheel', handleZoomWheel, { passive: false, capture: true })
})

onUnmounted(() => {
  shellRef.value?.removeEventListener('wheel', handleZoomWheel, true)
})
</script>

<template>
  <div ref="shellRef" class="code-editor-shell" @keydown.capture="handleZoomKeydown">
    <VueMonacoEditor
      :value="modelValue"
      @update:value="(v: string) => emit('update:modelValue', v)"
      :language="monacoLang"
      :theme="editorTheme"
      :height="height || '760px'"
      :options="options"
    />
  </div>
</template>

<style scoped>
.code-editor-shell {
  width: 100%;
}

.code-editor-shell :deep(.monaco-editor .line-numbers) {
  left: 6px !important;
  width: calc(100% - 6px) !important;
  text-align: left !important;
}
</style>
