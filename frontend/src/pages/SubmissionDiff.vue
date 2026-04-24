<script setup lang="ts">
import { NCard, NSpace, NTag, useMessage } from 'naive-ui'
import { VueMonacoDiffEditor } from '@guolao/vue-monaco-editor'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { http } from '../api/http'
import { verdictType } from '../api/verdict'
import { handleEditorZoomShortcut, useCodeEditorZoomStore } from '../stores/editorZoom'
import { useThemeStore } from '../stores/theme'
import { t } from '../i18n'

const route = useRoute()
const msg = useMessage()
const theme = useThemeStore()
const editorZoom = useCodeEditorZoomStore()

const a = ref<any>(null)
const b = ref<any>(null)

const load = async () => {
  const id = route.params.id
  const other = route.params.otherId || 0
  try {
    const { data } = await http.get(`/submissions/${id}/diff/${other}`)
    a.value = data.a
    b.value = data.b
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.submission.diffLoadFailed)
  }
}
onMounted(load)

const language = computed(() => {
  const l = a.value?.language
  return l === 'cpp' ? 'cpp' : l === 'c' ? 'c' : l === 'java' ? 'java' : l === 'python' ? 'python' : 'plaintext'
})

const editorTheme = computed(() => (theme.isDark ? 'vs-dark' : 'vs'))
const editorOptions = computed(() => ({
  readOnly: true,
  automaticLayout: true,
  minimap: { enabled: false },
  fontSize: editorZoom.monacoFontSize,
}))

const handleZoomKeydown = (event: KeyboardEvent) => {
  handleEditorZoomShortcut(event, (action) => editorZoom.apply(action))
}
</script>

<template>
  <NCard :title="t.submission.diffTitle">
    <NSpace v-if="a && b" class="mb-2">
      <span>
        {{ t.submission.diffLeft(b.id) }}
        <NTag :type="verdictType(b.verdict)" size="small">{{ b.verdict }}</NTag>
      </span>
      <span>
        {{ t.submission.diffRight(a.id) }}
        <NTag :type="verdictType(a.verdict)" size="small">{{ a.verdict }}</NTag>
      </span>
    </NSpace>
    <NCard v-else-if="a && !b" embedded>{{ t.submission.diffNoOther }}</NCard>

    <div v-if="a && b" class="submission-diff-editor" @keydown.capture="handleZoomKeydown">
      <VueMonacoDiffEditor
        :original="b.code || ''"
        :modified="a.code || ''"
        :language="language"
        :theme="editorTheme"
        height="600px"
        :options="editorOptions"
      />
    </div>
  </NCard>
</template>

<style scoped>
.submission-diff-editor {
  width: 100%;
}
</style>
