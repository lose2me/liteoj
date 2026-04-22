<script setup lang="ts">
import {
  NCard, NDescriptions, NDescriptionsItem, NStatistic, NGrid, NGridItem,
  NDataTable, NTag, NSelect, NSpace, NButton,
} from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { h, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../../api/http'
import { verdictType, verdictLabel } from '../../api/verdict'
import VerdictPie from '../../components/VerdictPie.vue'
import { t } from '../../i18n'

// Admin-only read-only profile for another user. Mirrors Me.vue's layout
// (profile card + stats + verdict pie + submission table) but drops the
// password change card and reads uid from the URL instead of session.

const route = useRoute()
const router = useRouter()
const profile = ref<any>(null)
const subs = ref<any[]>([])
const subsTotal = ref(0)
const subsPage = ref(1)
const subsVerdict = ref('')

const userId = () => Number(route.params.id)

const loadProfile = async () => {
  const { data } = await http.get(`/admin/users/${userId()}/profile`)
  profile.value = data
}

const loadSubs = async () => {
  const { data } = await http.get('/submissions', {
    params: {
      user_id: userId(),
      verdict: subsVerdict.value || undefined,
      page: subsPage.value,
      page_size: 16,
    },
  })
  subs.value = data.items
  subsTotal.value = data.total
}

onMounted(async () => {
  await loadProfile()
  await loadSubs()
})

const verdictOptions: SelectOption[] = [
  { label: t.common.all, value: '' },
  ...Object.keys(verdictLabel).map((k) => ({ label: k, value: k })),
]

const subsColumns = [
  { title: t.submission.colId, key: 'id', width: 80 },
  {
    title: t.submission.colProblem,
    key: 'problem_id',
    width: 80,
    render: (r: any) =>
      h('span', {
        class: 'cursor-pointer text-green-400',
        onClick: (e: Event) => { e.stopPropagation(); router.push(`/problems/${r.problem_id}`) },
      }, `#${r.problem_id}`),
  },
  {
    title: t.submission.colResult,
    key: 'verdict',
    width: 110,
    render: (r: any) =>
      h(NTag, { type: verdictType(r.verdict), size: 'small' }, {
        default: () => r.verdict,
      }),
  },
  { title: t.submission.colLang, key: 'language', width: 80 },
  { title: t.submission.colTime, key: 'time_used_ms', width: 80, render: (r: any) => `${r.time_used_ms} ms` },
  { title: t.submission.colMemory, key: 'memory_used_kb', width: 100, render: (r: any) => `${r.memory_used_kb} KB` },
  { title: t.submission.colCreatedAt, key: 'created_at', width: 170 },
  {
    title: t.submission.colOp, key: 'op', width: 100,
    render: (r: any) => h(NButton, { size: 'tiny', onClick: () => router.push(`/submissions/${r.id}`) },
      { default: () => t.submission.opDetail }),
  },
]
</script>

<template>
  <div v-if="profile" class="grid grid-cols-3 gap-4">
    <NCard :title="t.usersAdmin.profileTitle(profile.user?.name || profile.user?.username)" class="col-span-1">
      <NSpace class="mb-3">
        <NButton size="small" @click="router.push('/admin/users')">{{ t.usersAdmin.backToList }}</NButton>
      </NSpace>
      <NDescriptions :column="1" bordered label-placement="left" size="small">
        <NDescriptionsItem :label="t.me.account">{{ profile.user?.username }}</NDescriptionsItem>
        <NDescriptionsItem :label="t.me.name">{{ profile.user?.name }}</NDescriptionsItem>
        <NDescriptionsItem :label="t.me.role">{{ profile.user?.role }}</NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NCard :title="t.me.stats" class="col-span-2">
      <NGrid :cols="5" :x-gap="12">
        <NGridItem><NStatistic :label="t.me.acProblems" :value="profile.distinct_ac" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.triedProblems" :value="profile.distinct_tried" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.totalSubs" :value="profile.total_submissions" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.acRate" :value="Math.round((profile.ac_rate || 0) * 100)" suffix="%" /></NGridItem>
        <NGridItem><NStatistic :label="t.me.ak" :value="profile.ak || 0" /></NGridItem>
      </NGrid>

      <h3 class="mt-6">{{ t.me.verdictDistribution }}</h3>
      <VerdictPie :distribution="profile.distribution || {}" />
    </NCard>

    <NCard :title="t.me.mySubs" class="col-span-3">
      <NSpace class="mb-3">
        <NSelect
          v-model:value="subsVerdict"
          :options="verdictOptions"
          class="w-40"
          :placeholder="t.submission.filterVerdict"
        />
        <NButton type="primary" @click="subsPage = 1; loadSubs()">{{ t.me.filter }}</NButton>
      </NSpace>
      <NDataTable
        :columns="subsColumns"
        :data="subs"
        :pagination="{
          page: subsPage, pageSize: 16, itemCount: subsTotal, showSizePicker: false,
          onChange: (p: number) => { subsPage = p; loadSubs() },
        }"
        remote
        striped
      />
    </NCard>
  </div>
</template>
