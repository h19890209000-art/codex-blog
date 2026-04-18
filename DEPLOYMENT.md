# codex-blog 部署说明

本文档整理了当前项目的完整部署方式，并补充了“每日资讯英语学习”功能的上线注意事项。

## 1. 项目组成

- `backend`
  - Go + Gin API 服务
  - 默认端口 `8080`
- `frontend/reader`
  - 读者端
  - 默认开发端口 `5173`
- `frontend/admin`
  - 管理后台
  - 构建后部署到 `/admin/`
  - 默认开发端口 `5174`
- `scripts/fetch-daily-briefings.ps1`
  - 每日资讯抓取脚本

## 2. 当前主要功能

- 博客文章、分类、标签、评论、站内问答
- 每日资讯抓取、展示、后台管理
- 每日资讯英语学习
  - 原文抓取
  - 中文译文
  - 单词悬浮释义
  - 句子结构分析
  - 10 分钟学习流
  - 半开放角色扮演纠错
  - 本地复习卡

## 3. 环境要求

推荐版本：

- Go `1.23.x`
- Node.js `20.x`
- npm `10.x`
- MySQL `8.0+`
- Nginx `1.20+`

Windows 本地开发建议：

- PowerShell 5.1 或 PowerShell 7
- 已安装 `go`、`node`、`npm`

## 4. 配置文件

后端默认读取：

```text
backend/configs/config.local.json
```

示例文件：

```text
backend/configs/config.example.json
```

如果你要自定义位置，可以设置环境变量：

```bash
APP_CONFIG=/opt/codex-blog/backend/configs/config.local.json
```

至少要检查这些字段：

- `server.host`
- `server.port`
- `database.host`
- `database.port`
- `database.username`
- `database.password`
- `database.database_name`
- `auth.token_secret`
- `auth.default_admin_username`
- `auth.default_admin_password`
- `ai.routing.*`
- `ai.providers.*`

## 5. AI 配置建议

英语学习功能会调用以下能力：

- 英文原文翻译
- 单词释义
- 句子结构分析
- 学习计划生成
- 角色扮演纠错

推荐至少配置一个可用的聊天模型。你当前项目已经接入了：

- `xiaomi`
- `glm`
- `minimax`
- `openai`
- `ollama`

如果你要使用小米 Mimo，建议在 `config.local.json` 里确认：

```json
{
  "ai": {
    "routing": {
      "analyze_title": "xiaomi",
      "summary": "xiaomi",
      "chat": "xiaomi",
      "moderate": "xiaomi"
    },
    "providers": {
      "xiaomi": {
        "type": "xiaomi",
        "base_url": "https://api.xiaomimimo.com/v1",
        "api_key": "你的 key",
        "model": "mimo-v2-pro"
      }
    }
  }
}
```

说明：

- 结构化学习任务会自动优先走更稳的 `mimo-v2-flash`
- 如果上游返回不稳定，系统会自动切到本地兜底结果，不会整页空掉

## 6. 数据库初始化

先创建数据库：

```sql
CREATE DATABASE ai_blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

项目使用 GORM 自动迁移。首次启动后端时会自动创建或补齐表结构。

当前版本会自动迁移这些关键数据：

- 文章、分类、标签、评论、用户
- 每日资讯表
- 系统配置表
- 每日资讯英语学习字段
  - `source_content`
  - `translated_content`
  - `content_fetched_at`
  - `translated_at`

## 7. 本地运行

### 7.1 启动后端

```powershell
cd D:\codex-blog\backend
go mod tidy
go run ./cmd/server
```

### 7.2 启动读者端

```powershell
cd D:\codex-blog\frontend\reader
npm install
npm run dev
```

### 7.3 启动后台

```powershell
cd D:\codex-blog\frontend\admin
npm install
npm run dev
```

默认开发地址：

- 读者端: [http://127.0.0.1:5173](http://127.0.0.1:5173)
- 后台: [http://127.0.0.1:5174/admin/](http://127.0.0.1:5174/admin/)
- 后端: [http://127.0.0.1:8080](http://127.0.0.1:8080)

## 8. 生产环境构建

### 8.1 Linux 构建后端

```bash
cd /opt/codex-blog/backend
go mod tidy
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
go build -o /opt/codex-blog/bin/dailybriefing-fetcher ./cmd/dailybriefing_fetcher
```

### 8.2 Linux 构建前端

如果前后端同域部署，通常不需要设置 `VITE_API_ORIGIN`。

```bash
cd /opt/codex-blog/frontend/reader
npm install
npm run build

cd /opt/codex-blog/frontend/admin
npm install
npm run build
```

如果前端和 API 不是同域，请在构建读者端前设置：

```bash
export VITE_API_ORIGIN=https://api.your-domain.com
```

### 8.3 Windows 构建

```powershell
cd D:\codex-blog\backend
go build -o D:\codex-blog\bin\codex-blog-server.exe .\cmd\server
go build -o D:\codex-blog\bin\dailybriefing-fetcher.exe .\cmd\dailybriefing_fetcher

cd D:\codex-blog\frontend\reader
npm install
npm run build

cd D:\codex-blog\frontend\admin
npm install
npm run build
```

## 9. systemd 配置

示例：

```bash
sudo tee /etc/systemd/system/codex-blog-backend.service > /dev/null << 'EOF'
[Unit]
Description=codex-blog backend service
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/codex-blog/backend
Environment=GIN_MODE=release
ExecStart=/opt/codex-blog/bin/codex-blog-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

启用：

```bash
sudo systemctl daemon-reload
sudo systemctl enable codex-blog-backend
sudo systemctl start codex-blog-backend
sudo systemctl status codex-blog-backend
```

## 10. Nginx 配置

推荐同域部署：

- 读者端: `https://your-domain.com/`
- 后台: `https://your-domain.com/admin/`
- API: `https://your-domain.com/api/...`

示例：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /opt/codex-blog/frontend/reader/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /admin/assets/ {
        alias /opt/codex-blog/frontend/admin/dist/assets/;
        access_log off;
        expires 7d;
    }

    location /admin/ {
        alias /opt/codex-blog/frontend/admin/dist/;
        try_files $uri $uri/ /admin/index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用配置：

```bash
sudo ln -sf /etc/nginx/sites-available/codex-blog.conf /etc/nginx/sites-enabled/codex-blog.conf
sudo nginx -t
sudo systemctl reload nginx
```

如果需要 HTTPS，再接 Let’s Encrypt 或你自己的证书。

## 11. 每日资讯抓取

### 11.1 Windows 手动抓取

```powershell
cd D:\codex-blog
.\scripts\fetch-daily-briefings.ps1 -date 2026-04-18 -limit 10
```

### 11.2 Linux 手动抓取

```bash
cd /opt/codex-blog/backend
/opt/codex-blog/bin/dailybriefing-fetcher
```

### 11.3 定时任务建议

Linux `crontab` 示例：

```bash
0 8 * * * cd /opt/codex-blog/backend && /opt/codex-blog/bin/dailybriefing-fetcher >> /var/log/codex-blog-briefing.log 2>&1
```

## 12. 每日资讯英语学习功能部署说明

### 12.1 页面入口

- 每日资讯列表页进入英语精读
- 路由：
  - `/briefings/:id/study`

### 12.2 后端接口

- `GET /api/public/daily-briefings/:id/study`
- `POST /api/public/daily-briefings/:id/word-explanation`
- `POST /api/public/daily-briefings/:id/sentence-analysis`
- `POST /api/public/daily-briefings/:id/learning-plan`
- `POST /api/public/daily-briefings/:id/roleplay`

### 12.3 首次打开为什么会慢

学习页第一次打开时，系统会依次执行：

1. 抓取资讯原文正文
2. 生成中文译文
3. 缓存到数据库

因此首次打开常见耗时在 `20-30 秒`。当前后端已经内置了更高的写超时底线，避免返回体被截断。

如果你反向代理层也设置了超时，建议一并放宽：

- Nginx 可设置：
  - `proxy_read_timeout 120s;`
  - `proxy_send_timeout 120s;`

### 12.4 浏览器能力说明

- 单词释义、句法分析、学习计划、角色扮演都依赖后端 API
- “听句块”按钮依赖浏览器 `SpeechSynthesis`
- 复习卡保存在当前浏览器 `localStorage`，不是后端数据库

这意味着：

- 换浏览器或清缓存后，本地复习卡会消失
- 如果你想做跨设备同步，后续需要再补一张后端复习卡表

### 12.5 英语学习功能验证清单

部署后建议按这个顺序检查：

1. 打开读者端首页，确认每日资讯能正常显示
2. 点一条资讯进入英语精读页
3. 首次打开等待学习页完整加载
4. 悬浮英文单词，确认能返回释义
5. 点击英文句子，确认能返回主谓宾和语法点
6. 输入中文学习目标，点击“生成今日 10 分钟练习”
7. 在角色扮演区输入一句中文或中英混合，确认 AI 会给纠正结果
8. 把句块卡或角色纠正结果导入复习卡，刷新页面后确认本地卡片还在

## 13. 部署后快速自检

### 13.1 后端健康检查

```bash
curl http://127.0.0.1:8080/health
```

期望返回：

```json
{"message":"ok","success":true}
```

### 13.2 关键 API 自检

```bash
curl http://127.0.0.1:8080/api/public/daily-briefings
curl http://127.0.0.1:8080/api/public/daily-briefings/71/study
```

### 13.3 前端资源检查

- 读者端首页能打开
- `/admin/` 能打开
- `/briefings/:id/study` 能打开

## 14. 常见问题

### 14.1 学习页报空响应或 JSON 错误

优先排查：

- 后端是否已更新到新版本
- Nginx 代理超时是否太短
- 首次打开是否还在抓原文和生成译文

### 14.2 角色扮演或学习计划没有内容

优先排查：

- `config.local.json` 中的 AI provider 是否可用
- 当前 key 是否还有额度
- 服务端日志里是否有上游 provider 错误

### 14.3 后台能打开但接口报错

优先排查：

- `config.local.json` 路径是否正确
- MySQL 是否能连通
- 数据库是否完成自动迁移

## 15. 建议的上线顺序

1. 先部署后端和数据库
2. 跑一次健康检查和每日资讯接口
3. 构建并部署读者端
4. 构建并部署后台
5. 手动抓取一次每日资讯
6. 打开英语精读页完成一次完整自测
