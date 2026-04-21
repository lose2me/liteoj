<script setup lang="ts">
import { NCard, NDataTable, NTag, NProgress, NTabs, NTabPane, NSpace, NButton, NInput, useMessage } from 'naive-ui'
import { computed, h, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Ranking from './Ranking.vue'
import ProblemsetSubs from './ProblemsetSubs.vue'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import { statusLabel, statusTagType } from '../api/verdict'
import { t } from '../i18n'

const route = useRoute()
const router = useRouter()
const msg = useMessage()
const ps = ref<any>(null)
const problems = ref<any[]>([])
const myStatus = ref<Record<number, string>>({})
const isMember = ref(false)
const isBanned = ref(false)
const itemCount = ref(0)
const pwdInput = ref('')
let off: (() => void) | null = null

// 每秒 tick 的时间戳，驱动倒计时 computed 刷新。
const nowTs = ref(Date.now())
let tickTimer: number | null = null

// stats：成员用 problems+myStatus 计算真实进度；非成员用后端给出的 item_count
// 作为总数，AC 数固定 0，界面仍能渲染出一致的"进度条+x/y"外观。
const stats = computed(() => {
  if (isMember.value) {
    const total = problems.value.length
    const ac = problems.value.filter((p) => {
      const s = myStatus.value[p.id]
      return s === 'AC' || s === 'AC_FADED'
    }).length
    return { total, ac, pct: total ? Math.round((ac / total) * 100) : 0 }
  }
  return { total: itemCount.value, ac: 0, pct: 0 }
})

const load = async () => {
  const { data } = await http.get(`/problemsets/${route.params.id}`)
  ps.value = {
    ...data.problemset,
    has_password: !!data.has_password,
    top_ac_name: data.top_ac_name || '',
    top_ac_count: data.top_ac_count || 0,
  }
  isMember.value = !!data.is_member
  isBanned.value = !!data.is_banned || data.lock_reason === 'banned'
  problems.value = data.problems || []
  myStatus.value = data.my_status || {}
  itemCount.value = typeof data.item_count === 'number' ? data.item_count : problems.value.length
}

const joinSet = async () => {
  try {
    await http.post(`/problemsets/${route.params.id}/join`, ps.value?.has_password ? { password: pwdInput.value } : {})
    msg.success(t.problemset.joinOk)
    pwdInput.value = ''
    await load()
  } catch (e: any) {
    const code = e?.response?.data?.error
    if (code === 'banned') {
      msg.error(t.problemset.bannedLocked)
      isBanned.value = true
    } else if (code === 'password_incorrect' || e?.response?.status === 403) {
      msg.error(t.problemset.pwdWrong)
    } else {
      msg.error(code || t.problemset.joinFailed)
    }
  }
}

onMounted(() => {
  load()
  off = onEvent((ev) => {
    if (ev.type === 'submission:done' && ev.data?.verdict === 'AC') load()
    // admin 改题单规则、改题目列表、踢人加人 → 粗粒度 reload 当前页。
    // 仅当事件的 problemset id 与当前页一致时触发。
    const psid = Number(route.params.id)
    if (
      (ev.type === 'problemset:changed' || ev.type === 'problemset:members:changed')
      && ev.data?.id === psid
    ) {
      load()
    }
    // 题目本身被改（例如 admin 改了题目标题/限制），影响本题单内的题目列表。
    if (ev.type === 'problem:changed') load()
  })
  tickTimer = window.setInterval(() => { nowTs.value = Date.now() }, 1000)
})
onUnmounted(() => {
  off?.()
  if (tickTimer) clearInterval(tickTimer)
})

const fmt = (s?: string | null) => (s ? s.replace('T', ' ').slice(0, 16) : t.common.empty)

// 倒计时：相对于 start/end 三种状态 pending/running/ended。仅在 end_time
// 存在时渲染——没有截止时间就不是"比赛"题单，不需要倒计时干扰。
const countdown = computed(() => {
  if (!ps.value?.end_time) return null
  const end = new Date(ps.value.end_time).getTime()
  const start = ps.value.start_time ? new Date(ps.value.start_time).getTime() : 0
  if (start && nowTs.value < start) return { type: 'info' as const, label: t.problemset.cdPending, ms: start - nowTs.value }
  if (nowTs.value < end) return { type: 'success' as const, label: t.problemset.cdRunning, ms: end - nowTs.value }
  return { type: 'error' as const, label: t.problemset.cdEnded, ms: 0 }
})

const fmtCd = (ms: number) => {
  if (ms <= 0) return '0s'
  const total = Math.floor(ms / 1000)
  const d = Math.floor(total / 86400)
  const h = Math.floor((total % 86400) / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const pad = (x: number) => String(x).padStart(2, '0')
  return d > 0 ? `${d}d ${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(h)}:${pad(m)}:${pad(s)}`
}

const columns = [
  {
    title: t.problem.status,
    key: 'my_status',
    width: 110,
    render: (p: any) => {
      const s = myStatus.value[p.id]
      if (!s) return h('span', { class: 'opacity-40' }, t.common.empty)
      return h(NTag, { type: statusTagType(s), size: 'small' }, { default: () => statusLabel(s) })
    },
  },
  // Show the per-set code (A/B/C…) instead of the global problem.ID — students
  // navigating inside a set expect ICPC-style labels, not raw DB ids. Routing
  // still uses problem.id under the hood via rowProps.
  { title: t.problem.id, key: 'code', width: 70 },
  { title: t.problem.title, key: 'title' },
  { title: t.problem.difficulty, key: 'difficulty', width: 80 },
]

const rowProps = (row: any) => ({
  style: 'cursor: pointer;',
  onClick: () => router.push({
    path: `/problems/${row.id}`,
    query: { problemset: String(route.params.id) },
  }),
})

const rankingEndpoint = computed(() => `/problemsets/${route.params.id}/ranking`)

// 当前激活的 tab 绑在 URL 上。学生从"提交记录" tab 点某个提交 → 浏览器
// 历史里这一页的 URL 是 /problemsets/:id?tab=submissions；跳到详情后浏览器
// 后退会原样回到该 URL，这里读出 tab 值再激活同名 tab，实现"返回到我刚才
// 看的那个 tab"。默认 problems。
const activeTab = computed<string>({
  get: () => {
    const v = route.query.tab
    return typeof v === 'string' && v ? v : 'problems'
  },
  set: (v) => {
    router.replace({ query: { ...route.query, tab: v } })
  },
})
</script>

<template>
  <div v-if="ps">
    <NCard :title="ps.title">
      <NSpace class="mb-3">
        <NTag v-if="ps.has_password" type="warning">{{ t.problemset.pwdLock }}</NTag>
        <NTag v-if="ps.allowed_langs?.length" type="warning">
          {{ t.problemset.langRestricted }}
        </NTag>
        <NTag v-if="ps.top_ac_name" type="success">
          {{ t.problemset.crown(ps.top_ac_name, ps.top_ac_count) }}
        </NTag>
        <NTag v-if="ps.start_time">{{ t.problemset.startAt(fmt(ps.start_time)) }}</NTag>
        <NTag v-if="ps.end_time" type="warning">{{ t.problemset.endAt(fmt(ps.end_time)) }}</NTag>
        <NTag v-if="countdown" :type="countdown.type">
          {{ countdown.label }} {{ fmtCd(countdown.ms) }}
        </NTag>
      </NSpace>

      <!-- 进度条 + 总数：成员/非成员都渲染，保证标题区视觉一致；非成员
           此时 ac=0、total=item_count，条形为空但"x / y AC"仍然读得通。 -->
      <div v-if="!isBanned" class="flex items-center gap-3 mt-2">
        <NProgress type="line" :percentage="stats.pct" class="flex-1" />
        <span class="text-sm">{{ t.problemset.acOfTotal(stats.ac, stats.total) }}</span>
      </div>

      <template v-if="!isMember">
        <template v-if="isBanned">
          <div class="opacity-80 mb-2 text-red-400">{{ t.problemset.bannedLocked }}</div>
        </template>
        <template v-else>
          <!-- 两条提示文字合到一行；按钮 + 密码框另起一行。 -->
          <div class="mt-3 mb-2 text-sm">
            <span class="opacity-80">{{ t.problemset.needJoin }}</span>
            <span class="opacity-60 ml-2">{{ t.problemset.joinHintNoLeave }}</span>
          </div>
          <NSpace>
            <NInput
              v-if="ps.has_password"
              v-model:value="pwdInput"
              type="password"
              show-password-on="click"
              :placeholder="t.problemset.pwdInputPlaceholder"
              class="w-60"
              @keyup.enter="joinSet"
            />
            <NButton type="primary" @click="joinSet">{{ t.problemset.joinBtn }}</NButton>
          </NSpace>
        </template>
      </template>
    </NCard>

    <NTabs v-if="isMember" v-model:value="activeTab" class="mt-4" type="line">
      <NTabPane name="problems" :tab="t.problemset.tabProblems">
        <NDataTable :columns="columns" :data="problems" :row-props="rowProps" striped />
      </NTabPane>
      <NTabPane name="ranking" :tab="t.problemset.tabRanking">
        <div class="opacity-60 text-xs mb-2">{{ t.problemset.penaltyHint }}</div>
        <Ranking :endpoint="rankingEndpoint" :hide-scope="true" :no-card="true" />
      </NTabPane>
      <NTabPane name="submissions" :tab="t.problemset.tabSubs">
        <ProblemsetSubs :problemset-id="Number(route.params.id)" />
      </NTabPane>
    </NTabs>
  </div>
</template>
