<script setup lang="ts">
import { NDataTable, NTag, NSpace, NProgress, NButton, NModal, NInput, useMessage } from 'naive-ui'
import { h, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { t } from '../i18n'

const router = useRouter()
const items = ref<any[]>([])
const msg = useMessage()
let off: (() => void) | null = null

// Join dialog state — 当用户点击带密码题单的"加入"按钮时弹出。
const showJoin = ref(false)
const joinRow = ref<any>(null)
const joinPwd = ref('')

const load = async () => {
  const { data } = await http.get('/problemsets')
  items.value = data.items
}

onMounted(() => {
  load()
  off = onEvent((ev) => {
    if (ev.type === 'submission:done' && ev.data?.verdict === 'AC') load()
    // admin 改题单（新增/改规则/可见性/复制/删除）或成员变动 → 列表重拉。
    if (ev.type === 'problemset:changed' || ev.type === 'problemset:members:changed') load()
  })
})
onUnmounted(() => { off?.() })

const fmt = (s?: string | null) => (s ? s.replace('T', ' ').slice(0, 16) : t.common.empty)

const statusTags = (r: any) => {
  const tags: any[] = []
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
  return tags.length
    ? h(NSpace, { size: 4 }, { default: () => tags })
    : h('span', { class: 'opacity-40' }, t.common.empty)
}

const renderProgress = (r: any) => {
  const total = r.item_count || 0
  const ac = r.my_ac_count || 0
  const pct = total > 0 ? Math.round((ac / total) * 100) : 0
  return h('div', { style: 'display:flex; align-items:center; gap:8px; width:160px' }, [
    h(NProgress, { type: 'line', percentage: pct, showIndicator: false, height: 8, style: 'flex:1' }),
    h('span', { style: 'font-size:12px; opacity:0.8; white-space:nowrap' }, `${ac}/${total}`),
  ])
}

const openJoin = (r: any) => {
  joinRow.value = r
  joinPwd.value = ''
  if (!r.has_password) {
    // 无密码题单直接加入，不弹框
    doJoin(r, '')
    return
  }
  showJoin.value = true
}

const doJoin = async (r: any, pwd: string) => {
  try {
    await http.post(`/problemsets/${r.id}/join`, pwd ? { password: pwd } : {})
    msg.success(t.problemset.joinOk)
    showJoin.value = false
    await load()
  } catch (e: any) {
    const code = e?.response?.data?.error
    if (code === 'banned') {
      msg.error(t.problemset.bannedLocked)
      showJoin.value = false
    } else if (code === 'password_incorrect' || e?.response?.status === 403) {
      msg.error(t.problemset.pwdWrong)
    } else {
      msg.error(code || t.problemset.joinFailed)
    }
  }
}

const renderJoin = (r: any) => {
  if (r.is_banned) {
    return h(NTag, { type: 'error', size: 'small' }, { default: () => t.problemset.bannedTag })
  }
  if (r.is_member) {
    return h(NTag, { type: 'success', size: 'small' }, { default: () => t.problemset.joined })
  }
  return h(NButton, {
    size: 'small',
    type: 'primary',
    onClick: (e: MouseEvent) => { e.stopPropagation(); openJoin(r) },
  }, { default: () => t.problemset.joinBtn })
}

const columns = [
  { title: t.problemset.colId, key: 'id', width: 70 },
  { title: t.problemset.colTitle, key: 'title' },
  { title: t.problemset.colStatus, key: 'status', width: 240, render: statusTags },
  { title: t.problemset.colProgress, key: 'progress', width: 180, render: renderProgress },
  {
    title: t.problemset.colTop, key: 'top', width: 160,
    render: (r: any) => r.top_ac_name
      ? h('span', { style: 'white-space:nowrap' }, t.problemset.topWith(r.top_ac_name, r.top_ac_count))
      : h('span', { class: 'opacity-40' }, t.common.empty),
  },
  {
    title: t.problemset.colStart, key: 'start_time', width: 140,
    render: (r: any) => h('span', { style: 'white-space:nowrap' }, fmt(r.start_time)),
  },
  {
    title: t.problemset.colEnd, key: 'end_time', width: 140,
    render: (r: any) => h('span', { style: 'white-space:nowrap' }, fmt(r.end_time)),
  },
  { title: t.problemset.colJoin, key: 'join', width: 100, render: renderJoin },
]

const rowProps = (row: any) => ({
  style: 'cursor: pointer;',
  onClick: (e: MouseEvent) => {
    // 点到"加入"按钮时不跳转详情
    const target = e.target as HTMLElement
    if (target.closest('.n-button')) return
    router.push(`/problemsets/${row.id}`)
  },
})
</script>

<template>
  <div>
    <NDataTable :columns="columns" :data="items" :row-props="rowProps" :pagination="{ pageSize: 16 }" striped />
    <NModal v-model:show="showJoin" preset="card" :title="t.problemset.joinTitle" :style="{ width: 'min(400px, 94vw)' }">
      <div class="mb-2 text-sm opacity-70">{{ t.problemset.joinPwdHint }}</div>
      <NInput v-model:value="joinPwd" type="password" show-password-on="click" :placeholder="t.problemset.pwdInputPlaceholder" @keyup.enter="doJoin(joinRow, joinPwd)" />
      <NSpace class="mt-3">
        <NButton type="primary" @click="doJoin(joinRow, joinPwd)">{{ t.problemset.joinBtn }}</NButton>
      </NSpace>
    </NModal>
  </div>
</template>
