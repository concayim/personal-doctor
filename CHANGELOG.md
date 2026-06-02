# Changelog

## 2026-06-01

- 配置后端本地模型连接参数，真实 API key 仅写入 `backend/.env`，不进入前端代码或示例配置。
- 新增 `.gitignore`，忽略后端 `.env`、本地 SQLite 数据库、前端依赖和构建产物。
- 扩展 README，补充项目结构、敏感配置说明和接口文档入口。
- 新增 `docs/API.md`，记录当前 REST API 请求和响应约定。
- 后端自动规范化 OpenAI-compatible base URL，允许环境变量填写服务根地址。
- 增加 base URL 规范化单元测试。
- 修复前端在列表接口返回空值或异常中间态时的 Vue 渲染报错，并保证发送按钮状态总能恢复。
