<script setup lang="ts">
import { NConfigProvider, NMessageProvider, NDialogProvider, darkTheme, zhCN, dateZhCN } from 'naive-ui'
import { computed, onMounted, watch } from 'vue'
import { useUserStore } from './stores/user'
import { useCodeEditorZoomStore } from './stores/editorZoom'
import { useThemeStore } from './stores/theme'

const user = useUserStore()
const theme = useThemeStore()
const codeEditorZoom = useCodeEditorZoomStore()

// 主题绑定：暗色走 Naive UI 的 darkTheme，亮色传 null 让 Naive 回落到默认
// 亮色主题。document.documentElement 上同步切换 `dark` class，供自定义
// CSS（例如 ContributionHeatmap 的 SVG 填色）按类选择不同颜色。
const activeTheme = computed(() => (theme.isDark ? darkTheme : null))

const applyHtmlClass = () => {
  document.documentElement.classList.toggle('dark', theme.isDark)
  document.documentElement.classList.toggle('light', !theme.isDark)
}

onMounted(() => {
  user.hydrate()
  theme.hydrate()
  codeEditorZoom.hydrate()
  applyHtmlClass()
})

// 切换主题时立即同步 html class，让纯 CSS 规则（非 Naive UI 管辖的自定义
// 组件样式）第一帧就呈现正确配色，无需等组件重绘。
watch(() => theme.mode, applyHtmlClass)
</script>

<template>
  <NConfigProvider :theme="activeTheme" :locale="zhCN" :date-locale="dateZhCN">
    <NMessageProvider>
      <NDialogProvider>
        <RouterView />
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>
