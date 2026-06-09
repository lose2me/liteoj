<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import {
  NButton,
  NInput,
  NInputNumber,
  NScrollbar,
  NSpace,
  useMessage,
} from "naive-ui";
import { http } from "../api/http";
import { onEvent } from "../api/events";
import { t } from "../i18n";

type BonusRow = {
  user_id: number;
  username: string;
  name: string;
  scores: Record<string, number>;
};

const props = withDefaults(
  defineProps<{ problemsetId: number; readonly?: boolean }>(),
  {
    readonly: false,
  },
);

const msg = useMessage();
const loading = ref(false);
const dates = ref<string[]>([]);
const rows = ref<BonusRow[]>([]);
const savingDate = ref<string | null>(null);
const accountInput = ref("");
const scoreInput = ref<number | null>(null);
let off: (() => void) | null = null;

function startOfToday() {
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  return now.getTime();
}

function toYMD(ts: number) {
  const d = new Date(ts);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function normalizeScores(raw: any) {
  const out: Record<string, number> = {};
  for (const [date, score] of Object.entries(raw || {})) {
    out[date] = Number(score || 0);
  }
  return out;
}

const todayDate = computed(() => toYMD(startOfToday()));
const isBusy = computed(() => loading.value || savingDate.value !== null);
const loadEndpoint = computed(() =>
  props.readonly
    ? `/problemsets/${props.problemsetId}/bonus`
    : `/admin/problemsets/${props.problemsetId}/bonus`,
);
const hintText = computed(() =>
  props.readonly ? t.problemset.bonusHint : t.problemsetAdmin.bonusHint,
);

function shortMonth(date: string) {
  return date.slice(5, 7);
}

function shortDay(date: string) {
  return date.slice(8, 10);
}

function getScore(row: BonusRow, date: string) {
  return Number(row.scores?.[date] || 0);
}

function displayScore(row: BonusRow, date: string) {
  const score = getScore(row, date);
  return score > 0 ? String(score) : "";
}

function ensureDate(date: string) {
  if (dates.value.includes(date)) return;
  dates.value = [...dates.value, date].sort();
  rows.value.forEach((row) => {
    if (!(date in row.scores)) row.scores[date] = 0;
  });
}

function setScore(userId: number, date: string, score: number) {
  const row = rows.value.find((item) => item.user_id === userId);
  if (!row) return;
  row.scores[date] = score;
}

const load = async () => {
  if (!props.problemsetId) return;
  loading.value = true;
  try {
    const { data } = await http.get(loadEndpoint.value);
    dates.value = Array.isArray(data.dates) ? [...data.dates].sort() : [];
    rows.value = (data.items || []).map((row: any) => ({
      user_id: row.user_id,
      username: row.username,
      name: row.name,
      scores: normalizeScores(row.scores),
    }));
    if (!props.readonly) ensureDate(todayDate.value);
    accountInput.value = "";
    scoreInput.value = null;
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.opFailed);
  } finally {
    loading.value = false;
  }
};

const saveDateColumn = async (date: string) => {
  if (!props.problemsetId || props.readonly) return;
  savingDate.value = date;
  try {
    await http.put(`/admin/problemsets/${props.problemsetId}/bonus`, {
      date,
      items: rows.value.map((row) => ({
        user_id: row.user_id,
        score: getScore(row, date),
      })),
    });
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.saveFailed);
    throw e;
  } finally {
    savingDate.value = null;
  }
};

const addBonus = async () => {
  if (props.readonly || isBusy.value) return;

  const username = accountInput.value.trim();
  const delta = Number(scoreInput.value);
  if (!username) {
    msg.error(t.problemsetAdmin.bonusAccountRequired);
    return;
  }
  if (!Number.isInteger(delta) || delta <= 0) {
    msg.error(t.problemsetAdmin.bonusInvalidScore);
    return;
  }

  const row = rows.value.find((item) => item.username === username);
  if (!row) {
    msg.error(t.problemsetAdmin.bonusAccountNotFound(username));
    return;
  }

  const date = todayDate.value;
  ensureDate(date);
  const prevScore = getScore(row, date);
  setScore(row.user_id, date, prevScore + delta);
  try {
    await saveDateColumn(date);
    accountInput.value = "";
    scoreInput.value = null;
    msg.success(t.problemsetAdmin.bonusSavedOk);
  } catch {
    setScore(row.user_id, date, prevScore);
  }
};

watch([() => props.problemsetId, () => props.readonly], load, {
  immediate: true,
});

onMounted(() => {
  off = onEvent((ev) => {
    if (
      (ev.type === "problemset:changed" ||
        ev.type === "problemset:members:changed") &&
      ev.data?.id === props.problemsetId
    ) {
      load();
    }
  });
});

onUnmounted(() => {
  off?.();
});
</script>

<template>
  <div>
    <NSpace vertical size="small" class="mb-3">
      <div class="text-sm opacity-70">{{ hintText }}</div>
      <NSpace v-if="!readonly && rows.length" align="end" wrap>
        <NInput
          v-model:value="accountInput"
          class="bonus-account-input"
          :disabled="isBusy"
          :placeholder="t.problemsetAdmin.bonusAccountPlaceholder"
          @keyup.enter="addBonus"
        />
        <NInputNumber
          v-model:value="scoreInput"
          class="bonus-score-input"
          :disabled="isBusy"
          :min="1"
          :precision="0"
          :placeholder="t.problemsetAdmin.bonusScorePlaceholder"
        />
        <NButton
          type="primary"
          :disabled="isBusy"
          :loading="savingDate === todayDate"
          @click="addBonus"
        >
          {{ t.problemsetAdmin.bonusAdd }}
        </NButton>
      </NSpace>
    </NSpace>

    <div v-if="loading" class="opacity-60 text-sm">
      {{ t.common.loadingDots }}
    </div>
    <div v-else-if="!rows.length" class="opacity-60 text-sm">
      {{ t.problemsetAdmin.bonusNoMembers }}
    </div>
    <div v-else-if="!dates.length" class="opacity-60 text-sm">
      {{ t.problemset.bonusEmpty }}
    </div>
    <NScrollbar v-else style="max-height: 520px">
      <div class="bonus-table-wrap">
        <table class="bonus-table">
          <thead>
            <tr>
              <th class="bonus-user-head">{{ t.problemset.bonusStudents }}</th>
              <th v-for="date in dates" :key="date" class="bonus-date-head">
                <div class="bonus-date-month">{{ shortMonth(date) }}</div>
                <div class="bonus-date-day">{{ shortDay(date) }}</div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.user_id">
              <th scope="row" class="bonus-user-cell">
                <div class="font-medium">{{ row.name || row.username }}</div>
                <div class="opacity-60 text-xs">{{ row.username }}</div>
              </th>
              <td v-for="date in dates" :key="date" class="bonus-cell-wrap">
                <div
                  class="bonus-cell"
                  :class="{
                    'is-filled': getScore(row, date) > 0,
                    'is-saving': savingDate === date,
                  }"
                >
                  {{ displayScore(row, date) }}
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </NScrollbar>
  </div>
</template>

<style scoped>
.bonus-account-input {
  width: 220px;
}

.bonus-score-input {
  width: 140px;
}

.bonus-table-wrap {
  min-width: max-content;
}

.bonus-table {
  border-collapse: separate;
  border-spacing: 0;
}

.bonus-user-head,
.bonus-user-cell {
  position: sticky;
  left: 0;
  z-index: 1;
  min-width: 180px;
  max-width: 180px;
  padding: 10px 12px;
  text-align: left;
  background: var(--lo-page-bg);
  border: 1px solid var(--lo-subtle-border);
}

.bonus-user-head {
  z-index: 2;
}

.bonus-date-head,
.bonus-cell-wrap {
  padding: 4px;
  border: 1px solid var(--lo-subtle-border);
  background: var(--lo-subtle-bg);
}

.bonus-date-head {
  width: 56px;
  min-width: 56px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.bonus-date-month {
  font-size: 11px;
  opacity: 0.7;
}

.bonus-date-day {
  font-size: 16px;
  font-weight: 700;
  line-height: 1.1;
}

.bonus-cell {
  width: 48px;
  height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid var(--lo-subtle-border);
  background: transparent;
  text-align: center;
  font-size: 15px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.bonus-cell.is-filled {
  background: var(--lo-accent-bg);
  color: var(--lo-accent-fg);
}

.bonus-cell.is-saving {
  opacity: 0.6;
}
</style>
