<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, ref } from 'vue'
import {
  NCard, NForm, NFormItem, NInput, NInputNumber, NSelect, NSwitch,
  NButton, NSpace, NDataTable, NModal, NTabs, NTabPane, useMessage, NPopconfirm, NAlert,
} from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import MarkdownView from '../../components/MarkdownView.vue'
import MarkdownEditor from '../../components/MarkdownEditor.vue'
import TagPicker from '../../components/TagPicker.vue'
import { http } from '../../api/http'
import { onEvent } from '../../api/events'
import { t } from '../../i18n'

const route = useRoute()
const router = useRouter()
const msg = useMessage()
const p = ref<any>(null)
const tagIDs = ref<number[]>([])
// 三个 markdown 区域用 tabs 切换，默认停在"描述"。
const contentTab = ref<'desc' | 'idea' | 'solution'>('desc')
const testcases = ref<any[]>([])
const showPreview = ref(false)
const aiBusy = ref<Record<string, boolean>>({})
const aiTagSuggestion = ref<any>(null)
// AI import textarea — intentionally a local ref, NOT part of `p`. The
// backend Problem model no longer persists raw_source; admins paste here
// only to feed one-shot AI calls, and the text is discarded on reload.
const rawSource = ref('')
// Running-count of AI tasks targeting this problem. When >0 every AI button
// is disabled so overlapping flows can't race to apply to the same fields.
const aiRunning = ref(0)
// Locally-busy flag covers the window between our POST and the backend
// `ai:tasks:changed` broadcast arriving — aiRunning lags by one round trip,
// so without this we briefly allow double-submits from the same tab.
const aiLocalBusy = computed(() => Object.values(aiBusy.value).some(Boolean))
const aiAnyRunning = computed(() => aiRunning.value > 0 || aiLocalBusy.value)

const showTC = ref(false)
const tcForm = ref({ id: 0, input: '', expected_output: '', order_index: 0 })

const load = async () => {
  const { data } = await http.get(`/problems/${route.params.id}`)
  p.value = data.problem
  tagIDs.value = data.tag_ids || []
  const tc = await http.get(`/admin/problems/${route.params.id}/testcases`)
  testcases.value = tc.data.items
}

const fetchRunning = async () => {
  try {
    const { data } = await http.get(`/admin/problems/${route.params.id}/ai-running`)
    aiRunning.value = data.running || 0
  } catch {
    // Non-fatal — failing this check should never block the page.
  }
}

let offEvt: (() => void) | null = null
onMounted(async () => {
  await load()
  await fetchRunning()
  // Any AI task state change anywhere triggers a refetch of our scoped count;
  // cheap and lets other admins' in-flight tasks lock this page too.
  offEvt = onEvent((ev) => {
    if (ev.type === 'ai:tasks:changed') fetchRunning()
    if (ev.type === 'ai:task:done') handleAITaskDone(ev.data)
  })
})
onUnmounted(() => { offEvt?.() })

const save = async () => {
  await http.put(`/admin/problems/${route.params.id}`, { ...p.value, tag_ids: tagIDs.value })
  msg.success(t.common.savedOk)
}

// runAI wraps a POST to /admin/problems/:id/<path>. Since the backend is
// async, POST returns {task_id} immediately; the result arrives later via
// the `ai:task:done` SSE event. `pendingTasks` maps task_id → the callback
// that will apply the result once we fetch it from GET /admin/ai/tasks/:id.
// `aiBusy[key]` stays true across the full round-trip to keep the button in
// its loading state.
interface PendingTask {
  key: string
  apply: (data: any) => void
}
const pendingTasks = ref<Record<number, PendingTask>>({})

const runAI = async (
  key: string,
  path: string,
  apply: (data: any) => void,
  // Tag 流程需要 admin 在本页审阅 suggestion 再手动应用，不跳列表；其它
  // 一键解析 / 单项生成后端会自动把结果回写到 Problem 行，直接跳回列表。
  navigateAway = true,
) => {
  if (aiAnyRunning.value) {
    msg.warning(t.problemAdmin.aiAlreadyRunning)
    return
  }
  const raw = rawSource.value.trim()
  if (!raw) {
    msg.warning(t.problemAdmin.aiNeedRaw)
    return
  }
  aiBusy.value[key] = true
  try {
    const { data } = await http.post(
      `/admin/problems/${route.params.id}/${path}`,
      { raw },
    )
    if (data?.task_id) {
      if (navigateAway) {
        msg.info(t.problemAdmin.aiDispatched)
        router.push('/admin/problems')
        return
      }
      // Tag 路径沿用原先的 pending task 机制，SSE 回来在本页显示 suggestion。
      pendingTasks.value[data.task_id] = { key, apply }
      return
    }
    // No task_id means the backend fell back to the synchronous path.
    apply(data)
    aiBusy.value[key] = false
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.problemAdmin.aiFailed)
    aiBusy.value[key] = false
  }
}

// handleAITaskDone finalises a pending AI task: fetches the full row for its
// structured `result`, applies it, and clears the busy flag. Invoked by the
// SSE listener whenever `ai:task:done` lands.
const handleAITaskDone = async (ev: { id: number; kind: string; status: string }) => {
  const pending = pendingTasks.value[ev.id]
  if (!pending) return
  delete pendingTasks.value[ev.id]
  try {
    if (ev.status !== 'done') {
      const { data } = await http.get(`/admin/ai/tasks/${ev.id}`)
      msg.error(data?.error || t.problemAdmin.aiFailed)
      return
    }
    const { data } = await http.get(`/admin/ai/tasks/${ev.id}`)
    const parsed = data.result ? JSON.parse(data.result) : null
    if (!parsed) {
      msg.error(t.problemAdmin.aiFailed)
      return
    }
    pending.apply(parsed)
  } catch {
    msg.error(t.problemAdmin.aiFailed)
  } finally {
    aiBusy.value[pending.key] = false
  }
}

const runAITag = () => runAI(
  'tag', 'ai-tag',
  (data) => { aiTagSuggestion.value = data },
  false,
)

const applyAITag = () => {
  if (!aiTagSuggestion.value) return
  const ids = aiTagSuggestion.value.tag_ids || []
  tagIDs.value = Array.from(new Set([...tagIDs.value, ...ids]))
  // Apply suggested difficulty too — backend already filters to the 4 valid
  // values (入门/简单/中等/困难), so anything present is safe to write.
  const diff = aiTagSuggestion.value.difficulty
  if (diff && p.value) {
    p.value.difficulty = diff
  }
  msg.success(t.problemAdmin.mergedTags(ids.length))
}

const runAIGenTitle = () => runAI(
  'title', 'ai-gen-title',
  (d) => { p.value.title = d.title },
)
const runAIGenDesc = () => runAI(
  'desc', 'ai-gen-desc',
  (d) => { p.value.description = d.description },
)
const runAIGenIdea = () => runAI(
  'idea', 'ai-gen-idea',
  (d) => { p.value.solution_idea_md = d.solution_idea_md },
)
const runAIGenExplain = () => runAI(
  'explain', 'ai-gen-explain',
  (d) => { p.value.solution_md = d.solution_md },
)
const runAIGenAll = () => runAI(
  'all', 'ai-gen-all',
  (d) => {
    if (d.title) p.value.title = d.title
    if (d.description) p.value.description = d.description
    if (d.solution_idea_md !== undefined) p.value.solution_idea_md = d.solution_idea_md
    if (d.solution_md !== undefined) p.value.solution_md = d.solution_md
  },
)

const openTC = (row?: any) => {
  if (row) tcForm.value = { ...row }
  else tcForm.value = { id: 0, input: '', expected_output: '', order_index: testcases.value.length }
  showTC.value = true
}
const saveTC = async () => {
  if (tcForm.value.id) {
    await http.put(`/admin/problems/${route.params.id}/testcases/${tcForm.value.id}`, tcForm.value)
  } else {
    await http.post(`/admin/problems/${route.params.id}/testcases`, tcForm.value)
  }
  showTC.value = false
  await load()
}
const delTC = async (id: number) => {
  await http.delete(`/admin/problems/${route.params.id}/testcases/${id}`)
  await load()
}

const tcColumns = [
  { title: t.problemAdmin.tcOrderCol, key: 'order_index', width: 60 },
  { title: t.problemAdmin.tcInput, key: 'input', ellipsis: { tooltip: true } },
  { title: t.problemAdmin.tcExpected, key: 'expected_output', ellipsis: { tooltip: true } },
  {
    title: t.problemsetAdmin.colOp,
    key: 'op',
    width: 160,
    render: (r: any) => [
      h(NButton, { size: 'tiny', onClick: () => openTC(r) }, { default: () => t.common.edit }),
      ' ',
      h(NPopconfirm, { onPositiveClick: () => delTC(r.id) }, {
        trigger: () => h(NButton, { size: 'tiny', type: 'error' }, { default: () => t.common.delete }),
        default: () => t.problemAdmin.confirmDeleteTc,
      }),
    ],
  },
]
</script>

<template>
  <div v-if="p">
    <NButton class="mb-3" @click="router.back()">{{ t.problemAdmin.backHint }}</NButton>

    <NCard :title="t.problemAdmin.rawSourceLabel" class="mb-4">
      <NAlert v-if="aiAnyRunning" type="info" size="small" class="mb-2">
        {{ t.problemAdmin.aiRunningHint(aiRunning) }}
      </NAlert>
      <NInput
        v-model:value="rawSource"
        type="textarea"
        :autosize="{ minRows: 6, maxRows: 20 }"
        :placeholder="t.problemAdmin.rawSourcePlaceholder"
      />
      <NSpace class="mt-3" size="small">
        <NPopconfirm @positive-click="runAIGenAll">
          <template #trigger>
            <NButton type="primary" :disabled="aiAnyRunning" :loading="aiBusy.all">
              {{ t.problemAdmin.aiGenAll }}
            </NButton>
          </template>
          {{ t.problemAdmin.aiGenAllConfirm }}
        </NPopconfirm>
        <NButton :disabled="aiAnyRunning" :loading="aiBusy.title" @click="runAIGenTitle">
          {{ t.problemAdmin.aiGenTitle }}
        </NButton>
        <NButton :disabled="aiAnyRunning" :loading="aiBusy.desc" @click="runAIGenDesc">
          {{ t.problemAdmin.aiGenDesc }}
        </NButton>
        <NButton :disabled="aiAnyRunning" :loading="aiBusy.idea" @click="runAIGenIdea">
          {{ t.problemAdmin.aiGenIdea }}
        </NButton>
        <NButton :disabled="aiAnyRunning" :loading="aiBusy.explain" @click="runAIGenExplain">
          {{ t.problemAdmin.aiGenExplain }}
        </NButton>
        <NButton :disabled="aiAnyRunning" :loading="aiBusy.tag" @click="runAITag">
          {{ t.problemAdmin.aiTag }}
        </NButton>
      </NSpace>

      <NAlert v-if="aiTagSuggestion" class="mt-3" type="info" :title="t.problemAdmin.aiTagSuggestTitle">
        <div>{{ t.problemAdmin.matchedLabel }}{{ (aiTagSuggestion.matched || []).join('，') || t.common.none }}</div>
        <div v-if="aiTagSuggestion.unmatched?.length" class="mt-1 opacity-80">
          {{ t.problemAdmin.unmatchedLabel }}{{ aiTagSuggestion.unmatched.join('，') }}
        </div>
        <div v-if="aiTagSuggestion.difficulty" class="mt-1">
          {{ t.problemAdmin.difficultyLabel }}{{ aiTagSuggestion.difficulty }}
        </div>
        <NButton class="mt-2" size="small" type="primary" @click="applyAITag">
          {{ t.problemAdmin.appliedToForm((aiTagSuggestion.tag_ids || []).length) }}
        </NButton>
      </NAlert>
    </NCard>

    <NCard :title="t.problemAdmin.editTitle">
      <NForm label-placement="left" label-width="120">
        <NFormItem :label="t.problemAdmin.titleLabel"><NInput v-model:value="p.title" /></NFormItem>
        <NFormItem :label="t.problem.difficulty">
          <NSelect v-model:value="p.difficulty" :options="[
            { label: t.problem.diffEntry, value: t.problem.diffEntry },
            { label: t.problem.diffEasy, value: t.problem.diffEasy },
            { label: t.problem.diffMedium, value: t.problem.diffMedium },
            { label: t.problem.diffHard, value: t.problem.diffHard },
          ]" clearable />
        </NFormItem>
        <NFormItem :label="t.problemAdmin.timeLimit"><NInputNumber v-model:value="p.time_limit_ms" :min="100" :max="10000" /></NFormItem>
        <NFormItem :label="t.problemAdmin.memoryLimit"><NInputNumber v-model:value="p.memory_limit_mb" :min="16" :max="1024" /></NFormItem>
        <NFormItem :label="t.problemAdmin.visibleLabel"><NSwitch v-model:value="p.visible" /></NFormItem>
        <NFormItem :label="t.problemAdmin.tagLabel">
          <TagPicker v-model="tagIDs" />
        </NFormItem>
      </NForm>
      <!-- 三段 markdown 放进 tabs 切换，避免一屏铺开过长；naive-ui 默认会
           lazy 渲染但保留已激活 tab 的 DOM，切换不会丢编辑内容。 -->
      <NTabs v-model:value="contentTab" type="line" animated class="mt-2">
        <NTabPane name="desc" :tab="t.problemAdmin.descLabel">
          <MarkdownEditor v-model="p.description" height="520px" />
        </NTabPane>
        <NTabPane name="idea" :tab="t.problemAdmin.solutionIdeaLabel">
          <MarkdownEditor v-model="p.solution_idea_md" height="520px" />
        </NTabPane>
        <NTabPane name="solution" :tab="t.problemAdmin.solutionLabel">
          <MarkdownEditor v-model="p.solution_md" height="520px" />
        </NTabPane>
      </NTabs>
      <NSpace>
        <NButton type="primary" @click="save">{{ t.common.save }}</NButton>
        <NButton @click="showPreview = true">{{ t.problemAdmin.previewDesc }}</NButton>
      </NSpace>
    </NCard>

    <NCard class="mt-4" :title="t.problemAdmin.testcases">
      <NSpace class="mb-3"><NButton type="primary" @click="openTC()">{{ t.problemAdmin.testcaseNew }}</NButton></NSpace>
      <NDataTable :columns="tcColumns" :data="testcases" striped />
    </NCard>

    <NModal v-model:show="showPreview" preset="card" :title="t.problemAdmin.previewTitle" :style="{ width: 'min(800px, 96vw)' }">
      <MarkdownView :content="p.description" />
    </NModal>

    <NModal v-model:show="showTC" preset="card" :title="tcForm.id ? t.problemAdmin.testcaseEdit : t.problemAdmin.testcaseCreate" :style="{ width: 'min(640px, 96vw)' }">
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="t.problemAdmin.tcInput">
          <NInput v-model:value="tcForm.input" type="textarea" :autosize="{ minRows: 3, maxRows: 10 }" />
        </NFormItem>
        <NFormItem :label="t.problemAdmin.tcExpected">
          <NInput v-model:value="tcForm.expected_output" type="textarea" :autosize="{ minRows: 3, maxRows: 10 }" />
        </NFormItem>
        <NFormItem :label="t.problemAdmin.tcOrder"><NInputNumber v-model:value="tcForm.order_index" /></NFormItem>
        <NButton type="primary" @click="saveTC">{{ t.common.save }}</NButton>
      </NForm>
    </NModal>
  </div>
</template>
