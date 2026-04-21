<script setup lang="ts">
import {
  NCard, NDescriptions, NDescriptionsItem, NForm, NFormItem, NInput, NButton,
  NStatistic, NGrid, NGridItem, NSelect, NSpace, useMessage,
} from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { http } from '../api/http'
import { useUserStore } from '../stores/user'
import { verdictLabel } from '../api/verdict'
import { onEvent } from '../api/events'
import VerdictPie from '../components/VerdictPie.vue'
import SubmissionTable from '../components/SubmissionTable.vue'
import { t } from '../i18n'

const user = useUserStore()
const msg = useMessage()
const oldPwd = ref('')
const newPwd = ref('')
const loading = ref(false)

const stats = ref<any>({ total_submissions: 0, distinct_ac: 0, distinct_tried: 0, ac_rate: 0, ak: 0, distribution: {} })

// "我的提交" — 列、分页、刷新都交给 SubmissionTable；这里只负责挑条件。
// verdict filter 通过 query prop 透传，watch 会自动重拉，filter 按钮也省了。
const subsVerdict = ref('')
const subsQuery = computed(() => ({
  user_id: user.user?.id,
  verdict: subsVerdict.value || undefined,
}))

let off: (() => void) | null = null

const loadStats = async () => {
  const { data } = await http.get('/me/stats')
  stats.value = data
}

onMounted(async () => {
  await loadStats()
  // SubmissionTable 自己订阅 submission:* 和 ai:task:done 刷表——这里只需
  // 监听 submission:done 来同步更新上方的统计卡（AC 数、提交总数等）。
  off = onEvent((ev) => {
    if (ev.type === 'submission:done' && ev.data?.user_id === user.user?.id) {
      loadStats()
    }
  })
})
onUnmounted(() => { off?.() })

const submit = async () => {
  if (!oldPwd.value || !newPwd.value || newPwd.value.length < 6) {
    msg.warning(t.me.pwdTooShort)
    return
  }
  loading.value = true
  try {
    await http.post('/me/password', { old_password: oldPwd.value, new_password: newPwd.value })
    msg.success(t.me.changeOk)
    user.logout()
    location.href = '/'
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.me.changeFailed)
  } finally {
    loading.value = false
  }
}

const verdictOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  ...Object.keys(verdictLabel).map((k) => ({ label: k, value: k })),
]
</script>

<template>
  <div class="grid grid-cols-3 gap-4">
    <NCard :title="t.me.profile" class="col-span-1">
      <NDescriptions :column="1" bordered label-placement="left" size="small">
        <NDescriptionsItem :label="t.me.account">{{ user.user?.username }}</NDescriptionsItem>
        <NDescriptionsItem :label="t.me.name">{{ user.user?.name }}</NDescriptionsItem>
        <NDescriptionsItem :label="t.me.role">{{ user.user?.role }}</NDescriptionsItem>
      </NDescriptions>

      <h3 class="mt-6">{{ t.me.changePassword }}</h3>
      <NForm label-placement="left" label-width="80">
        <NFormItem :label="t.me.oldPassword">
          <NInput v-model:value="oldPwd" type="password" show-password-on="click" />
        </NFormItem>
        <NFormItem :label="t.me.newPassword">
          <NInput v-model:value="newPwd" type="password" show-password-on="click" />
        </NFormItem>
        <NButton type="primary" :loading="loading" @click="submit">{{ t.common.submit }}</NButton>
      </NForm>
    </NCard>

    <NCard :title="t.me.stats" class="col-span-2">
      <NGrid :cols="5" :x-gap="12">
        <NGridItem><NStatistic :label="t.me.acProblems" :value="stats.distinct_ac" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.triedProblems" :value="stats.distinct_tried" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.totalSubs" :value="stats.total_submissions" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.acRate" :value="Math.round((stats.ac_rate || 0) * 100)" suffix="%" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.ak" :value="stats.ak || 0" /></NGridItem>
      </NGrid>

      <h3 class="mt-6">{{ t.me.verdictDistribution }}</h3>
      <VerdictPie :distribution="stats.distribution || {}" />
    </NCard>

    <NCard :title="t.me.mySubs" class="col-span-3">
      <NSpace class="mb-3">
        <NSelect
          v-model:value="subsVerdict"
          :options="verdictOptions"
          class="w-40"
          :placeholder="t.submission.filterVerdict"
        />
      </NSpace>
      <SubmissionTable v-if="user.user?.id" :query="subsQuery" :page-size="10" />
    </NCard>
  </div>
</template>
