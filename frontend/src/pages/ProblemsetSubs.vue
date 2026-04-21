<script setup lang="ts">
import { NDataTable, NTag, NSpace, NButton, NSpin } from 'naive-ui'
import { h, onMounted, onUnmounted, ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../api/http'
import { verdictType } from '../api/verdict'
import { onEvent } from '../api/events'
import { useUserStore } from '../stores/user'
import { t } from '../i18n'

// Single-purpose submissions table scoped to one problemset. Reuses the same
// column layout as MySubmissions.vue so the look-and-feel stays consistent
// between 题单 Tab 和全局 提交列表.
const props = defineProps<{ problemsetId: number }>()
const router = useRouter()
const user = useUserStore()
const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
let off: (() => void) | null = null

const load = async () => {
  const { data } = await http.get('/submissions', {
    params: {
      problemset_id: props.problemsetId,
      page: page.value,
      page_size: 16,
    },
  })
  items.value = data.items
  total.value = data.total
}

onMounted(() => {
  load()
  off = onEvent((ev) => {
    if (ev.type === 'submission:new' || ev.type === 'submission:done') load()
    else if (ev.type === 'ai:task:done') load()
    // 踢人/加人（属于本题单）→ 提交记录可能刚被清空或新成员刚产生提交，
    // 直接重拉最稳。
    else if (ev.type === 'problemset:members:changed' && ev.data?.id === props.problemsetId) load()
  })
})
onUnmounted(() => { off?.() })
watch(() => props.problemsetId, () => { page.value = 1; load() })

// 学生侧隐藏账号列；进入详情带 from=ps query，详情页据此展示"返回题单"按钮。
// 他人提交对非 admin 不可点——后端 Detail 本身会 403，避免误点。
// AI 任务在跑时本行整行禁用，让 SSE 回来后再自然解锁。
const canView = (r: any) => (user.isAdmin || r.user_id === user.user?.id) && !r.ai_pending
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
    {
      title: t.submission.colOp, key: 'op', width: 110,
      render: (r: any) => {
        if (!canView(r)) {
          return h('span', { class: 'opacity-40 text-xs' }, r.ai_pending ? t.submission.aiPendingHint : t.common.empty)
        }
        return h(NSpace, { size: 4 }, {
          default: () => [
            h(NButton, { size: 'tiny', onClick: () => router.push({ path: `/submissions/${r.id}`, query: { from: 'ps' } }) },
              { default: () => t.submission.opDetail }),
          ],
        })
      },
    },
  )
  return cols
})

const rowProps = (row: any) => ({
  style: canView(row) ? 'cursor: pointer;' : 'cursor: default;',
  onClick: (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('.n-button')) return
    if (!canView(row)) return
    router.push({ path: `/submissions/${row.id}`, query: { from: 'ps' } })
  },
})
</script>

<template>
  <NDataTable
    :columns="columns"
    :data="items"
    :row-props="rowProps"
    :pagination="{
      page, pageSize: 16, itemCount: total, showSizePicker: false,
      onChange: (p: number) => { page = p; load() },
    }"
    remote
    striped
  />
</template>
