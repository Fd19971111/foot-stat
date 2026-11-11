<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NSelect, NSpin } from 'naive-ui';
import { fetchFootballMatches } from '@/service/api/football';
import type { MatchSummary } from '@/utils/football';
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
const average = computed(() => history.value.length ? history.value.reduce((sum, item) => sum + item[metric.value], 0) / history.value.length : 0);
const metricValue = (value: number) => metric.value === 'homeRate' || metric.value === 'over25Rate' ? `${value.toFixed(1)}%` : value.toFixed(2);
const chartPoints = computed(() => {
  if (!history.value.length) return [];
  const values = history.value.map(item => item[metric.value]);
  const minimum = Math.min(...values, average.value);
  const maximum = Math.max(...values, average.value);
  const range = maximum - minimum || 1;
  return values.map((value, index) => ({
    label: String(history.value[index].year),
    value,
    x: history.value.length === 1 ? 50 : 8 + (index / (history.value.length - 1)) * 84,
    y: 180 - ((value - minimum) / range) * 145
  }));
});
const chartPath = computed(() => chartPoints.value.map((point, index) => `${index ? 'L' : 'M'} ${point.x} ${point.y}`).join(' '));

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
      <div><span class="fw-eyebrow">Football / Analysis</span><h2>数据分析</h2><p>按年份聚合本地完赛记录，查看联赛赛果和比赛指标变化。</p></div>
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
        <NSpin :show="loading || historyLoading"><div class="fw-panel-body">
          <div class="fw-summary-line"><div><span>所选年度均值</span><strong>{{ metricValue(average) }}</strong></div><span class="fw-tag">当前赛季 {{ metricValue(currentMetric) }}</span></div>
          <svg v-if="chartPoints.length" class="fw-chart" viewBox="0 0 100 200" preserveAspectRatio="none" role="img" aria-label="年度数据趋势">
            <line v-for="y in [35, 75, 115, 155, 190]" :key="y" x1="5" x2="95" :y1="y" :y2="y" class="fw-chart-grid" />
            <path :d="chartPath" class="fw-chart-line" />
            <circle v-for="point in chartPoints" :key="point.label" :cx="point.x" :cy="point.y" r="1.7" class="fw-chart-dot"><title>{{ point.label }} · {{ metricValue(point.value) }}</title></circle>
          </svg>
          <div v-if="chartPoints.length" class="fw-chart-labels"><span v-for="point in chartPoints" :key="point.label">{{ point.label }}</span></div>
          <div v-else class="fw-empty"><strong>当前赛季没有可汇总的完赛数据</strong><span>请调整月份或选择其他赛季</span></div>
        </div></NSpin>
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