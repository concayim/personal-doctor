# Personal Doctor Agent

一个 Go + Eino + Vue 的个人医生聊天 agent 原型。它支持管理多人病历、录入病情/药方、按病人维度聊天，并把病历上下文注入到医生助手提示词中。

> 重要：这是健康信息管理和问诊辅助工具，不替代医生诊断、处方或急救服务。

## 功能

- 多病人档案管理
- 每个病人独立病历/药方记录
- 每个病人独立聊天历史
- 后端使用 Go，模型编排预留 Eino
- 前端使用 Vue 3 + Vite
- 未配置 API 时可本地模拟回复，方便先体验流程
- 前端会把病人、病历、消息列表响应统一规范成数组，避免接口空值导致页面崩溃。

## 项目结构

```text
.
├── backend/          # Go API、SQLite、Eino 模型调用
├── frontend/         # Vue 3 工作台
├── docs/API.md       # REST API 文档
└── CHANGELOG.md      # 修改记录
```

## 启动

后端：

```bash
cp backend/.env.example backend/.env
cd backend
go run ./cmd/server
```

前端：

```bash
cd frontend
npm install
npm run dev
```

如果后端不是 `http://localhost:8080`，可以这样指定：

```bash
VITE_API_BASE=http://localhost:8081 npm run dev
```

默认地址：

- API: `http://localhost:8080`
- Web: `http://localhost:5173`

## 模型配置

模型配置只放在后端本地 `backend/.env`，不要写入前端代码、接口响应或 `.env.example`。

修改 `backend/.env`：

```bash
LLM_PROVIDER=openai
OPENAI_API_KEY=your_api_key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
```

`OPENAI_BASE_URL` 可以填写服务根地址或 `/v1` 地址；后端会在需要时自动补齐 `/v1`。

也可以先保留 `LLM_PROVIDER=mock`，后端会生成本地占位回复。

当前 `.gitignore` 已忽略 `backend/.env` 和 `backend/data/`，避免本地密钥和病历数据库被误提交。

## 接口文档

REST API 见 [docs/API.md](docs/API.md)。

## 修改规范

每次修改都需要同步更新：

- [CHANGELOG.md](CHANGELOG.md)
- [README.md](README.md)，如果启动、配置或功能发生变化
- [docs/API.md](docs/API.md)，如果接口发生变化

敏感信息只能保存在后端环境配置中，不能暴露到前端包、接口文档或示例配置。
