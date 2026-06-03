# API 文档

默认 API 地址为 `http://localhost:8080`。如果本机 8080 已被占用，可以通过 `APP_ADDR=:8081` 启动后端，并让前端使用 `VITE_API_BASE=http://localhost:8081`。

所有接口均返回 JSON。错误响应格式：

```json
{
  "error": "错误信息"
}
```

列表接口正常情况下返回数组；前端也会兼容空值响应并按空数组处理。

## 健康检查

`GET /api/health`

响应：

```json
{
  "status": "ok"
}
```

## 病人

### 获取病人列表

`GET /api/patients`

响应：

```json
[
  {
    "id": "uuid",
    "name": "张三",
    "gender": "男",
    "birthday": "1990-01-01",
    "phone": "13800000000",
    "allergies": "青霉素过敏",
    "notes": "",
    "createdAt": "2026-06-01T14:00:00Z",
    "updatedAt": "2026-06-01T14:00:00Z",
    "lastRecordAt": "2026-06-01T14:05:00Z"
  }
]
```

### 创建病人

`POST /api/patients`

请求：

```json
{
  "name": "张三",
  "gender": "男",
  "birthday": "1990-01-01",
  "phone": "13800000000",
  "allergies": "青霉素过敏",
  "notes": "高血压随访"
}
```

响应：返回创建后的病人对象。

## 病历和药方

### 获取病历列表

`GET /api/patients/{patientId}/records`

响应：

```json
[
  {
    "id": "uuid",
    "patientId": "uuid",
    "kind": "condition",
    "title": "近期咳嗽",
    "content": "咳嗽三天，无发热。",
    "recordedAt": "2026-06-01T14:05:00Z",
    "createdAt": "2026-06-01T14:05:00Z"
  }
]
```

`kind` 当前支持：

- `condition`：病情
- `prescription`：药方
- `exam`：检查
- `note`：备注

### 创建病历

`POST /api/patients/{patientId}/records`

请求：

```json
{
  "kind": "prescription",
  "title": "门诊处方",
  "content": "药品名称、剂量、频次、疗程等",
  "recordedAt": "2026-06-01T14:05:00Z"
}
```

`recordedAt` 可省略，后端会使用当前时间。

> 注意：`patientId` 必须对应一个已存在的病人。后端启用了 SQLite 外键约束，
> 如果传入不存在的 patientId，请求会返回 400 错误：
> 
> ```json
> {
>   "error": "constraint failed: FOREIGN KEY constraint failed (787)"
> }
> ```
响应：返回创建后的病历对象。

## 聊天

### 获取聊天记录

`GET /api/patients/{patientId}/messages`

响应：

```json
[
  {
    "id": "uuid",
    "patientId": "uuid",
    "role": "user",
    "content": "我这个咳嗽需要注意什么？",
    "createdAt": "2026-06-01T14:10:00Z"
  }
]
```

`role` 当前支持：

- `user`
- `assistant`

### 发送消息

`POST /api/patients/{patientId}/chat`

请求：

```json
{
  "message": "我这个咳嗽需要注意什么？"
}
```

响应：

```json
{
  "user": {
    "id": "uuid",
    "patientId": "uuid",
    "role": "user",
    "content": "我这个咳嗽需要注意什么？",
    "createdAt": "2026-06-01T14:10:00Z"
  },
  "assistant": {
    "id": "uuid",
    "patientId": "uuid",
    "role": "assistant",
    "content": "医生助手回复内容",
    "createdAt": "2026-06-01T14:10:01Z"
  }
}
```

后端会把当前病人信息、病历/药方和历史聊天注入医生助手上下文。模型 API key 和 base URL 只从后端环境变量读取，不通过任何前端接口返回。
