<script setup lang="ts">
import { NLayout, NLayoutHeader, NMenu, NButton, NDrawer, NDrawerContent } from 'naive-ui'
import { computed, ref } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import LoginCard from './LoginCard.vue'
import ThemeToggleButton from './ThemeToggleButton.vue'
import { t } from '../i18n'

const user = useUserStore()
const router = useRouter()
const route = useRoute()

const isWide = computed(() => route.matched.some((m) => m.meta.wide))

// 登录抽屉：仅由用户点击导航栏「登录」按钮触发。
// 未登录访问受保护页面会被 router guard 带着 ?next= 跳回首页，但不自动弹抽屉——
// 让用户先看到首页内容，确认要登录再手动点按钮。
const showLogin = ref(false)
const onLoginSuccess = () => {
  showLogin.value = false
  // 登录后跳回 ?next= 目标（如果有）。放在 AppLayout 里是因为它不会随抽屉卸载，
  // router.push 的 navigation 不会被 LoginCard 卸载过程打断。
  const next = typeof route.query.next === 'string' ? route.query.next : ''
  if (next && next !== '/' && !next.startsWith('//')) {
    router.push(next)
  }
}

// Use plain-text labels + @update:value for navigation. An earlier version
// embedded `<RouterLink>` inside each option's label, which caused two click
// handlers (RouterLink's + NMenu's own selection handler) to fire against
// the same event and left the router stuck after landing on /me — the only
// recovery was a hard refresh. Single handler → single source of truth.
const menuOptions = computed(() => [
  { label: t.nav.home, key: '/' },
  { label: t.nav.problems, key: '/problems' },
  { label: t.nav.problemsets, key: '/problemsets' },
  { label: t.nav.submissions, key: '/submissions' },
  { label: t.nav.ranking, key: '/ranking' },
  { label: t.nav.me, key: '/me' },
])

const activeKey = computed(() => {
  const p = route.path
  if (p === '/') return '/'
  if (p.startsWith('/problemsets')) return '/problemsets'
  if (p.startsWith('/problems')) return '/problems'
  if (p.startsWith('/submissions')) return '/submissions'
  if (p.startsWith('/ranking')) return '/ranking'
  if (p.startsWith('/me')) return '/me'
  // 未匹配任何导航项时返回空串，避免误高亮。
  return ''
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
  <NLayout class="min-h-screen">
    <NLayoutHeader bordered class="app-header">
      <div class="header-inner">
        <div class="text-lg font-bold mr-8 cursor-pointer" @click="router.push('/')">{{ t.nav.appName }}</div>
        <NMenu
          mode="horizontal"
          :options="menuOptions"
          :value="activeKey"
          class="flex-1 app-nav-menu"
          @update:value="onMenuSelect"
        />
        <div class="header-actions">
          <ThemeToggleButton />
          <template v-if="user.isLoggedIn">
            <span class="user-nick text-sm opacity-80">
              {{ user.user?.name || user.user?.username }}
            </span>
            <NButton v-if="user.isAdmin" size="small" type="primary" ghost @click="router.push('/admin')">
              {{ t.nav.adminPanel }}
            </NButton>
            <NButton size="small" @click="logout">{{ t.nav.logout }}</NButton>
          </template>
          <template v-else>
            <NButton size="small" type="primary" @click="showLogin = true">{{ t.nav.login }}</NButton>
          </template>
        </div>
      </div>
    </NLayoutHeader>
    <div class="content-wrap" :class="{ wide: isWide }">
      <RouterView />
    </div>
    <NDrawer v-model:show="showLogin" :width="360" placement="right">
      <NDrawerContent :title="t.auth.loginTitle" closable>
        <LoginCard @success="onLoginSuccess" />
      </NDrawerContent>
    </NDrawer>
  </NLayout>
</template>

<style scoped>
.app-header {
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
.app-nav-menu {
  height: 100%;
}
/* Naive horizontal NMenu defaults item height to ~42px; force it to match the
 * 56px header so labels sit on the header's vertical midline instead of near
 * the top. Targets both the menu row and each item's inner content box. */
:deep(.app-nav-menu.n-menu--horizontal),
:deep(.app-nav-menu .n-menu-item),
:deep(.app-nav-menu .n-menu-item-content),
:deep(.app-nav-menu .n-submenu > .n-menu-item) {
  height: 56px;
  line-height: 56px;
}
:deep(.app-nav-menu .n-menu-item-content) {
  display: flex;
  align-items: center;
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
.content-wrap {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}
.content-wrap.wide {
  max-width: none;
}
</style>
