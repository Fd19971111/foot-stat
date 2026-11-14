<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import * as XLSX from 'xlsx';
import { createFootballImport, fetchFootballImportDetails, fetchFootballImports, fetchFootballRemoteMatches } from '@/service/api/football';
import type { ImportBatch, ImportMatch, MatchPrediction } from '@/service/api/football';
import { displayTeamName } from '@/utils/football';

defineOptions({ name: 'FootballImport' });

const file = ref<File | null>(null);
const matches = ref<ImportMatch[]>([]);
const batches = ref<ImportBatch[]>([]);
const errors = ref<string[]>([]);
const parsing = ref(false);
const fetching = ref(false);
const saving = ref(false);
const sourceName = ref('');
const activeBatch = ref<ImportBatch | null>(null);
const predictions = ref<MatchPrediction[]>([]);

const fieldAliases: Record<string, string[]> = {
  date: ['date', '比赛日期', '日期'],
  time: ['time', '开赛时间', '时间'],
  league: ['league', 'div', '联赛'],
  homeTeam: ['hometeam', '主队'],
  awayTeam: ['awayteam', '客队'],
  homeOdds: ['homeodds', 'b365h', 'avgh', '主胜赔率', '主胜', '主队赔率'],
  drawOdds: ['drawodds', 'b365d', 'avgd', '平局赔率', '平赔', '和局赔率'],
  awayOdds: ['awayodds', 'b365a', 'avga', '客胜赔率', '客胜', '客队赔率'],
  homeGoals: ['fthg', 'homegoals', '主队进球', '主队得分', '主进球'],
  awayGoals: ['ftag', 'awaygoals', '客队进球', '客队得分', '客进球']
};

function normalizedKey(value: unknown): string {
  return String(value ?? '').replace(/[\s_\-]/g, '').toLocaleLowerCase();
}

function findCell(row: Record<string, unknown>, field: string): string {
  const accepted = (fieldAliases[field] ?? []).map(normalizedKey);
  const key = Object.keys(row).find(header => accepted.includes(normalizedKey(header)));
  return key ? String(row[key] ?? '').trim() : '';
}

function oddsCell(row: Record<string, unknown>, field: string): number | undefined {
  const value = Number(findCell(row, field));
  return Number.isFinite(value) && value > 1 && value <= 1000 ? value : undefined;
}

function goalsCell(row: Record<string, unknown>, field: string): number | null {
  const text = findCell(row, field);
  if (!text) return null;
  const value = Number(text);
  return Number.isInteger(value) && value >= 0 && value <= 20 ? value : null;
}

function normalizeDate(value: string): string {
  const text = value.trim();
  if (/^\d{4}-\d{1,2}-\d{1,2}$/.test(text)) return text;
  if (/^\d{1,2}\/\d{1,2}\/\d{4}$/.test(text)) {
    const [day, month, year] = text.split('/');
    return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
  }
  const excelDate = XLSX.SSF.parse_date_code(Number(text));
  if (excelDate) return `${excelDate.y}-${String(excelDate.m).padStart(2, '0')}-${String(excelDate.d).padStart(2, '0')}`;
  return text;
}

function downloadTemplate() {
  const worksheet = XLSX.utils.aoa_to_sheet([
    ['比赛日期', '开赛时间', '联赛', '主队', '客队', '主胜赔率', '平局赔率', '客胜赔率', '主队进球', '客队进球'],
    ['2026-11-15', '20:30', '英超', '阿森纳', '切尔西', '2.10', '3.40', '3.30', '', ''],
    ['2026-11-08', '23:00', '英超', '曼城', '利物浦', '2.05', '3.50', '3.20', '2', '1']
  ]);
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, '比赛');
  XLSX.writeFile(workbook, '比赛导入模板.xlsx');
}

async function loadBatches() {
  const { data, error } = await fetchFootballImports();
  if (error || !data) return;
  batches.value = data.batches ?? [];
}

async function parseFile(selectedFile: File | null) {
  file.value = selectedFile;
  matches.value = [];
  errors.value = [];
  sourceName.value = selectedFile?.name ?? '';
  if (selectedFile) {
    activeBatch.value = null;
    predictions.value = [];
  }
  if (!selectedFile) return;
  if (selectedFile.size > 10 * 1024 * 1024) {
    errors.value = ['文件不能超过 10 MB'];
    return;
  }
  parsing.value = true;
  try {
    const buffer = await selectedFile.arrayBuffer();
    const workbook = XLSX.read(buffer, { type: 'array', cellDates: true });
    const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
    const sourceRows = XLSX.utils.sheet_to_json<Record<string, unknown>>(firstSheet, { defval: '', raw: false, dateNF: 'yyyy-mm-dd' });
    if (!sourceRows.length) {
      errors.value = ['文件没有比赛数据'];
      return;
    }
    const parsed: ImportMatch[] = [];
    const fixtures = new Set<string>();
    sourceRows.forEach((row, index) => {
      const match: ImportMatch = {
        date: normalizeDate(findCell(row, 'date')),
        time: findCell(row, 'time'),
        league: findCell(row, 'league'),
        homeTeam: findCell(row, 'homeTeam'),
        awayTeam: findCell(row, 'awayTeam')
      };
      const homeOdds = oddsCell(row, 'homeOdds');
      const drawOdds = oddsCell(row, 'drawOdds');
      const awayOdds = oddsCell(row, 'awayOdds');
      if (homeOdds !== undefined && drawOdds !== undefined && awayOdds !== undefined) {
        match.homeOdds = homeOdds;
        match.drawOdds = drawOdds;
        match.awayOdds = awayOdds;
      }
      const homeGoals = goalsCell(row, 'homeGoals');
      const awayGoals = goalsCell(row, 'awayGoals');
      if (homeGoals !== null && awayGoals !== null) {
        match.homeGoals = homeGoals;
        match.awayGoals = awayGoals;
      } else if (homeGoals !== null || awayGoals !== null) {
        errors.value.push(`第 ${index + 2} 行的主队进球与客队进球须同时提供`);
        return;
      }
      const missing = (['date', 'league', 'homeTeam', 'awayTeam'] as const).filter(field => !match[field]);
      if (missing.length) {
        errors.value.push(`第 ${index + 2} 行缺少${missing.map(field => ({ date: '比赛日期', league: '联赛', homeTeam: '主队', awayTeam: '客队' })[field]).join('、')}`);
        return;
      }
      if (match.homeTeam === match.awayTeam) {
        errors.value.push(`第 ${index + 2} 行主队和客队不能相同`);
        return;
      }
      const fixtureKey = [match.date, match.time, match.league, match.homeTeam, match.awayTeam].join('|').toLocaleLowerCase();
      if (fixtures.has(fixtureKey)) {
        errors.value.push(`第 ${index + 2} 行与文件中的其他比赛重复`);
        return;
      }
      fixtures.add(fixtureKey);
      parsed.push(match);
    });
    if (sourceRows.length > 500) errors.value.push('单次最多导入 500 场比赛');
    matches.value = parsed.slice(0, 500);
  } catch {
    errors.value = ['无法解析文件，请检查 CSV 或 Excel 格式'];
  } finally {
    parsing.value = false;
  }
}

async function fetchRemoteMatches() {
  if (fetching.value) return;
  fetching.value = true;
  file.value = null;
  matches.value = [];
  errors.value = [];
  sourceName.value = '';
  activeBatch.value = null;
  predictions.value = [];
  const { data, error } = await fetchFootballRemoteMatches();
  fetching.value = false;
  if (error || !data) {
    errors.value = ['拉取中国竞彩网赛程失败，请稍后重试'];
    return;
  }
  sourceName.value = `竞彩网 ${data.fetchedAt.slice(0, 10)}`;
  matches.value = (data.matches ?? []).slice(0, 500).map(item => ({
    date: item.date,
    time: item.time,
    league: item.league,
    homeTeam: item.homeTeam,
    awayTeam: item.awayTeam,
    ...(item.homeOdds > 1 && item.drawOdds > 1 && item.awayOdds > 1
      ? { homeOdds: item.homeOdds, drawOdds: item.drawOdds, awayOdds: item.awayOdds }
      : {})
  }));
  if (!matches.value.length) {
    window.$message?.info('中国竞彩网当前没有在售比赛');
    return;
  }
  window.$message?.success(`已从中国竞彩网获取 ${matches.value.length} 场比赛`);
}

async function saveImport() {
  if (!matches.value.length || errors.value.length) return;
  saving.value = true;
  const { data, error } = await createFootballImport({ fileName: sourceName.value || '手动导入.xlsx', matches: matches.value });
  saving.value = false;
  if (error || !data) return;
  activeBatch.value = data.batch;
  predictions.value = data.predictions ?? [];
  window.$message?.success(`已完成 ${data.batch.matchCount} 场比赛分析`);
  file.value = null;
  sourceName.value = '';
  matches.value = [];
  errors.value = [];
  await loadBatches();
}

async function showBatchPredictions(batch: ImportBatch) {
  const { data, error } = await fetchFootballImportDetails(batch.id);
  if (error || !data) return;
  activeBatch.value = data.batch;
  predictions.value = data.predictions ?? [];
}

const marketCount = computed(() => predictions.value.filter(item => item.marketUsed).length);
const historyCount = computed(() => predictions.value.filter(item => item.historyMatches > 0).length);
const insufficientCount = computed(() => predictions.value.filter(item => item.dataQuality === 'insufficient').length);
const resultCount = computed(() => matches.value.filter(item => item.homeGoals !== undefined && item.homeGoals !== null).length);

onMounted(loadBatches);
</script>

<template>
  <div class="football-workspace">
    <header class="fw-heading">
      <div><span class="fw-eyebrow">足球 / 比赛导入</span><h2>比赛导入</h2><p>赛程自动带上胜平负赔率；带比分的历史 CSV 会累积为建模样本，让预测真正随对阵变化。</p></div>
      <div class="fw-heading-actions"><button class="fw-button primary" :disabled="fetching" @click="fetchRemoteMatches">{{ fetching ? '正在获取…' : '从竞彩网获取' }}</button><button class="fw-button" @click="downloadTemplate">下载导入模板</button></div>
    </header>

    <div class="fw-import-layout">
      <section class="fw-panel">
        <div class="fw-panel-head"><div><h3>获取比赛</h3><p>支持竞彩网一键拉取，也可上传 CSV / Excel · 最多 500 场</p></div></div>
        <div class="fw-panel-body">
          <label class="fw-upload-zone" @dragover.prevent @drop.prevent="parseFile($event.dataTransfer?.files[0] ?? null)">
            <input type="file" accept=".csv,.xlsx,.xls" @change="parseFile(($event.target as HTMLInputElement).files?.[0] ?? null)" />
            <span class="fw-upload-icon">↑</span><strong>{{ file?.name ?? (matches.length ? sourceName : '选择 CSV 或 Excel 文件') }}</strong><span>{{ parsing ? '正在解析文件…' : fetching ? '正在连接中国竞彩网…' : matches.length ? `已获取 ${matches.length} 场比赛` : '点击右上角从竞彩网获取，或拖放文件' }}</span><em>浏览文件</em>
          </label>
          <div v-if="errors.length" class="fw-notice" style="margin-top: 14px; border-color: rgb(185 74 72 / 28%); background: rgb(185 74 72 / 8%); color: var(--fw-red);">
            <strong>发现 {{ errors.length }} 项问题</strong><ul><li v-for="error in errors.slice(0, 8)" :key="error">{{ error }}</li></ul><span v-if="errors.length > 8">其余 {{ errors.length - 8 }} 项未显示</span>
          </div>
          <div v-else-if="matches.length" class="fw-notice" style="margin-top: 14px;">已就绪：{{ matches.length }} 场比赛（含 {{ resultCount }} 场已完赛历史赛果），核对下方表格后即可导入。</div>
          <div class="fw-actions" style="margin-top: 14px;"><span style="color: var(--fw-muted); font-size: 11px;">数据来源：中国竞彩网在售赛程 / 本地 CSV、Excel</span><span class="fw-tag">导入前可核对表格</span></div>
        </div>
      </section>
      <aside class="fw-panel"><div class="fw-panel-head"><div><h3>使用说明</h3><p>推荐使用竞彩网一键拉取</p></div></div><div class="fw-guide"><div class="fw-guide-row"><b>01</b><span><strong>获取赛程</strong><p>点击“从竞彩网获取”，自动拉取当前在售比赛。</p></span></div><div class="fw-guide-row"><b>02</b><span><strong>核对表格</strong><p>确认日期、联赛、主客队与开赛时间。</p></span></div><div class="fw-guide-row"><b>03</b><span><strong>导入并分析</strong><p>赛程按赔率校准；上传带「主队进球 / 客队进球」的 CSV 可累积历史样本。</p></span></div><div class="fw-guide-row"><b>04</b><span><strong>补齐历史再分析</strong><p>可直接把 data/ 目录下带 FTHG、FTAG 的历史 CSV 拖进来，赛果会跨批次累积为建模样本；竞彩赛程本身不含历史。</p></span></div></div></aside>
    </div>

    <section v-if="matches.length" class="fw-panel">
      <div class="fw-panel-head"><div><h3>待导入比赛</h3><p>{{ sourceName }} · {{ matches.length }} 场 · 确认无误后导入</p></div><div class="fw-filters"><span v-if="sourceName.includes('竞彩网')" class="fw-tag fw-result-home">竞彩网数据</span><button class="fw-button primary" :disabled="saving || errors.length > 0" @click="saveImport">{{ saving ? '导入中…' : '确认导入并分析' }}</button></div></div>
      <div class="fw-table-wrap">
        <table class="fw-table">
          <thead><tr><th>比赛</th><th>日期</th><th>开赛时间</th><th>联赛</th><th>胜平负赔率</th><th>全场比分</th></tr></thead>
          <tbody>
            <tr v-for="(match, index) in matches" :key="`${match.date}-${match.homeTeam}-${index}`">
              <td><strong>{{ displayTeamName(match.homeTeam) }} 对 {{ displayTeamName(match.awayTeam) }}</strong></td>
              <td>{{ match.date }}</td>
              <td>{{ match.time || '-' }}</td>
              <td>{{ match.league }}</td>
              <td>{{ match.homeOdds ? `${match.homeOdds} / ${match.drawOdds} / ${match.awayOdds}` : '自动补齐' }}</td>
              <td>{{ match.homeGoals !== undefined && match.homeGoals !== null ? `${match.homeGoals} - ${match.awayGoals}` : '待赛' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="activeBatch" class="fw-panel">
      <div class="fw-panel-head"><div><h3>自动预测报告</h3><p>{{ activeBatch.fileName }} · {{ predictions.length }} 场 · 赔率校准 {{ marketCount }} 场 · 有历史样本 {{ historyCount }} 场</p></div><div class="fw-filters"><span class="fw-tag">预测仅供数据分析参考</span><RouterLink class="fw-link-button" to="/analysis">查看完整分析</RouterLink></div></div>
      <div v-if="insufficientCount" class="fw-notice" style="margin: 0 18px 14px; border-color: rgb(167 107 23 / 32%); background: rgb(215 164 74 / 12%); color: var(--fw-amber);">
        <strong>数据不足提示</strong>：{{ insufficientCount }} 场比赛既没有历史赛果，也没有胜平负赔率，预测只能退化为通用先验，结果会高度相似。建议导入带比分的历史 CSV，或导入带赔率的赛程后重新查看预测。
      </div>
      <div class="fw-table-wrap"><table class="fw-table"><thead><tr><th>比赛</th><th>最可能比分</th><th>预期进球</th><th>综合概率</th><th>大 2.5 球</th><th>数据依据</th></tr></thead><tbody>
        <tr v-for="(prediction, index) in predictions" :key="`${prediction.matchDate}-${prediction.homeTeam}-${index}`">
          <td><strong>{{ displayTeamName(prediction.homeTeam) }} - {{ displayTeamName(prediction.awayTeam) }}</strong><small>{{ prediction.matchDate }} {{ prediction.kickoffTime }} · {{ prediction.league }}</small></td>
          <td><strong>{{ prediction.predictedScore }}</strong></td>
          <td>{{ prediction.expectedHomeGoals.toFixed(2) }} - {{ prediction.expectedAwayGoals.toFixed(2) }}</td>
          <td>
            <div class="fw-prob-bar"><i class="fw-prob-home" :style="{ width: `${prediction.homeWinProbability}%` }" /><i class="fw-prob-draw" :style="{ width: `${prediction.drawProbability}%` }" /><i class="fw-prob-away" :style="{ width: `${prediction.awayWinProbability}%` }" /></div>
            <div class="fw-prob-labels"><span>主 <b>{{ prediction.homeWinProbability }}%</b></span><span>平 <b>{{ prediction.drawProbability }}%</b></span><span>客 <b>{{ prediction.awayWinProbability }}%</b></span></div>
          </td>
          <td>{{ prediction.over25Probability }}%</td><td class="fw-prediction-basis">{{ prediction.modelBasis }}</td>
        </tr>
      </tbody></table></div>
    </section>

    <section class="fw-panel"><div class="fw-panel-head"><div><h3>最近导入记录</h3><p>本地数据库已保存批次 · 最近 100 条</p></div></div>
      <div class="fw-table-wrap"><table v-if="batches.length" class="fw-table"><thead><tr><th>文件名称</th><th>导入时间</th><th>比赛数量</th><th>状态</th><th>操作</th></tr></thead><tbody><tr v-for="batch in batches" :key="batch.id"><td><strong>{{ batch.fileName }}</strong></td><td>{{ batch.importedAt }}</td><td>{{ batch.matchCount }} 场</td><td><span class="fw-tag fw-result-home">已分析</span></td><td><button class="fw-link-button" @click="showBatchPredictions(batch)">查看预测</button></td></tr></tbody></table><div v-else class="fw-empty"><strong>还没有导入记录</strong><span>完成首次导入后，预测结果会保存在这里。</span></div></div>
    </section>
  </div>
</template>