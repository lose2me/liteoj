<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NTag, NTabs, NTabPane, NSelect, NButton, NAlert, useMessage } from 'naive-ui'
import MarkdownView from '../components/MarkdownView.vue'
import CodeEditor from '../components/CodeEditor.vue'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { verdictType } from '../api/verdict'
import { t } from '../i18n'

const route = useRoute()
const msg = useMessage()

const problem = ref<any>(null)
const tags = ref<any[]>([])
const languages = ref<string[]>([])
const lang = ref('cpp')
const code = ref('')
const submitting = ref(false)
const result = ref<any>(null)
const pollTimer = ref<number | null>(null)
// 题单上下文下，后端返回当前题单是否禁用思路/题解/AI 解析。前端在标题栏
// 渲染对应标签告知学生"这个 tab 为什么不见了"——比默默隐藏更友好。
const restrictions = ref({ idea: false, solution: false, ai: false })
// 上次 AI 结果：后端按上下文回传——独立页取 problem_set_id IS NULL 的提交，
// 题单页取 problem_set_id=当前题单 的提交。题单禁用 AI 时后端不返回；这里
// 再以 type 字段区分"上一次解析"（非 AC）与"上一次优化"（AC）。
const myLatestAI = ref<{ submission_id: number; verdict: string; explanation: string; type: 'analyze' | 'optimize' } | null>(null)
const detailTab = ref<'desc' | 'idea' | 'solution' | 'ai'>('desc')
// 本次判完提交的 AI 状态——按钮直接落在 result 区，不跳 SubmissionDetail。
// ai_busy=true：已发起、等 SSE。aiPendingTaskId 用于匹配 ai:task:done 事件。
const aiBusy = ref(false)
const aiPendingTaskId = ref<number | null>(null)

// If opened from a problem set, we carry ?problemset=<id> so the language list
// is narrowed and the backend enforces the whitelist on submit. Absent means
// standalone access (no restriction).
const psid = computed(() => {
  const v = route.query.problemset
  return typeof v === 'string' && v ? Number(v) : null
})

type AlertType = 'default' | 'info' | 'success' | 'warning' | 'error'
const alertType = computed<AlertType>(() => {
  const tv = verdictType(result.value?.verdict)
  return (tv === 'primary' ? 'info' : tv) as AlertType
})

const codeKey = (pid: string | number, l: string) => `liteoj.code.${pid}.${l}`

const normalizeDetailTab = () => {
  if (detailTab.value === 'idea' && !problem.value?.solution_idea_md) {
    detailTab.value = 'desc'
  } else if (detailTab.value === 'solution' && !problem.value?.solution_md) {
    detailTab.value = 'desc'
  } else if (detailTab.value === 'ai' && !myLatestAI.value) {
    detailTab.value = 'desc'
  }
}

const load = async () => {
  const { data } = await http.get(`/problems/${route.params.id}`, {
    params: psid.value ? { problemset_id: psid.value } : {},
  })
  problem.value = data.problem
  tags.value = data.tags || []
  languages.value = data.languages
  if (!languages.value.includes(lang.value)) lang.value = languages.value[0] || 'cpp'
  code.value = localStorage.getItem(codeKey(route.params.id as string, lang.value)) || ''
  restrictions.value = {
    idea: !!data.disable_idea,
    solution: !!data.disable_solution,
    ai: !!data.disable_ai,
  }
  myLatestAI.value = data.my_latest_ai || null
  normalizeDetailTab()
}

const onLangChange = (v: string) => {
  lang.value = v
  code.value = localStorage.getItem(codeKey(route.params.id as string, v)) || ''
}

const saveDraft = () =>
  localStorage.setItem(codeKey(route.params.id as string, lang.value), code.value)

const stopPoll = () => {
  if (pollTimer.value) {
    clearInterval(pollTimer.value)
    pollTimer.value = null
  }
}

const startPoll = (submissionID: number) => {
  stopPoll()
  pollTimer.value = window.setInterval(async () => {
    try {
      const { data } = await http.get(`/submissions/${submissionID}`)
      if (data.verdict && data.verdict !== 'PENDING') {
        stopPoll()
        const cases = data.testcase_result_json ? JSON.parse(data.testcase_result_json) : []
        result.value = {
          submission_id: data.id,
          verdict: data.verdict,
          message: data.message,
          time_used_ms: data.time_used_ms,
          memory_used_kb: data.memory_used_kb,
          cases,
          ai_disabled: !!data.ai_disabled,
          ai_rejected: !!data.ai_rejected,
          ai_reject_reason: data.ai_reject_reason || '',
          ai_explanation: data.ai_explanation || '',
        }
        if (data.verdict === 'AC') msg.success(t.problem.acCongrats)
        else msg.warning(data.verdict)
      }
    } catch {
      stopPoll()
    }
  }, 1200)
}

// 本次提交结果区的 AI 按钮条件 —— 两套对称：
//   canAnalyze：判完非 AC & 未生成 & 未被拒 → "AI 解析"
//   canOptimize：判完 AC              & 未生成          → "AI 优化"
// 两者都要求题单未禁用 AI（ai_disabled）。
const canAnalyze = computed(() =>
  result.value
  && result.value.submission_id
  && result.value.verdict
  && result.value.verdict !== 'AC'
  && result.value.verdict !== 'PENDING'
  && !result.value.ai_disabled
  && !result.value.ai_explanation
  && !result.value.ai_rejected,
)
const canOptimize = computed(() =>
  result.value
  && result.value.submission_id
  && result.value.verdict === 'AC'
  && !result.value.ai_disabled
  && !result.value.ai_explanation,
)
const refreshCurrentSubmission = async () => {
  if (!result.value?.submission_id) return
  const { data } = await http.get(`/submissions/${result.value.submission_id}`)
  result.value = {
    ...result.value,
    ai_disabled: !!data.ai_disabled,
    ai_rejected: !!data.ai_rejected,
    ai_reject_reason: data.ai_reject_reason || '',
    ai_explanation: data.ai_explanation || '',
  }
}

const focusLatestAITab = async () => {
  await load()
  if (myLatestAI.value) detailTab.value = 'ai'
}

const runAI = async (endpoint: 'analyze' | 'optimize') => {
  if (!result.value?.submission_id) return
  aiBusy.value = true
  const failedMsg = endpoint === 'optimize'
    ? t.submission.aiOptimizeFailed
    : t.submission.aiAnalyzeFailed
  const dispatchedMsg = endpoint === 'optimize'
    ? t.submission.aiOptimizeDispatched
    : t.submission.aiAnalyzeDispatched
  try {
    const { data } = await http.post(`/submissions/${result.value.submission_id}/${endpoint}`)
    if (data?.cached || data?.explanation) {
      // 提交区不再内联展示 AI 正文；命中缓存时直接刷新左侧“上一次解析/优化”。
      result.value.ai_explanation = data.explanation
      await focusLatestAITab()
      aiBusy.value = false
      return
    }
    if (data?.task_id) {
      aiPendingTaskId.value = data.task_id
      msg.info(dispatchedMsg)
      // 保留 aiBusy=true，等 ai:task:done 到来再 refetch & 解除。
      return
    }
    // 兜底：响应格式异常（不含 task_id 也不含 explanation），别让按钮卡在 loading。
    msg.warning(failedMsg)
    aiBusy.value = false
  } catch (e: any) {
    msg.error(e?.response?.data?.error || failedMsg)
    aiBusy.value = false
  }
}
const analyze = () => runAI('analyze')
const optimize = () => runAI('optimize')

let offEvt: (() => void) | null = null

const submit = async () => {
  if (!code.value.trim()) {
    msg.warning(t.problem.codeEmpty)
    return
  }
  saveDraft()
  submitting.value = true
  // 新提交开始——清空上一次的 AI 状态。
  aiBusy.value = false
  aiPendingTaskId.value = null
  result.value = { verdict: 'PENDING', cases: [], submission_id: 0 }
  try {
    const { data } = await http.post(`/problems/${route.params.id}/submit`, {
      language: lang.value, code: code.value,
      problemset_id: psid.value || undefined,
    })
    startPoll(data.submission_id)
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.problem.submitFailed)
    result.value = null
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await load()
  offEvt = onEvent(async (ev) => {
    // admin 改题目/题单规则/踢人加人 → 粗粒度刷新当前页面。
    if (ev.type === 'problem:changed' && ev.data?.id === Number(route.params.id)) {
      await load()
      return
    }
    if (ev.type === 'problemset:changed' && psid.value && ev.data?.id === psid.value) {
      await load()
      return
    }
    if (ev.type === 'problemset:members:changed' && psid.value && ev.data?.id === psid.value) {
      await load()
      return
    }
    if (ev.type !== 'ai:task:done') return
    const d = ev.data as { id: number; status: string }
    if (aiPendingTaskId.value !== d.id) return
    aiPendingTaskId.value = null
    try {
      if (d.status === 'done') {
        await refreshCurrentSubmission()
        await focusLatestAITab()
      } else {
        msg.error(result.value?.verdict === 'AC'
          ? t.submission.aiOptimizeFailed
          : t.submission.aiAnalyzeFailed)
      }
    } finally {
      aiBusy.value = false
    }
  })
})
onUnmounted(() => {
  stopPoll()
  offEvt?.()
})
</script>

<template>
  <!-- 宽屏两列（题面左、代码右），半屏以下（< xl=1280）单列：
       代码块自然排到题面下方，避免半分屏时两栏挤成一条线读不顺。
       items-start 让两张卡按各自内容高度独立撑开，不互相拉到同高。 -->
  <div v-if="problem" class="grid grid-cols-1 xl:grid-cols-2 gap-4 items-start">
    <NCard :title="`#${problem.id}  ${problem.title}`">
      <NSpace class="mb-3">
        <NTag v-if="psid" type="warning" size="small">
          {{ t.problem.langLimitedInSet(psid) }}
        </NTag>
        <NTag v-if="psid && restrictions.idea" type="warning" size="small">
          {{ t.problem.disableIdeaTag }}
        </NTag>
        <NTag v-if="psid && restrictions.solution" type="warning" size="small">
          {{ t.problem.disableSolutionTag }}
        </NTag>
        <NTag v-if="psid && restrictions.ai" type="warning" size="small">
          {{ t.problem.disableAITag }}
        </NTag>
        <NTag>{{ problem.difficulty || t.problem.notRated }}</NTag>
        <NTag type="warning">{{ problem.time_limit_ms }} ms</NTag>
        <NTag type="warning">{{ problem.memory_limit_mb }} MB</NTag>
        <NTag v-for="tag in tags" :key="tag.id" size="small">{{ tag.name }}</NTag>
      </NSpace>

      <NTabs v-model:value="detailTab" type="line" animated>
        <NTabPane name="desc" :tab="t.problem.descTitle">
          <MarkdownView :content="problem.description" />
        </NTabPane>
        <NTabPane
          v-if="problem.solution_idea_md"
          name="idea"
          :tab="t.problem.solutionIdea"
        >
          <MarkdownView :content="problem.solution_idea_md" />
        </NTabPane>
        <NTabPane
          v-if="problem.solution_md"
          name="solution"
          :tab="t.problem.solution"
        >
          <MarkdownView :content="problem.solution_md" />
        </NTabPane>
        <!-- 上一次 AI tab：只在当前上下文里呈现（独立页看独立提交、题单页
             看该题单内提交；题单禁用 AI 时后端直接不返回 my_latest_ai）。
             type 区分解析 / 优化，tab 标题和正文提示跟着走。 -->
        <NTabPane
          v-if="myLatestAI"
          name="ai"
          :tab="myLatestAI.type === 'optimize' ? t.problem.aiLastOptimize : t.problem.aiLastAnalysis"
        >
          <div class="text-xs opacity-60 mb-2">
            {{ t.problem.aiLastHint(myLatestAI.submission_id, myLatestAI.verdict, myLatestAI.type) }}
          </div>
          <MarkdownView :content="myLatestAI.explanation" />
        </NTabPane>
      </NTabs>
    </NCard>

    <NCard :title="t.problem.submitCode">
      <NSpace class="mb-2">
        <NSelect
          :value="lang"
          :options="languages.map((l) => ({ label: l, value: l }))"
          class="w-32"
          @update:value="onLangChange"
        />
        <NButton :loading="submitting" type="primary" @click="submit">{{ t.common.submit }}</NButton>
        <NButton @click="saveDraft">{{ t.problem.saveDraft }}</NButton>
      </NSpace>
      <CodeEditor v-model="code" :language="lang" />
      <div v-if="result" class="mt-3">
        <NAlert :type="alertType">
          <b>{{ result.verdict }}</b>
          <span v-if="result.verdict !== 'PENDING'">
            · {{ t.problem.timeUsedMs(result.time_used_ms) }} · {{ t.problem.memoryUsedKb(result.memory_used_kb) }}
          </span>
          <span v-else class="ml-2 opacity-70">{{ t.problem.judging }}</span>
        </NAlert>
        <div v-if="result.cases?.length" class="mt-2">
          <div v-for="c in result.cases" :key="c.index" class="text-sm py-1">
            <NTag :type="verdictType(c.verdict)" size="small">{{ t.problem.caseLabel(c.index) }}</NTag>
            {{ c.verdict }} · {{ c.time_ms }} ms / {{ c.memory_kb }} KB
            <span v-if="c.message" class="opacity-70 ml-2">{{ c.message }}</span>
          </div>
        </div>
        <!-- 判完后的 AI 入口：非 AC → AI 解析；AC → AI 优化；被拒则用灰
             tag 展示原因；题单 disable_ai 时按钮都不出现。 -->
        <NSpace v-if="canAnalyze || canOptimize || result.ai_rejected" class="mt-3" align="center">
          <NButton v-if="canAnalyze" :loading="aiBusy" type="primary" size="small" @click="analyze">
            {{ t.submission.aiAnalyze }}
          </NButton>
          <NButton v-if="canOptimize" :loading="aiBusy" type="primary" size="small" @click="optimize">
            {{ t.submission.aiOptimize }}
          </NButton>
          <NButton v-if="result.ai_rejected" disabled size="small">
            {{ t.submission.aiRejectedBadge(result.ai_reject_reason || '') }}
          </NButton>
        </NSpace>
      </div>
    </NCard>
  </div>
</template>
