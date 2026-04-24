<script setup lang="ts">
import { computed } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { useThemeStore } from '../stores/theme'

const props = defineProps<{
  modelValue: string
  height?: string
  placeholder?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const theme = useThemeStore()
const mdTheme = computed(() => theme.mode)
const shellStyle = {
  '--liteoj-md-font-size': '16px',
  '--liteoj-md-code-font-size': '14px',
  '--liteoj-md-small-font-size': '12px',
  '--liteoj-md-line-height': '20px',
}
const editorStyle = computed(() => ({ height: props.height || '420px' }))
</script>

<template>
  <div class="markdown-editor-shell" :style="shellStyle">
    <MdEditor
      :model-value="modelValue"
      @update:model-value="(v: string) => emit('update:modelValue', v)"
      :placeholder="placeholder"
      :style="editorStyle"
      :theme="mdTheme"
      preview-theme="github"
      language="zh-CN"
      :show-code-row-number="false"
      :toolbars-exclude="['save', 'htmlPreview', 'github', 'mermaid', 'prettier', 'pageFullscreen']"
    />
  </div>
</template>

<style scoped>
.markdown-editor-shell {
  width: 100%;
}

.markdown-editor-shell :deep(.md-editor .cm-editor) {
  font-size: var(--liteoj-md-code-font-size);
}

.markdown-editor-shell :deep(.md-editor .cm-scroller),
.markdown-editor-shell :deep(.md-editor .cm-scroller .cm-line) {
  line-height: var(--liteoj-md-line-height);
}

.markdown-editor-shell :deep(.md-editor .md-editor-preview),
.markdown-editor-shell :deep(.md-editor .md-editor-html) {
  font-size: var(--liteoj-md-font-size);
}

.markdown-editor-shell :deep(.md-editor .md-editor-preview .md-editor-code),
.markdown-editor-shell :deep(.md-editor .md-editor-preview .md-editor-code .md-editor-code-head),
.markdown-editor-shell :deep(.md-editor .md-editor-preview .md-editor-code pre code),
.markdown-editor-shell :deep(.md-editor .md-editor-preview .md-editor-admonition) {
  font-size: var(--liteoj-md-code-font-size);
}

.markdown-editor-shell :deep(.md-editor .md-editor-toolbar-item-name),
.markdown-editor-shell :deep(.md-editor .md-editor-footer),
.markdown-editor-shell :deep(.md-editor .md-editor-catalog-editor),
.markdown-editor-shell :deep(.md-editor .md-editor-input),
.markdown-editor-shell :deep(.md-editor .md-editor-btn),
.markdown-editor-shell :deep(.md-editor .md-editor-modal-body),
.markdown-editor-shell :deep(.md-editor .md-editor-modal-header) {
  font-size: var(--liteoj-md-small-font-size);
}
</style>
