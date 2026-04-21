<script setup lang="ts">
import { NCard, NGrid, NGridItem, NStatistic, NButton, NSpace, NDataTable, NTag, useMessage } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../../api/http'
import MarkdownEditor from '../../components/MarkdownEditor.vue'
import { t } from '../../i18n'

const router = useRouter()
const route = useRoute()
const msg = useMessage()
const stats = ref({ problems: 0, problemsets: 0, users: 0, submissions: 0, online_users: 0 })
const online = ref<any[]>([])
const sys = ref<any>(null)
// `judge` is the raw latest /status response. `judgeView` is what the UI
// actually renders — we only switch it to "offline" after two consecutive
// misses so a single timeout doesn't make the badge flap red/green every
// poll cycle (happens often when go-judge sits behind Windows netsh portproxy).
const judge = ref<any>(null)
const judgeView = ref<any>(null)
let judgeMissStreak = 0

// 轮询周期。Dashboard 里这三条请求占了后端日志的绝大多数噪音，30s 足以用
// 作"大概看一眼"的粒度，真正需要实时状态时手动刷新页面即可。
const POLL_MS = 30000

// 用模块级单例 timer 而非组件级 let——Vue 理论上不会双挂 Dashboard，但线上
// 见过热更新残留/快速切换路由导致定时器并存的现象。模块级单例保证"新来的
// 挂载先把上一份 kill 掉再装自己的"，而不是两份并行。如果你看到后端日志里
// 每秒好几条 /admin/stats，一般意味着浏览器还有一个打开的 /admin 标签页 →
// 关掉其他标签或 Ctrl+F5；这段代码只能管好当前标签。
let dashTimer: number | null = null

// 首页 markdown 编辑态。load 时 GET /home 填充 homeContent；保存走 PUT
// /admin/home，成功后后端 broadcast home:changed，学生端立即刷新。
const homeContent = ref('')
const homeSaving = ref(false)
const loadHome = async () => {
  const { data } = await http.get('/home')
  homeContent.value = data.content || ''
}
const saveHome = async () => {
  homeSaving.value = true
  try {
    await http.put('/admin/home', { content: homeContent.value })
    msg.success(t.common.savedOk)
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.saveFailed)
  } finally {
    homeSaving.value = false
  }
}
let dashVisHandler: (() => void) | null = null

const load = async () => {
  const [s, o, st] = await Promise.all([
    http.get('/admin/stats'),
    http.get('/admin/online'),
    http.get('/admin/system/status'),
  ])
  stats.value = s.data
  online.value = o.data.items || []
  sys.value = st.data.system
  judge.value = st.data.judge
  if (st.data.judge?.reachable) {
    judgeView.value = st.data.judge
    judgeMissStreak = 0
  } else {
    judgeMissStreak++
    // On first load we have nothing to show, so adopt the miss immediately.
    // Otherwise wait for 2 consecutive misses before surfacing "offline".
    if (judgeView.value === null || judgeMissStreak >= 2) {
      judgeView.value = st.data.judge
    }
  }
}

const startTimer = () => {
  if (dashTimer !== null) return
  if (route.path !== '/admin') return
  if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return
  dashTimer = window.setInterval(load, POLL_MS)
}
const stopTimer = () => {
  if (dashTimer !== null) {
    clearInterval(dashTimer)
    dashTimer = null
  }
}

onMounted(async () => {
  // 防御：任何残留的 timer（理论上 onUnmounted 已清，但热更新 / 异常路径
  // 可能留着）在进新实例前先一律杀掉。
  stopTimer()
  if (dashVisHandler) {
    document.removeEventListener('visibilitychange', dashVisHandler)
    dashVisHandler = null
  }
  await load()
  loadHome()
  startTimer()
  dashVisHandler = () => {
    if (document.visibilityState === 'hidden') stopTimer()
    else startTimer()
  }
  document.addEventListener('visibilitychange', dashVisHandler)
})
onUnmounted(() => {
  stopTimer()
  if (dashVisHandler) {
    document.removeEventListener('visibilitychange', dashVisHandler)
    dashVisHandler = null
  }
})
watch(() => route.path, (p) => {
  if (p === '/admin') startTimer()
  else stopTimer()
})

const fmt = (s: string) => (s || '').replace('T', ' ').slice(0, 19)

const uptime = computed(() => {
  const s = sys.value?.uptime_seconds || 0
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (d > 0) return `${d}d ${h}h ${m}m`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m ${s % 60}s`
})

// go-judge /version 返回 JSON，直接透传到 UI 会塞满一屏花括号；解析出几个
// 关键字段渲染更好看。parse 失败时退化为原始字符串。
const judgeVersion = computed(() => {
  const raw = judgeView.value?.version
  if (!raw) return null
  try {
    const v = JSON.parse(raw)
    if (v && typeof v === 'object') {
      return {
        build: v.buildVersion || v.version || '?',
        go: v.goVersion || '',
        platform: [v.os, v.platform].filter(Boolean).join('/'),
      }
    }
  } catch { /* not JSON, fall through */ }
  return { build: String(raw), go: '', platform: '' }
})

const columns = [
  { title: t.adminDashboard.colAccount, key: 'username', width: 140 },
  { title: t.adminDashboard.colName, key: 'name' },
  {
    title: t.adminDashboard.colRole, key: 'role', width: 100,
    render: (r: any) => r.role === 'admin' ? t.adminDashboard.roleAdmin : t.adminDashboard.roleStudent,
  },
  {
    title: t.adminDashboard.colLastSeen, key: 'last_seen_at',
    render: (r: any) => fmt(r.last_seen_at),
  },
]
</script>

<template>
  <div>
    <NCard :title="t.adminDashboard.overview">
      <NGrid :cols="5" :x-gap="12" :y-gap="12">
        <NGridItem><NStatistic :label="t.adminDashboard.problems" :value="stats.problems" /></NGridItem>
        <NGridItem><NStatistic :label="t.adminDashboard.problemsets" :value="stats.problemsets" /></NGridItem>
        <NGridItem><NStatistic :label="t.adminDashboard.users" :value="stats.users" /></NGridItem>
        <NGridItem><NStatistic :label="t.adminDashboard.totalSubs" :value="stats.submissions" /></NGridItem>
        <NGridItem>
          <NStatistic :label="t.adminDashboard.onlineUsers" :value="stats.online_users" />
        </NGridItem>
      </NGrid>

      <NSpace class="mt-6">
        <NButton type="primary" @click="router.push('/admin/users')">{{ t.adminDashboard.quickUsers }}</NButton>
        <NButton @click="router.push('/admin/tags')">{{ t.adminDashboard.quickTags }}</NButton>
        <NButton @click="router.push('/admin/problems')">{{ t.adminDashboard.quickProblems }}</NButton>
        <NButton @click="router.push('/admin/problemsets')">{{ t.adminDashboard.quickProblemsets }}</NButton>
        <NButton @click="router.push('/admin/submissions')">{{ t.adminDashboard.quickSubmissions }}</NButton>
        <NButton @click="router.push('/admin/ai')">{{ t.adminDashboard.quickAi }}</NButton>
      </NSpace>
    </NCard>

    <div class="status-row">
      <NCard :title="t.adminDashboard.systemStatus" class="card-fill">
        <div v-if="sys" class="status-grid">
          <div><span class="k">Go</span><span class="v">{{ sys.go_version }}</span></div>
          <div><span class="k">{{ t.adminDashboard.platform }}</span><span class="v">{{ sys.os }}/{{ sys.arch }} · {{ sys.num_cpu }} CPU</span></div>
          <div><span class="k">{{ t.adminDashboard.goroutines }}</span><span class="v">{{ sys.goroutines }}</span></div>
          <div><span class="k">{{ t.adminDashboard.memory }}</span><span class="v">{{ sys.alloc_mb }} MB / {{ sys.sys_mb }} MB</span></div>
          <div><span class="k">{{ t.adminDashboard.uptime }}</span><span class="v">{{ uptime }}</span></div>
          <div><span class="k">{{ t.adminDashboard.startedAt }}</span><span class="v">{{ fmt(sys.started_at) }}</span></div>
        </div>
        <div v-else class="opacity-60 text-sm">{{ t.common.loadingDots }}</div>
      </NCard>

      <NCard :title="t.adminDashboard.judgeStatus" class="card-fill">
        <div v-if="judgeView" class="status-grid">
          <div>
            <span class="k">{{ t.adminDashboard.reachable }}</span>
            <span class="v">
              <NTag v-if="judgeView.reachable" type="success" size="small">{{ t.adminDashboard.online }}</NTag>
              <NTag v-else type="error" size="small">{{ t.adminDashboard.offline }}</NTag>
            </span>
          </div>
          <div><span class="k">{{ t.adminDashboard.addr }}</span><span class="v mono">{{ judgeView.base_url }}</span></div>
          <div v-if="judgeView.reachable && judgeVersion">
            <span class="k">{{ t.adminDashboard.version }}</span>
            <span class="v mono">
              {{ judgeVersion.build }}
              <span v-if="judgeVersion.platform" class="opacity-60"> · {{ judgeVersion.platform }}</span>
              <span v-if="judgeVersion.go" class="opacity-60"> · {{ judgeVersion.go }}</span>
            </span>
          </div>
          <div v-else-if="!judgeView.reachable">
            <span class="k">{{ t.adminDashboard.error }}</span>
            <span class="v mono opacity-70">{{ judgeView.error || t.common.empty }}</span>
          </div>
          <div><span class="k">{{ t.adminDashboard.queue }}</span><span class="v">{{ judgeView.queue_len }} / {{ judgeView.queue_cap }}</span></div>
          <div><span class="k">{{ t.adminDashboard.workers }}</span><span class="v">{{ judgeView.workers }}</span></div>
        </div>
        <div v-else class="opacity-60 text-sm">{{ t.common.loadingDots }}</div>
      </NCard>
    </div>

    <NCard class="mt-4" :title="t.adminDashboard.onlineTitle(online.length)">
      <div v-if="!online.length" class="opacity-60 text-sm">{{ t.adminDashboard.noOnline }}</div>
      <NDataTable v-else :columns="columns" :data="online" :pagination="{ pageSize: 10 }" />
    </NCard>

    <NCard class="mt-4" :title="t.adminDashboard.homeEditTitle">
      <MarkdownEditor v-model="homeContent" height="420px" />
      <NSpace class="mt-3">
        <NButton type="primary" :loading="homeSaving" @click="saveHome">{{ t.common.save }}</NButton>
      </NSpace>
    </NCard>
  </div>
</template>

<style scoped>
.status-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 16px;
  align-items: stretch;
}
.card-fill {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.status-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 6px;
  font-size: 13px;
}
.status-grid > div {
  display: flex;
  align-items: center;
}
.k {
  display: inline-block;
  width: 90px;
  opacity: 0.6;
}
.v {
  flex: 1;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
}
</style>
