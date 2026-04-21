# LiteOJ

轻量化 OJ 平台：Go 后端 + Vue 3 前端，SQLite 数据库，通过 go-judge 完成沙箱判题。单二进制部署，点开即用，支持内网与本地化运行。

## 已实现功能

Phase 1（MVP）：登录 / 改密 / 用户 & Admin CRUD / 题目 + 测试用例 CRUD（Markdown + LaTex）/ go-judge 同步判题 / 提交历史 / 基础题单 / 单二进制构建。

Phase 2：
- 一级/二级标签字典 + 题目标签选择器 + 按标签筛选
- AI 三件套（基于 Bifrost OpenAI 协议）：非 AC 提交错误分析、题目自动打标签（预定义字典匹配）、题目解析生成
- 学生 Excel 批量导入（首行表头 `username/name/password`，支持中文）
- Monaco Diff 视图，支持两次提交代码并排对比
- OI/ACM 双榜 + 时间范围（总/年/月/周）；题单内同样支持排名
- 个人中心：AC 数 / AC 率 / 饼图分布 + 一年 Contribution 热力图
- 题目四态颜色反馈（未做 / 尝试过 / AC / AC 后有回退）
- 通知公告（首页默认展示）
- 进程内判题队列（goroutine worker + channel，单 worker 时 SQLite 并发写入无锁冲突）
- 简易 TTL 缓存（题目列表、题单、标签字典、公告 30 秒 TTL；写路径显式失效）

## 技术栈

- 后端：Go 1.22+ / Gin / GORM / `modernc.org/sqlite`（无 CGO）/ JWT / bcrypt
- 前端：Vue 3 + TypeScript + Vite + Naive UI + UnoCSS + Pinia + Monaco Editor + markdown-it + KaTeX
- 判题：[go-judge](https://github.com/criyle/go-judge)（独立进程，HTTP 对接）
- AI：兼容 OpenAI `/v1/chat/completions` 协议，部署时指向你的 Bifrost 实例
- 打包：前端构建产物通过 `go:embed` 嵌入单二进制

## 目录结构

```
liteoj/
├── backend/                Go 后端
│   ├── cmd/liteoj/main.go  入口
│   └── internal/           config / db / models / handlers / services/{judge,ai}
├── frontend/               Vue 前端（pnpm）
├── scripts/                dev.sh / build.sh
├── .env.example
└── Makefile
```

## 快速开始（仅支持 Linux；Windows 请在 WSL 中运行）

### 1. 启动 go-judge

从 [criyle/go-judge](https://github.com/criyle/go-judge/releases) 下载对应 Linux 二进制：

```bash
./go-judge -http-addr :5050
```

保持该进程常驻。LiteOJ 会通过 `JUDGE_BASE_URL` 调用它。

### 2. 准备配置

```bash
cp .env.example .env
# 编辑 JWT_SECRET、ADMIN_INIT_PASSWORD、（可选）BIFROST_BASE_URL / BIFROST_API_KEY
```

### 3. 开发模式（前后端分离）

```bash
# 终端 A：后端
cd backend
go run ./cmd/liteoj       # 访问 http://127.0.0.1:8080/api/health

# 终端 B：前端
cd frontend
pnpm install
pnpm dev                  # 访问 http://127.0.0.1:5173
```

Vite 会把 `/api/*` 代理到后端 `:8080`。首次启动时后端会依据 `.env` 中的 `ADMIN_INIT_*` 写入初始管理员。

### 4. 生产模式（单二进制）

```bash
./scripts/build.sh        # 产出 ./liteoj
./liteoj                  # 默认监听 :8080，同时提供 API 与前端静态资源
```

或使用 Make：

```bash
make install     # pnpm install + go mod tidy
make build       # 单二进制
make run         # 构建并运行
```

## 冒烟验证

1. 浏览器打开 `http://127.0.0.1:8080`，用 `.env` 里的初始管理员登录。
2. 「用户管理」→ 新增学生账号。
3. 「题目管理」→ 新建题目：
   - 标题 `A + B`
   - 描述（可包含 Markdown 与 `$a+b$` LaTex 公式）
   - 时间 1000 ms / 内存 256 MB / 可见=是
   - 「测试用例」栏新增两条（输入 `1 2`→输出 `3`，输入 `100 200`→输出 `300`）
4. 登出，用学生账号重新登录。
5. 进入题目详情页，选择语言（例 `cpp`），写 A+B 代码 → 提交，应返回 **AC**。
6. 故意写错（`cout << a;`），再次提交应返回 **WA**，并列出失败用例。
7. 「提交」菜单查看历史记录；「题单」菜单查看题单与完成进度。
8. 「个人中心」修改自身密码后重新登录有效。

## 常见问题

- **判题一直是 SE / 连接失败**：检查 go-judge 是否在 `JUDGE_BASE_URL` 监听；`curl $JUDGE_BASE_URL/version` 可验证。
- **Windows 运行报错**：判题依赖 Linux 沙箱能力，请在 WSL 中运行整个项目（后端、go-judge、前端可全部在 WSL 内）。
- **切换 PostgreSQL**：改 `.env` 的 `DB_DRIVER=postgres`、`DB_DSN=...`，重启即可。
- **修改 Admin 初始密码**：仅在数据库为空（即首次启动）时生效。已有 admin 后请在「用户管理」或「个人中心」里改密。

## Phase 2 路线图

Phase 2 已落地。下阶段可选扩展：比赛/作业模式、PG 驱动持续联调、判题队列持久化（进程重启不丢 PENDING）、管理员批量重置密码 UI、AI 模板管理 UI（当前只走 `.env`）。

## 切换 PostgreSQL

1. 准备 PG 数据库，创建空库 `liteoj`（或任意名字）。
2. `.env` 里改：
   ```
   DB_DRIVER=postgres
   DB_DSN=host=127.0.0.1 user=liteoj password=liteoj dbname=liteoj port=5432 sslmode=disable TimeZone=Asia/Shanghai
   ```
3. 重启后端，`AutoMigrate` 会在 PG 上建表；首启依旧按 `ADMIN_INIT_*` 写入 admin。
4. 注意 `StatsHandler.Contribution` 使用了 SQLite 专属函数 `strftime`；切到 PG 时需改成 `to_char(created_at, 'YYYY-MM-DD')`。本仓库保留该接口为 SQLite 默认实现，PG 场景在 Phase 3 做双路适配。
