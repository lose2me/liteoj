<script setup lang="ts">
import { computed } from "vue";
import { NButton, NModal, NScrollbar, NTabPane, NTabs } from "naive-ui";
import { t } from "../../i18n";

const props = defineProps<{
  show: boolean;
  members: any[];
  bans: any[];
  membersTab: "members" | "bans";
}>();

const emit = defineEmits<{
  (e: "update:show", v: boolean): void;
  (e: "update:membersTab", v: "members" | "bans"): void;
  (e: "kick-member", member: any): void;
  (e: "unban", ban: any): void;
}>();

const showModel = computed({
  get: () => props.show,
  set: (v: boolean) => emit("update:show", v),
});

const tabModel = computed({
  get: () => props.membersTab,
  set: (v: "members" | "bans") => emit("update:membersTab", v),
});
</script>

<template>
  <NModal
    v-model:show="showModel"
    preset="card"
    :title="t.problemsetAdmin.membersTitle"
    :style="{ width: 'min(640px, 96vw)' }"
  >
    <NTabs v-model:value="tabModel" type="line" style="min-height: 480px">
      <NTabPane
        name="members"
        :tab="`${t.problemsetAdmin.tabMembers} (${members.length})`"
      >
        <div class="opacity-60 text-xs mb-2">
          {{ t.problemsetAdmin.memberRemoveHint }}
        </div>
        <div v-if="!members.length" class="opacity-60 text-sm">
          {{ t.problemsetAdmin.noMembers }}
        </div>
        <NScrollbar v-else style="max-height: 420px">
          <div class="members-list">
            <div v-for="m in members" :key="m.user_id" class="members-row">
              <div>
                <span class="font-medium">{{ m.name || m.username }}</span>
                <span class="ml-2 opacity-60 text-xs">{{ m.username }}</span>
                <span class="ml-2 opacity-60 text-xs">{{
                  (m.joined_at || "").replace("T", " ").slice(0, 16)
                }}</span>
              </div>
              <NButton size="tiny" type="error" @click="emit('kick-member', m)">
                {{ t.problemsetAdmin.memberRemove }}
              </NButton>
            </div>
          </div>
        </NScrollbar>
      </NTabPane>
      <NTabPane
        name="bans"
        :tab="`${t.problemsetAdmin.tabBans} (${bans.length})`"
      >
        <div v-if="!bans.length" class="opacity-60 text-sm">
          {{ t.problemsetAdmin.noBans }}
        </div>
        <NScrollbar v-else style="max-height: 420px">
          <div class="members-list">
            <div v-for="b in bans" :key="b.user_id" class="members-row">
              <div>
                <span class="font-medium">{{ b.name || b.username }}</span>
                <span class="ml-2 opacity-60 text-xs">{{ b.username }}</span>
                <span class="ml-2 opacity-60 text-xs">{{
                  (b.banned_at || "").replace("T", " ").slice(0, 16)
                }}</span>
              </div>
              <NButton size="tiny" type="warning" @click="emit('unban', b)">
                {{ t.problemsetAdmin.banUnban }}
              </NButton>
            </div>
          </div>
        </NScrollbar>
      </NTabPane>
    </NTabs>
  </NModal>
</template>

<style scoped>
.members-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.members-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border: 1px solid var(--lo-subtle-border);
  border-radius: 6px;
}
</style>
