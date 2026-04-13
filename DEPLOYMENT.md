# codex-blog 部署文档

## 1. 当前项目组成

- `backend`：Go + Gin API 服务，默认端口 `8080`
- `frontend/reader`：读者端，Vite + Vue 3
- `frontend/admin`：管理后台，Vite + Vue 3，部署路径固定为 `/admin/`
- `scripts/fetch-daily-briefings.ps1`：每日简讯抓取脚本
- `scripts/package-release.ps1`：Windows 一键打包脚本

## 2. 这版新增内容

- 读者端首页标题、简介、搜索按钮等文案已支持后台配置
- 后台新增“系统设置 -> 系统配置”模块
- 后端启动时会自动迁移 `system_configs` 表
- 远程数据库第一次启动新版本后，会自动补齐默认系统配置

## 3. 环境要求

- Go `1.23.3` 或更高
- Node.js `20.x`
- npm `10.x` 或更高
- MySQL `8.0+`
- Nginx `1.20+`

Windows 本地建议：

- PowerShell 5.1 或 PowerShell 7
- 已安装 `go`、`node`、`npm`

## 4. 配置文件

后端配置文件位置：

- 示例文件：`backend/configs/config.example.json`
- 实际运行文件：`backend/configs/config.local.json`

如果没有显式设置 `APP_CONFIG`，后端默认读取：

```text
backend/configs/config.local.json
```

至少需要确认这些字段：

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

重要说明：

- 不要把 `config.local.json` 提交到 Git
- 不要把真实密钥写进公开文档
- 如果密钥已经暴露，先旋转再部署

## 5. 数据库初始化

先创建数据库：

```sql
CREATE DATABASE ai_blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

当前项目使用 GORM 自动迁移。后端首次启动时会自动创建或更新这些表：

- 用户、文章、分类、标签、评论相关表
- 每日简讯表
- `system_configs` 系统配置表

如果你是从旧版本升级上来，只要正常启动一次后端即可，不需要手工建 `system_configs`。

## 6. 本地启动

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

- 读者端：`http://127.0.0.1:5173/`
- 管理后台：`http://127.0.0.1:5174/admin/`
- 后端：`http://127.0.0.1:8080/`

## 7. 一键打包

Windows 下可以直接执行：

```powershell
cd D:\codex-blog
powershell -ExecutionPolicy Bypass -File .\scripts\package-release.ps1
```

如果你想指定版本号：

```powershell
cd D:\codex-blog
powershell -ExecutionPolicy Bypass -File .\scripts\package-release.ps1 -Version 20260413
```

打包后会在 `release/` 目录下生成：

- 一个发布目录：`release/codex-blog-版本号`
- 一个压缩包：`release/codex-blog-版本号.zip`

发布包包含：

- `backend/codex-blog-server.exe`
- `backend/dailybriefing-fetcher.exe`
- `backend/configs/config.example.json`
- `frontend/reader/dist`
- `frontend/admin/dist`
- `scripts/fetch-daily-briefings.ps1`
- `docs/DEPLOYMENT.md`
- `docs/README.md`
- `database/ai_blog.sql`

## 8. Linux 生产部署

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

- 数据库连接正确
- `auth.token_secret` 已替换为随机长串
- 后台默认密码已修改
- AI 与 OSS 的密钥正确

### 8.4 构建后端

```bash
cd /opt/codex-blog/backend
go mod tidy
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
go build -o /opt/codex-blog/bin/dailybriefing-fetcher ./cmd/dailybriefing_fetcher
```

### 8.5 构建前端

```bash
cd /opt/codex-blog/frontend/reader
npm install
npm run build

cd /opt/codex-blog/frontend/admin
npm install
npm run build
```

## 9. systemd 配置

创建后端服务：

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

## 10. Nginx 配置

推荐同域名部署：

- 读者端：`https://your-domain.com/`
- 后台：`https://your-domain.com/admin/`
- API：`https://your-domain.com/api/...`

配置示例：

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

如果你需要 HTTPS，再接 Let’s Encrypt 或你自己的证书。

## 11. Windows 服务器部署

### 11.1 构建后端

```powershell
cd D:\codex-blog\backend
go build -o D:\codex-blog\bin\codex-blog-server.exe .\cmd\server
go build -o D:\codex-blog\bin\dailybriefing-fetcher.exe .\cmd\dailybriefing_fetcher
```

### 11.2 构建前端

```powershell
cd D:\codex-blog\frontend\reader
npm install
npm run build

cd D:\codex-blog\frontend\admin
npm install
npm run build
```

### 11.3 启动建议

- 后端建议用 NSSM、WinSW 或计划任务托管
- 工作目录指向 `D:\codex-blog\backend`
- 静态文件可交给 IIS 或 Nginx for Windows

## 12. 每日简讯抓取

### 12.1 手动执行

Windows：

```powershell
cd D:\codex-blog
.\scripts\fetch-daily-briefings.ps1 -date 2026-04-13 -limit 10
```

Linux：

```bash
cd /opt/codex-blog/backend
/opt/codex-blog/bin/dailybriefing-fetcher -date 2026-04-13 -limit 10
```

### 12.2 Linux cron

```cron
30 8 * * * cd /opt/codex-blog/backend && /opt/codex-blog/bin/dailybriefing-fetcher -limit 10 >> /var/log/codex-blog-dailybriefing.log 2>&1
```

## 13. 部署后的检查项

至少检查这些地址和能力：

- `GET /health` 正常返回
- 读者端首页可以打开
- 管理后台可以登录
- `GET /api/public/system-configs` 返回系统配置
- 后台“系统设置 -> 系统配置”可以看到默认文案
- 修改一条系统配置并保存后，读者端刷新能看到变更
- 每日简讯列表可见
- OSS 同步状态可见

推荐检查命令：

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/public/articles
curl http://127.0.0.1:8080/api/public/daily-briefings
curl http://127.0.0.1:8080/api/public/system-configs
```

## 14. 更新流程

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

升级到当前版本时额外注意：

- 后端至少要重启一次，让 `system_configs` 自动迁移生效
- 如果用了旧版管理端静态文件，记得重新覆盖 `frontend/admin/dist`
- 读者端首页文案不再只写死在前端，优先读数据库

## 15. 常见问题

### 15.1 后台能打开但资源 404

优先检查：

- `frontend/admin/vite.config.js` 是否保留 `base: '/admin/'`
- Nginx 是否同时配置了 `/admin/` 和 `/admin/assets/`

### 15.2 后台报 `[vue/compiler-sfc] Missing semicolon`

这通常不是接口问题，而是前端开发服务缓存了旧的异常文件状态。处理方式：

- 先确认 `frontend/admin/src/App.vue` 本身是正常文件
- 重新执行一次 `npm run build`
- 重启管理端开发服务
- 浏览器使用 `Ctrl + F5` 强刷

### 15.3 系统配置修改后前台没变化

优先检查：

- 后台保存是否成功
- 读者端请求的是否是同一个后端
- 浏览器是否缓存旧页面
- `GET /api/public/system-configs` 是否已经返回新值

### 15.4 配置文件泄露

处理建议：

- 立刻旋转数据库密码、OSS 密钥、AI Key
- 清理仓库和部署机上的旧配置副本
- 不要再把真实配置发到公开聊天、截图或仓库里
