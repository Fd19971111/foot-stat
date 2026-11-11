import type { MatchRow } from '@/service/api/football';

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