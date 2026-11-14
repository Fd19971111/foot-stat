import { request } from '../request';

export interface SeasonItem {
  value: string;
  label: string;
  current: boolean;
}

export interface LeagueItem {
  value: string;
  label: string;
}

export interface MatchRow {
  [key: string]: string;
}

export interface MatchesResp {
  columns: string[];
  rows: MatchRow[];
}

export interface ImportMatch {
  date: string;
  time: string;
  league: string;
  homeTeam: string;
  awayTeam: string;
  /** 胜平负赔率，可缺省；系统会自动从竞彩网补齐 */
  homeOdds?: number;
  drawOdds?: number;
  awayOdds?: number;
  /** 已完赛的全场比分；提供后会进入历史样本池，供后续比赛建模 */
  homeGoals?: number | null;
  awayGoals?: number | null;
}

export interface ImportBatch {
  id: number;
  fileName: string;
  matchCount: number;
  importedAt: string;
}

export interface MatchPrediction {
  matchDate: string;
  kickoffTime: string;
  league: string;
  homeTeam: string;
  awayTeam: string;
  predictedScore: string;
  expectedHomeGoals: number;
  expectedAwayGoals: number;
  homeWinProbability: number;
  drawProbability: number;
  awayWinProbability: number;
  over25Probability: number;
  historyMatches: number;
  homeTeamHistory: number;
  awayTeamHistory: number;
  modelBasis: string;
  marketUsed: boolean;
  dataQuality: 'both' | 'market' | 'history' | 'insufficient' | string;
}

/** 获取可用赛季列表 */
export function fetchFootballSeasons() {
  return request<{ seasons: SeasonItem[] }>({ url: '/football/seasons' });
}

/** 获取某赛季下的联赛列表 */
export function fetchFootballLeagues(season: string) {
  return request<{ leagues: LeagueItem[] }>({ url: '/football/leagues', params: { season } });
}

/** 获取某赛季某联赛某月的比赛数据 */
export function fetchFootballMatches(params: {
  season: string;
  league: string;
  year: number;
  month: number;
}) {
  return request<MatchesResp>({ url: '/football/matches', params });
}

/** Save user-uploaded fixtures and list persisted import batches. */
export function createFootballImport(payload: { fileName: string; matches: ImportMatch[] }) {
  return request<{ batch: ImportBatch; predictions: MatchPrediction[] }>({ url: '/football/imports', method: 'post', data: payload });
}

export function fetchFootballImports() {
  return request<{ batches: ImportBatch[] }>({ url: '/football/imports' });
}

export function fetchFootballImportDetails(id: number) {
  return request<{ batch: ImportBatch; predictions: MatchPrediction[] }>({ url: `/football/imports/${id}` });
}

export interface AnalysisProbability {
  homeWin: number;
  draw: number;
  awayWin: number;
}

export interface AnalysisOdds {
  home: number;
  draw: number;
  away: number;
}

export interface AnalysisTeamForm {
  matches: number;
  wins: number;
  draws: number;
  losses: number;
  points: number;
  goalsFor: number;
  goalsAgainst: number;
  sequence: string;
  elo: number;
}

export interface AnalysisScore {
  score: string;
  probability: number;
}

export interface MatchAnalysis {
  matchDate: string;
  kickoffTime: string;
  league: string;
  homeTeam: string;
  awayTeam: string;
  expectedHomeGoals: number;
  expectedAwayGoals: number;
  poisson: AnalysisProbability;
  elo: AnalysisProbability;
  final: AnalysisProbability;
  over25Probability: number;
  bttsProbability: number;
  fairOdds: AnalysisOdds;
  predictedScore: string;
  recommendation: string;
  recommendationType: 'home' | 'draw' | 'away';
  goalsAdvice: string;
  goalsType: 'over' | 'under';
  bttsAdvice: string;
  confidence: number;
  confidenceLevel: '高' | '中' | '低';
  historyMatches: number;
  homeTeamHistory: number;
  awayTeamHistory: number;
  homeForm: AnalysisTeamForm;
  awayForm: AnalysisTeamForm;
  scoreLines: AnalysisScore[];
  modelBasis: string;
}

export interface AnalysisSummary {
  matchCount: number;
  highConfidence: number;
  homeLean: number;
  overLean: number;
  averageConfidence: number;
  algorithm: string;
}

export interface RemoteMatch {
  matchId: string;
  matchNum: string;
  date: string;
  time: string;
  league: string;
  homeTeam: string;
  awayTeam: string;
  status: string;
  homeOdds: number;
  drawOdds: number;
  awayOdds: number;
}

/** 拉取中国竞彩网当前在售赛程 */
export function fetchFootballRemoteMatches() {
  return request<{ source: string; fetchedAt: string; matches: RemoteMatch[] }>({ url: '/football/remote-matches' });
}

/** 对导入批次应用 Dixon-Coles 泊松、Elo 与近期状态做综合分析 */
export function fetchFootballAnalysis(id: number) {
  return request<{ batch: ImportBatch; summary: AnalysisSummary; matches: MatchAnalysis[] }>({ url: `/football/analysis/${id}` });
}
