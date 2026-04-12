# codex-blog 部署文档

## 1. 项目结构

本项目当前包含 4 个主要部分：

- `backend`：Go + Gin API 服务，默认端口 `8080`
- `frontend/reader`：读者端，Vite + Vue 3
- `frontend/admin`：管理后台，Vite + Vue 3
- `scripts/fetch-daily-briefings.ps1`：每日简讯抓取脚本

当前已经集成的关键能力：

- 文章、分类、标签、评论、用户管理
- AI 能力路由
- 读者端“每日简讯”
- 后台“每日简讯管理”
- 自动抓取每日 10 条全球 AI 资讯

## 2. 推荐部署方式

推荐使用下面这套结构：

- 后端二进制部署为常驻服务
- MySQL 单独部署
- 读者端静态文件由 Nginx 托管在 `/`
- 管理端静态文件由 Nginx 托管在 `/admin/`
- Nginx 反向代理 `/api/` 和 `/health`
- 每日简讯抓取脚本通过 Windows 计划任务或 Linux `cron` 定时执行

## 3. 环境要求

建议版本：

- Go `1.23.3` 或更高
- Node.js `20.x`
- npm `10.x` 或更高
- MySQL `8.0+`
- Nginx `1.20+`

Windows 本地开发建议：

- PowerShell 5.1 或 PowerShell 7
- 已安装 `go`、`node`、`npm`

## 4. 配置文件说明

后端配置文件路径：

- 示例文件：`backend/configs/config.example.json`
- 本地实际配置：`backend/configs/config.local.json`

启动时如果没有显式设置 `APP_CONFIG`，后端会默认读取：

```text
backend/configs/config.local.json
```

### 4.1 必改项

至少要检查这些字段：

- `database.host`
- `database.port`
- `database.username`
- `database.password`
- `database.database_name`
- `auth.token_secret`
- `auth.default_admin_password`
- `ai.providers.*.api_key`
- `oss.access_key_id`
- `oss.access_key_secret`

### 4.2 小米 MiMo 配置

本项目里小米 Provider 现在已经按可用方式接通，建议这样配置：

```json
"xiaomi": {
  "type": "xiaomi",
  "base_url": "https://api.xiaomimimo.com/v1",
  "api_key": "replace-with-your-xiaomi-key",
  "model": "mimo-v2-pro"
}
```

补充说明：

- `base_url` 必须带上 `/v1`
- 模型 `mimo-v2-pro` 可用于正常对话
- 本项目的“评论审核”会自动走兼容封装，并优先使用 `mimo-v2-flash` 输出短 JSON 结果
- 小米 MiMo 当前不能按标准 OpenAI 方式直接调用 `/moderations`，所以项目里做了审核兼容层，不要再把它改回裸 `/moderations`

## 5. 数据库初始化

先创建数据库：

```sql
CREATE DATABASE ai_blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

本项目使用 GORM 自动迁移，后端首次启动时会自动建表和补表结构。

## 6. 本地开发启动

### 6.1 启动后端

```powershell
cd D:\codex-blog\backend
go run ./cmd/server
```

### 6.2 启动读者端

```powershell
cd D:\codex-blog\frontend\reader
npm install
npm run dev
```

### 6.3 启动管理端

```powershell
cd D:\codex-blog\frontend\admin
npm install
npm run dev
```

默认开发地址：

- 读者端：`http://localhost:5173`
- 管理端：`http://localhost:5174`
- 后端：`http://localhost:8080`

## 7. 生产构建

### 7.1 构建后端

```bash
cd /opt/codex-blog/backend
go mod tidy
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
go build -o /opt/codex-blog/bin/dailybriefing-fetcher ./cmd/dailybriefing_fetcher
```

### 7.2 构建读者端

```bash
cd /opt/codex-blog/frontend/reader
npm install
npm run build
```

### 7.3 构建管理端

```bash
cd /opt/codex-blog/frontend/admin
npm install
npm run build
```

说明：

- 管理端已经配置为 `base: '/admin/'`
- 因此同域名子路径部署时，静态资源会从 `/admin/assets/...` 加载
- 如果你改成后台独立子域名部署，可以再把 `base` 改回 `/`

## 8. Linux 服务器部署

以下示例基于 Ubuntu 22.04。

### 8.1 安装环境

```bash
sudo apt update
sudo apt install -y git curl nginx mysql-server
```

安装 Go：

```bash
curl -LO https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

安装 Node.js：

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
node -v
npm -v
```

### 8.2 拉取代码

```bash
cd /opt
sudo git clone <your-repo-url> codex-blog
sudo chown -R $USER:$USER /opt/codex-blog
cd /opt/codex-blog
```

### 8.3 准备配置

```bash
cd /opt/codex-blog/backend
cp ./configs/config.example.json ./configs/config.local.json
```

然后修改：

```text
/opt/codex-blog/backend/configs/config.local.json
```

重点确认：

- 数据库连接是否正确
- `auth.token_secret` 是否替换为随机长串
- 管理员默认密码是否已修改
- 小米 MiMo、MiniMax、GLM、OpenAI、OSS 的密钥是否正确

### 8.4 后端 systemd 服务

创建服务文件：

```bash
sudo tee /etc/systemd/system/codex-blog-backend.service > /dev/null << 'EOF'
[Unit]
Description=codex-blog backend service
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/codex-blog/backend
ExecStart=/opt/codex-blog/bin/codex-blog-server
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
```

启用并启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable codex-blog-backend
sudo systemctl start codex-blog-backend
sudo systemctl status codex-blog-backend
```

## 9. Nginx 配置

推荐同域名部署：

- 读者端：`https://your-domain.com/`
- 管理端：`https://your-domain.com/admin/`
- API：`https://your-domain.com/api/...`

Nginx 配置示例：

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

启用站点：

```bash
sudo ln -sf /etc/nginx/sites-available/codex-blog.conf /etc/nginx/sites-enabled/codex-blog.conf
sudo nginx -t
sudo systemctl reload nginx
```

如果你要上 HTTPS，建议再接上 Let's Encrypt 或你自己的证书。

## 10. Windows 服务器部署

如果你准备直接部署在 Windows 机器上，建议这样做：

### 10.1 构建后端

```powershell
cd D:\codex-blog\backend
go build -o D:\codex-blog\bin\codex-blog-server.exe .\cmd\server
go build -o D:\codex-blog\bin\dailybriefing-fetcher.exe .\cmd\dailybriefing_fetcher
```

### 10.2 启动后端

开发环境临时启动：

```powershell
cd D:\codex-blog\backend
D:\codex-blog\bin\codex-blog-server.exe
```

正式环境建议：

- 使用 NSSM、WinSW 或任务计划程序将 `codex-blog-server.exe` 托管为常驻服务
- 工作目录要指向 `D:\codex-blog\backend`
- 保证 `backend/configs/config.local.json` 可读

### 10.3 构建前端

```powershell
cd D:\codex-blog\frontend\reader
npm install
npm run build

cd D:\codex-blog\frontend\admin
npm install
npm run build
```

然后把：

- `frontend/reader/dist`
- `frontend/admin/dist`

交给 IIS、Nginx for Windows 或其他静态文件服务托管。

## 11. 每日简讯自动抓取

### 11.1 手动执行

Windows：

```powershell
cd D:\codex-blog
.\scripts\fetch-daily-briefings.ps1 -date 2026-04-12 -limit 10
```

Linux：

```bash
cd /opt/codex-blog/backend
/opt/codex-blog/bin/dailybriefing-fetcher -date 2026-04-12 -limit 10
```

不传 `-date` 时，会按当天日期抓取。

### 11.2 Windows 计划任务

可以设置一个每天早上 08:30 执行的任务，命令示例：

```powershell
powershell.exe -ExecutionPolicy Bypass -File D:\codex-blog\scripts\fetch-daily-briefings.ps1 -limit 10
```

建议：

- 起始目录设为 `D:\codex-blog`
- 失败后允许重试
- 日志重定向到单独文件

### 11.3 Linux cron

每天 08:30 自动抓取：

```cron
30 8 * * * cd /opt/codex-blog/backend && /opt/codex-blog/bin/dailybriefing-fetcher -limit 10 >> /var/log/codex-blog-dailybriefing.log 2>&1
```

## 12. 部署完成后的检查项

至少检查下面这些地址和能力：

- `GET /health` 返回正常
- 读者端首页可打开
- 管理后台可登录
- 文章列表、分类、标签加载正常
- 每日简讯列表可见
- 后台“自动抓取”能成功执行
- 评论审核返回 `provider=xiaomi`

推荐自检命令：

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/public/articles
curl http://127.0.0.1:8080/api/public/daily-briefings
```

## 13. 更新流程

```bash
cd /opt/codex-blog
git pull

cd /opt/codex-blog/backend
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
go build -o /opt/codex-blog/bin/dailybriefing-fetcher ./cmd/dailybriefing_fetcher
sudo systemctl restart codex-blog-backend

cd /opt/codex-blog/frontend/reader
npm install
npm run build

cd /opt/codex-blog/frontend/admin
npm install
npm run build

sudo systemctl reload nginx
```

## 14. 常见问题

### 14.1 管理后台能打开但样式或资源 404

先检查两件事：

- `frontend/admin/vite.config.js` 是否包含 `base: '/admin/'`
- Nginx 是否同时配置了 `location /admin/` 和 `location /admin/assets/`

### 14.2 小米 MiMo 调用 404

通常是下面两个原因：

- `base_url` 没带 `/v1`
- 把审核错误地走到了 `/moderations`

本项目现在的正确方式是：

- 对话走 `https://api.xiaomimimo.com/v1/chat/completions`
- 审核走项目内兼容层，不直接请求 `/moderations`

### 14.3 每日简讯抓取失败

优先检查：

- 服务器是否能访问外部 RSS
- MySQL 是否正常
- 后端日志是否有超时
- 定时任务工作目录是否正确

### 14.4 配置文件泄露

强烈建议：

- 不要把 `config.local.json` 提交到 Git
- 不要把真实 API Key 写进文档
- 如果密钥曾经在聊天、截图或公开仓库里暴露，立即旋转

## 15. 当前实测结论

这次已经在项目内实测通过的内容：

- 后端可正常编译和启动
- 管理端 `/admin/` 子路径构建已调整
- 小米 MiMo 已接通
- 评论审核已通过项目接口联调
- 正常评论会返回 `provider=xiaomi` 且 `flagged=false`
- 风险内容会返回 `provider=xiaomi` 且 `flagged=true`

如果你后面把你自己的旧部署文档贴给我，我可以继续按你原有格式再帮你做一版“最终对外发布版”。
