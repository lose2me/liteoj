<script setup lang="ts">
import { h, computed, onMounted, ref } from 'vue'
import {
  NDataTable, NButton, NSpace, NModal, NForm, NFormItem, NInput, NSelect,
  NPopconfirm, useMessage, NAlert,
} from 'naive-ui'
import { useRouter } from 'vue-router'
import { http } from '../../api/http'
import { t } from '../../i18n'

interface Row {
  id: number
  username: string
  name: string
  role: string
  distinct_ac: number
  distinct_tried: number
  total_submissions: number
  ac_rate: number
  ak: number
  last_seen_at?: string | null
}

// 将 ISO 字符串格式化成 YYYY-MM-DD HH:MM；null/空串显示一个淡色占位符。
const fmtSeen = (s?: string | null) => {
  if (!s) return h('span', { class: 'opacity-40' }, t.common.empty)
  return h('span', { style: 'white-space:nowrap' }, s.replace('T', ' ').slice(0, 16))
}

const users = ref<Row[]>([])
const searchQ = ref('')
const msg = useMessage()
const router = useRouter()
const showEdit = ref(false)
const editMode = ref<'create' | 'update'>('create')
const form = ref({ id: 0, username: '', name: '', password: '', role: 'student' })

// Bulk import state — three aligned text columns pasted by the admin.
const showBulk = ref(false)
const bulkName = ref('')
const bulkUser = ref('')
const bulkPwd = ref('')
const bulkResult = ref<any>(null)
const bulkSubmitting = ref(false)

// `trim` at the split level so a trailing newline doesn't produce a phantom empty row.
const splitLines = (s: string) => s.split(/\r?\n/).filter((line) => line.trim() !== '')
const namesArr = computed(() => splitLines(bulkName.value))
const usersArr = computed(() => splitLines(bulkUser.value))
const pwdsArr = computed(() => splitLines(bulkPwd.value))
const bulkReady = computed(() => {
  const n = namesArr.value.length
  return n > 0 && n === usersArr.value.length && n === pwdsArr.value.length
})
const bulkHint = computed(() =>
  t.usersAdmin.bulkHint(namesArr.value.length, usersArr.value.length, pwdsArr.value.length),
)

const openBulk = () => {
  bulkName.value = ''
  bulkUser.value = ''
  bulkPwd.value = ''
  bulkResult.value = null
  showBulk.value = true
}

const bulkSubmit = async () => {
  if (!bulkReady.value) return
  bulkSubmitting.value = true
  try {
    const payload = {
      users: namesArr.value.map((name, i) => ({
        name: name.trim(),
        username: usersArr.value[i].trim(),
        password: pwdsArr.value[i],
      })),
    }
    const { data } = await http.post('/admin/users/bulk', payload)
    bulkResult.value = data
    msg.success(data.summary || t.usersAdmin.bulkSubmitOk)
    await load()
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.usersAdmin.bulkSubmitFailed)
  } finally {
    bulkSubmitting.value = false
  }
}

const load = async () => {
  const { data } = await http.get('/admin/users', {
    params: { q: searchQ.value || undefined },
  })
  users.value = data.items
}
onMounted(load)

const openCreate = () => {
  editMode.value = 'create'
  form.value = { id: 0, username: '', name: '', password: '', role: 'student' }
  showEdit.value = true
}
const openUpdate = (r: Row) => {
  editMode.value = 'update'
  form.value = { id: r.id, username: r.username, name: r.name, password: '', role: r.role }
  showEdit.value = true
}

const submit = async () => {
  try {
    if (editMode.value === 'create') {
      await http.post('/admin/users', form.value)
      msg.success(t.usersAdmin.createOk)
    } else {
      const { id, username, ...body } = form.value
      void username
      await http.put(`/admin/users/${id}`, body)
      msg.success(t.common.updatedOk)
    }
    showEdit.value = false
    await load()
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.opFailed)
  }
}

const remove = async (r: Row) => {
  await http.delete(`/admin/users/${r.id}`)
  msg.success(t.common.deletedOk)
  await load()
}

const columns = [
  { title: t.submission.colId, key: 'id', width: 70 },
  {
    title: t.submission.colAccount, key: 'username', width: 140,
    render: (r: Row) => h('span', {
      class: 'cursor-pointer text-green-400',
      onClick: (e: Event) => { e.stopPropagation(); router.push(`/admin/users/${r.id}`) },
    }, r.username),
  },
  { title: t.submission.colName, key: 'name' },
  { title: t.adminDashboard.colRole, key: 'role', width: 90 },
  { title: t.usersAdmin.colAC, key: 'distinct_ac', width: 80 },
  { title: t.usersAdmin.colTried, key: 'distinct_tried', width: 80 },
  { title: t.usersAdmin.colTotalSubs, key: 'total_submissions', width: 90 },
  {
    title: t.usersAdmin.colACRate, key: 'ac_rate', width: 80,
    render: (r: Row) => `${Math.round((r.ac_rate || 0) * 100)}%`,
  },
  { title: t.usersAdmin.colAK, key: 'ak', width: 70 },
  {
    title: t.usersAdmin.colLastSeen, key: 'last_seen_at', width: 150,
    render: (r: Row) => fmtSeen(r.last_seen_at),
  },
  {
    title: t.usersAdmin.colOp,
    key: 'op',
    width: 220,
    render: (r: Row) =>
      h('div', { class: 'row-actions', onClick: (e: Event) => e.stopPropagation() }, [
        h(NSpace, {}, {
          default: () => [
            h(NButton, { size: 'tiny', onClick: () => openUpdate(r) }, { default: () => t.common.edit }),
            h(NPopconfirm, { onPositiveClick: () => remove(r) }, {
              trigger: () => h(NButton, { size: 'tiny', type: 'error' }, { default: () => t.common.delete }),
              default: () => t.usersAdmin.confirmDeleteUser(r.username),
            }),
          ],
        }),
      ]),
  },
]

const rowProps = (row: Row) => ({
  style: 'cursor: pointer;',
  onClick: (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('.row-actions')) return
    router.push(`/admin/users/${row.id}`)
  },
})

// 客户端分页，固定 16/页。数据一次拉完后本地翻页即可。
const pagination = { pageSize: 16 }
</script>

<template>
  <div>
    <NSpace class="mb-3">
      <NButton type="primary" @click="openCreate">{{ t.usersAdmin.listNew }}</NButton>
      <NButton @click="openBulk">{{ t.usersAdmin.bulkImport }}</NButton>
      <NInput v-model:value="searchQ" :placeholder="t.usersAdmin.searchNamePlaceholder" clearable @keyup.enter="load" class="w-60" />
      <NButton @click="load">{{ t.common.search }}</NButton>
    </NSpace>
    <NDataTable :columns="columns" :data="users" :row-props="rowProps" :pagination="pagination" striped />

    <NModal v-model:show="showEdit" preset="card" :title="editMode === 'create' ? t.usersAdmin.modalCreateTitle : t.usersAdmin.modalUpdateTitle" class="w-120">
      <NForm label-placement="left" label-width="80">
        <NFormItem :label="t.submission.colAccount">
          <NInput v-model:value="form.username" :disabled="editMode === 'update'" />
        </NFormItem>
        <NFormItem :label="t.submission.colName">
          <NInput v-model:value="form.name" />
        </NFormItem>
        <NFormItem :label="t.auth.password">
          <NInput v-model:value="form.password" type="password"
            :placeholder="editMode === 'update' ? t.usersAdmin.pwdPlaceholderOnUpdate : ''" />
        </NFormItem>
        <NFormItem :label="t.adminDashboard.colRole">
          <NSelect v-model:value="form.role" :options="[
            { label: t.usersAdmin.roleStudent, value: 'student' },
            { label: t.usersAdmin.roleAdmin, value: 'admin' },
          ]" />
        </NFormItem>
        <NButton type="primary" @click="submit">{{ t.common.submit }}</NButton>
      </NForm>
    </NModal>

    <NModal v-model:show="showBulk" preset="card" :title="t.usersAdmin.bulkImportTitle" :style="{ width: 'min(820px, 96vw)' }">
      <div class="bulk-grid">
        <div>
          <div class="bulk-col-title">{{ t.usersAdmin.bulkColName }}</div>
          <NInput v-model:value="bulkName" type="textarea" :rows="10" />
        </div>
        <div>
          <div class="bulk-col-title">{{ t.usersAdmin.bulkColUser }}</div>
          <NInput v-model:value="bulkUser" type="textarea" :rows="10" />
        </div>
        <div>
          <div class="bulk-col-title">{{ t.usersAdmin.bulkColPwd }}</div>
          <NInput v-model:value="bulkPwd" type="textarea" :rows="10" />
        </div>
      </div>
      <div class="mt-2 text-xs opacity-70">{{ bulkHint }}</div>
      <NButton class="mt-3" type="primary" :disabled="!bulkReady" :loading="bulkSubmitting" @click="bulkSubmit">
        {{ t.usersAdmin.bulkImport }}
      </NButton>
      <NAlert v-if="bulkResult" class="mt-3" type="info" :title="bulkResult.summary">
        <div v-if="bulkResult.failures?.length">
          {{ t.usersAdmin.failedRows }}
          <ul>
            <li v-for="(f, i) in bulkResult.failures" :key="i">
              #{{ f.row }}: {{ f.error }}
            </li>
          </ul>
        </div>
      </NAlert>
    </NModal>
  </div>
</template>

<style scoped>
.bulk-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 12px;
}
.bulk-col-title {
  font-size: 12px;
  opacity: 0.7;
  margin-bottom: 4px;
}
</style>
