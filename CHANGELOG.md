# Changelog

## 2026-06-01

- 配置后端本地模型连接参数，真实 API key 仅写入 `backend/.env`，不进入前端代码或示例配置。
- 新增 `.gitignore`，忽略后端 `.env`、本地 SQLite 数据库、前端依赖和构建产物。
- 扩展 README，补充项目结构、敏感配置说明和接口文档入口。
- 新增 `docs/API.md`，记录当前 REST API 请求和响应约定。
- 后端自动规范化 OpenAI-compatible base URL，允许环境变量填写服务根地址。
- 增加 base URL 规范化单元测试。
- 修复前端在列表接口返回空值或异常中间态时的 Vue 渲染报错，并保证发送按钮状态总能恢复。

## 2026-06-03

- 修复：SQLite 外键约束未启用，`FOREIGN KEY` 和 `ON DELETE CASCADE` 不生效。
  - `backend/internal/store/store.go` 中 `Open()` 在连接建立后执行 `PRAGMA foreign_keys = ON`。
- 修复：前端发送消息时输入框过早清空，API 失败后用户丢失输入内容。
  - `frontend/src/App.vue` `submitMessage()` 的 `draft.value = ""` 移至 `guard` 回调内（API 成功后才清空）。
- 修复：前端刷新按钮（`loadPatients`）仅在当前病人消失时才重载数据，未更新当前病人的病历和聊天记录。
  - `frontend/src/App.vue` `loadPatients()` 在当前病人仍存在时也主动调用 `loadPatientData()`。
- 新增 `docs/REQUIREMENTS.md` 需求文档，覆盖功能需求、非功能需求、数据模型。
- 同步更新 `docs/API.md`，补全外键约束说明。
- 同步更新 `README.md`，项目结构包含需求文档入口，修改规范加入 REQUIREMENTS.md。

## 2026-06-03

- 修复：前端 502 Bad Gateway 错误。
  - 根因：`backend/.env` 中 `LLM_PROVIDER=openai` 指向本地 OrbStack 代理（`:8317`），Eino 调用模型时该端点未正常响应，导致聊天接口无限挂起，前端请求超时返回 502。
  - 修复：临时切换至 `LLM_PROVIDER=mock`，后端使用本地占位回复，聊天接口瞬时返回。
  - 后续：配置好本地模型代理后可恢复 `LLM_PROVIDER=openai`。
