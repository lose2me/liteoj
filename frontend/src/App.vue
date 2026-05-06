<script setup lang="ts">
import hljsAtomOneDarkCss from 'highlight.js/styles/atom-one-dark.css?url'
import hljsAtomOneLightCss from 'highlight.js/styles/atom-one-light.css?url'
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
const HIGHLIGHT_THEME_LINK_ID = 'liteoj-hljs-theme'

const applyHtmlClass = () => {
  document.documentElement.classList.toggle('dark', theme.isDark)
  document.documentElement.classList.toggle('light', !theme.isDark)
}

const applyHighlightThemeCss = () => {
  let link = document.getElementById(HIGHLIGHT_THEME_LINK_ID) as HTMLLinkElement | null
  if (!link) {
    link = document.createElement('link')
    link.id = HIGHLIGHT_THEME_LINK_ID
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }
  const href = theme.isDark ? hljsAtomOneDarkCss : hljsAtomOneLightCss
  if (link.getAttribute('href') !== href) {
    link.href = href
  }
}

const applyRuntimeTheme = () => {
  applyHtmlClass()
  applyHighlightThemeCss()
}

onMounted(() => {
  user.hydrate()
  theme.hydrate()
  codeEditorZoom.hydrate()
  applyRuntimeTheme()
})

// 切换主题时立即同步 html class，让纯 CSS 规则（非 Naive UI 管辖的自定义
// 组件样式）第一帧就呈现正确配色，无需等组件重绘。
watch(() => theme.mode, applyRuntimeTheme)
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
