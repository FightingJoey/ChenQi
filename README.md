# 晨启 · ChenQi

> Miracle Morning SAVERS — guided morning practice web app

## 简介

晨启是一款帮助你深度践行《早起的奇迹》S.A.V.E.R.S. 人生拯救计划的 Web 应用。起床后一键开启「今日会话」，按顺序完成静默、肯定语、观想、运动、阅读与书写；每一步都有计时与引导文案，完成后可打卡并查看连续天数。你还可以管理个人肯定语、调整各步骤时长，把清晨例行真正落到日常。

## 功能

- 注册 / 登录（JWT）
- 今日引导会话：计时、完成 / 跳过、内置引导内容
- 肯定语 CRUD：系统文案只读；个人新建 / 编辑 / 删除；可「复制为我的」
- 书写步骤自动保存日记，支持日记列表查看
- 打卡日历与连续天数统计
- 例行设置：各步骤开关与时长（分钟）
- 健康检查：`GET /api/health`

## 技术栈

| 层 | 选型 |
|----|------|
| 前端 | Vue 3 + Vite + TypeScript + Naive UI + Tailwind CSS（zh-CN） |
| 后端 | Go + Chi + SQLite |
| 部署 | Docker Compose 一键构建；SQLite 数据卷持久化 |

## 一键部署

```bash
cd savers-app
docker compose up --build
# 若本机 Docker 守护进程需权限：sudo docker compose up --build
```

浏览器打开 <http://localhost:8080> ，注册账号后即可使用。

数据保存在 Docker volume `savers-data`（路径由环境变量 `DB_PATH` 控制，默认 `/data/savers.db`）。

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `ADDR` | `:8080` | 监听地址 |
| `DB_PATH` | `./data/savers.db` | SQLite 路径 |
| `STATIC_DIR` | _(空)_ | 前端静态目录；生产镜像为 `/app/static` |
| `JWT_SECRET` | 开发默认值 | **生产请务必修改** |
| `TZ` | `Asia/Shanghai` | 时区 |

## 本地开发

### 后端

```bash
cd backend
go run ./cmd/server
# 默认监听 :8080，数据库 ./data/savers.db
```

### 前端

```bash
cd frontend
npm install
npm run dev
# http://localhost:5173 ，已将 /api 代理到 :8080
```

### API 冒烟

```bash
./scripts/smoke.sh http://127.0.0.1:8080/api
```

## 界面截图

位于 `artifacts/`：

| 文件 | 说明 |
|------|------|
| `01-login.png` | 登录页 |
| `02-today-empty.png` | 今日会话（未开始） |
| `03-today-session.png` | 今日会话进行中 |
| `04-settings.png` | 例行设置 |
| `05-calendar.png` | 打卡日历 |
| `06-affirmations.png` | 肯定语列表 |
| `06-affirmations-editor.png` | 肯定语（新建后） |

## API 摘要

鉴权：除注册 / 登录 / 健康检查外，请求头需带 `Authorization: Bearer <token>`；可选 `X-Timezone`（如 `Asia/Shanghai`）用于「今日」日期。

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/auth/register` | 注册 |
| `POST` | `/api/auth/login` | 登录 |
| `GET` | `/api/auth/me` | 当前用户 |
| `GET` / `PUT` | `/api/routine` | 例行配置 |
| `GET` | `/api/sessions/today` | 今日会话 |
| `POST` | `/api/sessions/start` | 开始今日会话 |
| `PATCH` | `/api/sessions/{id}/steps/{key}` | 更新步骤状态 |
| `POST` | `/api/sessions/{id}/complete` | 完成打卡 |
| `GET` | `/api/content?step_key=affirmations` | 内容列表（含系统 + 个人） |
| `POST` | `/api/content` | 新建内容（如肯定语） |
| `PUT` / `DELETE` | `/api/content/{id}` | 更新 / 删除（仅个人） |
| `POST` | `/api/content/{id}/clone` | 复制系统文案为我的 |
| `GET` / `POST` | `/api/journals` | 日记列表 / 新建 |
| `GET` | `/api/stats` | 连续天数等统计 |
| `GET` | `/api/calendar` | 月历打卡 |
| `GET` | `/api/health` | 健康检查 |

系统内置内容（`user_id` 为空）只读；个人内容可完整 CRUD。今日会话的肯定语步骤会同时展示系统与个人肯定语。

## 目录结构

```
savers-app/
├── backend/          # Go API
├── frontend/         # Vue 应用
├── artifacts/        # 截图
├── scripts/smoke.sh  # API 冒烟脚本
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## License

私人 / 学习项目，按需使用。
