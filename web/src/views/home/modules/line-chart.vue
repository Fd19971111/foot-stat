<script setup lang="ts">
import { watch } from 'vue';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useEcharts } from '@/hooks/common/echarts';

defineOptions({
  name: 'LineChart'
});

const appStore = useAppStore();

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'line', lineStyle: { color: '#646cff', type: 'dashed' } }
  },
  legend: {
    data: [$t('page.home.downloadCount'), $t('page.home.registerCount')],
    right: '0',
    top: '0',
    icon: 'roundRect',
    itemWidth: 12,
    itemHeight: 8
  },
  grid: { left: '2%', right: '2%', bottom: '2%', top: '18%', containLabel: true },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: [] as string[],
    axisLine: { show: false },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { type: 'dashed', opacity: 0.5 } }
  },
  series: [
    {
      name: $t('page.home.downloadCount'),
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 7,
      lineStyle: { width: 3, color: '#646cff' },
      itemStyle: { color: '#646cff' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(100,108,255,0.32)' },
            { offset: 1, color: 'rgba(100,108,255,0.02)' }
          ]
        }
      },
      data: [] as number[]
    },
    {
      name: $t('page.home.registerCount'),
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 7,
      lineStyle: { width: 3, color: '#26deca' },
      itemStyle: { color: '#26deca' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(38,222,202,0.24)' },
            { offset: 1, color: 'rgba(38,222,202,0.02)' }
          ]
        }
      },
      data: [] as number[]
    }
  ]
}));

const days = ['11/04', '11/05', '11/06', '11/07', '11/08', '11/09', '11/10'];
const importData = [4623, 6145, 6268, 6411, 5890, 7251, 7978];
const validateData = [3208, 4016, 4916, 4512, 5281, 6008, 6963];

function fillData() {
  updateOptions(opts => {
    opts.xAxis.data = days;
    opts.series[0].data = importData;
    opts.series[1].data = validateData;
    return opts;
  });
}

function updateLocale() {
  updateOptions((opts, factory) => {
    const originOpts = factory();
    opts.legend.data = originOpts.legend.data;
    opts.series[0].name = originOpts.series[0].name;
    opts.series[1].name = originOpts.series[1].name;
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
        <h3 class="chart-title">{{ $t('page.home.trendTitle') }}</h3>
        <p class="chart-subtitle">{{ $t('page.home.overviewDesc') }}</p>
      </div>
      <NTag size="small" :bordered="false" round type="info">7 DAYS</NTag>
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
