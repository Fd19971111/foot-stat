<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NSelect, NSpin } from 'naive-ui';
import { fetchFootballMatches } from '@/service/api/football';
import type { MatchSummary } from '@/utils/football';
import { useEcharts } from '@/hooks/common/echarts';
import type { ECOption } from '@/hooks/common/echarts';
import { useFootballData } from '@/hooks/business/useFootballData';
import { downloadCsv, summarizeMatches } from '@/utils/football';

defineOptions({ name: 'FootballAnalysis' });

const { seasons, season, league, month, rows, loading, seasonOptions, leagueOptions, monthOptions } = useFootballData();
const metric = ref<'homeRate' | 'averageGoals' | 'over25Rate' | 'averageCorners' | 'averageCards'>('over25Rate');
const summary = computed(() => summarizeMatches(rows.value));
const historyLoading = ref(false);
const history = ref<Array<{ year: string } & MatchSummary>>([]);
const metricOptions = [
  { label: '主胜比例', value: 'homeRate' },
  { label: '场均进球', value: 'averageGoals' },
  { label: '大于 2.5 球', value: 'over25Rate' },
  { label: '场均角球', value: 'averageCorners' },
  { label: '场均黄牌', value: 'averageCards' }
];
async function loadHistory() {
  if (!seasons.value.length || !league.value) {
    history.value = [];
    return;
  }
  historyLoading.value = true;
  const recentSeasons = seasons.value.slice(0, 5);
  const results = await Promise.all(recentSeasons.map(async item => {
    const startYear = Number(item.label.split('-')[0]);
    const year = month.value ? startYear + (month.value < 8 ? 1 : 0) : 0;
    const { data, error } = await fetchFootballMatches({ season: item.value, league: league.value, year, month: month.value });
    return { label: item.label, rows: error ? [] : data?.rows ?? [] };
  }));
  history.value = results.map(item => ({ year: item.label, ...summarizeMatches(item.rows) })).reverse();
  historyLoading.value = false;
}

watch([() => seasons.value.length, league, month], loadHistory, { immediate: true });
const currentMetric = computed(() => summary.value[metric.value]);
const average = computed(() =>
  history.value.length ? history.value.reduce((sum, item) => sum + item[metric.value], 0) / history.value.length : 0
);
const isRate = computed(() => metric.value === 'homeRate' || metric.value === 'over25Rate');
const metricLabel = computed(() => metricOptions.find(item => item.value === metric.value)?.label ?? '');
const metricValue = (value: number) => (isRate.value ? `${value.toFixed(1)}%` : value.toFixed(2));
const axisValue = (value: number) => (isRate.value ? `${value.toFixed(0)}%` : value.toFixed(1));
const delta = computed(() => currentMetric.value - average.value);
const deltaTone = computed(() => (Math.abs(delta.value) < 0.005 ? 'is-flat' : delta.value > 0 ? 'is-up' : 'is-down'));
const deltaText = computed(() => {
  if (Math.abs(delta.value) < 0.005) return '与均值持平';
  const unit = isRate.value ? ' pp' : '';
  return `较均值${delta.value > 0 ? '高' : '低'} ${Math.abs(delta.value).toFixed(isRate.value ? 1 : 2)}${unit}`;
});

function buildChartOption(): ECOption {
  const labels = history.value.map(item => item.year);
  const values = history.value.map(item => item[metric.value]);
  return {
    animationDuration: 600,
    animationEasing: 'cubicOut',
    grid: { left: 18, right: 24, top: 34, bottom: 6, containLabel: true },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.92)',
      borderColor: '#dbe7f2',
      borderWidth: 1,
      padding: [10, 13],
      textStyle: { color: '#3d5468', fontSize: 12 },
      extraCssText: 'border-radius: 8px; backdrop-filter: blur(10px); box-shadow: 0 10px 28px rgba(22, 62, 105, 0.14);',
      valueFormatter: value => metricValue(Number(Array.isArray(value) ? value[0] : value)),
      axisPointer: { type: 'line', lineStyle: { color: '#0d7fa8', width: 1, type: 'dashed' } }
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: labels,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#7089a0', fontSize: 10, margin: 14, hideOverlap: true }
    },
    yAxis: {
      type: 'value',
      scale: true,
      splitNumber: 4,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        color: '#7089a0',
        fontSize: 10,
        formatter: (value: number | string) => axisValue(Number(value))
      },
      splitLine: { lineStyle: { color: '#e2ecf6', type: 'dashed' } }
    },
    series: [
      {
        name: metricLabel.value,
        type: 'line',
        smooth: 0.35,
        symbol: 'circle',
        symbolSize: 10,
        clip: false,
        z: 3,
        lineStyle: {
          width: 3,
          color: '#0d7fa8',
          shadowColor: 'rgba(13, 127, 168, 0.22)',
          shadowBlur: 12,
          shadowOffsetY: 7
        },
        itemStyle: { color: '#38bdf8', borderColor: '#fff', borderWidth: 2 },
        emphasis: {
          focus: 'series',
          itemStyle: {
            color: '#0a6285',
            borderColor: '#fff',
            borderWidth: 2,
            shadowBlur: 12,
            shadowColor: 'rgba(13, 127, 168, 0.35)'
          }
        },
        label: {
          show: true,
          position: 'top',
          distance: 12,
          color: '#12263a',
          fontSize: 11,
          fontWeight: 600,
          formatter: (params: { value: unknown }) => metricValue(Number(params.value))
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(56, 189, 248, 0.28)' },
              { offset: 1, color: 'rgba(56, 189, 248, 0.01)' }
            ]
          }
        },
        data: values
      },
      {
        name: '均值参考线',
        type: 'line',
        data: labels.map(() => Number(average.value.toFixed(2))),
        symbol: 'none',
        silent: true,
        z: 1,
        lineStyle: { color: '#a9bfd4', width: 1.5, type: 'dashed' }
      }
    ]
  };
}

const { domRef, updateOptions } = useEcharts(() => buildChartOption(), {
  onRender: () => {},
  onUpdated: () => {}
});

watch([history, metric], () => {
  updateOptions(() => buildChartOption());
});

function exportAnalysis() {
  downloadCsv('football-analysis.csv', [
    ['年份', '场次', '主胜率(%)', '平局率(%)', '客胜率(%)', '场均进球', '大于2.5球(%)', '场均角球', '场均黄牌'],
    ...history.value.map(item => [item.year, item.games, item.homeRate.toFixed(2), item.drawRate.toFixed(2), item.awayRate.toFixed(2), item.averageGoals.toFixed(2), item.over25Rate.toFixed(2), item.averageCorners.toFixed(2), item.averageCards.toFixed(2)])
  ]);
}
</script>

<template>
  <div class="football-workspace">
    <header class="fw-heading">
      <div><span class="fw-eyebrow">足球 / 数据分析</span><h2>数据分析</h2><p>按年份聚合本地完赛记录，查看联赛赛果和比赛指标变化。</p></div>
      <div class="fw-heading-actions">
        <label class="fw-field">赛季<NSelect v-model:value="season" :options="seasonOptions" /></label>
        <label class="fw-field">联赛<NSelect v-model:value="league" :options="leagueOptions" /></label>
        <label class="fw-field">月份<NSelect v-model:value="month" :options="monthOptions" /></label>
      </div>
    </header>

    <section class="fw-hero">
      <div class="fw-hero-copy"><span>真实赛果 · 年度汇总</span><h2>{{ leagueOptions.find(item => item.value === league)?.label ?? '联赛分析' }}</h2><p>只统计 CSV 中存在的完赛字段，缺失指标按可用样本计算。</p></div>
      <div class="fw-hero-stats"><div><span>有效年份</span><strong>{{ history.length }}</strong></div><div><span>总场次</span><strong>{{ summary.games.toLocaleString() }}</strong></div><div><span>当前指标</span><strong>{{ metricValue(currentMetric) }}</strong></div></div>
    </section>

    <section class="fw-analysis-layout">
      <article class="fw-panel">
        <div class="fw-panel-head fw-chart-head"><div><h3>近五赛季趋势</h3><p>{{ leagueOptions.find(item => item.value === league)?.label }} · 同一月份筛选</p></div><label class="fw-field">分析指标<NSelect v-model:value="metric" :options="metricOptions" /></label></div>
        <NSpin :show="loading || historyLoading">
          <div class="fw-panel-body">
            <div class="fw-chart-summary">
              <div class="fw-chart-summary-main"><span>近 {{ history.length }} 季均值 · {{ metricLabel }}</span><strong>{{ metricValue(average) }}</strong></div>
              <div class="fw-chart-summary-side"><span class="fw-tag">当前赛季 {{ metricValue(currentMetric) }}</span><span class="fw-delta" :class="deltaTone">{{ deltaText }}</span></div>
            </div>
            <div class="fw-chart-legend"><span><i class="fw-legend-line"></i>{{ metricLabel }}</span><span><i class="fw-legend-mean"></i>均值参考线</span></div>
            <div class="fw-chart-frame">
              <div ref="domRef" class="fw-chart"></div>
              <div v-if="!history.length" class="fw-chart-empty"><strong>当前赛季没有可汇总的完赛数据</strong><span>请调整月份或选择其他赛季</span></div>
            </div>
          </div>
        </NSpin>
      </article>
      <section class="fw-analysis-kpis">
        <article class="fw-stat"><span>主胜 / 平 / 客胜</span><strong>{{ summary.homeRate.toFixed(1) }} / {{ summary.drawRate.toFixed(1) }} / {{ summary.awayRate.toFixed(1) }}<small>%</small></strong></article>
        <article class="fw-stat"><span>场均进球</span><strong>{{ summary.averageGoals.toFixed(2) }}<small>球</small></strong></article>
        <article class="fw-stat"><span>大于 2.5 球</span><strong>{{ summary.over25Rate.toFixed(1) }}<small>%</small></strong></article>
        <article class="fw-stat"><span>场均角球 / 黄牌</span><strong>{{ summary.averageCorners.toFixed(1) }} / {{ summary.averageCards.toFixed(1) }}</strong></article>
      </section>
    </section>

    <section class="fw-panel">
      <div class="fw-panel-head fw-toolbar"><div><h3>赛季明细</h3><p>最近五个有数据的赛季，同一联赛和月份区间</p></div><button class="fw-button" @click="exportAnalysis">导出 CSV</button></div>
      <div class="fw-table-wrap"><table class="fw-table"><thead><tr><th>年份</th><th>场次</th><th>主胜</th><th>平局</th><th>客胜</th><th>场均进球</th><th>大于 2.5 球</th><th>场均角球</th><th>场均黄牌</th></tr></thead><tbody><tr v-for="item in [...history].reverse()" :key="item.year"><td><strong>{{ item.year }}</strong></td><td>{{ item.games }}</td><td>{{ item.homeRate.toFixed(1) }}%</td><td>{{ item.drawRate.toFixed(1) }}%</td><td>{{ item.awayRate.toFixed(1) }}%</td><td>{{ item.averageGoals.toFixed(2) }}</td><td><strong>{{ item.over25Rate.toFixed(1) }}%</strong></td><td>{{ item.averageCorners.toFixed(1) }}</td><td>{{ item.averageCards.toFixed(1) }}</td></tr></tbody></table></div>
    </section>
  </div>
</template>
