<script setup lang="ts">
import { NSpace, NSelect, NDataTable, NTag } from 'naive-ui'
import { computed, h, onMounted, onUnmounted, ref, watch } from 'vue'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { verdictType } from '../api/verdict'
import { useUserStore } from '../stores/user'
import { t } from '../i18n'

const props = defineProps<{
  endpoint?: string
  hideScope?: boolean        // 题单内：不需要时间范围选择
  noCard?: boolean           // 保留旧 prop 兼容 ProblemsetDetail 的调用；无实际影响
}>()

const user = useUserStore()
const scope = ref<'week' | 'month' | 'year' | 'all'>('all')
const items = ref<any[]>([])
// Problems list comes back from the backend for problemset-mode only; it drives
// the per-problem A/B/C… cells.
const problems = ref<Array<{ id: number; code: string; title: string }>>([])
let off: (() => void) | null = null

const isProblemsetMode = () => !!props.endpoint
const endpoint = () => props.endpoint || '/ranking'

const load = async () => {
  const params: Record<string, string> = {}
  if (!props.hideScope) params.scope = scope.value
  const { data } = await http.get(endpoint(), { params })
  items.value = data.items || []
  problems.value = data.problems || []
}

onMounted(() => {
  load()
  // Only AC / done submissions can change rankings — but any `submission:done`
  // may alter the per-problem verdict grid (e.g. a fresh WA becomes the latest
  // non-AC), so reload on all done events in problemset mode.
  off = onEvent((ev) => {
    if (ev.type === 'submission:done') {
      if (isProblemsetMode() || ev.data?.verdict === 'AC') load()
      return
    }
    // 题单模式下：成员/题目列表变动（踢人、加题、改排列）都影响排名结构。
    if (isProblemsetMode()
      && (ev.type === 'problemset:changed' || ev.type === 'problemset:members:changed')) {
      load()
    }
  })
})
onUnmounted(() => { off?.() })
watch(scope, load)

// 学生侧统一隐藏账号列。admin 保留账号列用于定位。
const globalColumns = computed(() => {
  const cols: any[] = [
    {
      title: t.ranking.rankCol,
      key: 'rank',
      width: 60,
      render: (_: any, idx: number) => h('span', {}, `${idx + 1}`),
    },
  ]
  if (user.isAdmin) {
    cols.push({
      title: t.ranking.account, key: 'username', width: 140,
      render: (r: any) => h('span', { style: 'white-space:nowrap' }, r.username),
    })
  }
  cols.push(
    {
      title: t.ranking.name, key: 'name', width: 140,
      render: (r: any) => h('span', { style: 'white-space:nowrap' }, r.name),
    },
    { title: t.ranking.acCount, key: 'ac_count', width: 90 },
    { title: t.ranking.ak, key: 'ak', width: 80, render: (r: any) => r.ak || 0 },
    {
      title: t.ranking.acRate,
      key: 'ac_rate',
      width: 90,
      render: (r: any) => `${Math.round((r.ac_rate || 0) * 100)}%`,
    },
    {
      title: t.ranking.lastActive,
      key: 'last_active_at',
      width: 150,
      render: (r: any) => r.last_active_at
        ? String(r.last_active_at).replace('T', ' ').slice(0, 16)
        : h('span', { class: 'opacity-40' }, '-'),
    },
  )
  return cols
})

const buildProblemsetColumns = () => {
  const perProblem = problems.value.map((p) => ({
    title: p.code,
    key: `problem_${p.id}`,
    width: 64,
    render: (r: any) => {
      const v = r.results?.[p.id]
      if (!v) return h('span', { class: 'opacity-30' }, '—')
      return h(NTag, { type: verdictType(v), size: 'small' }, { default: () => v })
    },
  }))
  const head: any[] = [
    {
      title: t.ranking.rankCol,
      key: 'rank',
      width: 60,
      render: (_: any, idx: number) => h('span', {}, `${idx + 1}`),
    },
  ]
  if (user.isAdmin) {
    head.push({
      title: t.ranking.account, key: 'username', width: 140,
      render: (r: any) => h('span', { style: 'white-space:nowrap' }, r.username),
    })
  }
  head.push(
    {
      title: t.ranking.name, key: 'name', width: 140,
      render: (r: any) => h('span', { style: 'white-space:nowrap' }, r.name),
    },
    { title: t.ranking.acCount, key: 'ac_count', width: 90 },
    { title: t.ranking.penalty, key: 'penalty_min', width: 90, render: (r: any) => `${r.penalty_min || 0} min` },
  )
  return [...head, ...perProblem]
}
</script>

<template>
  <div>
    <NSpace v-if="!hideScope" class="mb-3">
      <NSelect
        :value="scope"
        @update:value="(v: 'week' | 'month' | 'year' | 'all') => (scope = v)"
        :options="[
          { label: t.ranking.scopeAll, value: 'all' },
          { label: t.ranking.scopeYear, value: 'year' },
          { label: t.ranking.scopeMonth, value: 'month' },
          { label: t.ranking.scopeWeek, value: 'week' },
        ]"
        class="w-28"
      />
    </NSpace>
    <NDataTable
      :columns="isProblemsetMode() ? buildProblemsetColumns() : globalColumns"
      :data="items"
      :pagination="{ pageSize: 16 }"
      striped
    />
  </div>
</template>
