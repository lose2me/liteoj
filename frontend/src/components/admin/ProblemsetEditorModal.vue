<script setup lang="ts">
import { computed, toRef } from "vue";
import {
  NButton,
  NDatePicker,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSwitch,
} from "naive-ui";
import { t } from "../../i18n";

type ProblemsetEditMode = "create" | "update";

interface ProblemsetEditorForm {
  id: number;
  title: string;
  password: string;
  allowed_langs: string[];
  start_ts: number | null;
  end_ts: number | null;
  visible: boolean;
  enable_idea: boolean;
  enable_solution: boolean;
  enable_ai: boolean;
  enable_bonus: boolean;
}

const props = defineProps<{
  show: boolean;
  mode: ProblemsetEditMode;
  form: ProblemsetEditorForm;
  allLangs: string[];
}>();

const emit = defineEmits<{
  (e: "update:show", v: boolean): void;
  (e: "submit"): void;
}>();

const showModel = computed({
  get: () => props.show,
  set: (v: boolean) => emit("update:show", v),
});

const form = toRef(props, "form");

const title = computed(() =>
  props.mode === "create"
    ? t.problemsetAdmin.modalCreateTitle
    : t.problemsetAdmin.modalUpdateTitle,
);

const langOptions = computed(() =>
  props.allLangs.map((lang) => ({ label: lang, value: lang })),
);
</script>

<template>
  <NModal
    v-model:show="showModel"
    preset="card"
    :title="title"
    :style="{ width: 'min(560px, 96vw)' }"
  >
    <NForm label-placement="left" label-width="100">
      <NFormItem :label="t.problemsetAdmin.formTitle">
        <NInput v-model:value="form.title" />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formStart">
        <NDatePicker
          v-model:value="form.start_ts"
          type="datetime"
          clearable
          style="width: 100%"
        />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formEnd">
        <div style="width: 100%">
          <NDatePicker
            v-model:value="form.end_ts"
            type="datetime"
            clearable
            style="width: 100%"
          />
          <div class="text-xs opacity-60 mt-1">
            {{ t.problemsetAdmin.formTsHint }}
          </div>
        </div>
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formPwd">
        <NInput
          v-model:value="form.password"
          :placeholder="t.problemsetAdmin.formPwdPlaceholder"
        />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formAllowedLangs">
        <NSelect
          v-model:value="form.allowed_langs"
          multiple
          :options="langOptions"
          :placeholder="t.problemsetAdmin.formAllowedLangsPlaceholder"
        />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formDisableIdea">
        <NSwitch v-model:value="form.enable_idea" />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formDisableSolution">
        <NSwitch v-model:value="form.enable_solution" />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formDisableAI">
        <NSwitch v-model:value="form.enable_ai" />
      </NFormItem>
      <NFormItem :label="t.problemsetAdmin.formEnableBonus">
        <NSwitch v-model:value="form.enable_bonus" />
      </NFormItem>
      <NButton type="primary" @click="emit('submit')">
        {{ t.common.save }}
      </NButton>
    </NForm>
  </NModal>
</template>
