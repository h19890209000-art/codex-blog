# AI 智能博客系统

这个仓库现在已经完成了第一版可运行骨架，重点是：

- 后端按 `router -> controller -> service -> repository -> model` 分层
- AI Provider 已支持 `GLM`、`MiniMax`、`小米`、`OpenAI 兼容接口`、`Ollama`
- 前端拆成 `reader` 和 `admin` 两个 Vue 应用
- 代码尽量按 Go 新手友好的方式组织

## 目录说明

- `backend`：Go + Gin 后端
- `frontend/reader`：读者端 Vue 页面骨架
- `frontend/admin`：后台管理端 Vue 页面骨架

## 快速开始

### 后端

```powershell
cd .\backend
Copy-Item .\configs\config.example.json .\configs\config.local.json
go mod tidy
go run ./cmd/server
```

当前本地默认数据库配置：

- Host: `127.0.0.1`
- Port: `3306`
- Username: `root`
- Password: `123456`
- Database: `ai_blog`

当前默认后台账号：

- Username: `admin`
- Password: `admin123456`

### 读者端

```powershell
cd .\frontend\reader
npm install
npm run dev
```

### 管理端

```powershell
cd .\frontend\admin
npm install
npm run dev
```

## 已实现接口

### 读者端

- `GET /api/public/articles`
- `GET /api/public/articles/:id`
- `POST /api/public/ai/analyze-title`
- `POST /api/public/qa/article/:id`
- `POST /api/public/qa/site`

### 管理端

- `GET /api/admin/dashboard`
- `POST /api/admin/ai/generate-summary`
- `POST /api/admin/ai/suggest-tags`
- `POST /api/admin/ai/brainstorm`
- `POST /api/admin/ai/rewrite`
- `POST /api/admin/ai/generate-cover`
- `POST /api/admin/comments/:id/reply-suggestions`
- `POST /api/admin/comments/moderate`
