<script setup lang="ts">
import { computed, ref } from 'vue';
import { NEmpty, NSelect, NSpin } from 'naive-ui';
import { useFootballData } from '@/hooks/business/useFootballData';
import { displayDate, sortMatchesNewest, summarizeMatches } from '@/utils/football';

defineOptions({ name: 'MatchResults' });

const { season, league, month, rows, loading, seasonOptions, leagueOptions, monthOptions } = useFootballData();
const search = ref('');
const onlySettled = ref(true);
const summary = computed(() => summarizeMatches(rows.value));
const filteredRows = computed(() => {
  const keyword = search.value.trim().toLocaleLowerCase();
  return sortMatchesNewest(rows.value.filter(row => {
    const settled = ['H', 'D', 'A'].includes(row.FTR);
    const hasKeyword = !keyword || `${row.HomeTeam} ${row.AwayTeam} ${row.Date}`.toLocaleLowerCase().includes(keyword);
    return hasKeyword && (!onlySettled.value || settled);
  }));
});

const resultLabel: Record<string, string> = { H: '主胜', D: '平局', A: '客胜' };
const resultClass: Record<string, string> = { H: 'fw-result-home', D: 'fw-result-draw', A: 'fw-result-away' };
</script>

<template>
  <div class="football-workspace">
    <header class="fw-heading">
      <div><span class="fw-eyebrow">Football / Results</span><h2>赛果记录</h2><p>逐场核对已结束比赛的全场比分和技术统计。</p></div>
      <div class="fw-heading-actions">
        <label class="fw-field">赛季<NSelect v-model:value="season" :options="seasonOptions" /></label>
        <label class="fw-field">联赛<NSelect v-model:value="league" :options="leagueOptions" /></label>
        <label class="fw-field">月份<NSelect v-model:value="month" :options="monthOptions" /></label>
      </div>
    </header>

    <section class="fw-result-grid">
      <article class="fw-stat"><span>比赛记录</span><strong>{{ summary.games.toLocaleString() }}<small>场</small></strong></article>
      <article class="fw-stat"><span>已完赛</span><strong>{{ summary.homeWins + summary.draws + summary.awayWins }}<small>场</small></strong></article>
      <article class="fw-stat"><span>主胜比例</span><strong>{{ summary.homeRate.toFixed(1) }}<small>%</small></strong></article>
      <article class="fw-stat"><span>大于 2.5 球</span><strong>{{ summary.over25Rate.toFixed(1) }}<small>%</small></strong></article>
    </section>

    <section class="fw-panel">
      <div class="fw-panel-head fw-toolbar">
        <div><h3>比赛记录</h3><p>{{ seasonOptions.find(item => item.value === season)?.label }} · {{ leagueOptions.find(item => item.value === league)?.label }} · {{ filteredRows.length }} 场</p></div>
        <div class="fw-filters">
          <input v-model="search" class="fw-search" placeholder="搜索球队或日期" />
          <label class="fw-field"><span>显示范围</span><select v-model="onlySettled"><option :value="true">仅已完赛</option><option :value="false">全部比赛</option></select></label>
        </div>
      </div>
      <NSpin :show="loading">
        <div class="fw-table-wrap">
          <table v-if="filteredRows.length" class="fw-table">
            <thead><tr><th>日期 / 时间</th><th>主队</th><th>比分</th><th>客队</th><th>赛果</th><th>射门</th><th>射正</th><th>角球</th><th>黄牌</th><th>红牌</th></tr></thead>
            <tbody>
              <tr v-for="(row, index) in filteredRows" :key="`${row.Date}-${row.HomeTeam}-${index}`">
                <td>{{ displayDate(row.Date) }}<small v-if="row.Time"> {{ row.Time }}</small></td>
                <td><strong>{{ row.HomeTeam }}</strong></td>
                <td><strong>{{ row.FTHG || '-' }} : {{ row.FTAG || '-' }}</strong></td>
                <td><strong>{{ row.AwayTeam }}</strong></td>
                <td><span class="fw-tag" :class="resultClass[row.FTR] ?? 'fw-result-pending'">{{ resultLabel[row.FTR] ?? '未完赛' }}</span></td>
                <td>{{ row.HS || '-' }} : {{ row.AS || '-' }}</td>
                <td>{{ row.HST || '-' }} : {{ row.AST || '-' }}</td>
                <td>{{ row.HC || '-' }} : {{ row.AC || '-' }}</td>
                <td>{{ row.HY || '-' }} : {{ row.AY || '-' }}</td>
                <td>{{ row.HR || '-' }} : {{ row.AR || '-' }}</td>
              </tr>
            </tbody>
          </table>
          <NEmpty v-else description="没有符合条件的赛果" class="fw-empty" />
        </div>
      </NSpin>
    </section>
  </div>
</template>