<script setup lang="ts">
import { NDataTable, NTag, NInput, NSelect, NSpace, NButton } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { h, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { statusLabel, statusTagType } from '../api/verdict'
import { t } from '../i18n'

interface ProblemRow {
  id: number
  title: string
  difficulty: string
  my_status: string
  ac_users: number
  tag_names?: string[]
  restricted_idea?: boolean
  restricted_solution?: boolean
  restricted_ai?: boolean
}

const router = useRouter()
const items = ref<ProblemRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(16)
const q = ref('')
const difficulty = ref<string>('')
const tagId = ref<string>('')
const tagOptions = ref<SelectOption[]>([{ label: t.problem.tagAll, value: '' }])
let off: (() => void) | null = null

const load = async () => {
  const { data } = await http.get('/problems', {
    params: {
      q: q.value || undefined,
      difficulty: difficulty.value || undefined,
      tag_id: tagId.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    },
  })
  items.value = data.items
  total.value = data.total
}

onMounted(async () => {
  const { data: tagsData } = await http.get('/tags')
  const opts: SelectOption[] = [{ label: t.problem.tagAll, value: '' }]
  for (const g of tagsData.groups || []) {
    for (const tg of g.tags || []) {
      opts.push({ label: `${g.name} / ${tg.name}`, value: String(tg.id) })
    }
  }
  tagOptions.value = opts
  await load()
  // admin 侧改题目/改题单规则/踢人 加人 都要反映到列表——粗粒度：任一事件
  // 到达就 reload 当前页，受影响的 chip/状态自然同步。
  // submission:done 也要订阅——学生自己 AC 后 my_status 列应即时变绿。
  off = onEvent((ev) => {
    if (
      ev.type === 'problem:changed'
      || ev.type === 'problemset:changed'
      || ev.type === 'problemset:members:changed'
      || ev.type === 'submission:done'
    ) {
      load()
    }
  })
})
onUnmounted(() => { off?.() })

const difficultyOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  { label: t.problem.diffEntry, value: t.problem.diffEntry },
  { label: t.problem.diffEasy, value: t.problem.diffEasy },
  { label: t.problem.diffMedium, value: t.problem.diffMedium },
  { label: t.problem.diffHard, value: t.problem.diffHard },
]

const columns = [
  {
    title: t.problem.status,
    key: 'my_status',
    width: 110,
    render: (r: ProblemRow) => {
      if (!r.my_status) return h('span', { class: 'opacity-40' }, t.common.empty)
      return h(NTag, { type: statusTagType(r.my_status), size: 'small' }, { default: () => statusLabel(r.my_status) })
    },
  },
  { title: t.problem.id, key: 'id', width: 70 },
  {
    title: t.problem.title,
    key: 'title',
    render: (r: ProblemRow) => h('span', {}, r.title),
  },
  {
    title: '',
    key: 'restriction',
    width: 104,
    render: (r: ProblemRow) => {
      if (r.restricted_idea || r.restricted_solution || r.restricted_ai) {
        return h(NTag, { size: 'small', type: 'warning' }, { default: () => t.problem.restrictedByProblemsetTag })
      }
      return null
    },
  },
  {
    title: t.problem.tag,
    key: 'tag_names',
    render: (r: ProblemRow) =>
      (r.tag_names || []).map((n) => h(NTag, { size: 'small', class: 'mr-1' }, { default: () => n })),
  },
  {
    title: t.problem.difficulty,
    key: 'difficulty',
    width: 88,
  },
  {
    title: t.problem.acUsers,
    key: 'ac_users',
    width: 100,
  },
]

const rowProps = (row: ProblemRow) => ({
  style: 'cursor: pointer;',
  onClick: () => router.push(`/problems/${row.id}`),
})
</script>

<template>
  <div>
    <NSpace class="mb-4">
      <NInput v-model:value="q" :placeholder="t.problem.searchPlaceholder" clearable @keyup.enter="load" />
      <NSelect v-model:value="difficulty" :options="difficultyOptions" :placeholder="t.problem.difficulty" class="w-32" />
      <NSelect v-model:value="tagId" :options="tagOptions" :placeholder="t.problem.tag" class="w-48" filterable />
      <NButton type="primary" @click="page = 1; load()">{{ t.common.search }}</NButton>
    </NSpace>
    <NDataTable
      class="problem-list-table"
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

<style scoped>
.problem-list-table :deep(table) {
  table-layout: auto;
}

.problem-list-table :deep(thead th:nth-child(3)),
.problem-list-table :deep(tbody td:nth-child(3)) {
  width: 1%;
  padding-right: 8px;
}

.problem-list-table :deep(thead th:nth-child(4)),
.problem-list-table :deep(tbody td:nth-child(4)) {
  width: 1%;
  padding-left: 0;
  padding-right: 14px;
  white-space: nowrap;
}

.problem-list-table :deep(thead th:nth-child(5)),
.problem-list-table :deep(tbody td:nth-child(5)) {
  padding-left: 24px;
}
</style>
