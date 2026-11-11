<script setup lang="ts">
import { watch } from 'vue';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useEcharts } from '@/hooks/common/echarts';

defineOptions({
  name: 'PieChart'
});

const appStore = useAppStore();

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  title: {
    text: '97,925',
    subtext: $t('page.home.totalSamples'),
    left: 'center',
    top: '36%',
    textStyle: { fontSize: 22, fontWeight: 600, color: '#646cff' },
    subtextStyle: { fontSize: 12, color: '#9aa0ae' }
  },
  legend: { bottom: '0', left: 'center', icon: 'circle', itemWidth: 8, itemHeight: 8 },
  series: [
    {
      name: $t('page.home.schedule'),
      type: 'pie',
      radius: ['56%', '76%'],
      center: ['50%', '46%'],
      avoidLabelOverlap: false,
      color: ['#646cff', '#38bdf8', '#fbbf24', '#34d399'],
      itemStyle: { borderRadius: 8, borderWidth: 3, borderColor: 'transparent' },
      label: { show: false },
      labelLine: { show: false },
      data: [] as { name: string; value: number }[]
    }
  ]
}));

function getData() {
  return [
    { name: $t('page.home.study'), value: 36 },
    { name: $t('page.home.work'), value: 28 },
    { name: $t('page.home.rest'), value: 22 },
    { name: $t('page.home.entertainment'), value: 14 }
  ];
}

function fillData() {
  updateOptions(opts => {
    opts.series[0].data = getData();
    return opts;
  });
}

function updateLocale() {
  updateOptions((opts, factory) => {
    const originOpts = factory();
    opts.title.text = originOpts.title.text;
    opts.title.subtext = originOpts.title.subtext;
    opts.series[0].name = originOpts.series[0].name;
    opts.series[0].data = getData();
    return opts;
  });
}

watch(
  () => appStore.locale,
  () => updateLocale()
);

fillData();
</script>

<template>
  <div class="chart-wrap">
    <header class="chart-header">
      <div>
        <h3 class="chart-title">{{ $t('page.home.distributionTitle') }}</h3>
        <p class="chart-subtitle">{{ $t('page.home.schedule') }}</p>
      </div>
    </header>
    <div ref="domRef" class="h-300px w-full"></div>
  </div>
</template>

<style scoped>
.chart-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.chart-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.chart-subtitle {
  margin: 4px 0 0;
  font-size: 12px;
  color: #9aa0ae;
}
</style>
