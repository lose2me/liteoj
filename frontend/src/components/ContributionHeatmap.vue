<script setup lang="ts">
// 年度贡献热力图：52 周 × 7 天的 SVG 网格，灰→绿五档按当日提交数上色。
// 数据来源 /me/contribution ({items: [{date: 'YYYY-MM-DD', count, ac}]})。
// 手写 SVG 而非引入第三方库——依赖最小化，样式与站内暗色主题贴合。
import { computed, onMounted, ref } from 'vue'
import { NTooltip } from 'naive-ui'
import { http } from '../api/http'
import { t } from '../i18n'

interface Bucket { date: string; count: number; ac: number }
interface Cell { date: string; count: number; ac: number; level: 0 | 1 | 2 | 3 | 4 }

const items = ref<Bucket[]>([])

onMounted(async () => {
  const { data } = await http.get('/me/contribution')
  items.value = data.items || []
})

// 把 365 天铺成 [7][~53] 网格。起点对齐到今天所在周往前数 52 周的周日。
// weeks[i][j] = 第 i 周 第 j 天（0=周日…6=周六）。
const CELL = 11
const GAP = 3
const ROWS = 7

const weeks = computed<Cell[][]>(() => {
  const byDate: Record<string, Bucket> = {}
  for (const b of items.value) byDate[b.date] = b
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const end = today
  // 起点：一年前的周日。
  const start = new Date(end)
  start.setDate(end.getDate() - 364)
  while (start.getDay() !== 0) start.setDate(start.getDate() - 1)
  const days: Cell[] = []
  const d = new Date(start)
  while (d <= end) {
    const iso = fmt(d)
    const b = byDate[iso] || { date: iso, count: 0, ac: 0 }
    days.push({ ...b, level: levelOf(b.count) })
    d.setDate(d.getDate() + 1)
  }
  // 切成 7 行 N 列。
  const rows: Cell[][] = Array.from({ length: ROWS }, () => [])
  for (let i = 0; i < days.length; i++) {
    rows[i % ROWS].push(days[i])
  }
  return rows
})

const cols = computed(() => weeks.value[0]?.length || 53)
const width = computed(() => cols.value * (CELL + GAP))
const height = ROWS * (CELL + GAP)

// 月份标签：每一列首行是一个周日，若这一天所在的月份在上一列首周日所属月份
// 之后就在这一列画月份文字。
const monthLabels = computed(() => {
  const labels: { x: number; text: string }[] = []
  let lastMonth = -1
  for (let ci = 0; ci < cols.value; ci++) {
    const cell = weeks.value[0]?.[ci]
    if (!cell) continue
    const m = new Date(cell.date).getMonth()
    if (m !== lastMonth) {
      labels.push({ x: ci * (CELL + GAP), text: t.contribution.month(m + 1) })
      lastMonth = m
    }
  }
  return labels
})

const levelClass = (lv: number) => `lv-${lv}`

function fmt(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${dd}`
}

function levelOf(n: number): 0 | 1 | 2 | 3 | 4 {
  if (n <= 0) return 0
  if (n <= 2) return 1
  if (n <= 5) return 2
  if (n <= 9) return 3
  return 4
}
</script>

<template>
  <div class="heatmap">
    <svg :width="width + 28" :height="height + 18" class="heatmap-svg">
      <!-- 月份标签 -->
      <g transform="translate(24, 0)">
        <text
          v-for="(lbl, i) in monthLabels"
          :key="i"
          :x="lbl.x"
          :y="10"
          class="lbl"
        >{{ lbl.text }}</text>
      </g>
      <!-- 星期标签（仅一三五） -->
      <g transform="translate(0, 18)">
        <text :y="(CELL + GAP) * 1 + CELL - 2" class="lbl">{{ t.contribution.weekdays[1] }}</text>
        <text :y="(CELL + GAP) * 3 + CELL - 2" class="lbl">{{ t.contribution.weekdays[3] }}</text>
        <text :y="(CELL + GAP) * 5 + CELL - 2" class="lbl">{{ t.contribution.weekdays[5] }}</text>
      </g>
      <g transform="translate(24, 18)">
        <template v-for="(row, ri) in weeks" :key="ri">
          <template v-for="(cell, ci) in row" :key="cell.date">
            <NTooltip :delay="150">
              <template #trigger>
                <rect
                  :x="ci * (CELL + GAP)"
                  :y="ri * (CELL + GAP)"
                  :width="CELL"
                  :height="CELL"
                  :class="['cell', levelClass(cell.level)]"
                  rx="2"
                />
              </template>
              {{ t.contribution.tip(cell.date, cell.ac, cell.count) }}
            </NTooltip>
          </template>
        </template>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.heatmap {
  overflow-x: auto;
}
.heatmap-svg {
  display: block;
}
/* 暗 / 亮主题由 html.dark / html.light 类切换；SVG fill/stroke 不支持 CSS
 * var 继承到 :deep，所以直接在两个类下写两套值。 */
.lbl {
  font-size: 10px;
  fill: rgba(255, 255, 255, 0.45);
}
:global(html.light) .lbl {
  fill: rgba(0, 0, 0, 0.55);
}
.cell {
  stroke: rgba(255, 255, 255, 0.06);
  stroke-width: 1;
}
:global(html.light) .cell {
  stroke: rgba(0, 0, 0, 0.08);
}
.lv-0 { fill: rgba(255, 255, 255, 0.06); }
:global(html.light) .lv-0 { fill: #ebedf0; }
.lv-1 { fill: #0e4429; }
:global(html.light) .lv-1 { fill: #9be9a8; }
.lv-2 { fill: #006d32; }
:global(html.light) .lv-2 { fill: #40c463; }
.lv-3 { fill: #26a641; }
:global(html.light) .lv-3 { fill: #30a14e; }
.lv-4 { fill: #39d353; }
:global(html.light) .lv-4 { fill: #216e39; }
</style>
