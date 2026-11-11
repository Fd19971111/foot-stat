# FootStat 前端

本目录包含 FootStat 的 Vue 3 单页应用，使用 Naive UI、Pinia、Vue Router、ECharts 和 UnoCSS。通用管理界面基于 Soybean Admin 定制，足球业务页面使用 `src/styles/football-workspace.css` 统一样式。

## 本地开发

要求 Node.js 20.19+ 与 pnpm 10.5+。在项目根目录运行：

```powershell
pnpm --dir web install
pnpm --dir web dev
```

开发服务器默认监听 `http://localhost:9527`。`web/.env.test` 中的 `VITE_SERVICE_BASE_URL` 默认指向 `http://127.0.0.1:8005`，需要同时启动 Go 后端。

## 页面

- `/home`：比赛中心，按赛季、联赛和月份查看近期赛果。
- `/history`：对比近五个赛季的联赛指标。
- `/results`：筛选并浏览逐场赛果及技术统计。
- `/import`：导入 CSV/Excel 赛程或拉取可用赛程。
- `/analysis`：查看批次预测、模型概率和比分分布。

## 环境配置提示

生产构建读取 `web/.env.prod`。该文件当前使用 Apifox 示例服务地址；正式部署前，应把 `VITE_SERVICE_BASE_URL` 配置为实际后端 API 地址，并验证跨域/反向代理设置。

## 检查

```powershell
pnpm --dir web typecheck
pnpm --dir web build
```

前端基础模板来源和许可信息保留于 `README.md`、`README.en_US.md` 与 `LICENSE`。本文件补充 FootStat 的实际页面和开发方式，不替代上游来源说明。
