<script setup lang="ts">
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

defineProps<{
  modelValue: string
  height?: string
  placeholder?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()
</script>

<template>
  <MdEditor
    :model-value="modelValue"
    @update:model-value="(v: string) => emit('update:modelValue', v)"
    :placeholder="placeholder"
    :style="{ height: height || '420px' }"
    theme="dark"
    preview-theme="github"
    language="zh-CN"
    :show-code-row-number="false"
    :toolbars-exclude="['save', 'htmlPreview', 'github', 'mermaid', 'prettier', 'pageFullscreen']"
  />
</template>
