# LiteOJ

轻量化在线判题平台：Go 后端 + Vue 3 前端，modernc.org/sqlite 无 CGO 数据库，go-judge
沙箱判题。点开即用、单二进制部署，适合校内 / 社团 / 班级的算法刷题场景。

## 栈

- **后端**：Go 1.22+ / Gin / GORM / `modernc.org/sqlite` / JWT / bcrypt / pelletier/go-toml/v2
- **前端**：Vue 3 + TypeScript + Vite + Naive UI + UnoCSS + Pinia + Monaco + markdown-it + KaTeX + pnpm
- **判题**：[go-judge](https://github.com/criyle/go-judge)（Linux 沙箱，Windows 场景走 WSL）
- **AI**：OpenAI 兼容接口（DeepSeek / Bifrost / 自部署网关），错因解析 + AC 优化 + 题目生成
- **部署**：`liteoj.exe` + `config.toml` + `./dist/`（前端产物，**不内嵌** 进二进制）

## 核心约束

- **仅中文** i18n，**仅暗色** 主题，**仅 PC** 布局（移动端只显示"请用 PC 布局"提示）。
- 用户**只能登录**，由 admin 统一创建 / 改密；批量导入仅支持三栏粘贴（姓名/账号/密码）。
- 判题语言白名单：`c / cpp / java / python`（无 go，go-judge 在 WSL 里编译器只装这 4 种）。
- 所有列表实时更新**走 SSE**（`/api/events/stream`），**禁止轮询**。
- 所有滚动条**走 Naive UI** (`NScrollbar`)，禁止浏览器原生滚动条。
- 题单内做题 / AC 与独立题目页**完全独立**；题单内额外计 **罚时**：
  `(首次 AC 前错题数 + 首次 AC 前成功 AI 解析数) × 20 分钟`，rejected 的 AI 不计。
- 踢出题单的学生：`/submissions` 全站仍可见其提交；**不贡献**题单排名、不计 AK、
  不能再次加入。

## 目录结构

```
liteoj/
├── backend/
│   ├── cmd/liteoj/main.go          入口
│   └── internal/
│       ├── auth/                   JWT + bcrypt
│       ├── cache/                  轻量 TTL 缓存
│       ├── config/                 config.toml 解析
│       ├── db/                     Open + Migrate + PRAGMA
│       ├── events/                 内存 pub/sub broker（SSE 源头）
│       ├── handlers/               HTTP 层，按资源分文件
│       ├── i18n/                   后端错误文案常量
│       ├── middleware/             Auth / OptionalAuth / AdminOnly
│       ├── models/                 GORM 模型（含索引 tag）
│       ├── seed/                   启动 seed：admin / 字典 / 示例题 / 样本提交
│       ├── services/
│       │   ├── ai/                 prompts / client / runner / queue
│       │   └── judge/              client / runner / lang / queue
│       └── web/                    static.go 托管 ./dist/ 前端
├── frontend/                       Vue 前端（pnpm）
│   └── src/{pages, components, stores, router, api, i18n, styles}
├── scripts/                        dev.sh / build.sh / audit-e2e.sh / ...
├── docs/                           审计报告 + e2e 日志
├── config.toml                     运行时配置
├── config.example.toml             模板
└── data/                           SQLite 库 + 上传文件
```

## 启动步骤

### 0. 前置 —— 启 go-judge（WSL）

go-judge 只能在 Linux 沙箱里跑。Windows 用户用随项目附带的启动脚本：

```
C:\WSL\start-go-judge.bat   # UAC 提权 → 启 WSL liteoj (Debian 13) → 启 go-judge →
                            # 刷 netsh portproxy → 重启 iphlpsvc
```

详见 `C:\WSL\README.md`。探活：

```
curl http://127.0.0.1:5050/version
```

关闭：`C:\WSL\stop-go-judge.bat`。**不要** 修改 `C:\WSL\*` 脚本，它们是用户手工
维护的环境层。

### 1. 前端构建一次

```
cd frontend
pnpm install
pnpm build            # 产物：./dist/（项目根）
```

开发模式：`pnpm dev`，访问 `http://127.0.0.1:5173`，`/api/*` 会代理到 `:8080`。

### 2. 后端构建 & 启

```
cd backend
go build -o ../liteoj.exe ./cmd/liteoj
cd ..
./liteoj.exe         # 默认 :8080，读取 ./config.toml
```

首次启动会按 `config.toml [admin_init]` 写入管理员（默认 `admin / admin123`，务必
上线前改）并运行 seed：22 组知识点字典 + 示例题 + 样本学生 / 提交 / AI 任务。

## 配置：`config.toml`

拷贝 `config.example.toml → config.toml`，关键字段：

| 段 | 字段 | 说明 |
|---|---|---|
| `[app]` | `port` / `mode` | HTTP 端口 / `dev` or `prod` |
| `[db]` | `driver` / `dsn` | `sqlite`（默认 `./data/liteoj.db`）or `postgres` |
| `[jwt]` | `secret` / `ttl_hours` | **务必改 secret** |
| `[admin_init]` | `username` / `password` / `name` | 仅首次有效；之后改密走后台 |
| `[judge]` | `base_url` | 指向 go-judge (`http://127.0.0.1:5050`) |
|  | `langs` | 白名单 `["c","cpp","java","python"]` |
|  | `queue_workers` | 建议保持 1（单 worker → SQLite 单写手，不抖动） |
|  | `max_wait_seconds` | 单条提交从 worker 取出到落 verdict 的上限秒数；go-judge 掉线时能让队列可控恢复 |
| `[ai]` | `enabled` / `bifrost_*` | OpenAI 兼容网关 |
|  | `max_wait_seconds` | 覆盖所有 kind 的单次调用上限 |
|  | `prompt_*` | 见 `config.example.toml` 注释，JSON schema 由代码强制（见 `services/ai/prompts.go`） |

## 前后端路由速查

前端路由：
- `/`                     首页（admin 后台维护的 markdown）
- `/problems`             题目列表（**匿名可访问**）
- `/problems/:id`         题目详情（需登录）
- `/problemsets[/...]`    题单（需登录；进入题单需先点"加入"）
- `/submissions[/...]`    提交列表 / 详情 / diff
- `/ranking`              全站排行榜
- `/me`                   个人中心（统计 + 饼图 + 贡献热力图）
- `/admin/**`             仅管理员

后端 API：`/api/health` / `/api/meta` / `/api/home` / `/api/events/stream` +
`/api/problems` 匿名可读；其余 `authed` 需登录，`authed/admin` 需管理员。完整
清单见 `backend/cmd/liteoj/main.go`。

## AI 提示词

- `prompt_wrong_answer` / `prompt_optimize` / `prompt_tag` / `prompt_gen_{title,desc,idea,explain,all}`
  全部在 `config.toml [ai]` 段，admin 可随时调"判定/输出内容"。
- **JSON 结构约束** 在 `services/ai/prompts.go` 的 `analyzeOutputSuffix /
  tagOutputSuffix / genAllOutputSuffix`，代码级，不给 admin 权限改——避免模型返回被破坏
  让 parser 炸掉。
- "乱写判定" 规则（Hello World / 原样输出等入门题豁免、CE 不判乱写）在 config 的
  `prompt_wrong_answer` 注释里。
- 管理员手动触发的 Analyze 自动**绕过**乱写判定（handlers/ai.go 的 `ForceAnalyze=true`）。

## 判题 verdict

`AC / WA / TLE / MLE / OLE / RE / CE / PE / SE / UKE / PENDING`。详见
`backend/internal/models/submission.go` 开头注释。

## 脚本

| 脚本 | 作用 |
|---|---|
| `scripts/dev.sh` | 拉起前端 dev + 后端 |
| `scripts/build.sh` | 全量打包 → `./liteoj` + `./dist/` |
| `scripts/reset-and-seed.sh` | 备份 + 删库 + 重启让 seed 重建 |
| `scripts/audit-e2e.sh` | 端到端冒烟（要求 go-judge 已启） |
| `scripts/smoke-judge.sh` | 只打 go-judge，不依赖 liteoj |

## 常见问题

- **判题一直 SE**：先 `curl http://127.0.0.1:5050/version` 确认 go-judge 活着；
  若长时间无响应而 `max_wait_seconds` 到期，队列会主动落 SE 并释放 worker。
- **Windows 原生运行 go-judge**：不支持；沙箱依赖 Linux namespace，必走 WSL。
- **切换 PostgreSQL**：改 `[db].driver=postgres` 与 `dsn`；`stats.Contribution` 里
  `strftime` 是 SQLite 专属，PG 场景需要改成 `to_char(created_at, 'YYYY-MM-DD')`。
- **删库重来**：`scripts/reset-and-seed.sh` 会自动备份到 `./data/backups/`。
