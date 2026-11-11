import { computed, onMounted, ref, watch } from 'vue';
import { fetchFootballLeagues, fetchFootballMatches, fetchFootballSeasons } from '@/service/api/football';
import type { LeagueItem, MatchRow, SeasonItem } from '@/service/api/football';

export function useFootballData() {
  const seasons = ref<SeasonItem[]>([]);
  const leagues = ref<LeagueItem[]>([]);
  const season = ref('');
  const league = ref('');
  const month = ref(0);
  const rows = ref<MatchRow[]>([]);
  const loading = ref(false);
  const seasonOptions = computed(() => seasons.value.map(item => ({ label: item.label, value: item.value })));
  const leagueOptions = computed(() => leagues.value.map(item => ({ label: item.label, value: item.value })));
  const monthOptions = [
    { label: '全部月份', value: 0 },
    ...Array.from({ length: 12 }, (_, index) => ({ label: `${index + 1} 月`, value: index + 1 }))
  ];

  async function loadMatches() {
    if (!season.value || !league.value) {
      rows.value = [];
      return;
    }
    loading.value = true;
    const seasonInfo = seasons.value.find(item => item.value === season.value);
    const startYear = Number(seasonInfo?.label.split('-')[0]);
    const year = month.value ? startYear + (month.value < 8 ? 1 : 0) : 0;
    const { data, error } = await fetchFootballMatches({
      season: season.value,
      league: league.value,
      year,
      month: month.value
    });
    loading.value = false;
    if (error || !data) {
      rows.value = [];
      window.$message?.error('比赛数据加载失败');
      return;
    }
    rows.value = data.rows ?? [];
  }

  async function loadLeagues() {
    if (!season.value) {
      leagues.value = [];
      league.value = '';
      return;
    }
    const { data, error } = await fetchFootballLeagues(season.value);
    if (error || !data) {
      leagues.value = [];
      window.$message?.error('联赛列表加载失败');
      return;
    }
    leagues.value = data.leagues ?? [];
    league.value = leagues.value.find(item => item.value === 'E0')?.value ?? leagues.value[0]?.value ?? '';
  }

  watch(season, async () => {
    await loadLeagues();
    await loadMatches();
  });
  watch([league, month], () => loadMatches());

  onMounted(async () => {
    const { data, error } = await fetchFootballSeasons();
    if (error || !data) {
      window.$message?.error('赛季列表加载失败');
      return;
    }
    seasons.value = data.seasons ?? [];
    season.value = seasons.value.find(item => item.current)?.value ?? seasons.value[0]?.value ?? '';
  });

  return { seasons, leagues, season, league, month, rows, loading, seasonOptions, leagueOptions, monthOptions };
}