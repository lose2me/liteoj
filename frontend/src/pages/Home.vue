<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { NCard, NSpin } from 'naive-ui'
import { http } from '../api/http'
import { onEvent } from '../api/events'
import MarkdownView from '../components/MarkdownView.vue'

// 首页（路由 /）——仅渲染 admin 在后台维护的一段 markdown。
// 登录入口收敛到导航栏右上角按钮 + 右侧抽屉，详见 AppLayout.vue。
// 订阅 home:changed 事件，admin 改完推送到位。
const content = ref('')
const loaded = ref(false)
let off: (() => void) | null = null

const load = async () => {
  const { data } = await http.get('/home')
  content.value = data.content || ''
  loaded.value = true
}

onMounted(() => {
  load()
  off = onEvent((ev) => {
    if (ev.type === 'home:changed') load()
  })
})
onUnmounted(() => { off?.() })
</script>

<template>
  <div class="home-wrap">
    <NCard v-if="loaded">
      <MarkdownView :content="content" />
    </NCard>
    <div v-else class="flex justify-center py-10"><NSpin /></div>
  </div>
</template>

<style scoped>
.home-wrap {
  max-width: 960px;
  margin: 0 auto;
}
</style>
