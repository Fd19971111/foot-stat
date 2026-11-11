<script setup lang="ts">
import { computed, h, ref } from 'vue';
import { NDrawer, NDrawerContent, NEmpty, NSelect, NSpin, NTag } from 'naive-ui';
import {
  displayDate,
  displayMatchFieldLabel,
  displayMatchFieldValue,
  displayTeamName,
  sortMatchesNewest,
  summarizeMatches,
  teamSearchText
} from '@/utils/football';
import type { MatchRow } from '@/service/api/football';
import { useFootballData } from '@/hooks/business/useFootballData';

defineOptions({ name: 'MatchCenter' });

const { season, league, month, rows, loading, seasonOptions, leagueOptions, monthOptions } = useFootballData();
const search = ref('');
const selected = ref<MatchRow | null>(null);
const summary = computed(() => summarizeMatches(rows.value));
const filteredRows = computed(() => {
  const keyword = search.value.trim().toLocaleLowerCase();
  const matches = keyword
    ? rows.value.filter(row => `${teamSearchText(row.HomeTeam, row.AwayTeam)} ${row.Date}`.toLocaleLowerCase().includes(keyword))
    : rows.value;
  return sortMatchesNewest(matches).slice(0, 12);
});

const resultInfo: Record<string, { label: string; type: 'success' | 'warning' | 'error' }> = {
  H: { label: '主胜', type: 'success' },
  D: { label: '平局', type: 'warning' },
  A: { label: '客胜', type: 'error' }
};

function resultTag(row: MatchRow) {
  const result = resultInfo[row.FTR];
  return result
    ? h(NTag, { size: 'small', bordered: false, type: result.type }, { default: () => result.label })
    : h(NTag, { size: 'small', bordered: false }, { default: () => '未完赛' });
}

function handleDrawerVisibility(visible: boolean) {
  if (!visible) selected.value = null;
}
</script>

<template>
  <div class="football-workspace">
    <header class="fw-heading">
      <div>
        <span class="fw-eyebrow">足球 / 比赛中心</span>
        <h2>比赛中心</h2>
        <p>从本地赛季数据快速查看赛果、进球、角球和牌数。</p>
      </div>
      <div class="fw-heading-actions">
        <label class="fw-field">
          赛季
          <NSelect v-model:value="season" :options="seasonOptions" />
        </label>
        <label class="fw-field">
          联赛
          <NSelect v-model:value="league" :options="leagueOptions" />
        </label>
        <label class="fw-field">
          月份
          <NSelect v-model:value="month" :options="monthOptions" />
        </label>
      </div>
    </header>

    <section class="fw-hero">
      <div class="fw-hero-copy">
        <span>本地赛果数据库 · {{ seasonOptions.find(item => item.value === season)?.label ?? '加载中' }}</span>
        <h2>{{ leagueOptions.find(item => item.value === league)?.label ?? '选择联赛' }}</h2>
        <p>统计来自 data 目录 CSV，未完赛数据不计入赛果率。</p>
      </div>
      <div class="fw-hero-stats">
        <div>
          <span>当前筛选场次</span>
          <strong>{{ summary.games.toLocaleString() }}</strong>
        </div>
        <div>
          <span>已完赛</span>
          <strong>{{ (summary.homeWins + summary.draws + summary.awayWins).toLocaleString() }}</strong>
        </div>
        <div>
          <span>主胜比例</span>
          <strong>{{ summary.homeRate.toFixed(1) }}%</strong>
        </div>
      </div>
    </section>

    <section class="fw-stat-grid">
      <article class="fw-stat">
        <span>主胜 / 平 / 客胜</span>
        <strong>{{ summary.homeWins }} / {{ summary.draws }} / {{ summary.awayWins }}</strong>
        <em>赛果分布</em>
      </article>
      <article class="fw-stat">
        <span>场均进球</span>
        <strong>
          {{ summary.averageGoals.toFixed(2) }}
          <small>球</small>
        </strong>
        <em>全场总进球</em>
      </article>
      <article class="fw-stat">
        <span>大于 2.5 球</span>
        <strong>
          {{ summary.over25Rate.toFixed(1) }}
          <small>%</small>
        </strong>
        <em>完赛样本</em>
      </article>
      <article class="fw-stat">
        <span>场均角球 / 黄牌</span>
        <strong>{{ summary.averageCorners.toFixed(1) }} / {{ summary.averageCards.toFixed(1) }}</strong>
        <em>CSV 可用字段</em>
      </article>
    </section>

    <section class="fw-panel">
      <div class="fw-panel-head">
        <div>
          <h3>比赛记录</h3>
          <p>最近日期优先 · 当前显示 {{ filteredRows.length }} / {{ rows.length }} 场</p>
        </div>
        <div class="fw-filters"><input v-model="search" class="fw-search" placeholder="搜索主队、客队或日期" /></div>
      </div>
      <NSpin :show="loading">
        <div v-if="filteredRows.length" class="fw-match-list">
          <article
            v-for="(row, index) in filteredRows"
            :key="`${row.Date}-${row.HomeTeam}-${index}`"
            class="fw-match-row"
          >
            <div class="fw-match-date">
              <strong>
                {{ displayDate(row.Date) }}
                <span v-if="row.Time">· {{ row.Time }}</span>
              </strong>
              <span>{{ leagueOptions.find(item => item.value === league)?.label ?? league }}</span>
            </div>
            <div class="fw-match-teams">
              <strong>{{ displayTeamName(row.HomeTeam) }}</strong>
              <b class="fw-score-pill">{{ row.FTHG || '-' }} : {{ row.FTAG || '-' }}</b>
              <strong>{{ displayTeamName(row.AwayTeam) }}</strong>
            </div>
            <div class="fw-summary-line">
              <NTag size="small" :bordered="false" :type="resultInfo[row.FTR]?.type ?? 'default'">
                {{ resultInfo[row.FTR]?.label ?? '未完赛' }}
              </NTag>
              <span class="fw-tag">角球 {{ row.HC || 0 }} : {{ row.AC || 0 }}</span>
              <span class="fw-tag">黄牌 {{ row.HY || 0 }} : {{ row.AY || 0 }}</span>
            </div>
            <button class="fw-link-button" @click="selected = row">全部数据</button>
          </article>
        </div>
        <NEmpty v-else description="没有符合条件的比赛" class="fw-empty" />
      </NSpin>
    </section>

    <NDrawer :show="Boolean(selected)" :width="520" placement="right" @update:show="handleDrawerVisibility">
      <NDrawerContent v-if="selected" :title="`${displayTeamName(selected.HomeTeam)} 对 ${displayTeamName(selected.AwayTeam)}`" closable>
        <p class="fw-panel-head">
          {{ displayDate(selected.Date) }} {{ selected.Time }} ·
          {{ leagueOptions.find(item => item.value === league)?.label ?? league }}
        </p>
        <div class="fw-drawer-grid">
          <div v-for="(value, key) in selected" :key="key" class="fw-drawer-item">
            <span>{{ displayMatchFieldLabel(String(key)) }}</span>
            <strong>{{ displayMatchFieldValue(String(key), value) }}</strong>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>
