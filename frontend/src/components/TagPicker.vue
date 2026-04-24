<script setup lang="ts">
import { NSelect, NEmpty } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { http } from '../api/http'
import { t } from '../i18n'

interface Tag { id: number; name: string }
interface Group { id: number; name: string; tags: Tag[] }

const groups = ref<Group[]>([])
const modelValue = defineModel<number[]>({ default: () => [] })

onMounted(async () => {
  const { data } = await http.get('/tags')
  groups.value = data.groups || []
})

// 扁平化成 "分组 / 标签"，filterable 时可按分组名或标签名快速定位——237 项也
// 能即打即筛。value 仍是 tag.id，直接对应后端返回的 tag_ids。
const options = computed<SelectOption[]>(() => {
  const out: SelectOption[] = []
  for (const g of groups.value) {
    for (const tg of g.tags) {
      out.push({ label: `${g.name} / ${tg.name}`, value: tg.id })
    }
  }
  return out
})
</script>

<template>
  <div class="tag-picker">
    <NEmpty v-if="!groups.length" :description="t.tagPicker.emptyHint" size="small" />
    <!-- width: 100% 让输入框与上方标题等输入框同宽；不限 max-tag-count，
         已选标签超出一行会自动 wrap 到第二、第三行，高度自适应。 -->
    <NSelect
      v-else
      v-model:value="modelValue"
      :options="options"
      multiple
      filterable
      clearable
      :placeholder="t.tagPicker.placeholder"
      class="tag-select"
    />
  </div>
</template>

<style scoped>
.tag-picker {
  width: 100%;
}
.tag-select {
  width: 100%;
}
</style>
