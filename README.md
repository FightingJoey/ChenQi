# 早起奇迹 · SAVERS

基于《早起的奇迹》S.A.V.E.R.S. 人生拯救计划的晨间例行练习 Web 应用。

静默 (Silence) → 肯定语 (Affirmations) → 观想 (Visualization) → 运动 (Exercise) → 阅读 (Reading) → 书写 (Scribing)

## 技术栈

- 前端：Vue 3 + Vite + TypeScript + Naive UI + Tailwind CSS（中文界面）
- 后端：Go + Chi + SQLite
- 部署：Docker Compose 一键启动

## 一键部署

```bash
cd savers-app
docker compose up --build
```

浏览器打开 <http://localhost:8080> ，注册账号后即可开始「今日会话」。

数据持久化在 Docker volume `savers-data`（SQLite）。

## 本地开发

### 后端

```bash
cd backend
go run ./cmd/server
# 默认 :8080，DB_PATH=./data/savers.db
```

### 前端

```bash
cd frontend
npm install
npm run dev
# http://localhost:5173 ，已代理 /api → :8080
```

## 主要功能

- 注册 / 登录（JWT）
- 可配置各步骤时长与开关
- 今日引导会话：计时、完成/跳过、内置引导文案
- 书写步骤自动保存日记
- 连续天数与打卡日历
- 健康检查：`GET /api/health`

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `ADDR` | `:8080` | 监听地址 |
| `DB_PATH` | `./data/savers.db` | SQLite 路径 |
| `STATIC_DIR` | _(空)_ | 前端静态目录；非空则同端口托管 SPA |
| `JWT_SECRET` | 开发默认值 | 生产请务必修改 |
| `TZ` | `Asia/Shanghai` | 时区 |

## 截图

见 `artifacts/` 目录（登录、今日会话、日历、设置）。
