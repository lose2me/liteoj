<script setup lang="ts">
import { NTag, NDataTable, NSelect, NInput, NSpace, NButton, NModal, NScrollbar, useMessage } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { h, onMounted, onUnmounted, ref } from 'vue'
import { http } from '../../api/http'
import { onEvent } from '../../api/events'
import { t } from '../../i18n'

interface Row {
  id: number
  kind: string
  user_id: number
  username: string
  subject: string
  status: string
  started_at: string
  finished_at: string | null
  duration_ms: number
  error: string
}

const items = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 16
const kindFilter = ref<string>('')
const statusFilter = ref<string>('')
const userFilter = ref<string>('')
let off: (() => void) | null = null

type DetailMode = 'prompt' | 'output'
const showDetail = ref(false)
const detailMode = ref<DetailMode>('prompt')
const detailRow = ref<Row | null>(null)
const detailText = ref<string>('')
const detailLoading = ref(false)
const msg = useMessage()

const load = async () => {
  const { data } = await http.get('/admin/ai/tasks', {
    params: {
      page: page.value,
      page_size: pageSize,
      kind: kindFilter.value || undefined,
      status: statusFilter.value || undefined,
      username: userFilter.value || undefined,
    },
  })
  items.value = data.items || []
  total.value = data.total || 0
}

// One fetch per row click — the list endpoint omits the prompt/output blobs
// so the full row only loads when the admin actually opens the modal.
const openDetail = async (r: Row, mode: DetailMode) => {
  detailRow.value = r
  detailMode.value = mode
  detailText.value = ''
  showDetail.value = true
  detailLoading.value = true
  try {
    const { data } = await http.get(`/admin/ai/tasks/${r.id}`)
    const raw = (mode === 'prompt' ? data.prompt : data.output) || ''
    // Output blobs for Tag / GenAll flows are JSON — pretty-print so the
    // admin can spot structural issues at a glance. Prompts and failures
    // aren't JSON, so fall back to the raw text.
    detailText.value = mode === 'output' ? tryPrettyJSON(raw) : raw
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.opFailed)
  } finally {
    detailLoading.value = false
  }
}

// tryPrettyJSON finds the first top-level JSON object / array inside the
// string and re-serializes it with 2-space indent. Leaves non-JSON text
// untouched — critical for analysis outputs which are Markdown, not JSON.
const tryPrettyJSON = (raw: string): string => {
  if (!raw) return raw
  const trimmed = raw.trim()
  // Fast path: whole body is JSON.
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try {
      return JSON.stringify(JSON.parse(trimmed), null, 2)
    } catch {
      /* fall through to embedded scan */
    }
  }
  // Upstream occasionally wraps JSON in an OpenAI envelope like
  // {"choices":[{"message":{"content":"{...}"}}]}; unwrap one level and
  // pretty-print the content if possible.
  try {
    const env = JSON.parse(trimmed)
    const inner = env?.choices?.[0]?.message?.content
    if (typeof inner === 'string') {
      const s = inner.trim()
      if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
        return JSON.stringify(JSON.parse(s), null, 2)
      }
    }
    return JSON.stringify(env, null, 2)
  } catch {
    return raw
  }
}

onMounted(() => {
  load()
  off = onEvent((ev) => {
    if (ev.type === 'ai:tasks:changed') load()
  })
})
onUnmounted(() => { off?.() })

const fmt = (s: string | null) => (s ? s.replace('T', ' ').slice(0, 19) : t.common.empty)

const kindLabel = (k: string) => {
  if (k === 'analyze') return t.adminDashboard.aiKindAnalyze
  if (k === 'optimize') return t.adminDashboard.aiKindOptimize
  if (k === 'tag') return t.adminDashboard.aiKindTag
  if (k === 'gen_title') return t.adminDashboard.aiKindGenTitle
  if (k === 'gen_desc') return t.adminDashboard.aiKindGenDesc
  if (k === 'gen_idea') return t.adminDashboard.aiKindGenIdea
  if (k === 'gen_explain') return t.adminDashboard.aiKindGenExplain
  if (k === 'gen_all') return t.adminDashboard.aiKindGenAll
  return k
}

const statusTagType = (s: string): 'default' | 'info' | 'success' | 'warning' | 'error' => {
  if (s === 'running') return 'info'
  if (s === 'done') return 'success'
  if (s === 'aborted') return 'warning'
  if (s === 'failed') return 'error'
  return 'default'
}

const statusLabel = (s: string) => {
  if (s === 'running') return t.adminDashboard.aiStatusRunning
  if (s === 'done') return t.adminDashboard.aiStatusDone
  if (s === 'aborted') return t.adminDashboard.aiStatusAborted
  if (s === 'failed') return t.adminDashboard.aiStatusFailed
  return s
}

const fmtDuration = (ms: number, status: string) => {
  if (status === 'running') return '—'
  if (ms < 1000) return `${ms} ms`
  return `${(ms / 1000).toFixed(1)} s`
}

const kindOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  { label: t.adminDashboard.aiKindAnalyze, value: 'analyze' },
  { label: t.adminDashboard.aiKindOptimize, value: 'optimize' },
  { label: t.adminDashboard.aiKindTag, value: 'tag' },
  { label: t.adminDashboard.aiKindGenTitle, value: 'gen_title' },
  { label: t.adminDashboard.aiKindGenDesc, value: 'gen_desc' },
  { label: t.adminDashboard.aiKindGenIdea, value: 'gen_idea' },
  { label: t.adminDashboard.aiKindGenExplain, value: 'gen_explain' },
  { label: t.adminDashboard.aiKindGenAll, value: 'gen_all' },
]

const statusOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  { label: t.adminDashboard.aiStatusRunning, value: 'running' },
  { label: t.adminDashboard.aiStatusDone, value: 'done' },
  { label: t.adminDashboard.aiStatusAborted, value: 'aborted' },
  { label: t.adminDashboard.aiStatusFailed, value: 'failed' },
]

const columns = [
  { title: t.common.id, key: 'id', width: 70 },
  {
    title: t.adminDashboard.aiColKind, key: 'kind', width: 110,
    render: (r: Row) => h(NTag, { size: 'small', type: 'info' }, { default: () => kindLabel(r.kind) }),
  },
  { title: t.adminDashboard.aiColUser, key: 'username', width: 120 },
  { title: t.adminDashboard.aiColSubject, key: 'subject', width: 220, ellipsis: { tooltip: true } },
  {
    title: t.adminDashboard.aiColStatus, key: 'status', width: 90,
    render: (r: Row) => h(NTag, { size: 'small', type: statusTagType(r.status) },
      { default: () => statusLabel(r.status) }),
  },
  {
    title: t.adminDashboard.aiColStarted, key: 'started_at', width: 170,
    render: (r: Row) => fmt(r.started_at),
  },
  {
    title: t.adminDashboard.aiColDuration, key: 'duration_ms', width: 90,
    render: (r: Row) => fmtDuration(r.duration_ms, r.status),
  },
  {
    title: t.adminDashboard.aiColError, key: 'error',
    ellipsis: { tooltip: true },
  },
  {
    title: t.adminDashboard.aiColOp, key: 'op', width: 170,
    render: (r: Row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openDetail(r, 'prompt') },
            { default: () => t.adminDashboard.aiViewPrompt }),
          h(NButton, { size: 'tiny', onClick: () => openDetail(r, 'output') },
            { default: () => t.adminDashboard.aiViewOutput }),
        ],
      }),
  },
]

const detailTitle = () => {
  if (!detailRow.value) return ''
  return detailMode.value === 'prompt'
    ? t.adminDashboard.aiPromptTitle(detailRow.value.id)
    : t.adminDashboard.aiOutputTitle(detailRow.value.id)
}

const emptyHint = () =>
  detailMode.value === 'prompt' ? t.adminDashboard.aiPromptEmpty : t.adminDashboard.aiOutputEmpty
</script>

<template>
  <div>
    <NSpace class="mb-3">
      <NSelect v-model:value="kindFilter" :options="kindOptions" class="w-32" :placeholder="t.adminDashboard.aiFilterKind" />
      <NSelect v-model:value="statusFilter" :options="statusOptions" class="w-32" :placeholder="t.adminDashboard.aiFilterStatus" />
      <NInput v-model:value="userFilter" :placeholder="t.adminDashboard.aiFilterUser" class="w-40" clearable @keyup.enter="page = 1; load()" />
      <NButton type="primary" @click="page = 1; load()">{{ t.common.filter }}</NButton>
    </NSpace>

    <NDataTable
      :columns="columns"
      :data="items"
      :pagination="{
        page, pageSize, itemCount: total, showSizePicker: false,
        onChange: (p: number) => { page = p; load() },
      }"
      remote
      striped
    />

    <NModal
      v-model:show="showDetail"
      preset="card"
      :title="detailTitle()"
      :style="{ width: 'min(960px, 96vw)' }"
      class="ai-detail-modal"
    >
      <!-- 固定 height=65vh：不论"查看提示词"还是"查看输出"，Modal 体积都一致；
           NScrollbar 的高度必须走 :style（class 里的 max-height 不会转发到
           内部 scroll 容器）；内容短时留白，超长时内部滚动不会顶破屏幕。 -->
      <div class="detail-body">
        <div v-if="detailLoading" class="opacity-70 text-sm">{{ t.common.loadingDots }}</div>
        <NScrollbar v-else-if="detailText" :style="{ height: '100%' }" x-scrollable>
          <pre class="prompt-block">{{ detailText }}</pre>
        </NScrollbar>
      <div v-else class="opacity-60 text-sm">{{ emptyHint() }}</div>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.detail-body {
  height: 65vh;
  background: var(--lo-subtle-bg);
  border-radius: 6px;
}
.prompt-block {
  padding: 12px;
  margin: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
