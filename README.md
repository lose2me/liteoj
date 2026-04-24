<div align="center">

# LiteOJ

[![License: GPL](https://img.shields.io/badge/License-GPLv3-yellow.svg)](https://opensource.org/licenses/gpl-3-0)

</div>

LiteOJ 是一个面向弱网、内网、校内机房场景的轻量级 Online Judge。  
项目最初来自支教时的真实需求：乡村学校网络条件较差，传统 OJ 的部署、维护和更新成本偏高，  
由此这个项目诞生，同时内置了一组面向教师与学生的 AI 辅助能力。

## 与 Hydro、HUSTOJ 的功能对比

| 维度 | LiteOJ | Hydro | HUSTOJ |
| --- | --- | --- | --- |
| 项目定位 | 轻量校内 OJ | 平台型 OJ | 传统竞赛 OJ |
| 开发语言 | Go + Vue/TS | TypeScript | PHP + C/C++ |
| 默认数据库 | SQLite，可切 PostgreSQL | MongoDB | MySQL |
| 部署形态 | 单服务 | 平台化 | 经典 LAMP 风格 |
| 基础 OJ 能力 | 完整 | 完整 | 完整 |
| 题单/训练组织 | 支持 | 很强 | 成熟 |
| 多空间 / 多组织隔离 | 无 | 强 | 弱 |
| 插件生态 | 弱 | 强 | 一般 |
| 题库导入 / VJudge | 不是重点 | 强 | 强 |
| AI 错因解析 / 优化 | 支持 | 非主打 | 已支持 |
| AI 出题 / 题解 / 标签 | 支持 | 非主打 | 已支持 |
| AI 任务审计 | 强 | 非主打 | 一般 |
| 弱网 / 内网友好度 | 强 | 中 | 中 |
| 适合场景 | 校内教学 | 大平台 | 传统竞赛 |

## 项目目录结构

```text
liteoj/
├── backend/                         Go 后端
│   ├── cmd/liteoj/main.go           服务入口
│   └── internal/
│       ├── auth/                    JWT、密码哈希
│       ├── cache/                   轻量缓存
│       ├── config/                  TOML 配置解析
│       ├── db/                      数据库连接与表结构初始化
│       ├── events/                  SSE 事件总线
│       ├── handlers/                HTTP 接口
│       ├── i18n/                    中文文案
│       ├── middleware/              登录/管理员鉴权
│       ├── models/                  GORM 数据模型
│       ├── seed/                    管理员、标签体系、示例数据初始化
│       ├── services/
│       │   ├── ai/                  AI 客户端、提示词、队列、审计
│       │   └── judge/               判题语言、判题队列、判题执行
│       └── web/                     前端静态资源托管
├── frontend/                        Vue 3 前端
│   ├── src/api/                     前端 API 封装
│   ├── src/components/              通用组件
│   ├── src/pages/                   页面与后台页面
│   ├── src/router/                  路由
│   ├── src/stores/                  Pinia 状态管理
│   └── src/styles/                  全局样式
├── dist/                            已构建前端资源
├── data/                            SQLite、上传文件、运行期数据
├── scripts/                         开发、构建、审计、初始化脚本
├── config.toml                      运行配置
├── config.example.toml              配置模板
├── liteoj / liteoj.exe              构建产物
└── README.md
```