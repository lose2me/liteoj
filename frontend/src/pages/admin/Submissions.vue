<script setup lang="ts">
import { NButton, useMessage } from "naive-ui";
import { ref } from "vue";
import SubmissionTable from "../../components/SubmissionTable.vue";
import { http } from "../../api/http";
import { t } from "../../i18n";

const msg = useMessage();
const resuming = ref(false);
const tableKey = ref(0);

const resumePending = async () => {
  if (resuming.value) return;
  resuming.value = true;
  try {
    const { data } = await http.post("/admin/submissions/resume-pending");
    msg.success(
      t.adminDashboard.resumePendingOk(
        data.resumed_count || 0,
        data.pending_count || 0,
      ),
    );
    tableKey.value++;
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.adminDashboard.resumePendingFailed);
  } finally {
    resuming.value = false;
  }
};
</script>

<template>
  <div>
    <SubmissionTable :key="tableKey" :show-filters="true" :page-size="16">
      <template #filters-right>
        <NButton type="primary" :loading="resuming" @click="resumePending">
          {{ t.adminDashboard.resumePending }}
        </NButton>
      </template>
    </SubmissionTable>
  </div>
</template>
