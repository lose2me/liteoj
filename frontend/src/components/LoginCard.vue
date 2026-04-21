<script setup lang="ts">
import { NInput, NButton, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { http } from '../api/http'
import { useUserStore } from '../stores/user'
import { t } from '../i18n'

// 嵌在导航栏右侧抽屉里的登录面板。竖排紧凑布局。
// 登录成功后只抛 success，由父组件决定关抽屉 + 跳转 ?next=。
// 跳转放父组件是因为 emit 会触发抽屉关闭、LoginCard 卸载，若 router.push
// 写在这里会和卸载竞态，有时 navigation 会被丢掉。
const emit = defineEmits<{ (e: 'success'): void }>()

const user = useUserStore()
const route = useRoute()
const msg = useMessage()

const form = ref({ username: '', password: '' })
const loading = ref(false)

const submit = async () => {
  if (!form.value.username || !form.value.password) {
    msg.warning(t.auth.needUserAndPwd)
    return
  }
  loading.value = true
  try {
    const { data } = await http.post('/auth/login', form.value)
    user.setAuth(data.token, data.user)
    msg.success(t.auth.loginOk)
    emit('success')
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.auth.loginFailed)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-form">
    <NInput
      v-model:value="form.username"
      :placeholder="t.auth.username"
      autofocus
      @keyup.enter="submit"
    />
    <NInput
      v-model:value="form.password"
      type="password"
      show-password-on="click"
      :placeholder="t.auth.password"
      @keyup.enter="submit"
    />
    <NButton type="primary" block :loading="loading" @click="submit">
      {{ t.nav.login }}
    </NButton>
    <div v-if="route.query.next" class="redirect-hint">
      {{ t.auth.redirectHint }}<code>{{ route.query.next }}</code>
    </div>
  </div>
</template>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.redirect-hint {
  font-size: 12px;
  opacity: 0.6;
  word-break: break-all;
}
.redirect-hint code {
  margin-left: 4px;
}
</style>
