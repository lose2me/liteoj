<script setup lang="ts">
// Unified read-only renderer for both admin preview and student-facing views.
// Delegates to md-editor-v3's MdPreview so KaTeX / code-highlight / table /
// checkbox quirks are handled by the same pipeline as the admin edit pane.
// Rationale: markdown-it-katex@2.0.3 mishandled math delimiters adjacent to
// CJK punctuation, which left student-side formulas partially rendered; the
// editor's own preview doesn't have that bug.
//
// 跟随全站主题切换 md-editor-v3 的 dark / light 模式；否则 Naive UI 已经
// 切到亮色时，预览区仍会保持深色面板，视觉割裂。
import { computed } from 'vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { useThemeStore } from '../stores/theme'

defineProps<{ content: string }>()

const theme = useThemeStore()
const mdTheme = computed(() => theme.mode)
</script>

<template>
  <div class="markdown-body">
    <MdPreview
      :model-value="content || ''"
      :theme="mdTheme"
      language="zh-CN"
      preview-theme="github"
      :show-code-row-number="false"
    />
  </div>
</template>

<style scoped>
.markdown-body { line-height: 1.7; }
/* Make the preview blend with the surrounding NCard. The code-block header
   bar + diff-color reset live in src/styles/markdown-overrides.css so the
   rule also reaches MdEditor's internal preview pane in admin edit mode. */
.markdown-body :deep(.md-editor-preview-wrapper) {
  padding: 0;
  background: transparent;
}
.markdown-body :deep(.md-editor) {
  background: transparent;
}
.markdown-body :deep(.md-editor-code pre) {
  border-radius: var(--md-theme-code-block-radius, 6px);
}
</style>
