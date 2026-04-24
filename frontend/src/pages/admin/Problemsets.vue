<script setup lang="ts">
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  NDataTable, NButton, NSpace, NModal, NForm, NFormItem, NInput,
  NScrollbar, NSelect, NDatePicker, NTag, NProgress, NCheckbox, NSwitch, NTabs, NTabPane, NPagination, useMessage, useDialog,
} from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import Sortable from 'sortablejs'
import { useRouter } from 'vue-router'
import { http } from '../../api/http'
import { t } from '../../i18n'

interface Row {
  id: number; title: string;
  password?: string; start_time?: string | null; end_time?: string | null;
  allowed_langs?: string[];
  has_password?: boolean;
  visible?: boolean;
  item_count?: number;
  my_ac_count?: number;
  top_ac_count?: number;
  top_ac_name?: string;
}

const router = useRouter()
const items = ref<Row[]>([])
const allProblems = ref<any[]>([])
const allLangs = ref<string[]>([])
const msg = useMessage()
const dialog = useDialog()
const showEdit = ref(false)
const editMode = ref<'create' | 'update'>('create')
const form = ref<any>(({
  id: 0,
  title: '',
  password: '',
  allowed_langs: [] as string[],
  start_ts: null as number | null,
  end_ts: null as number | null,
  visible: true,
  disable_idea: false,
  disable_solution: false,
  disable_ai: false,
}))

const showItems = ref(false)
const currentSetId = ref(0)
const selected = ref<number[]>([])
// 左栏搜索（按 #id / title 模糊匹配）+ 标签筛选（OR 语义，任意命中即过）。
const pickSearch = ref('')
const pickTagNames = ref<string[]>([])
const tagOptions = ref<SelectOption[]>([])
// 候选区分页：一页 10。右栏（已选）不分页——保留整屏视图便于拖拽排序。
const PICK_PAGE_SIZE = 10
const pickPage = ref(1)
// 右栏列表 DOM 引用 + Sortable 实例——替换原生 DnD，体验更稳。
const rightListRef = ref<HTMLElement | null>(null)
let sortableInstance: Sortable | null = null

// 成员管理 Modal 状态
const showMembers = ref(false)
const membersSetId = ref(0)
const members = ref<any[]>([])
const bans = ref<any[]>([])
const membersTab = ref<'members' | 'bans'>('members')

const load = async () => {
  const { data } = await http.get('/problemsets')
  items.value = data.items
  // 题目候选区需要拉全量——/problems 后端上限 100/页，超过则循环翻页，
  // 否则题目 >100 的库里选择器会"消失"一部分题目（右栏也因此找不到 title）。
  const problems: any[] = []
  for (let page = 1; page < 50; page++) {
    const { data: pd } = await http.get('/problems', { params: { page, page_size: 100 } })
    const batch = pd.items || []
    problems.push(...batch)
    if (batch.length < 100) break
  }
  const [meta, tagsRes] = await Promise.all([
    http.get('/meta'),
    http.get('/tags'),
  ])
  allProblems.value = problems
  allLangs.value = meta.data.languages || []
  // 标签字典全量加载，options 再根据"候选题目里用到过的标签"过滤，避免字典
  // 里挂了但实际没题目打过的标签污染筛选 UI。
  const usedNames = new Set<string>()
  for (const p of problems) {
    for (const tn of p.tag_names || []) usedNames.add(tn)
  }
  const opts: SelectOption[] = []
  for (const g of tagsRes.data.groups || []) {
    for (const tg of g.tags || []) {
      if (!usedNames.has(tg.name)) continue
      opts.push({ label: `${g.name} / ${tg.name}`, value: tg.name })
    }
  }
  tagOptions.value = opts
}
onMounted(load)

const openCreate = () => {
  editMode.value = 'create'
  form.value = { id: 0, title: '', password: '', allowed_langs: [], start_ts: null, end_ts: null, visible: true, disable_idea: false, disable_solution: false, disable_ai: false }
  showEdit.value = true
}
const openUpdate = (r: Row) => {
  editMode.value = 'update'
  form.value = {
    id: r.id,
    title: r.title,
    password: r.password || '',
    allowed_langs: r.allowed_langs ? [...r.allowed_langs] : [],
    start_ts: r.start_time ? new Date(r.start_time).getTime() : null,
    end_ts: r.end_time ? new Date(r.end_time).getTime() : null,
    visible: r.visible !== false,
    disable_idea: !!(r as any).disable_idea,
    disable_solution: !!(r as any).disable_solution,
    disable_ai: !!(r as any).disable_ai,
  }
  showEdit.value = true
}

const submit = async () => {
  if (!form.value.title.trim()) {
    msg.warning(t.problemsetAdmin.titleEmpty)
    return
  }
  const body = {
    title: form.value.title,
    password: form.value.password,
    allowed_langs: form.value.allowed_langs || [],
    start_time: form.value.start_ts ? new Date(form.value.start_ts).toISOString() : null,
    end_time: form.value.end_ts ? new Date(form.value.end_ts).toISOString() : null,
    visible: form.value.visible !== false,
    disable_idea: !!form.value.disable_idea,
    disable_solution: !!form.value.disable_solution,
    disable_ai: !!form.value.disable_ai,
  }
  if (editMode.value === 'create') {
    await http.post('/admin/problemsets', body)
  } else {
    await http.put(`/admin/problemsets/${form.value.id}`, body)
  }
  showEdit.value = false
  await load()
  msg.success(t.common.savedOk)
}

const duplicate = async (r: Row) => {
  const { data } = await http.post(`/admin/problemsets/${r.id}/copy`)
  msg.success(t.problemsetAdmin.copiedAs(data.id))
  await load()
}

const remove = (r: Row) => {
  dialog.warning({
    title: t.problemsetAdmin.confirmDeleteTitle,
    content: t.problemsetAdmin.confirmDeleteBody(r.title),
    positiveText: t.common.delete,
    negativeText: t.common.cancel,
    onPositiveClick: async () => {
      await http.delete(`/admin/problemsets/${r.id}`)
      await load()
      msg.success(t.common.deletedOk)
    },
  })
}

const openItems = async (r: Row) => {
  currentSetId.value = r.id
  const { data } = await http.get(`/problemsets/${r.id}`)
  // 后端返回的 problems 已经按 order_index ASC 排好，直接用其顺序初始化。
  selected.value = (data.problems || []).map((p: any) => p.id)
  pickSearch.value = ''
  pickTagNames.value = []
  pickPage.value = 1
  showItems.value = true
}
const saveItems = async () => {
  await http.put(`/admin/problemsets/${currentSetId.value}/problems`, { problem_ids: selected.value })
  showItems.value = false
  msg.success(t.common.updatedOk)
}

// 左栏：按搜索 + 标签过滤的可选项。标签过滤取 OR 语义——选了多个标签时，
// 只要题目命中其中一个就留下（便于"把这几类题都拉出来一起选"）。
const pickFiltered = computed(() => {
  const kw = pickSearch.value.trim().toLowerCase()
  const tagFilter = pickTagNames.value
  return allProblems.value.filter((p: any) => {
    if (kw && !(String(p.id).includes(kw) || (p.title || '').toLowerCase().includes(kw))) return false
    if (tagFilter.length > 0) {
      const names: string[] = p.tag_names || []
      if (!tagFilter.some((tn: string) => names.includes(tn))) return false
    }
    return true
  })
})
// 分页切片：筛选结果变化（搜索/标签过滤）时重置到第 1 页，避免当前页越界。
watch([pickSearch, pickTagNames], () => { pickPage.value = 1 })
const pickPaged = computed(() => {
  const start = (pickPage.value - 1) * PICK_PAGE_SIZE
  return pickFiltered.value.slice(start, start + PICK_PAGE_SIZE)
})
const isPicked = (id: number) => selected.value.includes(id)
const togglePick = (id: number) => {
  if (isPicked(id)) {
    selected.value = selected.value.filter((x) => x !== id)
  } else {
    selected.value.push(id)
  }
}
// 右栏：把 id 列表还原成富数据，保持顺序。allProblems 是题目全集来源。
const pickedList = computed(() => {
  const byId = new Map<number, any>()
  for (const p of allProblems.value) byId.set(p.id, p)
  return selected.value.map((id) => byId.get(id)).filter(Boolean)
})
const removePick = (id: number) => {
  selected.value = selected.value.filter((x) => x !== id)
}

// ABCD 位置标签：A=第1题、B=第2题，拖拽换位后位置变，字母跟着位置走。
// Excel 式编码支持 26+ 题：A..Z, AA..AZ, BA..
const labelOf = (i: number) => {
  let n = i + 1
  let s = ''
  while (n > 0) {
    const r = (n - 1) % 26
    s = String.fromCharCode(65 + r) + s
    n = Math.floor((n - 1) / 26)
  }
  return s
}

// 右栏排序：Modal 打开后挂 Sortable，关闭销毁。Sortable onEnd 根据 DOM 移动
// 结果回写 selected.value——Vue re-render 会接管 DOM，两边状态保持一致。
watch(showItems, async (v) => {
  if (v) {
    await nextTick()
    if (rightListRef.value) {
      sortableInstance = Sortable.create(rightListRef.value, {
        handle: '.drag-handle',
        animation: 160,
        ghostClass: 'sortable-ghost',
        chosenClass: 'sortable-chosen',
        onEnd(ev) {
          const from = ev.oldIndex
          const to = ev.newIndex
          if (from === undefined || to === undefined || from === to) return
          const copy = [...selected.value]
          const [item] = copy.splice(from, 1)
          copy.splice(to, 0, item)
          selected.value = copy
        },
      })
    }
  } else {
    sortableInstance?.destroy()
    sortableInstance = null
  }
})
onBeforeUnmount(() => {
  sortableInstance?.destroy()
})

const openMembers = async (r: Row) => {
  membersSetId.value = r.id
  membersTab.value = 'members'
  await Promise.all([loadMembers(), loadBans()])
  showMembers.value = true
}
const loadMembers = async () => {
  const { data } = await http.get(`/admin/problemsets/${membersSetId.value}/members`)
  members.value = data.items || []
}
const loadBans = async () => {
  const { data } = await http.get(`/admin/problemsets/${membersSetId.value}/bans`)
  bans.value = data.items || []
}
const kickMember = (m: any) => {
  dialog.warning({
    title: t.problemsetAdmin.confirmKickTitle,
    content: t.problemsetAdmin.confirmKickBody(m.name || m.username),
    positiveText: t.problemsetAdmin.memberRemove,
    negativeText: t.common.cancel,
    onPositiveClick: async () => {
      await http.delete(`/admin/problemsets/${membersSetId.value}/members/${m.user_id}`)
      msg.success(t.problemsetAdmin.kickedOk)
      await Promise.all([loadMembers(), loadBans()])
    },
  })
}
const unban = async (b: any) => {
  await http.delete(`/admin/problemsets/${membersSetId.value}/bans/${b.user_id}`)
  msg.success(t.problemsetAdmin.unbannedOk)
  await loadBans()
}

// 关闭 / 开启题单：后端翻转 visible 位；学生端列表与详情据此决定是否可见。
// admin 自己不受 visible 过滤影响，所以这里刷新列表仍能看到自己刚关的题单。
const toggleVisible = async (r: Row) => {
  const { data } = await http.post(`/admin/problemsets/${r.id}/visibility`)
  msg.success(data?.visible ? t.problemsetAdmin.openedOk : t.problemsetAdmin.closedOk)
  await load()
}

const fmt = (s?: string | null) => (s ? s.replace('T', ' ').slice(0, 16) : t.common.empty)

const statusTags = (r: Row) => {
  const tags: any[] = []
  // 可见性优先展示——关闭的题单对学生不可见，admin 需要一眼能看出。
  if (r.visible === false) {
    tags.push(h(NTag, { type: 'error', size: 'small' }, { default: () => t.problemset.visibleClosed }))
  } else {
    tags.push(h(NTag, { type: 'success', size: 'small' }, { default: () => t.problemset.visibleOpen }))
  }
  if (r.has_password) {
    tags.push(h(NTag, { type: 'error', size: 'small' }, { default: () => t.problemset.pwdTag }))
  }
  if (r.allowed_langs && r.allowed_langs.length) {
    tags.push(h(
      NTag,
      { type: 'warning', size: 'small' },
      { default: () => t.problemset.langRestricted },
    ))
  }
  return h(NSpace, { size: 4 }, { default: () => tags })
}

const renderProgress = (r: Row) => {
  const total = r.item_count || 0
  const ac = r.my_ac_count || 0
  const pct = total > 0 ? Math.round((ac / total) * 100) : 0
  return h('div', { style: 'display:flex; align-items:center; gap:8px; width:160px' }, [
    h(NProgress, { type: 'line', percentage: pct, showIndicator: false, height: 8, style: 'flex:1' }),
    h('span', { style: 'font-size:12px; opacity:0.8; white-space:nowrap' }, `${ac}/${total}`),
  ])
}

const columns = [
  { title: t.problemset.colId, key: 'id', width: 70 },
  { title: t.problemset.colTitle, key: 'title' },
  { title: t.problemset.colStatus, key: 'status', width: 240, render: statusTags },
  { title: t.problemset.colProgress, key: 'progress', width: 180, render: renderProgress },
  {
    title: t.problemset.colTop, key: 'top', width: 160,
    render: (r: Row) => r.top_ac_name
      ? t.problemset.topWith(r.top_ac_name, r.top_ac_count || 0)
      : h('span', { class: 'opacity-40' }, t.common.empty),
  },
  {
    title: t.problemset.colStart, key: 'start_time', width: 140,
    render: (r: Row) => fmt(r.start_time as any),
  },
  {
    title: t.problemset.colEnd, key: 'end_time', width: 140,
    render: (r: Row) => fmt(r.end_time as any),
  },
  {
    title: t.problemsetAdmin.colOp, key: 'op', width: 380,
    render: (r: Row) =>
      h('div', { class: 'row-actions', onClick: (e: Event) => e.stopPropagation() }, [
        h(NSpace, {}, {
          default: () => [
            h(NButton, { size: 'tiny', onClick: () => openUpdate(r) }, { default: () => t.problemsetAdmin.opEdit }),
            h(NButton, { size: 'tiny', onClick: () => openItems(r) }, { default: () => t.problemsetAdmin.opProblems }),
            h(NButton, { size: 'tiny', onClick: () => openMembers(r) }, { default: () => t.problemsetAdmin.opMembers }),
            h(NButton, {
              size: 'tiny',
              type: r.visible === false ? 'primary' : 'warning',
              onClick: () => toggleVisible(r),
            }, { default: () => r.visible === false ? t.problemsetAdmin.opOpen : t.problemsetAdmin.opClose }),
            h(NButton, { size: 'tiny', onClick: () => duplicate(r) }, { default: () => t.problemsetAdmin.opCopy }),
            h(NButton, { size: 'tiny', type: 'error', onClick: () => remove(r) }, { default: () => t.problemsetAdmin.opDelete }),
          ],
        }),
      ]),
  },
]

const rowProps = (row: Row) => ({
  style: 'cursor: pointer;',
  onClick: (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('.row-actions')) return
    router.push(`/problemsets/${row.id}`)
  },
})
</script>

<template>
  <div>
    <NSpace class="mb-3">
      <NButton type="primary" @click="openCreate">{{ t.problemsetAdmin.listNew }}</NButton>
    </NSpace>
    <NDataTable :columns="columns" :data="items" :row-props="rowProps" :pagination="{ pageSize: 16 }" striped />

    <NModal v-model:show="showEdit" preset="card" :title="editMode === 'create' ? t.problemsetAdmin.modalCreateTitle : t.problemsetAdmin.modalUpdateTitle" :style="{ width: 'min(560px, 96vw)' }">
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="t.problemsetAdmin.formTitle">
          <NInput v-model:value="form.title" />
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formStart">
          <NDatePicker v-model:value="form.start_ts" type="datetime" clearable style="width: 100%" />
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formEnd">
          <!-- hint 紧贴结束时间下方，视觉上归属时间字段；挪出 FormItem 外会被
               下一条 formPwd 吸收，误导为密码的说明。 -->
          <div style="width: 100%">
            <NDatePicker v-model:value="form.end_ts" type="datetime" clearable style="width: 100%" />
            <div class="text-xs opacity-60 mt-1">{{ t.problemsetAdmin.formTsHint }}</div>
          </div>
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formPwd">
          <NInput v-model:value="form.password" :placeholder="t.problemsetAdmin.formPwdPlaceholder" />
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formAllowedLangs">
          <NSelect
            v-model:value="form.allowed_langs"
            multiple
            :options="allLangs.map((l) => ({ label: l, value: l }))"
            :placeholder="t.problemsetAdmin.formAllowedLangsPlaceholder"
          />
        </NFormItem>
        <!-- "对学生开放"开关已取消：跟列表行的"开启/关闭"按钮功能重复。
             新建题单默认 visible=true；后续关闭/开启走行按钮。 -->
        <NFormItem :label="t.problemsetAdmin.formDisableIdea">
          <NSwitch v-model:value="form.disable_idea" />
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formDisableSolution">
          <NSwitch v-model:value="form.disable_solution" />
        </NFormItem>
        <NFormItem :label="t.problemsetAdmin.formDisableAI">
          <NSwitch v-model:value="form.disable_ai" />
        </NFormItem>
        <NButton type="primary" @click="submit">{{ t.common.save }}</NButton>
      </NForm>
    </NModal>

    <NModal v-model:show="showItems" preset="card" :title="t.problemsetAdmin.pickProblems" :style="{ width: 'min(860px, 96vw)' }">
      <div class="picker">
        <div class="col">
          <div class="col-title">{{ t.problemsetAdmin.pickerAvailable(pickFiltered.length) }}</div>
          <NInput
            v-model:value="pickSearch"
            :placeholder="t.problemsetAdmin.pickerSearchPlaceholder"
            clearable
            class="mb-2"
          />
          <NSelect
            v-model:value="pickTagNames"
            multiple
            filterable
            :options="tagOptions"
            :placeholder="t.problemsetAdmin.pickerTagFilter"
            clearable
            class="mb-2"
          />
          <NScrollbar class="list-scroll">
            <div class="list">
              <label
                v-for="p in pickPaged"
                :key="p.id"
                class="row"
              >
                <NCheckbox :checked="isPicked(p.id)" @update:checked="togglePick(p.id)" />
                <span class="row-text">#{{ p.id }} {{ p.title }}</span>
              </label>
              <div v-if="!pickFiltered.length" class="opacity-60 text-sm p-2">{{ t.common.empty }}</div>
            </div>
          </NScrollbar>
          <NPagination
            v-if="pickFiltered.length > PICK_PAGE_SIZE"
            v-model:page="pickPage"
            :page-count="Math.ceil(pickFiltered.length / PICK_PAGE_SIZE)"
            size="small"
            class="mt-2"
          />
        </div>
        <div class="col">
          <div class="col-title">{{ t.problemsetAdmin.pickerSelected(pickedList.length) }}</div>
          <div class="opacity-60 text-xs mb-2">{{ t.problemsetAdmin.pickerDragHint }}</div>
          <NScrollbar class="list-scroll">
            <div ref="rightListRef" class="list sortable-list">
              <div
                v-for="(p, idx) in pickedList"
                :key="p.id"
                class="row ordered"
                :data-id="p.id"
              >
                <span class="drag-handle" aria-hidden="true">☰</span>
                <span class="row-index">{{ labelOf(idx) }}</span>
                <span class="row-text">#{{ p.id }} {{ p.title }}</span>
                <NButton size="tiny" text type="error" @click="removePick(p.id)">×</NButton>
              </div>
              <div v-if="!pickedList.length" class="opacity-60 text-sm p-2">{{ t.problemsetAdmin.pickerEmptySelected }}</div>
            </div>
          </NScrollbar>
        </div>
      </div>
      <NSpace class="mt-3">
        <NButton type="primary" @click="saveItems">{{ t.common.save }}</NButton>
      </NSpace>
    </NModal>

    <NModal v-model:show="showMembers" preset="card" :title="t.problemsetAdmin.membersTitle" :style="{ width: 'min(640px, 96vw)' }">
      <!-- 固定 min-height 避免切 tab 时 Modal 高度跳变；用 NScrollbar 替代原生
           overflow:auto，保持和全站一致的样式。 -->
      <NTabs v-model:value="membersTab" type="line" style="min-height: 480px">
        <NTabPane name="members" :tab="`${t.problemsetAdmin.tabMembers} (${members.length})`">
          <div class="opacity-60 text-xs mb-2">{{ t.problemsetAdmin.memberRemoveHint }}</div>
          <div v-if="!members.length" class="opacity-60 text-sm">{{ t.problemsetAdmin.noMembers }}</div>
          <NScrollbar v-else style="max-height: 420px">
            <div class="members-list">
              <div v-for="m in members" :key="m.user_id" class="members-row">
                <div>
                  <span class="font-medium">{{ m.name || m.username }}</span>
                  <span class="ml-2 opacity-60 text-xs">{{ m.username }}</span>
                  <span class="ml-2 opacity-60 text-xs">{{ (m.joined_at || '').replace('T', ' ').slice(0, 16) }}</span>
                </div>
                <NButton size="tiny" type="error" @click="kickMember(m)">
                  {{ t.problemsetAdmin.memberRemove }}
                </NButton>
              </div>
            </div>
          </NScrollbar>
        </NTabPane>
        <NTabPane name="bans" :tab="`${t.problemsetAdmin.tabBans} (${bans.length})`">
          <div v-if="!bans.length" class="opacity-60 text-sm">{{ t.problemsetAdmin.noBans }}</div>
          <NScrollbar v-else style="max-height: 420px">
            <div class="members-list">
              <div v-for="b in bans" :key="b.user_id" class="members-row">
                <div>
                  <span class="font-medium">{{ b.name || b.username }}</span>
                  <span class="ml-2 opacity-60 text-xs">{{ b.username }}</span>
                  <span class="ml-2 opacity-60 text-xs">{{ (b.banned_at || '').replace('T', ' ').slice(0, 16) }}</span>
                </div>
                <NButton size="tiny" type="warning" @click="unban(b)">
                  {{ t.problemsetAdmin.banUnban }}
                </NButton>
              </div>
            </div>
          </NScrollbar>
        </NTabPane>
      </NTabs>
    </NModal>
  </div>
</template>

<style scoped>
.members-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.members-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border: 1px solid var(--lo-subtle-border);
  border-radius: 6px;
}
.picker {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.picker .col {
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.picker .col-title {
  font-size: 13px;
  opacity: 0.75;
  margin-bottom: 6px;
}
.picker .list {
  border: 1px solid var(--lo-subtle-border);
  border-radius: 6px;
  padding: 4px;
}
.picker .list-scroll {
  max-height: 420px;
}
.picker .row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  user-select: none;
}
.picker .row:hover {
  background: var(--lo-subtle-bg);
}
.picker .row.ordered {
  cursor: grab;
}
.picker .row.ordered:active {
  cursor: grabbing;
  background: var(--lo-accent-bg-weak);
}
.picker .row-index {
  opacity: 0.7;
  min-width: 28px;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--lo-accent-bg);
  color: var(--lo-accent-fg);
}
.picker .row-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.picker .drag-handle {
  opacity: 0.5;
  cursor: grab;
}
/* Sortable 拖动态的视觉反馈：源行半透明、当前选中行轻微高亮。 */
.picker .sortable-ghost {
  opacity: 0.35;
  background: var(--lo-accent-bg-strong) !important;
}
.picker .sortable-chosen {
  background: var(--lo-accent-bg-weak);
}
</style>
