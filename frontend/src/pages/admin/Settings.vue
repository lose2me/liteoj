<script setup lang="ts">
import {
  NButton,
  NCard,
  NDynamicTags,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSpace,
  NSwitch,
  NTabs,
  NTabPane,
  useMessage,
} from "naive-ui";
import { onMounted, ref } from "vue";
import { http } from "../../api/http";
import MarkdownEditor from "../../components/MarkdownEditor.vue";
import { t } from "../../i18n";

interface SettingsForm {
  home: {
    content: string;
  };
  judge: {
    base_url: string;
    langs: string[];
    default_cpu_ms: number;
    default_mem_mb: number;
    queue_workers: number;
    queue_cap: number;
    max_wait_seconds: number;
    submit_interval_seconds: number;
  };
  ai: {
    enabled: boolean;
    bifrost_base_url: string;
    bifrost_api_key: string;
    bifrost_model: string;
    queue_workers: number;
    queue_cap: number;
    max_wait_seconds: number;
    prompt_wrong_answer: string;
    prompt_optimize: string;
    prompt_tag: string;
    prompt_gen_title: string;
    prompt_gen_desc: string;
    prompt_gen_idea: string;
    prompt_gen_explain: string;
    prompt_gen_testcases: string;
    prompt_gen_all: string;
  };
}

const msg = useMessage();
const loading = ref(false);
const saving = ref(false);

const createEmptyForm = (): SettingsForm => ({
  home: {
    content: "",
  },
  judge: {
    base_url: "",
    langs: [],
    default_cpu_ms: 1000,
    default_mem_mb: 256,
    queue_workers: 1,
    queue_cap: 256,
    max_wait_seconds: 120,
    submit_interval_seconds: 0,
  },
  ai: {
    enabled: false,
    bifrost_base_url: "",
    bifrost_api_key: "",
    bifrost_model: "",
    queue_workers: 2,
    queue_cap: 32,
    max_wait_seconds: 180,
    prompt_wrong_answer: "",
    prompt_optimize: "",
    prompt_tag: "",
    prompt_gen_title: "",
    prompt_gen_desc: "",
    prompt_gen_idea: "",
    prompt_gen_explain: "",
    prompt_gen_testcases: "",
    prompt_gen_all: "",
  },
});

const form = ref<SettingsForm>(createEmptyForm());

const load = async () => {
  loading.value = true;
  try {
    const { data } = await http.get("/admin/settings");
    form.value = data.settings || createEmptyForm();
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.opFailed);
  } finally {
    loading.value = false;
  }
};

const save = async () => {
  saving.value = true;
  try {
    await http.put("/admin/settings", form.value);
    msg.success(t.common.savedOk);
  } catch (e: any) {
    msg.error(e?.response?.data?.error || t.common.saveFailed);
  } finally {
    saving.value = false;
  }
};

onMounted(load);
</script>

<template>
  <div>
    <NCard :title="t.adminSettings.formTitle">
      <div v-if="loading" class="opacity-60">{{ t.common.loadingDots }}</div>
      <NTabs v-else type="line" animated>
        <NTabPane name="home" :tab="t.adminSettings.tabHome">
          <MarkdownEditor v-model="form.home.content" height="420px" />
        </NTabPane>

        <NTabPane name="judge" :tab="t.adminSettings.tabJudge">
          <NForm label-placement="top">
            <NFormItem :label="t.adminSettings.judgeBaseURL">
              <NInput v-model:value="form.judge.base_url" />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeLangs">
              <NDynamicTags v-model:value="form.judge.langs" />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeDefaultCPU">
              <NInputNumber
                v-model:value="form.judge.default_cpu_ms"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeDefaultMem">
              <NInputNumber
                v-model:value="form.judge.default_mem_mb"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeQueueWorkers">
              <NInputNumber
                v-model:value="form.judge.queue_workers"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeQueueCap">
              <NInputNumber
                v-model:value="form.judge.queue_cap"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeMaxWait">
              <NInputNumber
                v-model:value="form.judge.max_wait_seconds"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.judgeSubmitInterval">
              <NInputNumber
                v-model:value="form.judge.submit_interval_seconds"
                :min="0"
                style="width: 220px"
              />
            </NFormItem>
          </NForm>
        </NTabPane>

        <NTabPane name="ai" :tab="t.adminSettings.tabAI">
          <NForm label-placement="top">
            <NFormItem :label="t.adminSettings.aiEnabled">
              <NSwitch v-model:value="form.ai.enabled" />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiBaseURL">
              <NInput v-model:value="form.ai.bifrost_base_url" />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiAPIKey">
              <NInput
                v-model:value="form.ai.bifrost_api_key"
                type="password"
                show-password-on="click"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiModel">
              <NInput v-model:value="form.ai.bifrost_model" />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiQueueWorkers">
              <NInputNumber
                v-model:value="form.ai.queue_workers"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiQueueCap">
              <NInputNumber
                v-model:value="form.ai.queue_cap"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
            <NFormItem :label="t.adminSettings.aiMaxWait">
              <NInputNumber
                v-model:value="form.ai.max_wait_seconds"
                :min="1"
                style="width: 220px"
              />
            </NFormItem>
          </NForm>

          <NTabs class="mt-4" type="line" animated>
            <NTabPane
              name="wrong-answer"
              :tab="t.adminSettings.aiPromptTabWrongAnswer"
            >
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptWrongAnswer">
                  <NInput
                    v-model:value="form.ai.prompt_wrong_answer"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane name="optimize" :tab="t.adminSettings.aiPromptTabOptimize">
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptOptimize">
                  <NInput
                    v-model:value="form.ai.prompt_optimize"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane name="tag" :tab="t.adminSettings.aiPromptTabTag">
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptTag">
                  <NInput
                    v-model:value="form.ai.prompt_tag"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane
              name="gen-title"
              :tab="t.adminSettings.aiPromptTabGenTitle"
            >
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenTitle">
                  <NInput
                    v-model:value="form.ai.prompt_gen_title"
                    type="textarea"
                    :autosize="{ minRows: 8, maxRows: 18 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane name="gen-desc" :tab="t.adminSettings.aiPromptTabGenDesc">
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenDesc">
                  <NInput
                    v-model:value="form.ai.prompt_gen_desc"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane name="gen-idea" :tab="t.adminSettings.aiPromptTabGenIdea">
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenIdea">
                  <NInput
                    v-model:value="form.ai.prompt_gen_idea"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane
              name="gen-explain"
              :tab="t.adminSettings.aiPromptTabGenExplain"
            >
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenExplain">
                  <NInput
                    v-model:value="form.ai.prompt_gen_explain"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane
              name="gen-testcases"
              :tab="t.adminSettings.aiPromptTabGenTestcases"
            >
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenTestcases">
                  <NInput
                    v-model:value="form.ai.prompt_gen_testcases"
                    type="textarea"
                    :autosize="{ minRows: 10, maxRows: 24 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
            <NTabPane name="gen-all" :tab="t.adminSettings.aiPromptTabGenAll">
              <NForm label-placement="top">
                <NFormItem :label="t.adminSettings.aiPromptGenAll">
                  <NInput
                    v-model:value="form.ai.prompt_gen_all"
                    type="textarea"
                    :autosize="{ minRows: 12, maxRows: 28 }"
                  />
                </NFormItem>
              </NForm>
            </NTabPane>
          </NTabs>
        </NTabPane>

      </NTabs>

      <NSpace class="mt-4">
        <NButton
          type="primary"
          :loading="saving"
          :disabled="loading"
          @click="save"
        >
          {{ t.adminSettings.saveConfig }}
        </NButton>
      </NSpace>
    </NCard>
  </div>
</template>
