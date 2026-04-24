<script setup lang="ts">
import { NDataTable, NTag, NSelect, NInput, NSpace, NButton, NSpin } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { h, onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../api/http'
import { verdictType, verdictKeys } from '../api/verdict'
import { getLanguages } from '../api/languages'
import { onEvent } from '../api/events'
import { useUserStore } from '../stores/user'
import { t } from '../i18n'

// 共用的提交列表组件。所有调用点（/submissions、/me 的"我的提交"、
// /admin/submissions、题单提交 tab）走同一份 columns，区别只有三处：
//   1. 后端 query 固定部分：query prop 透传 (user_id / problemset_id / verdict …)
//   2. 是否显示内置筛选栏：showFilters prop
//   3. 点进详情时是否带 from=ps：detailFromPs prop（题单 tab 用）
// 视角差异（学生/admin）直接读 user store，不用外部传。
const props = withDefaults(defineProps<{
  query?: Record<string, any>
  showFilters?: boolean
  detailFromPs?: boolean
  pageSize?: number
}>(), {
  showFilters: false,
  detailFromPs: false,
  pageSize: 16,
})

const router = useRouter()
const user = useUserStore()

const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const verdict = ref('')
const userFilter = ref('')
const problemFilter = ref('')
const language = ref('')
const langOptions = ref<SelectOption[]>([{ label: t.common.all, value: '' }])

const load = async () => {
  const params: Record<string, any> = {
    page: page.value,
    page_size: props.pageSize,
    ...(props.query || {}),
  }
  if (props.showFilters) {
    if (verdict.value) params.verdict = verdict.value
    if (userFilter.value) params.username = userFilter.value
    if (problemFilter.value) params.problem_id = problemFilter.value
    if (language.value) params.language = language.value
  }
  const { data } = await http.get('/submissions', { params })
  items.value = data.items
  total.value = data.total
}

let off: (() => void) | null = null

onMounted(async () => {
  if (props.showFilters) {
    const langs = await getLanguages()
    langOptions.value = [{ label: t.common.all, value: '' }, ...langs.map((l) => ({ label: l, value: l }))]
  }
  load()
  off = onEvent((ev) => {
    if (ev.type === 'submission:new' || ev.type === 'submission:done' || ev.type === 'ai:task:done') {
      load()
    }
    // admin 改题目会触发我们清空该题所有 AI 字段——列表里的 "AI 解析" 标签
    // 应当即时消失。粒度上直接全刷：submissions 列表只是一页数据，不贵。
    if (ev.type === 'problem:changed') load()
  })
})
onUnmounted(() => { off?.() })

// 外部 query 变动（切题单 id、切 verdict 等）也要重拉。
watch(() => props.query, () => { page.value = 1; load() }, { deep: true })

const verdictOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  ...verdictKeys.map((k) => ({ label: k, value: k })),
]

// 只有自己 / admin 可点进详情；AI 转圈中也禁点（避免点进去发呆）。
const canView = (r: any) => (user.isAdmin || r.user_id === user.user?.id) && !r.ai_pending

const gotoDetail = (id: number) => {
  if (props.detailFromPs) {
    router.push({ path: `/submissions/${id}`, query: { from: 'ps' } })
  } else {
    router.push(`/submissions/${id}`)
  }
}

const columns = computed(() => {
  const cols: any[] = [{ title: t.submission.colId, key: 'id', width: 70 }]
  if (user.isAdmin) {
    cols.push({ title: t.submission.colAccount, key: 'username', width: 120 })
  }
  cols.push(
    { title: t.submission.colName, key: 'name', width: 120 },
    {
      title: t.submission.colProblem, key: 'problem_id', width: 80,
      render: (r: any) =>
        h('span', {
          class: 'cursor-pointer text-green-400',
          onClick: (e: Event) => { e.stopPropagation(); router.push(`/problems/${r.problem_id}`) },
        }, `#${r.problem_id}`),
    },
    {
      title: t.submission.colResult, key: 'verdict', width: 150,
      render: (r: any) => {
        const children: any[] = [
          h(NTag, { type: verdictType(r.verdict), size: 'small' }, { default: () => r.verdict }),
        ]
        if (r.ai_pending) {
          children.push(h(NSpin, { size: 14, stroke: '#63e2b7' }))
        } else if (r.ai_rejected) {
          children.push(h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => t.submission.aiRejectedTag }))
        } else if (r.has_ai_explanation) {
          children.push(h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => t.submission.aiTag }))
        }
        return h(NSpace, { size: 4, align: 'center' }, { default: () => children })
      },
    },
    { title: t.submission.colLang, key: 'language', width: 80 },
    { title: t.submission.colTime, key: 'time_used_ms', width: 80, render: (r: any) => `${r.time_used_ms} ms` },
    { title: t.submission.colMemory, key: 'memory_used_kb', width: 100, render: (r: any) => `${r.memory_used_kb} KB` },
    { title: t.submission.colCreatedAt, key: 'created_at', width: 170 },
  )
  return cols
})

const rowProps = (row: any) => ({
  style: canView(row) ? 'cursor: pointer;' : 'cursor: default;',
  onClick: () => {
    if (!canView(row)) return
    gotoDetail(row.id)
  },
})
</script>

<template>
  <div>
    <NSpace v-if="showFilters" class="mb-3">
      <NSelect v-model:value="verdict" :options="verdictOptions" class="w-40" :placeholder="t.submission.filterVerdict" />
      <NSelect v-model:value="language" :options="langOptions" class="w-32" :placeholder="t.submission.filterLang" />
      <NInput v-model:value="userFilter" :placeholder="t.submission.filterUser" class="w-40" clearable @keyup.enter="page = 1; load()" />
      <NInput v-model:value="problemFilter" :placeholder="t.submission.filterProblemId" class="w-32" clearable @keyup.enter="page = 1; load()" />
      <NButton type="primary" @click="page = 1; load()">{{ t.common.filter }}</NButton>
    </NSpace>

    <NDataTable
      :columns="columns"
      :data="items"
      :row-props="rowProps"
      :pagination="{
        page, pageSize, itemCount: total, showSizePicker: false,
        onChange: (p: number) => { page = p; load() },
      }"
      remote
      striped
    />
  </div>
</template>
