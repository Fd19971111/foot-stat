package football

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "server/api/football/v1"
)

// DataDir 数据根目录，可通过环境变量覆盖
func DataDir() string {
	if dir := os.Getenv("FOOT_DATA_DIR"); dir != "" {
		return dir
	}
	return "d:/phpstudy_pro/WWW/foot-stat/data"
}

// leagueName 联赛 code -> 名称映射（football-data.co.uk 编码）
var leagueName = map[string]string{
	"E0":  "英格兰超级联赛",
	"E1":  "英格兰冠军联赛",
	"E2":  "英格兰甲级联赛",
	"E3":  "英格兰乙级联赛",
	"EC":  "英格兰全国联赛",
	"D1":  "德国甲级联赛",
	"D2":  "德国乙级联赛",
	"I1":  "意大利甲级联赛",
	"I2":  "意大利乙级联赛",
	"SC0": "苏格兰超级联赛",
	"SC1": "苏格兰冠军联赛",
	"SC2": "苏格兰甲级联赛",
	"SC3": "苏格兰乙级联赛",
	"SP1": "西班牙甲级联赛",
	"SP2": "西班牙乙级联赛",
	"F1":  "法国甲级联赛",
	"F2":  "法国乙级联赛",
	"B1":  "比利时甲级联赛",
	"G1":  "希腊超级联赛",
	"N1":  "荷兰甲级联赛",
	"P1":  "葡萄牙超级联赛",
	"T1":  "土耳其超级联赛",
	"ARG": "阿根廷甲级联赛",
	"BRA": "巴西甲级联赛",
	"USA": "美国职业大联盟",
	"JPN": "日本职业联赛",
	"KOR": "韩国职业联赛",
	"CHN": "中国超级联赛",
}

// seasonStartYear 由文件夹名推导赛季起始年份，如 2526 -> 2025，9394 -> 1993
func seasonStartYear(folder string) int {
	v, err := strconv.Atoi(folder)
	if err != nil || len(folder) != 4 {
		return 0
	}
	prefix := v / 100
	if prefix >= 90 {
		return 1900 + prefix
	}
	return 2000 + prefix
}

func currentSeasonFolder() string {
	now := time.Now()
	start := now.Year()
	if now.Month() < 8 {
		start = now.Year() - 1
	}
	first := start % 100
	second := (start + 1) % 100
	return strconv.Itoa(first*100 + second)
}

// GetSeasons 扫描 data 目录下的赛季文件夹
func GetSeasons(_ context.Context) []v1.SeasonItem {
	root := DataDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		g.Log().Warningf(context.Background(), "read data dir failed: %v", err)
		return nil
	}

	current := currentSeasonFolder()
	type kv struct {
		folder string
		start  int
	}
	var list []kv
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		start := seasonStartYear(e.Name())
		if start == 0 {
			continue
		}
		list = append(list, kv{folder: e.Name(), start: start})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].start > list[j].start })

	items := make([]v1.SeasonItem, 0, len(list))
	for _, item := range list {
		items = append(items, v1.SeasonItem{
			Value:   item.folder,
			Label:   strconv.Itoa(item.start) + "-" + strconv.Itoa(item.start+1),
			Current: item.folder == current,
		})
	}
	return items
}

// GetLeagues 列出某赛季下的联赛
func GetLeagues(_ context.Context, season string) []v1.LeagueItem {
	dir := filepath.Join(DataDir(), season)
	files, err := os.ReadDir(dir)
	if err != nil {
		g.Log().Warningf(context.Background(), "read season dir failed: %v", err)
		return nil
	}

	var items []v1.LeagueItem
	for _, f := range files {
		if f.IsDir() || !strings.EqualFold(filepath.Ext(f.Name()), ".csv") {
			continue
		}
		code := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		name, ok := leagueName[code]
		if !ok {
			name = code
		}
		items = append(items, v1.LeagueItem{Value: code, Label: code + " - " + name})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Value < items[j].Value })
	return items
}

// GetMatches 读取某赛季某联赛的比赛，按月份筛选
func GetMatches(_ context.Context, season, league string, year, month int) (columns []string, rows []map[string]string) {
	path := filepath.Join(DataDir(), season, league+".csv")
	f, err := os.Open(path)
	if err != nil {
		g.Log().Warningf(context.Background(), "open csv failed: %v", err)
		return nil, nil
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // 容忍列数不一致
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		g.Log().Warningf(context.Background(), "read csv failed: %v", err)
		return nil, nil
	}
	if len(records) < 2 {
		return nil, nil
	}

	columns = records[0]
	for i := range columns {
		columns[i] = trimBOM(columns[i])
	}
	colIndex := make(map[string]int, len(columns))
	for i, c := range columns {
		colIndex[strings.TrimSpace(c)] = i
	}

	dateIdx, hasDate := colIndex["Date"]

	for _, rec := range records[1:] {
		if len(rec) == 0 || (len(rec) == 1 && strings.TrimSpace(rec[0]) == "") {
			continue
		}
		// 月份筛选
		if hasDate && year > 0 && month > 0 {
			_, m, y := parseCsvDate(rec[dateIdx])
			if y != year || m != month {
				continue
			}
		}

		row := make(map[string]string, len(columns))
		for i, c := range columns {
			if i < len(rec) {
				row[c] = strings.TrimSpace(rec[i])
			} else {
				row[c] = ""
			}
		}
		rows = append(rows, row)
	}
	return columns, rows
}

// trimBOM 去除可能存在的 UTF-8 BOM 头
func trimBOM(s string) string {
	return strings.TrimPrefix(s, "\uFEFF")
}

// parseCsvDate 解析 DD/MM/YYYY 或 YYYY-MM-DD
func parseCsvDate(s string) (day, month, year int) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) == 3 {
			day = gconv.Int(parts[0])
			month = gconv.Int(parts[1])
			year = gconv.Int(parts[2])
			if year >= 0 && year < 100 {
				if year >= 70 {
					year += 1900
				} else {
					year += 2000
				}
			}
		}
	} else if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 3 {
			year = gconv.Int(parts[0])
			month = gconv.Int(parts[1])
			day = gconv.Int(parts[2])
		}
	}
	return
}
