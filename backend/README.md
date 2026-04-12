# AI 智能博客系统后端

这个后端项目是专门按“Go 新手也能看懂”的思路写的。

## 你会看到什么

- 目录分层很直白：`router -> controller -> service -> repository -> model`
- 关键业务流程都尽量写了中文注释
- AI 适配层已经接好 5 类 Provider：
  - `GLM`
  - `MiniMax`
  - `小米`
  - `OpenAI 兼容接口`
  - `Ollama`

## 当前实现状态

- 已完成：
  - 文章列表和详情接口
  - 标题 AI 解析
  - 文章问答
  - 全站问答
  - AI 摘要
  - AI 标签推荐
  - AI 灵感风暴
  - AI 改写
  - AI 封面图生成
  - AI 评论回复建议
  - AI 评论审核
  - SSE 流式返回基础实现
- 当前为了让你容易跑起来，仓库层先用的是“内存版演示数据”
- 后续你可以把 `repository` 层替换成 `MySQL + GORM`

## 运行方式

1. 复制配置文件

```powershell
Copy-Item .\configs\config.example.json .\configs\config.local.json
```

2. 把你的大模型密钥填进 `config.local.json`

3. 安装依赖

```powershell
go mod tidy
```

4. 启动服务

```powershell
go run ./cmd/server
```

5. 访问健康检查

`http://127.0.0.1:8080/health`

## 推荐你先看的文件

- `cmd/server/main.go`
- `internal/bootstrap/app.go`
- `internal/router/router.go`
- `internal/controller/public_controller.go`
- `internal/service/ai_service.go`
- `internal/service/provider_registry.go`
