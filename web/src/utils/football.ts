import { teamNameZh } from '@/constants/team-names';
import type { MatchRow } from '@/service/api/football';

function teamNameKey(value: string): string {
  return value.toLocaleLowerCase().replace(/[^a-z0-9\u4e00-\u9fff]/g, '');
}

const normalizedTeamNameZh: Record<string, string> = Object.fromEntries(
  Object.entries(teamNameZh).map(([name, label]) => [teamNameKey(name), label])
);

/** 队名统一展示为中文；已是中文或未收录时原样返回 */
export function displayTeamName(value?: string | null): string {
  const name = String(value ?? '').trim();
  if (!name) return '';
  return teamNameZh[name] ?? normalizedTeamNameZh[teamNameKey(name)] ?? name;
}

/** 搜索文本同时包含英文原名与中文队名，便于两种关键字检索 */
export function teamSearchText(...values: (string | null | undefined)[]): string {
  return values.map(value => `${String(value ?? '')} ${displayTeamName(value)}`).join(' ');
}

const matchFieldLabels: Record<string, string> = {
  Div: '联赛代码',
  Date: '比赛日期',
  Time: '开赛时间',
  HomeTeam: '主队',
  AwayTeam: '客队',
  FTHG: '全场主队进球',
  FTAG: '全场客队进球',
  FTR: '全场赛果',
  HTHG: '半场主队进球',
  HTAG: '半场客队进球',
  HTR: '半场赛果',
  Referee: '主裁判',
  HxG: '主队预期进球',
  AxG: '客队预期进球',
  HS: '主队射门',
  AS: '客队射门',
  HST: '主队射正',
  AST: '客队射正',
  HF: '主队犯规',
  AF: '客队犯规',
  HC: '主队角球',
  AC: '客队角球',
  HY: '主队黄牌',
  AY: '客队黄牌',
  HR: '主队红牌',
  AR: '客队红牌',
  AHh: '亚洲让球盘口',
  AHCh: '收盘亚洲让球盘口'
};

const matchResultLabels: Record<string, string> = { H: '主胜', D: '平局', A: '客胜' };

/** 比赛详情中的字段名统一为中文说明 */
export function displayMatchFieldLabel(key: string): string {
  return matchFieldLabels[key] ?? `赔率指标 ${key}`;
}

/** 比赛详情中的队名、赛果与日期统一为中文展示 */
export function displayMatchFieldValue(key: string, value?: string): string {
  if (key === 'HomeTeam' || key === 'AwayTeam') return displayTeamName(value) || '-';
  if (key === 'FTR' || key === 'HTR') return matchResultLabels[value ?? ''] ?? value ?? '-';
  if (key === 'Date') return displayDate(value);
  return value || '-';
}

export interface MatchSummary {
  games: number;
  homeWins: number;
  draws: number;
  awayWins: number;
  homeRate: number;
  drawRate: number;
  awayRate: number;
  averageGoals: number;
  over25Rate: number;
  averageCorners: number;
  averageCards: number;
}

function numericValue(row: MatchRow, key: string): number | null {
  const value = Number(row[key]);
  return row[key]?.trim() && Number.isFinite(value) ? value : null;
}

export function summarizeMatches(rows: MatchRow[]): MatchSummary {
  const resultRows = rows.filter(row => ['H', 'D', 'A'].includes(row.FTR));
  const homeWins = resultRows.filter(row => row.FTR === 'H').length;
  const draws = resultRows.filter(row => row.FTR === 'D').length;
  const awayWins = resultRows.filter(row => row.FTR === 'A').length;
  const playedRows = rows.filter(row => numericValue(row, 'FTHG') !== null && numericValue(row, 'FTAG') !== null);
  const goalTotal = playedRows.reduce((total, row) => total + (numericValue(row, 'FTHG') ?? 0) + (numericValue(row, 'FTAG') ?? 0), 0);
  const over25 = playedRows.filter(row => (numericValue(row, 'FTHG') ?? 0) + (numericValue(row, 'FTAG') ?? 0) > 2.5).length;
  const cornerRows = rows.filter(row => numericValue(row, 'HC') !== null && numericValue(row, 'AC') !== null);
  const cardRows = rows.filter(row => numericValue(row, 'HY') !== null && numericValue(row, 'AY') !== null);
  const rate = (count: number) => resultRows.length ? (count / resultRows.length) * 100 : 0;

  return {
    games: rows.length,
    homeWins,
    draws,
    awayWins,
    homeRate: rate(homeWins),
    drawRate: rate(draws),
    awayRate: rate(awayWins),
    averageGoals: playedRows.length ? goalTotal / playedRows.length : 0,
    over25Rate: playedRows.length ? (over25 / playedRows.length) * 100 : 0,
    averageCorners: cornerRows.length
      ? cornerRows.reduce((total, row) => total + (numericValue(row, 'HC') ?? 0) + (numericValue(row, 'AC') ?? 0), 0) / cornerRows.length
      : 0,
    averageCards: cardRows.length
      ? cardRows.reduce((total, row) => total + (numericValue(row, 'HY') ?? 0) + (numericValue(row, 'AY') ?? 0), 0) / cardRows.length
      : 0
  };
}

export function sortMatchesNewest(rows: MatchRow[]): MatchRow[] {
  return [...rows].sort((a, b) => parseMatchDate(b.Date) - parseMatchDate(a.Date));
}

function parseMatchDate(value = ''): number {
  const parts = value.split(/[/-]/).map(Number);
  if (parts.length !== 3 || parts.some(part => !Number.isFinite(part))) return 0;
  const [first, second, third] = parts;
  const [year, month, day] = first > 31 ? [first, second, third] : [third, second, first];
  return new Date(year, month - 1, day).getTime();
}

export function displayDate(value = ''): string {
  const parts = value.split(/[/-]/);
  if (parts.length !== 3) return value || '-';
  if (parts[0].length === 4) return `${parts[0]}-${parts[1].padStart(2, '0')}-${parts[2].padStart(2, '0')}`;
  return `${parts[2]}-${parts[1].padStart(2, '0')}-${parts[0].padStart(2, '0')}`;
}

export function csvCell(value: unknown): string {
  const text = String(value ?? '');
  return /[",\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
}

export function downloadCsv(fileName: string, rows: unknown[][]): void {
  const content = `\ufeff${rows.map(row => row.map(csvCell).join(',')).join('\r\n')}`;
  const url = URL.createObjectURL(new Blob([content], { type: 'text/csv;charset=utf-8' }));
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = fileName;
  anchor.click();
  URL.revokeObjectURL(url);
}