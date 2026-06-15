# Gin API Integration for goInception

## 概览

本次改造新增了一个基于 `gin` 的 HTTP 服务入口，用于外部调用 goInception 的 SQL 审核和执行功能。核心逻辑仍使用现有的 `session.NewInception()`、`LoadOptions()`、`Audit()` 和 `RunExecute()`。

## 新增文件

- `server/gin_server.go`
  - 定义 `gin` 路由、请求参数和响应结构
  - 封装 SQL 审核/执行逻辑到 HTTP 接口
- `cmd/ginserver/main.go`
  - 新增独立命令行入口，用于启动 `gin` HTTP 服务
- `docs/gin_api.md`
  - 本文档

## 接口说明

### 健康检查

- `GET /healthz`
- 响应
  - `200` JSON: `{ "code": 0, "message": "ok" }`

### SQL 审核

- `POST /api/v1/audit`
- 请求 JSON
  - `host`：目标 MySQL 主机
  - `port`：目标 MySQL 端口
  - `user`：身份认证用户名
  - `password`：身份认证密码
  - `db`：默认数据库
  - `sql`：待审核的 SQL
  - `backup`：是否开启备份（可选）
  - `ignore_warnings`：是否忽略警告（可选）
- 响应 JSON
  - `code`：0 表示成功
  - `data.records`：审核结果列表
  - 每条 `record` 包含 `sql`、`err_level`、`error_message`、`ddl_rollback` 等字段

### SQL 执行

- `POST /api/v1/execute`
- 请求 JSON 同审核接口
- 返回结果结构相同，但内部执行模式为 `RunExecute()`

## 关键实现点

### `server/gin_server.go`

- 使用 `session.SourceOptions` 初始化目标 MySQL 连接信息
- `handleRequest()` 负责参数绑定和错误返回
- `processSQL()` 创建新 session 并调用 `Audit()` 或 `RunExecute()`
- `convertRecord()` 将内部 `session.Record` 映射为 JSON 友好的 `auditRecord`

### `cmd/ginserver/main.go`

- 提供命令行启动参数：
  - `-addr`：监听地址，默认 `:8080`
  - `-mode`：gin 模式，默认 `release`
- 启动 server 并输出日志

## 使用方式

1. 启动服务

```bash
cd /Users/yangkai/Documents/项目/goInception
go run ./cmd/ginserver -addr ":8080" -mode release
```

2. 调用审核接口

```bash
curl -X POST http://127.0.0.1:8080/api/v1/audit \
  -H 'Content-Type: application/json' \
  -d '{"host":"127.0.0.1","port":3306,"user":"root","password":"pwd","db":"test","sql":"ALTER TABLE t ADD COLUMN c INT;"}'
```

3. 调用执行接口

```bash
curl -X POST http://127.0.0.1:8080/api/v1/execute \
  -H 'Content-Type: application/json' \
  -d '{"host":"127.0.0.1","port":3306,"user":"root","password":"pwd","db":"test","sql":"CREATE TABLE t(id INT);"}'
```

## 编译验证

已完成以下验证：

- `go test ./server ./cmd/ginserver`

## 注意事项

- `gin` 接口仅做调用适配层，核心审核逻辑仍依赖项目现有 `session` 模块。
- `Backup` 参数会触发原 `goInception` 的备份检查逻辑，需确保配置文件中备份参数正确。
- 若需要更多参数支持（如 `ssl`, `tranBatch`, `split` 等），可继续扩展 `auditRequest` 和 `SourceOptions` 映射。