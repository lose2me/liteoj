<script setup lang="ts">
import { h, onMounted, onUnmounted, ref } from 'vue'
import {
  NDataTable, NButton, NSpace, NPopconfirm, NInput, NSelect, NSpin, NTag, useMessage,
} from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { useRouter } from 'vue-router'
import { http } from '../../api/http'
import { onEvent } from '../../api/events'
import { t } from '../../i18n'

const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 16
const q = ref('')
const tagId = ref<string>('')
const tagOptions = ref<SelectOption[]>([{ label: t.problem.tagAll, value: '' }])
const msg = useMessage()
const router = useRouter()

let off: (() => void) | null = null

const load = async () => {
  const { data } = await http.get('/problems', {
    params: {
      q: q.value || undefined,
      tag_id: tagId.value || undefined,
      page: page.value,
      page_size: pageSize,
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
  // AI 任务完成/状态变化时刷新，让正在转圈的行及时解锁并显示最新字段。
  off = onEvent((ev) => {
    if (ev.type === 'ai:task:done' || ev.type === 'ai:tasks:changed') load()
    // admin 改题目（可能是其他管理员或其它 tab）→ 列表刷新。
    else if (ev.type === 'problem:changed') load()
  })
})
onUnmounted(() => { off?.() })

const createBlank = async () => {
  const { data } = await http.post('/admin/problems', {
    title: t.problemAdmin.listPlaceholderTitle, description: '', time_limit_ms: 1000, memory_limit_mb: 256, visible: true,
  })
  router.push(`/admin/problems/${data.id}`)
}

const remove = async (id: number) => {
  await http.delete(`/admin/problems/${id}`)
  msg.success(t.common.deletedOk)
  await load()
}

// 有 AI 任务在跑的题目，不让进入编辑页——避免一键解析还没回写就被老字段
// 覆盖或 admin 误点重复操作。
const canEnter = (r: any) => !r.ai_pending

const columns = [
  { title: t.problem.id, key: 'id', width: 80 },
  {
    title: t.problem.title, key: 'title',
    render: (r: any) => {
      const nodes: any[] = [h('span', {}, r.title)]
      if (r.ai_pending) {
        nodes.push(h(NSpin, { size: 14, stroke: '#63e2b7' }))
        nodes.push(h(NTag, { size: 'small', type: 'warning', bordered: false }, { default: () => t.problemAdmin.aiPendingRow }))
      }
      // 后台视角：这题在任意题单里命中任一禁用开关就打一个 chip，提示
      // admin "这题在某题单里是受限状态"。合并成单个 chip 以免撑行。
      if (r.restricted_idea || r.restricted_solution || r.restricted_ai) {
        nodes.push(h(NTag, { size: 'small', type: 'warning' }, { default: () => t.problem.restrictedByProblemsetTag }))
      }
      return h(NSpace, { size: 6, align: 'center' }, { default: () => nodes })
    },
  },
  { title: t.problem.difficulty, key: 'difficulty', width: 80 },
  { title: t.problemAdmin.colVisible, key: 'visible', width: 70, render: (r: any) => (r.visible ? t.common.yes : t.common.no) },
  {
    title: t.problemsetAdmin.colOp,
    key: 'op',
    width: 200,
    render: (r: any) =>
      h('div', { class: 'row-actions', onClick: (e: Event) => e.stopPropagation() }, [
        h(NSpace, {}, {
          default: () => [
            h(NButton, {
              size: 'tiny',
              disabled: !canEnter(r),
              onClick: () => router.push(`/admin/problems/${r.id}`),
            }, { default: () => t.common.edit }),
            h(NPopconfirm, { onPositiveClick: () => remove(r.id) }, {
              trigger: () => h(NButton, { size: 'tiny', type: 'error', disabled: !canEnter(r) }, { default: () => t.common.delete }),
              default: () => t.problemAdmin.confirmDeleteRow(r.id),
            }),
          ],
        }),
      ]),
  },
]

const rowProps = (row: any) => ({
  style: canEnter(row) ? 'cursor: pointer;' : 'cursor: default; opacity: 0.7;',
  onClick: (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('.row-actions')) return
    if (!canEnter(row)) return
    router.push(`/admin/problems/${row.id}`)
  },
})
</script>

<template>
  <div>
    <NSpace class="mb-3">
      <NButton type="primary" @click="createBlank">{{ t.problemAdmin.listNew }}</NButton>
      <NInput v-model:value="q" :placeholder="t.problem.searchPlaceholder" clearable @keyup.enter="page = 1; load()" class="w-60" />
      <NSelect v-model:value="tagId" :options="tagOptions" :placeholder="t.problem.tag" class="w-48" filterable />
      <NButton @click="page = 1; load()">{{ t.common.search }}</NButton>
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
