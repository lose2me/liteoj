<script setup lang="ts">
// Lightweight, non-interactive verdict distribution: a stack of labeled bars.
// Replaces the old ECharts pie — no hover highlight, no legend toggling, no
// canvas — just enough visual to compare counts at a glance.
import { computed } from 'vue'
import { NTag } from 'naive-ui'
import { verdictType } from '../api/verdict'

const props = defineProps<{ distribution: Record<string, number> }>()

// AC pinned first; everything else by descending count, falling back to a
// stable canonical order so the layout doesn't reshuffle as counts change.
const order = ['AC', 'WA', 'PE', 'TLE', 'MLE', 'OLE', 'RE', 'CE', 'SE', 'UKE', 'PENDING']

const rows = computed(() => {
  const d = props.distribution || {}
  const total = Object.values(d).reduce((s, v) => s + (v || 0), 0)
  return order
    .filter((k) => (d[k] || 0) > 0)
    .map((k) => ({
      key: k,
      count: d[k],
      pct: total > 0 ? Math.max((d[k] / total) * 100, 2) : 0,
    }))
})

const total = computed(() =>
  Object.values(props.distribution || {}).reduce((s, v) => s + (v || 0), 0),
)
</script>

<template>
  <div v-if="rows.length === 0" class="empty">暂无提交</div>
  <div v-else class="vdist">
    <div v-for="r in rows" :key="r.key" class="row">
      <NTag :type="verdictType(r.key)" size="small" class="tag">{{ r.key }}</NTag>
      <div class="bar">
        <div class="fill" :class="`fill-${verdictType(r.key)}`" :style="{ width: r.pct + '%' }" />
      </div>
      <span class="count">{{ r.count }}</span>
    </div>
    <div class="total">总计 {{ total }}</div>
  </div>
</template>

<style scoped>
.vdist {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12.5px;
}
.tag {
  min-width: 56px;
  justify-content: center;
}
.bar {
  flex: 1;
  height: 8px;
  background: rgba(128, 128, 128, 0.15);
  border-radius: 4px;
  overflow: hidden;
}
.fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.25s ease;
}
.fill-success { background: #18a058; }
.fill-error   { background: #d03050; }
.fill-warning { background: #f0a020; }
.fill-info    { background: #2080f0; }
.fill-default { background: #909399; }
.fill-primary { background: #2080f0; }
.count {
  min-width: 36px;
  text-align: right;
  font-variant-numeric: tabular-nums;
  color: var(--n-text-color, #555);
}
.total {
  margin-top: 4px;
  font-size: 11.5px;
  opacity: 0.6;
  text-align: right;
}
.empty {
  font-size: 12.5px;
  opacity: 0.6;
  padding: 8px 0;
}
</style>
