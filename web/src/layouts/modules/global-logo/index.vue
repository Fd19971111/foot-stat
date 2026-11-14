<script setup lang="ts">
import { $t } from '@/locales';

defineOptions({
  name: 'GlobalLogo'
});

interface Props {
  /** Whether to show the title */
  showTitle?: boolean;
}

withDefaults(defineProps<Props>(), {
  showTitle: true
});
</script>

<template>
  <RouterLink to="/" class="footstat-brand nowrap-hidden">
    <span class="footstat-brand-mark"><SvgIcon icon="mdi:soccer" /></span>
    <span v-show="showTitle" class="footstat-brand-name">{{ $t('system.title') }}</span>
  </RouterLink>
</template>

<style scoped>
.footstat-brand {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 9px;
  overflow: hidden;
  text-decoration: none;
}

.footstat-brand-mark {
  display: grid;
  width: 30px;
  height: 30px;
  flex: 0 0 30px;
  place-items: center;
  border: 1px solid rgb(255 255 255 / 70%);
  border-radius: 8px;
  background: linear-gradient(135deg, #38bdf8, #0d7fa8);
  box-shadow: 0 4px 12px rgb(13 127 168 / 28%);
  color: #fff;
  font-size: 19px;
}

.footstat-brand-name {
  overflow: hidden;
  color: #12263a;
  font-size: 14px;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
