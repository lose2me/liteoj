<script setup lang="ts">
import { NLayout, NLayoutHeader, NLayoutSider, NLayoutContent, NMenu, NButton } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import ThemeToggleButton from './ThemeToggleButton.vue'
import { t } from '../i18n'

const user = useUserStore()
const router = useRouter()
const route = useRoute()

// Plain-text labels + @update:value navigation. Wrapping RouterLink inside
// each option's label caused double-click dispatch (RouterLink + NMenu own
// handler) which left the router stuck after visiting some pages. One
// handler, one source of truth.
const menuOptions: MenuOption[] = [
  { label: t.adminDashboard.overview, key: '/admin' },
  { label: t.adminDashboard.quickUsers, key: '/admin/users' },
  { label: t.adminDashboard.quickTags, key: '/admin/tags' },
  { label: t.adminDashboard.quickProblems, key: '/admin/problems' },
  { label: t.adminDashboard.quickProblemsets, key: '/admin/problemsets' },
  { label: t.adminDashboard.quickSubmissions, key: '/admin/submissions' },
  { label: t.adminDashboard.quickAi, key: '/admin/ai' },
]

const activeKey = computed(() => {
  const p = route.path
  if (p.startsWith('/admin/problemsets')) return '/admin/problemsets'
  if (p.startsWith('/admin/problems')) return '/admin/problems'
  if (p.startsWith('/admin/tags')) return '/admin/tags'
  if (p.startsWith('/admin/users')) return '/admin/users'
  if (p.startsWith('/admin/submissions')) return '/admin/submissions'
  if (p.startsWith('/admin/ai')) return '/admin/ai'
  return '/admin'
})

const onMenuSelect = (key: string) => {
  if (key !== route.path) router.push(key)
}

const logout = () => {
  user.logout()
  router.replace('/')
}
</script>

<template>
  <div class="admin-shell">
    <NLayoutHeader bordered class="admin-header">
      <div class="header-inner">
        <div class="text-lg font-bold mr-3">{{ t.nav.appName }}</div>
        <div class="text-sm opacity-70">{{ t.nav.adminConsole }}</div>
        <div class="flex-1" />
        <div class="header-actions">
          <ThemeToggleButton />
          <span class="user-nick text-sm opacity-80">
            {{ user.user?.name || user.user?.username }}
          </span>
          <NButton size="small" type="primary" ghost @click="router.push('/problems')">
            {{ t.nav.backToFront }}
          </NButton>
          <NButton size="small" @click="logout">{{ t.nav.logout }}</NButton>
        </div>
      </div>
    </NLayoutHeader>

    <NLayout has-sider class="admin-body">
      <NLayoutSider bordered :width="220" :native-scrollbar="false">
        <NMenu :options="menuOptions" :value="activeKey" :indent="18" @update:value="onMenuSelect" />
      </NLayoutSider>
      <!-- native-scrollbar=false 才会用 NScrollbar 包裹内容，去掉 Dashboard /
           Tags 等长页面在右侧出现的浏览器原生灰色滚动条，视觉上和学生端一致。 -->
      <NLayoutContent class="admin-content" :native-scrollbar="false">
        <RouterView />
      </NLayoutContent>
    </NLayout>
  </div>
</template>

<style scoped>
.admin-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
}
.admin-header {
  padding: 0;
}
.header-inner {
  display: flex;
  align-items: center;
  padding: 0 24px;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  height: 56px;
  box-sizing: border-box;
}
.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  min-width: 240px;
  justify-content: flex-end;
}
.user-nick {
  display: inline-flex;
  align-items: center;
  height: 100%;
}
.admin-body {
  flex: 1;
  min-height: 0;
}
.admin-content {
  padding: 24px;
}
</style>
