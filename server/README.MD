# FootStat 后端服务

GoFrame HTTP API 为 FootStat 前端提供赛季/联赛/比赛数据、赛程导入、预测和分析能力。历史赛果从项目根目录的 `data/` 加载，导入批次及预测结果保存至 MySQL。

## 环境要求

- Go 1.23 或更高版本
- MySQL
- 项目根目录中的赛季 CSV 数据；可使用 `FOOT_DATA_DIR` 指定其他数据路径

## 配置与启动

在 `manifest/config/config.yaml` 中配置 `server.address` 和 `database.default.link`。数据库 DSN 使用 GoFrame 的 `mysql:user:password@tcp(host:port)/database` 格式；请使用本机凭据，不要提交密码。

先创建配置中指定的数据库（默认数据库名为 `foot`），随后从项目根目录启动：

```powershell
go -C server run .
```

默认 API 地址为 `http://127.0.0.1:8005`。导入相关表会在首次访问时自动创建，SQL 定义见 `resource/sql/football_import.sql`。

## 主要接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/football/seasons` | 获取本地赛季列表 |
| GET | `/football/leagues` | 获取赛季下的联赛 |
| GET | `/football/matches` | 查询 CSV 比赛数据 |
| POST | `/football/imports` | 保存赛程并生成导入预测 |
| GET | `/football/imports` | 查询导入批次 |
| GET | `/football/imports/{id}` | 查询批次比赛与预测 |
| GET | `/football/analysis/{id}` | 生成 Dixon–Coles 与 Elo 综合分析 |
| GET | `/football/remote-matches` | 获取当前可用赛程 |

API 文档由 GoFrame 提供，配置路径为 `/api.json`，Swagger 页面为 `/swagger`。

## 离线回测

```powershell
go -C server run ./cmd/backtest -league E0 -from-year 2020
```

回测按日期滚动，只用比赛日前的历史数据，输出模型与赔率市场基线的命中率、Log Loss 和 Brier Score。请在读取新数据后重新运行，指标会随数据覆盖范围变化。

## 测试

```powershell
go -C server test ./...
```