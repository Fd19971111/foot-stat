<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { NDrawer, NDrawerContent, NEmpty, NSelect, NSpin } from 'naive-ui';
import { fetchFootballAnalysis, fetchFootballImports } from '@/service/api/football';
import type { AnalysisProbability, AnalysisSummary, ImportBatch, MatchAnalysis } from '@/service/api/football';
import { displayDate, displayTeamName, downloadCsv, teamSearchText } from '@/utils/football';

defineOptions({ name: 'AnalysisResults' });

const batches = ref<ImportBatch[]>([]);
const batchId = ref<number | null>(null);
const matches = ref<MatchAnalysis[]>([]);
const summary = ref<AnalysisSummary | null>(null);
const loading = ref(false);
const selected = ref<MatchAnalysis | null>(null);
const search = ref('');
const confidenceFilter = ref<'all' | '高' | '中' | '低'>('all');

const batchOptions = computed(() => batches.value.map(item => ({ label: `${item.fileName} · ${item.matchCount} 场`, value: item.id })));
const confidenceOptions = [
  { label: '全部信心', value: 'all' },
  { label: '高信心', value: '高' },
  { label: '中信心', value: '中' },
  { label: '低信心', value: '低' }
] as const;

const filteredMatches = computed(() => {
  const keyword = search.value.trim().toLocaleLowerCase();
  return matches.value.filter(item => {
    const hasKeyword = !keyword || `${teamSearchText(item.homeTeam, item.awayTeam)} ${item.league}`.toLocaleLowerCase().includes(keyword);
    const hasConfidence = confidenceFilter.value === 'all' || item.confidenceLevel === confidenceFilter.value;
    return hasKeyword && hasConfidence;
  });
});

const lean = computed(() => {
  const counts = { home: 0, draw: 0, away: 0, over: 0, under: 0 };
  matches.value.forEach(item => {
    counts[item.recommendationType] += 1;
    counts[item.goalsType] += 1;
  });
  return counts;
});

const selectedMaxScore = computed(() => Math.max(...(selected.value?.scoreLines.map(item => item.probability) ?? [1])));

const confidenceClass: Record<string, string> = {
  高: 'fw-confidence-high',
  中: 'fw-confidence-mid',
  低: 'fw-confidence-low'
};

const adviceClass: Record<string, string> = {
  home: 'fw-advice-home',
  draw: 'fw-advice-draw',
  away: 'fw-advice-away'
};

function probabilityRows(probability: AnalysisProbability) {
  return [
    { label: '主胜', value: probability.homeWin, className: 'fw-prob-home' },
    { label: '平局', value: probability.draw, className: 'fw-prob-draw' },
    { label: '客胜', value: probability.awayWin, className: 'fw-prob-away' }
  ];
}

function formSequence(sequence: string) {
  return sequence.split('').map(result => ({
    result,
    label: { W: '胜', D: '平', L: '负' }[result] ?? result,
    className: { W: 'fw-form-win', D: 'fw-form-draw', L: 'fw-form-loss' }[result] ?? ''
  }));
}

function scoreWidth(probability: number) {
  return `${Math.max(4, (probability / selectedMaxScore.value) * 100)}%`;
}

async function loadBatches() {
  const { data, error } = await fetchFootballImports();
  if (error || !data) {
    window.$message?.error('导入记录加载失败');
    return;
  }
  batches.value = data.batches ?? [];
  batchId.value = batches.value[0]?.id ?? null;
}

async function loadAnalysis(id: number) {
  loading.value = true;
  const { data, error } = await fetchFootballAnalysis(id);
  loading.value = false;
  if (error || !data) {
    matches.value = [];
    summary.value = null;
    window.$message?.error('分析结果加载失败');
    return;
  }
  matches.value = data.matches ?? [];
  summary.value = data.summary;
}

function exportAnalysis() {
  if (!filteredMatches.value.length) return;
  downloadCsv('football-analysis-result.csv', [
    ['日期', '时间', '联赛', '主队', '客队', '推荐', '最可能比分', '主胜(%)', '平局(%)', '客胜(%)', '大于2.5球(%)', '双方进球(%)', '信心'],
    ...filteredMatches.value.map(item => [
      displayDate(item.matchDate),
      item.kickoffTime,
      item.league,
      displayTeamName(item.homeTeam),
      displayTeamName(item.awayTeam),
      item.recommendation,
      item.predictedScore,
      item.final.homeWin.toFixed(1),
      item.final.draw.toFixed(1),
      item.final.awayWin.toFixed(1),
      item.over25Probability.toFixed(1),
      item.bttsProbability.toFixed(1),
      `${item.confidenceLevel} ${item.confidence}`
    ])
  ]);
}

watch(batchId, id => {
  selected.value = null;
  if (id) {
    loadAnalysis(id);
  } else {
    matches.value = [];
    summary.value = null;
  }
});

onMounted(loadBatches);
</script>

<template>
  <div class="football-workspace">
    <header class="fw-heading">
      <div>
        <span class="fw-eyebrow">足球 / 分析结果</span>
        <h2>分析结果</h2>
        <p>用 Dixon-Coles 泊松模型、Elo 等级分和近期状态，对导入的比赛给出综合判断。</p>
      </div>
      <div class="fw-heading-actions">
        <label class="fw-field">导入批次<NSelect v-model:value="batchId" :options="batchOptions" placeholder="选择导入批次" /></label>
        <button class="fw-button" :disabled="!filteredMatches.length" @click="exportAnalysis">导出结果</button>
      </div>
    </header>

    <template v-if="batches.length">
      <section class="fw-hero">
        <div class="fw-hero-copy">
          <span>综合分析报告 · {{ summary?.algorithm ?? 'Dixon-Coles 泊松 + Elo' }}</span>
          <h2>{{ batches.find(item => item.id === batchId)?.fileName ?? '选择导入批次' }}</h2>
          <p>由历史赛果与胜平负赔率共同推算，仅供数据分析参考；历史样本越充分，信心越高。</p>
        </div>
        <div class="fw-hero-stats">
          <div><span>分析场次</span><strong>{{ summary?.matchCount ?? 0 }}</strong></div>
          <div><span>平均信心</span><strong>{{ (summary?.averageConfidence ?? 0).toFixed(0) }}</strong></div>
          <div><span>高信心场次</span><strong>{{ summary?.highConfidence ?? 0 }}</strong></div>
        </div>
      </section>

      <section class="fw-result-grid">
        <article class="fw-stat"><span>主胜倾向</span><strong>{{ lean.home }}<small>场</small></strong><em>综合推荐</em></article>
        <article class="fw-stat"><span>平局倾向</span><strong>{{ lean.draw }}<small>场</small></strong><em>综合推荐</em></article>
        <article class="fw-stat"><span>客胜倾向</span><strong>{{ lean.away }}<small>场</small></strong><em>综合推荐</em></article>
        <article class="fw-stat"><span>大于 2.5 球</span><strong>{{ lean.over }}<small>场</small></strong><em>进球模型</em></article>
      </section>

      <section class="fw-panel">
        <div class="fw-panel-head fw-toolbar">
          <div><h3>逐场分析</h3><p>共 {{ matches.length }} 场 · 当前显示 {{ filteredMatches.length }} 场 · 点击右侧查看模型明细</p></div>
          <div class="fw-filters">
            <input v-model="search" class="fw-search" placeholder="搜索球队或联赛" />
            <label class="fw-field">信心筛选<NSelect v-model:value="confidenceFilter" :options="[...confidenceOptions]" /></label>
          </div>
        </div>
        <NSpin :show="loading">
          <div class="fw-table-wrap">
            <table v-if="filteredMatches.length" class="fw-table">
              <thead>
                <tr><th>比赛</th><th>综合概率</th><th>推荐</th><th>最可能比分</th><th>进球 / 双方进球</th><th>信心</th><th>操作</th></tr>
              </thead>
              <tbody>
                <tr v-for="(item, index) in filteredMatches" :key="`${item.matchDate}-${item.homeTeam}-${index}`">
                  <td>
                    <strong>{{ displayTeamName(item.homeTeam) }} 对 {{ displayTeamName(item.awayTeam) }}</strong>
                    <small class="fw-cell-sub">{{ displayDate(item.matchDate) }} {{ item.kickoffTime }} · {{ item.league }}</small>
                  </td>
                  <td>
                    <div class="fw-prob-bar">
                      <i class="fw-prob-home" :style="{ width: `${item.final.homeWin}%` }" />
                      <i class="fw-prob-draw" :style="{ width: `${item.final.draw}%` }" />
                      <i class="fw-prob-away" :style="{ width: `${item.final.awayWin}%` }" />
                    </div>
                    <div class="fw-prob-labels"><span>主 <b>{{ item.final.homeWin.toFixed(1) }}%</b></span><span>平 <b>{{ item.final.draw.toFixed(1) }}%</b></span><span>客 <b>{{ item.final.awayWin.toFixed(1) }}%</b></span></div>
                  </td>
                  <td><span class="fw-advice" :class="adviceClass[item.recommendationType]">{{ item.recommendation }}</span></td>
                  <td><strong>{{ item.predictedScore }}</strong><small class="fw-cell-sub">预期 {{ item.expectedHomeGoals.toFixed(2) }} - {{ item.expectedAwayGoals.toFixed(2) }}</small></td>
                  <td>
                    <div class="fw-outlook">
                      <span class="fw-tag" :class="item.goalsType === 'over' ? 'fw-result-home' : 'fw-result-pending'">{{ item.goalsAdvice }}</span>
                      <span class="fw-tag">{{ item.bttsAdvice }}</span>
                    </div>
                  </td>
                  <td>
                    <div class="fw-confidence">
                      <span class="fw-tag" :class="confidenceClass[item.confidenceLevel]">{{ item.confidenceLevel }}</span>
                      <div class="fw-progress"><i :style="{ width: `${item.confidence}%` }" /></div>
                      <small>{{ item.confidence }}</small>
                    </div>
                  </td>
                  <td><button class="fw-link-button" @click="selected = item">查看明细</button></td>
                </tr>
              </tbody>
            </table>
            <NEmpty v-else description="没有符合条件的分析结果" class="fw-empty" />
          </div>
        </NSpin>
      </section>
    </template>

    <section v-else class="fw-panel">
      <div class="fw-empty">
        <strong>还没有可分析的导入批次</strong>
        <span>先导入比赛，系统会在导入时保存比赛并生成分析结果。</span>
        <RouterLink class="fw-button primary" to="/import">去导入比赛</RouterLink>
      </div>
    </section>

    <NDrawer :show="Boolean(selected)" :width="560" placement="right" @update:show="visible => { if (!visible) selected = null; }">
      <NDrawerContent v-if="selected" :title="`${displayTeamName(selected.homeTeam)} 对 ${displayTeamName(selected.awayTeam)}`" closable>
        <div class="fw-drawer-content">
          <div class="fw-verdict">
            <div><span>综合推荐</span><strong class="fw-advice" :class="adviceClass[selected.recommendationType]">{{ selected.recommendation }}</strong></div>
            <div><span>最可能比分</span><strong>{{ selected.predictedScore }}</strong></div>
            <div><span>信心</span><strong>{{ selected.confidenceLevel }} · {{ selected.confidence }}</strong></div>
          </div>
          <p class="fw-drawer-note">{{ displayDate(selected.matchDate) }} {{ selected.kickoffTime }} · {{ selected.league }} · 预期进球 {{ selected.expectedHomeGoals.toFixed(2) }} - {{ selected.expectedAwayGoals.toFixed(2) }}</p>

          <h4 class="fw-drawer-title">模型概率对比</h4>
          <div class="fw-model-grid">
            <article v-for="model in [
              { title: '综合模型', probability: selected.final },
              { title: '泊松模型', probability: selected.poisson },
              { title: 'Elo 模型', probability: selected.elo }
            ]" :key="model.title" class="fw-model-card">
              <span>{{ model.title }}</span>
              <div v-for="row in probabilityRows(model.probability)" :key="row.label" class="fw-model-row">
                <em>{{ row.label }}</em>
                <div class="fw-progress"><i :class="row.className" :style="{ width: `${row.value}%` }" /></div>
                <strong>{{ row.value.toFixed(1) }}%</strong>
              </div>
            </article>
          </div>

          <h4 class="fw-drawer-title">进球与赔率参考</h4>
          <div class="fw-drawer-grid">
            <div class="fw-drawer-item"><span>大于 2.5 球</span><strong>{{ selected.over25Probability.toFixed(1) }}% · {{ selected.goalsAdvice }}</strong></div>
            <div class="fw-drawer-item"><span>双方进球</span><strong>{{ selected.bttsProbability.toFixed(1) }}% · {{ selected.bttsAdvice }}</strong></div>
            <div class="fw-drawer-item"><span>公平赔率（主 / 平 / 客）</span><strong>{{ selected.fairOdds.home.toFixed(2) }} / {{ selected.fairOdds.draw.toFixed(2) }} / {{ selected.fairOdds.away.toFixed(2) }}</strong></div>
            <div class="fw-drawer-item"><span>历史样本（主队 / 客队）</span><strong>{{ selected.homeTeamHistory }} 场 / {{ selected.awayTeamHistory }} 场</strong></div>
          </div>

          <h4 class="fw-drawer-title">最近五场状态</h4>
          <div class="fw-drawer-grid">
            <article v-for="side in [
              { team: displayTeamName(selected.homeTeam), form: selected.homeForm },
              { team: displayTeamName(selected.awayTeam), form: selected.awayForm }
            ]" :key="side.team" class="fw-form-card">
              <div class="fw-form-head"><strong>{{ side.team }}</strong><span>等级分 {{ side.form.elo.toFixed(0) }}</span></div>
              <div class="fw-form-sequence">
                <span v-for="(result, index) in formSequence(side.form.sequence)" :key="index" :class="result.className">{{ result.label }}</span>
                <em v-if="!side.form.matches">暂无赛前样本</em>
              </div>
              <p>{{ side.form.matches }} 场 · {{ side.form.wins }} 胜 {{ side.form.draws }} 平 {{ side.form.losses }} 负 · 进 {{ side.form.goalsFor }} 失 {{ side.form.goalsAgainst }}</p>
            </article>
          </div>

          <h4 class="fw-drawer-title">最可能比分</h4>
          <div class="fw-scorelines">
            <div v-for="line in selected.scoreLines" :key="line.score" class="fw-scoreline-row">
              <strong>{{ line.score }}</strong>
              <div class="fw-progress"><i :style="{ width: scoreWidth(line.probability) }" /></div>
              <span>{{ line.probability.toFixed(1) }}%</span>
            </div>
          </div>

          <p class="fw-drawer-note">{{ selected.modelBasis }}</p>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>
