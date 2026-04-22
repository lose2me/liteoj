import { defineStore } from 'pinia'

// 主题偏好：`dark` | `light`。默认暗色以保持历史默认行为；用户显式切换后
// 通过 localStorage 持久化，下次打开保留选择。App.vue 监听此 store，绑定
// Naive UI 的 ConfigProvider.theme + document.documentElement 的 dark class。
export type ThemeMode = 'dark' | 'light'

const KEY = 'liteoj.theme'

export const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: 'dark' as ThemeMode,
  }),
  getters: {
    isDark: (s) => s.mode === 'dark',
  },
  actions: {
    hydrate() {
      const saved = localStorage.getItem(KEY)
      if (saved === 'light' || saved === 'dark') {
        this.mode = saved
      }
    },
    toggle() {
      this.mode = this.mode === 'dark' ? 'light' : 'dark'
      localStorage.setItem(KEY, this.mode)
    },
    setMode(m: ThemeMode) {
      this.mode = m
      localStorage.setItem(KEY, m)
    },
  },
})
