<script setup lang="ts">
import {
  NButton, NInput, NInputNumber,
  NModal, NForm, NFormItem, useMessage, useDialog, NEmpty,
} from 'naive-ui'
import { onMounted, ref, computed } from 'vue'
import { http } from '../../api/http'
import { t } from '../../i18n'

interface Tag { id: number; name: string; order_index: number }
interface Group { id: number; name: string; tags: Tag[]; order_index: number }

const groups = ref<Group[]>([])
const selectedGroupId = ref<number | null>(null)
const msg = useMessage()
const dialog = useDialog()

const showGroup = ref(false)
const groupForm = ref({ id: 0, name: '', order_index: 0 })
const showTag = ref(false)
const tagForm = ref({ id: 0, group_id: 0, name: '', order_index: 0 })

const selectedGroup = computed(() =>
  groups.value.find((g) => g.id === selectedGroupId.value) || null,
)
const groupTags = computed(() => selectedGroup.value?.tags || [])

const load = async () => {
  const { data } = await http.get('/tags')
  groups.value = data.groups || []
  if (!groups.value.some((g) => g.id === selectedGroupId.value)) {
    selectedGroupId.value = groups.value[0]?.id ?? null
  }
}
onMounted(load)

const openCreateGroup = () => {
  groupForm.value = { id: 0, name: '', order_index: 0 }
  showGroup.value = true
}
const openEditGroup = (g: Group) => {
  groupForm.value = { id: g.id, name: g.name, order_index: g.order_index }
  showGroup.value = true
}
const saveGroup = async () => {
  if (groupForm.value.id) {
    await http.put(`/admin/taggroups/${groupForm.value.id}`, groupForm.value)
  } else {
    await http.post('/admin/taggroups', groupForm.value)
  }
  showGroup.value = false
  msg.success(t.common.savedOk)
  await load()
}
const confirmRemoveGroup = (g: Group) => {
  dialog.warning({
    title: t.tagsAdmin.confirmDeleteGroup,
    content: t.tagsAdmin.confirmDeleteGroupBody(g.name),
    positiveText: t.common.delete,
    negativeText: t.common.cancel,
    onPositiveClick: async () => {
      await http.delete(`/admin/taggroups/${g.id}`)
      if (selectedGroupId.value === g.id) selectedGroupId.value = null
      msg.success(t.common.deletedOk)
      await load()
    },
  })
}

const openCreateTag = () => {
  if (!selectedGroup.value) return
  tagForm.value = { id: 0, group_id: selectedGroup.value.id, name: '', order_index: 0 }
  showTag.value = true
}
const openEditTag = (tag: Tag) => {
  if (!selectedGroup.value) return
  tagForm.value = { id: tag.id, group_id: selectedGroup.value.id, name: tag.name, order_index: tag.order_index }
  showTag.value = true
}
const saveTag = async () => {
  if (tagForm.value.id) {
    await http.put(`/admin/tags/${tagForm.value.id}`, tagForm.value)
  } else {
    await http.post('/admin/tags', tagForm.value)
  }
  showTag.value = false
  msg.success(t.common.savedOk)
  await load()
}
const confirmRemoveTag = (tag: Tag) => {
  dialog.warning({
    title: t.tagsAdmin.confirmDeleteTag,
    content: t.tagsAdmin.confirmDeleteTagBody(tag.name),
    positiveText: t.common.delete,
    negativeText: t.common.cancel,
    onPositiveClick: async () => {
      await http.delete(`/admin/tags/${tag.id}`)
      msg.success(t.common.deletedOk)
      await load()
    },
  })
}
</script>

<template>
  <div class="grid grid-cols-2 gap-4">
    <section class="tags-section">
      <h3 class="tags-section-title">{{ t.tagsAdmin.groupTitle }}</h3>
      <NEmpty v-if="!groups.length" :description="t.tagsAdmin.noGroups" size="small" />
      <div class="tag-list">
        <div
          v-for="g in groups"
          :key="g.id"
          class="tag-row"
          :class="{ 'tag-row-active': selectedGroupId === g.id }"
          @click="selectedGroupId = g.id"
        >
          <div>
            <span class="font-medium">{{ g.name }}</span>
            <span class="ml-2 opacity-60 text-xs">#{{ g.order_index }}</span>
            <span class="ml-2 opacity-60 text-xs">（{{ g.tags.length }}）</span>
          </div>
          <div @click.stop>
            <NButton size="tiny" @click="openEditGroup(g)">{{ t.common.edit }}</NButton>
            <NButton size="tiny" type="error" class="ml-1" @click="confirmRemoveGroup(g)">{{ t.common.delete }}</NButton>
          </div>
        </div>
        <div class="tag-row tag-row-add" @click="openCreateGroup">
          <span>+ {{ t.tagsAdmin.createGroup }}</span>
        </div>
      </div>
    </section>

    <section class="tags-section">
      <h3 class="tags-section-title">
        {{ selectedGroup ? t.tagsAdmin.subTitleWith(selectedGroup.name) : t.tagsAdmin.pickGroupFirst }}
      </h3>
      <template v-if="selectedGroup">
        <NEmpty v-if="!groupTags.length" :description="t.tagsAdmin.noSubTags" size="small" />
        <div class="tag-list">
          <div v-for="tag in groupTags" :key="tag.id" class="tag-row">
            <div>
              <span>{{ tag.name }}</span>
              <span class="ml-2 opacity-60 text-xs">#{{ tag.order_index }}</span>
            </div>
            <div>
              <NButton size="tiny" @click="openEditTag(tag)">{{ t.common.edit }}</NButton>
              <NButton size="tiny" type="error" class="ml-1" @click="confirmRemoveTag(tag)">{{ t.common.delete }}</NButton>
            </div>
          </div>
          <div class="tag-row tag-row-add" @click="openCreateTag">
            <span>+ {{ t.tagsAdmin.createTag }}</span>
          </div>
        </div>
      </template>
    </section>

    <NModal v-model:show="showGroup" preset="card" :title="groupForm.id ? t.tagsAdmin.editGroup : t.tagsAdmin.createGroup" :style="{ width: 'min(480px, 96vw)' }">
      <NForm label-placement="left" label-width="80">
        <NFormItem :label="t.tagsAdmin.nameLabel"><NInput v-model:value="groupForm.name" /></NFormItem>
        <NFormItem :label="t.tagsAdmin.orderLabel"><NInputNumber v-model:value="groupForm.order_index" /></NFormItem>
        <NButton type="primary" @click="saveGroup">{{ t.common.save }}</NButton>
      </NForm>
    </NModal>

    <NModal v-model:show="showTag" preset="card" :title="tagForm.id ? t.tagsAdmin.editTag : t.tagsAdmin.createTag" :style="{ width: 'min(480px, 96vw)' }">
      <NForm label-placement="left" label-width="80">
        <NFormItem :label="t.tagsAdmin.nameLabel"><NInput v-model:value="tagForm.name" /></NFormItem>
        <NFormItem :label="t.tagsAdmin.orderLabel"><NInputNumber v-model:value="tagForm.order_index" /></NFormItem>
        <NButton type="primary" @click="saveTag">{{ t.common.save }}</NButton>
      </NForm>
    </NModal>
  </div>
</template>

<style scoped>
.tags-section {
  min-width: 0;
}
.tags-section-title {
  font-size: 14px;
  opacity: 0.7;
  margin: 0 0 8px;
  font-weight: 500;
}
.tag-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.tag-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border: 1px solid var(--lo-subtle-border);
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.15s;
}
.tag-row:hover {
  background-color: var(--lo-subtle-bg);
}
.tag-row-active {
  background-color: var(--lo-active-bg);
  border-color: var(--lo-active-border);
}
.tag-row-add {
  justify-content: center;
  border-style: dashed;
  opacity: 0.7;
}
.tag-row-add:hover {
  opacity: 1;
}
</style>
