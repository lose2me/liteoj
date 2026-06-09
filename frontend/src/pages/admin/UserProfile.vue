<script setup lang="ts">
import {
  NCard, NDescriptions, NDescriptionsItem, NStatistic, NGrid, NGridItem,
  NSpace, NButton,
} from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../../api/http'
import VerdictPie from '../../components/VerdictPie.vue'
import SubmissionTable from '../../components/SubmissionTable.vue'
import { t } from '../../i18n'

// Admin-only read-only profile for another user. Mirrors Me.vue's layout
// (profile card + stats + verdict pie + submission table) but drops the
// password change card and reads uid from the URL instead of session.

const route = useRoute()
const router = useRouter()
const profile = ref<any>(null)

const userId = () => Number(route.params.id)
const subsQuery = computed(() => ({
  user_id: userId(),
}))

const loadProfile = async () => {
  const { data } = await http.get(`/admin/users/${userId()}/profile`)
  profile.value = data
}

onMounted(async () => {
  await loadProfile()
})
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
      <SubmissionTable
        :query="subsQuery"
        :show-filters="true"
        :hide-filters-user="true"
        :page-size="16"
      />
    </NCard>
  </div>
</template>
