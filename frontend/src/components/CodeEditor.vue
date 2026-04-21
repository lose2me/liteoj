<script setup lang="ts">
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import { computed } from 'vue'

const props = defineProps<{ modelValue: string; language: string; height?: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const monacoLang = computed(() => {
  switch (props.language) {
    case 'cpp': return 'cpp'
    case 'c': return 'c'
    case 'java': return 'java'
    case 'python': return 'python'
    case 'go': return 'go'
    default: return 'plaintext'
  }
})

const options = {
  fontSize: 14,
  minimap: { enabled: false },
  automaticLayout: true,
  scrollBeyondLastLine: false,
  tabSize: 4,
}
</script>

<template>
  <VueMonacoEditor
    :value="modelValue"
    @update:value="(v: string) => emit('update:modelValue', v)"
    :language="monacoLang"
    theme="vs-dark"
    :height="height || '620px'"
    :options="options"
  />
</template>
