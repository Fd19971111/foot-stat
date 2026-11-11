<script setup lang="ts">
import { computed } from 'vue';
import { $t } from '@/locales';

defineOptions({
  name: 'CardData'
});

interface CardItem {
  key: string;
  title: string;
  value: number;
  trend: string;
  percent: string;
  icon: string;
  start: string;
  end: string;
}

const cardData = computed<CardItem[]>(() => [
  {
    key: 'sample',
    title: $t('page.home.visitCount'),
    value: 9725,
    trend: '+12.5%',
    percent: '76%',
    icon: 'mdi:database-outline',
    start: '#6366f1',
    end: '#8b5cf6'
  },
  {
    key: 'records',
    title: $t('page.home.turnover'),
    value: 970925,
    trend: '+8.2%',
    percent: '88%',
    icon: 'mdi:table-multiple',
    start: '#0ea5e9',
    end: '#38bdf8'
  },
  {
    key: 'import',
    title: $t('page.home.downloadCount'),
    value: 1026,
    trend: '+5.1%',
    percent: '62%',
    icon: 'mdi:tray-arrow-down',
    start: '#f59e0b',
    end: '#fbbf24'
  },
  {
    key: 'quality',
    title: $t('page.home.dealCount'),
    value: 926,
    trend: '+2.4%',
    percent: '92%',
    icon: 'mdi:shield-check-outline',
    start: '#10b981',
    end: '#34d399'
  }
]);
</script>

<template>
  <NGrid cols="s:1 m:2 l:4" responsive="screen" :x-gap="16" :y-gap="16">
    <NGi v-for="item in cardData" :key="item.key">
      <div class="data-card" :style="{ '--c-start': item.start, '--c-end': item.end }">
        <div class="data-top">
          <div class="data-icon"><SvgIcon :icon="item.icon" class="text-24px" /></div>
          <span class="data-trend">
            <SvgIcon icon="mdi:trending-up" class="text-14px" />
            {{ item.trend }}
          </span>
        </div>
        <div class="data-label">{{ item.title }}</div>
        <div class="data-value">
          <CountTo :start-value="1" :end-value="item.value" />
        </div>
        <div class="data-footer">
          <span>{{ $t('page.home.compareLastWeek') }}</span>
        </div>
        <div class="data-bar"><span :style="{ width: item.percent }"></span></div>
      </div>
    </NGi>
  </NGrid>
</template>

<style scoped>
.data-card {
  position: relative;
  padding: 20px 22px 16px;
  overflow: hidden;
  border: 1px solid rgba(100, 108, 255, 0.08);
  border-radius: 16px;
  background: var(--n-card-color, #fff);
  box-shadow: 0 4px 18px rgba(80, 90, 160, 0.05);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.data-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 14px 30px rgba(80, 90, 160, 0.14);
}

.data-card::after {
  position: absolute;
  top: -46px;
  right: -46px;
  width: 130px;
  height: 130px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--c-start), var(--c-end));
  opacity: 0.08;
  content: '';
}

.data-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.data-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 13px;
  color: #fff;
  background: linear-gradient(135deg, var(--c-start), var(--c-end));
  box-shadow: 0 8px 18px color-mix(in srgb, var(--c-start) 30%, transparent);
}

.data-trend {
  display: inline-flex;
  gap: 3px;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.data-label {
  margin-top: 16px;
  font-size: 13px;
  color: #8b91a1;
}

.data-value {
  margin-top: 4px;
  font-size: 28px;
  font-weight: 600;
  line-height: 1.3;
  font-variant-numeric: tabular-nums;
}

.data-footer {
  margin-top: 4px;
  font-size: 11px;
  color: #a3a8b5;
}

.data-bar {
  height: 4px;
  margin-top: 14px;
  overflow: hidden;
  border-radius: 4px;
  background: rgba(130, 140, 170, 0.14);
}

.data-bar span {
  display: block;
  height: 100%;
  border-radius: 4px;
  background: linear-gradient(90deg, var(--c-start), var(--c-end));
}
</style>
