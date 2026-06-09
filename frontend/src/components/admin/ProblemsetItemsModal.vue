<script setup lang="ts">
import { computed } from "vue";
import {
  NButton,
  NCheckbox,
  NInput,
  NModal,
  NPagination,
  NScrollbar,
  NSelect,
  NSpace,
} from "naive-ui";
import type { SelectOption } from "naive-ui";
import { t } from "../../i18n";

const props = defineProps<{
  show: boolean;
  pickFilteredCount: number;
  pickSearch: string;
  pickTagNames: string[];
  tagOptions: SelectOption[];
  pickPaged: any[];
  pickPage: number;
  pageCount: number;
  pageSize: number;
  pickedList: any[];
  bindRightListRef: (el: Element | null) => void;
  isPicked: (id: number) => boolean;
  labelOf: (i: number) => string;
}>();

const emit = defineEmits<{
  (e: "update:show", v: boolean): void;
  (e: "update:pickSearch", v: string): void;
  (e: "update:pickTagNames", v: string[]): void;
  (e: "update:pickPage", v: number): void;
  (e: "toggle-pick", id: number): void;
  (e: "remove-pick", id: number): void;
  (e: "save"): void;
}>();

const showModel = computed({
  get: () => props.show,
  set: (v: boolean) => emit("update:show", v),
});

const searchModel = computed({
  get: () => props.pickSearch,
  set: (v: string) => emit("update:pickSearch", v),
});

const tagNamesModel = computed({
  get: () => props.pickTagNames,
  set: (v: string[]) => emit("update:pickTagNames", v),
});

const pageModel = computed({
  get: () => props.pickPage,
  set: (v: number) => emit("update:pickPage", v),
});

const bindRightListElement = (
  el: Element | { $el?: Element | null } | null,
) => {
  if (el instanceof Element || el === null) {
    props.bindRightListRef(el);
    return;
  }
  props.bindRightListRef(el.$el ?? null);
};
</script>

<template>
  <NModal
    v-model:show="showModel"
    preset="card"
    :title="t.problemsetAdmin.pickProblems"
    :style="{ width: 'min(860px, 96vw)' }"
  >
    <div class="picker">
      <div class="col">
        <div class="col-title">
          {{ t.problemsetAdmin.pickerAvailable(pickFilteredCount) }}
        </div>
        <NInput
          v-model:value="searchModel"
          :placeholder="t.problemsetAdmin.pickerSearchPlaceholder"
          clearable
          class="mb-2"
        />
        <NSelect
          v-model:value="tagNamesModel"
          multiple
          filterable
          :options="tagOptions"
          :placeholder="t.problemsetAdmin.pickerTagFilter"
          clearable
          class="mb-2"
        />
        <NScrollbar class="list-scroll">
          <div class="list">
            <label v-for="p in pickPaged" :key="p.id" class="row">
              <NCheckbox
                :checked="isPicked(p.id)"
                @update:checked="emit('toggle-pick', p.id)"
              />
              <span class="row-text">#{{ p.id }} {{ p.title }}</span>
            </label>
            <div v-if="!pickFilteredCount" class="opacity-60 text-sm p-2">
              {{ t.common.empty }}
            </div>
          </div>
        </NScrollbar>
        <NPagination
          v-if="pickFilteredCount > pageSize"
          v-model:page="pageModel"
          :page-count="pageCount"
          size="small"
          class="mt-2"
        />
      </div>
      <div class="col">
        <div class="col-title">
          {{ t.problemsetAdmin.pickerSelected(pickedList.length) }}
        </div>
        <div class="opacity-60 text-xs mb-2">
          {{ t.problemsetAdmin.pickerDragHint }}
        </div>
        <NScrollbar class="list-scroll">
          <div :ref="bindRightListElement" class="list sortable-list">
            <div
              v-for="(p, idx) in pickedList"
              :key="p.id"
              class="row ordered"
              :data-id="p.id"
            >
              <span class="drag-handle" aria-hidden="true">☰</span>
              <span class="row-index">{{ labelOf(idx) }}</span>
              <span class="row-text">#{{ p.id }} {{ p.title }}</span>
              <NButton
                size="tiny"
                text
                type="error"
                @click="emit('remove-pick', p.id)"
              >
                ×
              </NButton>
            </div>
            <div v-if="!pickedList.length" class="opacity-60 text-sm p-2">
              {{ t.problemsetAdmin.pickerEmptySelected }}
            </div>
          </div>
        </NScrollbar>
      </div>
    </div>
    <NSpace class="mt-3">
      <NButton type="primary" @click="emit('save')">{{ t.common.save }}</NButton>
    </NSpace>
  </NModal>
</template>

<style scoped>
.picker {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.picker .col {
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.picker .col-title {
  font-size: 13px;
  opacity: 0.75;
  margin-bottom: 6px;
}
.picker .list {
  border: 1px solid var(--lo-subtle-border);
  border-radius: 6px;
  padding: 4px;
}
.picker .list-scroll {
  max-height: 420px;
}
.picker .row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  user-select: none;
}
.picker .row:hover {
  background: var(--lo-subtle-bg);
}
.picker .row.ordered {
  cursor: grab;
}
.picker .row.ordered:active {
  cursor: grabbing;
  background: var(--lo-accent-bg-weak);
}
.picker .row-index {
  opacity: 0.7;
  min-width: 28px;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--lo-accent-bg);
  color: var(--lo-accent-fg);
}
.picker .row-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.picker .drag-handle {
  opacity: 0.5;
  cursor: grab;
}
.picker .sortable-ghost {
  opacity: 0.35;
  background: var(--lo-accent-bg-strong) !important;
}
.picker .sortable-chosen {
  background: var(--lo-accent-bg-weak);
}
</style>
