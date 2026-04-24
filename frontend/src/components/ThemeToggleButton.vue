<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NTooltip } from 'naive-ui'
import { useThemeStore } from '../stores/theme'
import { t } from '../i18n'

const theme = useThemeStore()
const nextLabel = computed(() => (theme.isDark ? t.nav.themeLight : t.nav.themeDark))
</script>

<template>
  <NTooltip>
    <template #trigger>
      <NButton
        size="small"
        quaternary
        circle
        class="theme-toggle"
        :aria-label="nextLabel"
        :title="nextLabel"
        @click="theme.toggle"
      >
        <svg
          v-if="theme.isDark"
          viewBox="0 0 24 24"
          class="theme-icon"
          aria-hidden="true"
        >
          <circle cx="12" cy="12" r="4.5" fill="none" stroke="currentColor" stroke-width="1.8" />
          <path
            d="M12 2.75v2.5M12 18.75v2.5M21.25 12h-2.5M5.25 12h-2.5M18.54 5.46l-1.77 1.77M7.23 16.77l-1.77 1.77M18.54 18.54l-1.77-1.77M7.23 7.23L5.46 5.46"
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-width="1.8"
          />
        </svg>
        <svg
          v-else
          viewBox="0 0 24 24"
          class="theme-icon"
          aria-hidden="true"
        >
          <path
            d="M19.5 14.25A7.5 7.5 0 0 1 9.75 4.5a8.25 8.25 0 1 0 9.75 9.75Z"
            fill="none"
            stroke="currentColor"
            stroke-linejoin="round"
            stroke-width="1.8"
          />
        </svg>
      </NButton>
    </template>
    {{ nextLabel }}
  </NTooltip>
</template>

<style scoped>
.theme-toggle {
  flex-shrink: 0;
}

.theme-icon {
  width: 18px;
  height: 18px;
  display: block;
}
</style>
