<script setup lang="ts">
import { NCard, NGrid, NGridItem, NStatistic, NButton, NSpace, NDataTable, NModal, NInput, useMessage } from 'naive-ui'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../../api/http'
import MarkdownEditor from '../../components/MarkdownEditor.vue'
import { t } from '../../i18n'
import { useUserStore } from '../../stores/user'

const router = useRouter()
const route = useRoute()
const msg = useMessage()
const user = useUserStore()
const stats = ref({ problems: 0, problemsets: 0, users: 0, submissions: 0, online_users: 0 })
const online = ref<any[]>([])
const showReset = ref(false)
const resetPassword = ref('')
const resetSubmitting = ref(false)

// 轮询周期。Dashboard 只保留概览 + 在线用户的轻量刷新，不再探测系统状态和
// go-judge 版本，避免后台日志被低价值状态请求刷满。
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

const openReset = () => {
  resetPassword.value = ''
  showReset.value = true
}

const closeReset = () => {
  if (resetSubmitting.value) return
  showReset.value = false
  resetPassword.value = ''
}

const submitReset = async () => {
  const secondaryPassword = resetPassword.value.trim()
  if (!secondaryPassword) {
    msg.error(t.adminDashboard.resetDataPasswordRequired)
    return
  }
  resetSubmitting.value = true
  try {
    const { data } = await http.post('/admin/reset-data', { secondary_password: secondaryPassword })
    showReset.value = false
    resetPassword.value = ''
    msg.success(t.adminDashboard.resetDataOk)
    if (data?.warning) {
      msg.warning(data.warning)
    }
    user.logout()
    await router.replace('/')
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.opFailed)
  } finally {
    resetSubmitting.value = false
  }
}

let dashVisHandler: (() => void) | null = null

const load = async () => {
  const [s, o] = await Promise.all([
    http.get('/admin/stats'),
    http.get('/admin/online'),
  ])
  stats.value = s.data
  online.value = o.data.items || []
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

    <NCard class="mt-4" :title="t.adminDashboard.dangerTitle">
      <div class="text-sm opacity-75">{{ t.adminDashboard.dangerHint }}</div>
      <NSpace class="mt-3">
        <NButton type="error" @click="openReset">{{ t.adminDashboard.resetData }}</NButton>
      </NSpace>
    </NCard>

    <NModal v-model:show="showReset" preset="card" :title="t.adminDashboard.resetDataModalTitle" :style="{ width: 'min(520px, 96vw)' }" @after-leave="resetPassword = ''">
      <div class="text-sm">{{ t.adminDashboard.resetDataPasswordLabel }}</div>
      <NInput
        v-model:value="resetPassword"
        class="mt-2"
        type="password"
        show-password-on="click"
        :placeholder="t.adminDashboard.resetDataPasswordPlaceholder"
        @keyup.enter="submitReset"
      />
      <NSpace class="mt-4">
        <NButton type="error" :loading="resetSubmitting" @click="submitReset">{{ t.adminDashboard.resetData }}</NButton>
        <NButton :disabled="resetSubmitting" @click="closeReset">{{ t.common.cancel }}</NButton>
      </NSpace>
    </NModal>
  </div>
</template>
