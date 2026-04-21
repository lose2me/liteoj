<script setup lang="ts">
import { NCard, NTag, NDescriptions, NDescriptionsItem, NButton, NSpace, useMessage } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MarkdownView from '../components/MarkdownView.vue'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { verdictType } from '../api/verdict'
import { t } from '../i18n'

const route = useRoute()
const router = useRouter() // 保留：gotoProblem 用
const msg = useMessage()
const sub = ref<any>(null)
const aiBusy = ref(false)
const aiText = ref('')
// Task id returned by the async /analyze|/optimize endpoint. When the matching
// `ai:task:done` event arrives we refetch this submission (its
// ai_explanation column was populated by the worker) and drop the spinner.
const pendingTaskId = ref<number | null>(null)

const cases = computed<any[]>(() => {
  if (!sub.value?.testcase_result_json) return []
  try { return JSON.parse(sub.value.testcase_result_json) } catch { return [] }
})

// AC 已通过 → 不能走错因分析；非 AC 且非 PENDING → 不能走优化建议。
// 两个按钮互斥，同一个 ai_explanation 字段两者共用（由 verdict 区分语义）。
const canAnalyze = computed(() =>
  sub.value
  && sub.value.verdict !== 'AC'
  && sub.value.verdict !== 'PENDING'
  && !sub.value.ai_disabled
  && !sub.value.ai_explanation
  && !sub.value.ai_rejected,
)
const canOptimize = computed(() =>
  sub.value
  && sub.value.verdict === 'AC'
  && !sub.value.ai_disabled
  && !sub.value.ai_explanation,
)

// 代码块包裹：用 Markdown fenced code block 让 MarkdownView 走 highlight.js
// 管道，获得和题解一致的语法高亮外观。language 只做最基本的归一化。
const codeMd = computed(() => {
  const lang = String(sub.value?.language || '').toLowerCase().replace(/[^a-z0-9]/g, '')
  const code = sub.value?.code || ''
  return '```' + lang + '\n' + code + '\n```'
})

// ai_explanation 在 AC / 非 AC 下标题不同（"优化建议" vs "AI 解析"）。
const aiSectionTitle = computed(() =>
  sub.value?.verdict === 'AC' ? t.submission.aiOptimizeTitle : t.submission.aiAnalyzeTitle,
)

const refreshExplanation = async () => {
  const { data } = await http.get(`/submissions/${route.params.id}`)
  sub.value = data
  if (data.ai_explanation) aiText.value = data.ai_explanation
}

let offEvt: (() => void) | null = null
onMounted(async () => {
  await refreshExplanation()
  offEvt = onEvent(async (ev) => {
    // admin 改题目 → 该题所有 submission 的 ai_explanation 被清——当前页
    // 正好在看其中一条时要 refetch 让 UI 回到"未生成"状态。
    if (ev.type === 'problem:changed' && sub.value?.problem_id === ev.data?.id) {
      await refreshExplanation()
      return
    }
    if (ev.type !== 'ai:task:done') return
    const d = ev.data as { id: number; status: string }
    if (pendingTaskId.value !== d.id) return
    pendingTaskId.value = null
    try {
      if (d.status === 'done') {
        await refreshExplanation()
      } else {
        // No admin access to /admin/ai/tasks/:id for students — a generic
        // toast is the best we can do without leaking task history.
        msg.error(sub.value?.verdict === 'AC'
          ? t.submission.aiOptimizeFailed
          : t.submission.aiAnalyzeFailed)
      }
    } finally {
      aiBusy.value = false
    }
  })
})
onUnmounted(() => { offEvt?.() })

// 点击 AI 解析 / 优化建议后：立刻发起任务并留在当前详情页 —— 学生此时就在看
// 这一条提交，跳回列表反而打断上下文。aiBusy 令按钮进入 loading 态，等
// ai:task:done SSE 事件到达后 refreshExplanation 把 ai_explanation 渲染进页面。
const runAI = async (endpoint: 'analyze' | 'optimize') => {
  aiBusy.value = true
  const failedMsg = endpoint === 'optimize'
    ? t.submission.aiOptimizeFailed
    : t.submission.aiAnalyzeFailed
  const dispatchedMsg = endpoint === 'optimize'
    ? t.submission.aiOptimizeDispatched
    : t.submission.aiAnalyzeDispatched
  try {
    const { data } = await http.post(`/submissions/${route.params.id}/${endpoint}`)
    if (data?.cached || data?.explanation) {
      // 已有缓存——直接显示在本页即可。
      aiText.value = data.explanation
      aiBusy.value = false
      await refreshExplanation()
      return
    }
    if (data?.task_id) {
      pendingTaskId.value = data.task_id
      msg.info(dispatchedMsg)
      // 保留 aiBusy=true，等 SSE 完成再 refetch & 解除。
      return
    }
    // 兜底：响应格式异常，别让按钮卡 loading。
    msg.warning(failedMsg)
    aiBusy.value = false
  } catch (e: any) {
    msg.error(e?.response?.data?.error || failedMsg)
    aiBusy.value = false
  }
}
const analyze = () => runAI('analyze')
const optimize = () => runAI('optimize')

const gotoDiffPrev = () =>
  router.push(`/submissions/${route.params.id}/diff/0`)
</script>

<template>
  <NCard v-if="sub" :title="t.submission.detailTitle(sub.id)">
    <NDescriptions :column="3" bordered label-placement="left">
      <NDescriptionsItem :label="t.submission.fieldProblem">#{{ sub.problem_id }}</NDescriptionsItem>
      <NDescriptionsItem :label="t.submission.fieldLang">{{ sub.language }}</NDescriptionsItem>
      <NDescriptionsItem :label="t.submission.fieldVerdict">
        <NTag :type="verdictType(sub.verdict)">
          {{ sub.verdict }}
        </NTag>
      </NDescriptionsItem>
      <NDescriptionsItem :label="t.submission.fieldTime">{{ sub.time_used_ms }} ms</NDescriptionsItem>
      <NDescriptionsItem :label="t.submission.fieldMemory">{{ sub.memory_used_kb }} KB</NDescriptionsItem>
      <NDescriptionsItem :label="t.submission.fieldCreatedAt">{{ sub.created_at }}</NDescriptionsItem>
    </NDescriptions>

    <NSpace class="mt-3" align="center">
      <NButton @click="router.push(`/problems/${sub.problem_id}`)">
        {{ t.submission.gotoProblem }}
      </NButton>
      <NButton v-if="canAnalyze" :loading="aiBusy" type="primary" @click="analyze">{{ t.submission.aiAnalyze }}</NButton>
      <NButton v-if="canOptimize" :loading="aiBusy" type="primary" @click="optimize">{{ t.submission.aiOptimize }}</NButton>
      <NButton v-if="sub.has_prev" @click="gotoDiffPrev">{{ t.submission.diffPrev }}</NButton>
      <!-- rejected 状态放在"与上次对比"右侧，跟其它按钮等高对齐——灰
           disabled 态本来就适合表达"不可点的状态标"。 -->
      <NButton v-if="sub.ai_rejected" disabled>
        {{ t.submission.aiRejectedBadge(sub.ai_reject_reason || '') }}
      </NButton>
    </NSpace>

    <!-- 代码用 MarkdownView 渲染 fenced code block，拿到 highlight.js 语法高亮。 -->
    <h4 class="mt-4">{{ t.submission.codeSection }}</h4>
    <MarkdownView :content="codeMd" />

    <div v-if="aiText && !sub.ai_rejected" class="mt-3">
      <h4>{{ aiSectionTitle }}</h4>
      <NCard embedded>
        <MarkdownView :content="aiText" />
      </NCard>
    </div>

    <h4 class="mt-4">{{ t.submission.testcasesSection }}</h4>
    <div v-if="sub.message" class="mb-2 whitespace-pre-wrap opacity-80">{{ sub.message }}</div>
    <div v-for="c in cases" :key="c.index" class="text-sm py-1">
      <NTag :type="verdictType(c.verdict)" size="small">#{{ c.index }}</NTag>
      {{ c.verdict }} · {{ c.time_ms }} ms / {{ c.memory_kb }} KB
      <span v-if="c.message" class="opacity-70 ml-2">{{ c.message }}</span>
    </div>
  </NCard>
</template>
