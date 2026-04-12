# codex-blog 部署文档

## 1. 部署目标

本项目包含三部分：

- `backend`：Go + Gin API 服务
- `frontend/reader`：读者端 Vue 应用
- `frontend/admin`：管理端 Vue 应用

推荐部署方式：

- 后端作为 systemd 服务运行在 `127.0.0.1:8080`
- 两个前端构建为静态文件，由 Nginx 托管
- Nginx 同时提供反向代理，将 `/api/` 转发到后端

---

## 2. 服务器准备

以下示例基于 Ubuntu 22.04（其他 Linux 发行版可按等价命令调整）。

### 2.1 安装基础环境

```bash
sudo apt update
sudo apt install -y git curl nginx mysql-server
```

### 2.2 安装 Go（1.22+）

```bash
curl -LO https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2.3 安装 Node.js（20+）

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
node -v
npm -v
```

---

## 3. 拉取代码

```bash
cd /opt
sudo git clone git@github.com:h19890209000-art/codex-blog.git
sudo chown -R $USER:$USER /opt/codex-blog
cd /opt/codex-blog
```

---

## 4. 后端部署

## 4.1 配置文件

```bash
cd /opt/codex-blog/backend
cp ./configs/config.example.json ./configs/config.local.json
```

按实际环境修改 `configs/config.local.json`，重点字段：

- `database.host / port / username / password / database_name`
- `auth.token_secret`
- `ai.providers.*.api_key`
- `oss.access_key_id / access_key_secret / bucket`

## 4.2 初始化数据库

先在 MySQL 中创建数据库：

```sql
CREATE DATABASE ai_blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

## 4.3 编译后端

```bash
cd /opt/codex-blog/backend
go mod tidy
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
```

## 4.4 systemd 托管

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

启动并设置开机自启：

```bash
sudo systemctl daemon-reload
sudo systemctl enable codex-blog-backend
sudo systemctl start codex-blog-backend
sudo systemctl status codex-blog-backend
```

---

## 5. 前端部署

## 5.1 构建读者端

```bash
cd /opt/codex-blog/frontend/reader
npm install
npm run build
```

## 5.2 构建管理端

```bash
cd /opt/codex-blog/frontend/admin
npm install
npm run build
```

---

## 6. Nginx 配置

创建站点配置：

```bash
sudo tee /etc/nginx/sites-available/codex-blog.conf > /dev/null << 'EOF'
server {
    listen 80;
    server_name _;

    root /opt/codex-blog/frontend/reader/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /admin/ {
        alias /opt/codex-blog/frontend/admin/dist/;
        try_files $uri $uri/ /index.html;
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
EOF
```

启用配置并重载：

```bash
sudo ln -sf /etc/nginx/sites-available/codex-blog.conf /etc/nginx/sites-enabled/codex-blog.conf
sudo nginx -t
sudo systemctl reload nginx
```

---

## 7. 验证部署

```bash
curl http://127.0.0.1/health
curl http://127.0.0.1/api/public/articles
```

浏览器访问：

- `http://<服务器IP>/` 读者端
- `http://<服务器IP>/admin/` 管理端

---

## 8. 更新流程

```bash
cd /opt/codex-blog
git pull

cd /opt/codex-blog/backend
go build -o /opt/codex-blog/bin/codex-blog-server ./cmd/server
sudo systemctl restart codex-blog-backend

cd /opt/codex-blog/frontend/reader
npm install
npm run build

cd /opt/codex-blog/frontend/admin
npm install
npm run build

sudo systemctl reload nginx
```

---

## 9. 常用排查命令

```bash
sudo systemctl status codex-blog-backend
sudo journalctl -u codex-blog-backend -f
sudo nginx -t
sudo tail -f /var/log/nginx/error.log
```
