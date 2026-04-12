/*
 Navicat Premium Dump SQL

 Source Server         : 本地
 Source Server Type    : MySQL
 Source Server Version : 80045 (8.0.45)
 Source Host           : 127.0.0.1:3306
 Source Schema         : ai_blog

 Target Server Type    : MySQL
 Target Server Version : 80045 (8.0.45)
 File Encoding         : 65001

 Date: 12/04/2026 18:14:46
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for article_tags
-- ----------------------------
DROP TABLE IF EXISTS `article_tags`;
CREATE TABLE `article_tags`  (
  `article_id` bigint NOT NULL,
  `tag_id` bigint NOT NULL,
  PRIMARY KEY (`article_id`, `tag_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of article_tags
-- ----------------------------
INSERT INTO `article_tags` VALUES (1, 1);
INSERT INTO `article_tags` VALUES (1, 2);

-- ----------------------------
-- Table structure for articles
-- ----------------------------
DROP TABLE IF EXISTS `articles`;
CREATE TABLE `articles`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `cover_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `status` tinyint NULL DEFAULT 0 COMMENT '0草稿 1发布',
  `view_count` bigint NULL DEFAULT 0,
  `category_id` bigint NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `source_type` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'manual',
  `source_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_path` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `last_synced_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_article_source`(`source_type` ASC, `source_key` ASC) USING BTREE,
  INDEX `idx_articles_title`(`title` ASC) USING BTREE,
  INDEX `idx_articles_category_id`(`category_id` ASC) USING BTREE,
  INDEX `idx_articles_deleted_at`(`deleted_at` ASC) USING BTREE,
  CONSTRAINT `fk_articles_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of articles
-- ----------------------------
INSERT INTO `articles` VALUES (1, 'Go 新手如何理解 Gin 路由分层', 'Gin 的 router 负责定义路由，controller 负责接收参数，service 负责业务逻辑，repository 负责读写数据。这样分层后，你就能快速定位问题到底出在哪一层。', '从 PHP 控制器思维迁移到 Go 的 router/controller/service 分层。', 'https://placehold.co/1200x630?text=Go+Gin', 1, 128, 1, '2026-04-06 19:28:02.962', '2026-04-06 19:28:02.962', NULL, 'manual', '', '', '', NULL);
INSERT INTO `articles` VALUES (4, 'TEST', 'TESTTESTTESTTEST', 'TESTTESTTESTTEST', '', 1, 0, 1, '2026-04-06 20:43:46.900', '2026-04-12 17:42:01.625', NULL, 'oss', 'GO/TEST', 'blog/GO/TEST.md', '45ac10ab5e6c364a03cf0e550733b6f559c59b01b99068a41c92337c3e3cf0de', '2026-04-12 17:42:01.614');
INSERT INTO `articles` VALUES (5, 'phptest', 'testtesttest', 'testtesttest', '', 1, 0, 4, '2026-04-06 20:43:46.919', '2026-04-12 17:42:01.645', NULL, 'oss', 'PHP/phptest', 'blog/PHP/phptest.md', '5c02f48662cfc611909ccb9cb26982973c53ccf5a12c3c0e1965f507826b4f38', '2026-04-12 17:42:01.641');
INSERT INTO `articles` VALUES (6, 'OpenClaw 2026 安装与部署指南', '# OpenClaw 2026 安装与部署指南\n\n**版本：2.60**  \n**最后更新：2026年4月**  \n**适用系统：Windows、macOS、Linux**\n\n## 目录\n\n1. 系统要求\n2. 安装前准备\n3. 方法一：官方一键脚本安装\n4. 方法二：Windows专属一键部署包\n5. 首次配置与初始化\n6. 常见问题\n7. 验证安装\n![](https://nan-huai-obsidian.oss-cn-guangzhou.aliyuncs.com/blog/%E6%95%99%E7%A8%8B/DSC_3367.jpg)\n## 系统要求\n\n### 基础要求\n\n- **操作系统**：\n  - Windows 10/11 (64位)\n  - macOS 12 或更高版本\n  - 主流 Linux 发行版 (Ubuntu 22.04+, CentOS 7+)\n- **内存**：至少 4GB RAM\n- **磁盘空间**：至少 2GB 可用空间\n- **网络**：需要稳定的互联网连接\n\n### 推荐配置\n\n- **CPU**：Intel i5 或同等性能处理器\n- **内存**：8GB 或更高\n- **网络**：宽带连接，支持 HTTPS 访问\n\n## 安装前准备\n\n### 1. 环境检查\n\n在开始安装前，请确保您的系统满足以下条件：\n\n- 系统管理员权限（Windows 需要管理员权限）\n- 足够的磁盘空间\n- 稳定的网络连接\n\n### 2. 关闭安全软件（仅限Windows）\n\n为避免安装过程被拦截，请暂时关闭：\n\n- 360安全卫士\n- 火绒安全\n- Windows Defender\n- 其他实时防护软件\n\n## 方法一：官方一键脚本安装（推荐）\n\n### Windows 系统安装步骤\n\n1. **以管理员身份打开 PowerShell**  \n   - 在开始菜单搜索 \"PowerShell\"  \n   - 右键点击 \"Windows PowerShell\"  \n   - 选择 \"以管理员身份运行\"\n\n2. **设置执行策略**  \n   在 PowerShell 窗口中执行以下命令：  \n   `[命令缺失，请参考官方文档]`  \n   当提示 \"是否要更改执行策略？\" 时，输入 `Y` 并按回车确认。\n\n3. **执行安装脚本**  \n   复制并执行以下官方安装命令：  \n   `[命令缺失，请参考官方文档]`\n\n4. **等待安装完成**  \n   - 脚本将自动检测系统环境  \n   - 自动下载并安装 Node.js 等依赖  \n   - 全程无需手动干预  \n   - 安装时间：3-5 分钟\n\n### macOS / Linux 系统安装步骤\n\n1. **打开终端**  \n   在系统中找到并打开 \"终端\" 应用。\n\n2. **执行安装脚本**  \n   在终端中执行以下命令：  \n   `[命令缺失，请参考官方文档]`\n\n3. **输入管理员密码**  \n   根据系统提示，输入当前用户的管理员密码。\n\n## 方法二：Windows专属一键部署包\n\n1. **下载部署包**  \n   从官方渠道下载 `Openclaw-Windows-2.60.zip` 一键部署包。\n\n2. **解压文件**  \n   使用 WinRAR 或 7-Zip 将压缩包解压到目标文件夹。\n\n3. **运行启动器**  \n   进入解压后的 `Openclaw-win` 文件夹，双击运行：  \n   `OpenClaw-Setup.exe`\n\n4. **配置安装路径**  \n   - 选择一个 **纯英文、无空格** 的安装路径（如 `D:\\OpenClaw`）  \n   - 勾选 \"我同意软件许可协议\"  \n   - 点击 \"开始安装\"\n\n5. **等待部署完成**  \n   部署过程将自动完成，无需手动操作。\n\n## 首次配置与初始化\n\n1. **启动初始化向导**  \n   打开终端或 PowerShell，执行：  \n   `[命令缺失，请参考官方文档]`\n\n2. **按向导提示操作**\n\n   **选择模型提供商**  \n   使用方向键选择您要使用的 AI 模型：\n\n   **输入 API Key**  \n   根据选择的模型，输入对应的 API Key：  \n   - **Kimi**：从 Moonshot AI 官网获取 API Key  \n   - **OpenAI**：从 OpenAI 平台获取 API Key  \n   - **Claude**：从 Anthropic 平台获取 API Key\n\n   **选择配置模式**  \n   推荐新手选择：  \n   `[配置模式名称缺失，请参考官方文档]`\n\n   **完成配置**  \n   后续步骤按提示选择 Yes 或直接回车，直到出现：  \n   `[完成提示信息缺失]`\n\n## 验证安装\n\n1. **检查版本信息**  \n   执行以下命令验证安装是否成功：  \n   `[命令缺失，请参考官方文档]`  \n   正常输出应包含版本号信息。\n\n2. **运行健康检查**  \n   执行官方诊断命令：  \n   `[命令缺失，请参考官方文档]`  \n   当显示 \"All systems operational\" 或类似成功信息时，表示安装配置完成。\n\n3. **启动服务**  \n   启动 OpenClaw 服务：  \n   `[命令缺失，请参考官方文档]`\n\n## 常见问题\n\n**❓ 安装时提示权限错误怎么办？**\n\n- **Windows 解决方案**：  \n  - 确保以管理员身份运行 PowerShell  \n  - 检查执行策略设置是否正确\n\n- **macOS/Linux 解决方案**：  \n  - 在命令前加上 `sudo`  \n  - 检查用户是否在管理员组\n\n**❓ 杀毒软件误报或拦截怎么办？**  \n这是正常现象，因为安装脚本需要修改系统配置。请：  \n1. 暂时关闭实时防护  \n2. 将安装目录添加到信任白名单  \n3. 或使用一键部署包方式安装\n\n**❓ openclaw 命令找不到？**\n\n**解决方案**：  \n1. 关闭并重新打开终端  \n2. 检查环境变量 PATH 是否包含安装路径  \n3. 重启电脑后重试  \n4. 重新执行安装脚本\n\n**❓ 依赖安装慢或失败？**\n\n**使用国内镜像源**：  \n`[命令缺失，请参考官方文档]`\n\n**❓ 安装过程中断怎么办？**\n\n**清理残留文件**：  \n`[命令缺失，请参考官方文档]`\n\n## 卸载指南\n\n### 完全卸载 OpenClaw\n\n**Windows**  \n1. 通过控制面板卸载 Node.js  \n2. 删除安装目录：`C:\\Program Files\\OpenClaw`  \n3. 删除用户目录下的配置：`C:\\Users\\[用户名]\\.openclaw`\n\n**macOS/Linux**  \n`[命令缺失，请参考官方文档]`\n\n## 技术支持\n\n遇到问题时，可通过以下方式获取帮助：\n\n**官方文档**  \n- 访问：https://openclaw.ai/docs  \n- 查看详细的安装指南和 API 文档\n\n**社区支持**  \n- GitHub Issues：https://github.com/openclaw/openclaw/issues  \n- Discord 社区：加入官方开发者社区\n\n**商业支持**  \n如需企业级支持服务，请联系官方商务团队。\n\n---\n\n**文档版本：v2.60**  \n**最后更新：2026年4月6日**  \n**版权所有 © 2026 OpenClaw 项目组**  \n*(AI生成)*', '版本：2.60 最后更新：2026年4月 适用系统：Windows、macOS、Linux 1. 系统要求 2. 安装前准备 3. 方法一：官方一键脚本安装 4. 方法二：Windows专属一键部署包', 'https://nan-huai-obsidian.oss-cn-guangzhou.aliyuncs.com/blog/%E6%95%99%E7%A8%8B/DSC_3367.jpg', 1, 0, 5, '2026-04-06 20:43:46.936', '2026-04-06 22:56:49.009', '2026-04-12 16:09:01.126', 'oss', '教程/OpenClaw 2026 安装与部署指南', 'blog/教程/OpenClaw 2026 安装与部署指南.md', '94a0535868abaf2f3b34275d01de76e2ee2e0f82d6a16b67bd4707d14579b2d5', '2026-04-06 22:56:49.006');
INSERT INTO `articles` VALUES (7, 'Linux 安装 nvm', '## 安装教程\nnvm下载链接：curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash\n下载完成后，使用以下命令使 nvm 生效：source ~/.bashrc\n最后检查是否安装完成：nvm --version\n## 指令\n下载指定版本：nvm install <version>\n安装最新稳定版：nvm install node\n安装最新 LTS 版本：nvm install --lts\n安装特定代号的 LTS 版本（如 iron）：nvm install lts/iron\n卸载指定版本：\nnvm uninstall <version>\n使用指定版本：\nnvm use <version> # 如 nvm use 16.14.0\n使用系统自带的 Node.js：nvm use system\n‌查看当前使用的 Node.js 版本：nvm current\n‌列出所有已安装的 Node.js 版本：nvm ls\n列出所有可安装的远程版本：nvm ls-remote   nvm ls-remote --lts # 仅显示 LTS 版本\n\n', 'nvm下载链接：curl  o  https://raw.githubusercontent.com/nvm sh/nvm/v0.38.0/install.sh | bash 下载完成后，使用以下命令使 nvm 生效：source ~/.bashrc', '', 1, 0, 5, '2026-04-12 16:09:01.048', '2026-04-12 17:42:01.690', NULL, 'oss', '教程/Linux 安装 nvm', 'blog/教程/Linux 安装 nvm.md', '46878f1f5c29779e6d77c6bb6e23938518b37db5419381267501a432046b1362', '2026-04-12 17:42:01.686');
INSERT INTO `articles` VALUES (8, '指令', '[openclaw官网](https://openclaw.ai/)\n安装nvm 然后下载指定node版本：[[Linux 安装 nvm]]\nnpm全局下载龙虾：npm i -g openclaw\n初始化：openclaw onboard\n\n# 指令\n## 一、快速入门\n\n### 1.1 查看帮助信息\n\n```bash\n# 查看所有命令\nopenclaw --help\n\n# 查看版本号\nopenclaw --version\n\n# 查看特定命令的帮助\nopenclaw <command> --help\n\n# 示例：查看 config 命令帮助\nopenclaw config --help\n```\n\n### 1.2 初始化配置\n\n```bash\n# 首次安装后初始化配置\nopenclaw setup\n\n# 交互式引导配置（推荐新手）\nopenclaw onboard\n\n# 打开控制面板\nopenclaw dashboard\n```\n\n## 二、配置管理命令\n\n### 2.1 查看配置\n\n```bash\n# 查看完整配置\nopenclaw config get\n\n# 查看特定配置项\nopenclaw config get models.default\nopenclaw config get providers.mistral.apiKey\n\n# 查看特定部分配置\nopenclaw config get --section models\nopenclaw config get --section providers\n```\n\n**示例输出：**\n\n```bash\n{\n  \"models\": {\n    \"default\": \"mistral:mixtral-8x7b\"\n  },\n  \"providers\": {\n    \"mistral\": {\n      \"apiKey\": \"***REDACTED***\"\n    }\n  }\n}\n```\n\n### 2.2 设置配置\n\n```bash\n# 设置默认模型\nopenclaw config set models.default mistral:mixtral-8x7b\n\n# 设置快速模型\nopenclaw config set models.fast mistral:mistral-7b\n\n# 配置 Mistral API Key\nopenclaw config set providers.mistral.apiKey YOUR_API_KEY_HERE\n\n# 启用缓存\nopenclaw config set cache.enabled true\nopenclaw config set cache.maxSize 5000\n```\n\n### 2.3 删除配置\n\n```bash\n# 删除特定配置项\nopenclaw config unset models.fast\n\n# 重置某个节点\nopenclaw config unset models\n```\n\n### 2.4 配置向导\n\n```bash\n# 打开完整配置向导\nopenclaw configure\n\n# 打开特定部分配置向导\nopenclaw configure --section models\nopenclaw configure --section providers\nopenclaw configure --section channels\n```\n\n## 三、Gateway 控制命令\n\n### 3.1 启动/停止 Gateway\n\n```bash\n# 启动 Gateway（默认端口 18789）\nopenclaw gateway start\n\n# 自定义端口启动\nopenclaw gateway start --port 19000\n\n# 强制启动（杀死占用进程）\nopenclaw gateway start --force\n\n# 停止 Gateway\nopenclaw gateway stop\n\n# 重启 Gateway\nopenclaw gateway restart\n\n# 查看运行状态\nopenclaw gateway status\n```\n\n### 3.2 运行时 Gateway\n\n```bash\n# 前台运行 Gateway（调试用）\nopenclaw gateway\n\n# 开发模式运行（隔离状态）\nopenclaw --dev gateway\n\n# 查看健康状态\nopenclaw health\n```\n\n### 3.3 查看日志\n\n```bash\n# 查看实时日志\nopenclaw logs\n\n# 查看最近 50 行日志\nopenclaw logs --lines 50\n\n# 查看错误日志\nopenclaw logs --filter error\n\n# 持续监控日志\nopenclaw logs --follow\n```\n\n### 3.4 系统服务管理\n\n```bash\n# 使用 systemd 管理（推荐生产环境）\nsudo systemctl start openclaw-gateway\nsudo systemctl stop openclaw-gateway\nsudo systemctl restart openclaw-gateway\nsudo systemctl status openclaw-gateway\n\n# 开机自启动\nsudo systemctl enable openclaw-gateway\n```\n\n## 四、消息发送命令\n\n### 4.1 发送消息\n\n```bash\n# 发送消息到当前会话\nopenclaw message send --message \"Hello\"\n\n# 发送到特定目标（Telegram）\nopenclaw message send \\\n  --channel telegram \\\n  --target @mychat \\\n  --message \"Hello from OpenClaw\"\n\n# 发送到特定目标（WhatsApp）\nopenclaw message send \\\n  --channel whatsapp \\\n  --target +8613800138000 \\\n  --message \"您好\"\n\n# 发送到 Slack 频道\nopenclaw message send \\\n  --channel slack \\\n  --target C1234567890 \\\n  --message \"@channel 重要通知\"\n```\n\n### 4.2 发送媒体文件\n\n```bash\n# 发送图片\nopenclaw message send \\\n  --channel telegram \\\n  --target @mychat \\\n  --media /tmp/photo.jpg \\\n  --caption \"这是一张图片\"\n\n# 发送音频\nopenclaw message send \\\n  --channel whatsapp \\\n  --target +8613800138000 \\\n  --media /tmp/voice.mp3\n\n# 发送文档\nopenclaw message send \\\n  --channel telegram \\\n  --target @mychat \\\n  --media /tmp/report.pdf\n```\n\n### 4.3 高级消息功能\n\n```bash\n# 发送 JSON 格式（脚本自动化）\nopenclaw message send \\\n  --target @mychat \\\n  --message \"Hello\" \\\n  --json\n\n# 回复消息\nopenclaw message send \\\n  --target @mychat \\\n  --message \"收到\" \\\n  --replyTo 12345\n\n# 指定频道\nopenclaw message send \\\n  --channel discord \\\n  --target channel:1234567890 \\\n  --message \"Hello\"\n```\n\n### 4.4 频道动作（投票、反应等）\n\n```bash\n# 创建投票（Telegram）\nopenclaw message send \\\n  --channel telegram \\\n  --target @mychat \\\n  --pollQuestion \"OpenClaw 好用吗？\" \\\n  --pollOption 非常好用 \\\n  --pollOption 一般 \\\n  --pollOption 还需要改进 \\\n  --pollDurationHours 24\n\n# 发送反应（Discord）\nopenclaw message send \\\n  --channel discord \\\n  --messageId 1234567890 \\\n  --emoji 👍 \\\n  --action react\n```\n\n## 五、技能管理命令\n\n### 5.1 查看技能列表\n\n```bash\n# 查看所有已安装技能\nopenclaw skills list\n\n# 搜索技能\nopenclaw skills search weather\n\n# 查看技能详情\nopenclaw skills show weather\n```\n\n### 5.2 安装/卸载技能\n\n```bash\n# 安装技能\nopenclaw skills install weather\n\n# 从指定来源安装\nopenclaw skills install weather --source github\n\n# 指定版本安装\nopenclaw skills install weather@1.2.0\n\n# 卸载技能\nopenclaw skills uninstall weather\n```\n\n### 5.3 更新技能\n\n```bash\n# 更新所有技能\nopenclaw skills update\n\n# 更新特定技能\nopenclaw skills update weather\n\n# 同步技能\nopenclaw skills sync\n```\n\n### 5.4 技能开发\n\n```bash\n# 创建新技能\nopenclaw skills create my-skill\n\n# 验证技能\nopenclaw skills validate my-skill\n\n# 打包技能\nopenclaw skills pack my-skill\n```\n\n## 六、模型配置命令\n\n### 6.1 查看模型\n\n```bash\n# 查看所有配置的模型\nopenclaw models list\n\n# 查看默认模型\nopenclaw models default\n\n# 查看模型详情\nopenclaw models show mistral:mixtral-8x7b\n```\n\n### 6.2 配置模型\n\n```bash\n# 设置默认模型\nopenclaw models set-default mistral:mixtral-8x7b\n\n# 添加新模型\nopenclaw models add \\\n  --name my-model \\\n  --provider mistral \\\n  --id mistral-medium \\\n  --maxTokens 8192\n\n# 测试模型\nopenclaw models test \\\n  --model mistral:mixtral-8x7b \\\n  --prompt \"Hello, OpenClaw!\"\n```\n\n### 6.3 模型切换\n\n```bash\n# 临时指定模型\nopenclaw agent \\\n  --model mistral:mistral-7b \\\n  --message \"快速响应\"\n\n# 使用内置模型别名\nopenclaw agent --model fast --message \"这个用fast模型\"\nopenclaw agent --model premium --message \"这个用premium模型\"\n```\n\n## 七、频道管理命令\n\n### 7.1 查看频道\n\n```bash\n# 查看所有配置的频道\nopenclaw channels list\n\n# 查看频道状态\nopenclaw channels status\n\n# 查看特定频道详情\nopenclaw channels show telegram\n```\n\n### 7.2 登录频道\n\n```bash\n# Telegram 登录\nopenclaw channels login --channel telegram\n\n# WhatsApp 登录（会显示 QR 码）\nopenclaw channels login --channel whatsapp --verbose\n\n# Slack 登录\nopenclaw channels login --channel slack\n\n# Discord 登录\nopenclaw channels login --channel discord\n```\n\n### 7.3 频道测试\n\n```bash\n# 测试频道连接\nopenclaw channels test --channel telegram\n\n# 发送测试消息\nopenclaw channels test \\\n  --channel telegram \\\n  --target @mychat \\\n  --message \"测试消息\"\n```\n\n### 7.4 频道配置\n\n```bash\n# 配置频道\nopenclaw channels configure --channel telegram\n\n# 更新频道 Token\nopenclaw channels update \\\n  --channel telegram \\\n  --token NEW_TOKEN\n\n# 启用/禁用频道\nopenclaw channels enable telegram\nopenclaw channels disable telegram\n```\n\n## 八、会话管理命令\n\n### 8.1 查看会话\n\n```bash\n# 列出所有会话\nopenclaw sessions\n\n# 列出活跃会话\nopenclaw sessions --active\n\n# 列出特定频道的会话\nopenclaw sessions --channel telegram\n\n# 显示最近 10 个会话\nopenclaw sessions --limit 10\n```\n\n### 8.2 查看会话历史\n\n```text\n# 查看特定会话的历史\nopenclaw sessions history <session-key>\n\n# 查看最近的消息\nopenclaw sessions history <session-key> --limit 20\n\n# 导出会话历史\nopenclaw sessions history <session-key> --export > history.json\n```\n\n### 8.3 会话操作\n\n```bash\n# 发送消息到会话\nopenclaw sessions send \\\n  --session <session-key> \\\n  --message \"你好\"\n\n# 重置会话\nopenclaw sessions reset <session-key>\n\n# 删除会话\nopenclaw sessions delete <session-key>\n```\n\n## 九、节点管理命令（智能家居控制）\n\n### 9.1 查看节点\n\n```text\n# 查看所有配对的节点\nopenclaw nodes list\n\n# 查看节点状态\nopenclaw nodes status\n\n# 描述节点详情\nopenclaw nodes describe <node-id>\n```\n\n### 9.2 节点操作\n\n```text\n# 发送通知到节点\nopenclaw nodes notify \\\n  --node my-phone \\\n  --title \"提醒\" \\\n  --body \"该吃饭了\"\n\n# 设置推送优先级\nopenclaw nodes notify \\\n  --node my-phone \\\n  --priority timeSensitive \\\n  --title \"紧急通知\" \\\n  --body \"快递到了\"\n\n# 查看相册（手机）\nopenclaw nodes camera-list --node my-phone\n\n# 拍照\nopenclaw nodes camera-snap \\\n  --node my-phone \\\n  --facing back \\\n  --output /tmp/photo.jpg\n```\n\n### 9.3 节点配对\n\n```text\n# 启动配对\nopenclaw node pairing start\n\n# 查看待配对节点\nopenclaw nodes pending\n\n# 批准配对\nopenclaw nodes approve --node <node-id>\n\n# 拒绝配对\nopenclaw nodes reject --node <node-id>\n```\n\n## 十、记忆管理命令\n\n### 10.1 搜索记忆\n\n```text\n# 搜索记忆\nopenclaw memory search \"OpenClaw 配置\"\n\n# 搜索并显示多行上下文\nopenclaw memory search \"配置\" --lines 5\n\n# 搜索特定路径的记忆\nopenclaw memory search \"配置\" --path MEMORY.md\n\n# 限制结果数量\nopenclaw memory search \"配置\" --maxResults 10\n```\n\n### 10.2 记忆操作\n\n```text\n# 查看记忆统计\nopenclaw memory stats\n\n# 清理过期记忆\nopenclaw memory clean\n\n# 备份记忆\nopenclaw memory backup --output /tmp/memory-backup.json\n```\n\n## 十一、[Cron 定时任务](https://zhida.zhihu.com/search?content_id=270757527&content_type=Article&match_order=1&q=Cron+%E5%AE%9A%E6%97%B6%E4%BB%BB%E5%8A%A1&zd_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ6aGlkYV9zZXJ2ZXIiLCJleHAiOjE3NzYwNjczNzAsInEiOiJDcm9uIOWumuaXtuS7u-WKoSIsInpoaWRhX3NvdXJjZSI6ImVudGl0eSIsImNvbnRlbnRfaWQiOjI3MDc1NzUyNywiY29udGVudF90eXBlIjoiQXJ0aWNsZSIsIm1hdGNoX29yZGVyIjoxLCJ6ZF90b2tlbiI6bnVsbH0.OYYi89F6OU3DGHA-zaeePLlZVc41rRWc7k-mXKy7hoQ&zhida_source=entity)命令\n\n### 11.1 查看 Cron 任务\n\n```bash\n# 列出所有任务\nopenclaw cron list\n\n# 查看任务运行历史\nopenclaw cron runs <job-id>\n\n# 查看调度器状态\nopenclaw cron status\n```\n\n### 11.2 创建 Cron 任务\n\n```bash\n# 创建定时任务（每天凌晨触发）\nopenclaw cron add \\\n  --name \"daily-report\" \\\n  --schedule \"0 0 * * *\" \\\n  --text \"生成每日报告\"\n\n# 创建重复任务（每 30 分钟）\nopenclaw cron add \\\n  --name \"check-notifications\" \\\n  --schedule \"*/30 * * * *\" \\\n  --text \"检查通知\"\n\n# 创建单次任务（特定时间）\nopenclaw cron add \\\n  --name \"special-task\" \\\n  --schedule \"at\" \\\n  --at \"2026-03-01T10:00:00\" \\\n  --text \"执行特殊任务\"\n```\n\n### 11.3 Cron 任务操作\n\n```bash\n# 立即运行任务\nopenclaw cron run <job-id>\n\n# 更新任务\nopenclaw cron update <job-id> --schedule \"0 6 * * *\"\n\n# 删除任务\nopenclaw cron remove <job-id>\n\n# 发送唤醒事件\nopenclaw cron wake --text \"检查新消息\"\n```\n\n## 十二、[系统命令](https://zhida.zhihu.com/search?content_id=270757527&content_type=Article&match_order=1&q=%E7%B3%BB%E7%BB%9F%E5%91%BD%E4%BB%A4&zd_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ6aGlkYV9zZXJ2ZXIiLCJleHAiOjE3NzYwNjczNzAsInEiOiLns7vnu5_lkb3ku6QiLCJ6aGlkYV9zb3VyY2UiOiJlbnRpdHkiLCJjb250ZW50X2lkIjoyNzA3NTc1MjcsImNvbnRlbnRfdHlwZSI6IkFydGljbGUiLCJtYXRjaF9vcmRlciI6MSwiemRfdG9rZW4iOm51bGx9.XTMTb5oahajTL52bExqo7e8DRcdRzUHaHoLHXD8G4fI&zhida_source=entity)\n\n### 12.1 健康检查\n\n```text\n# 运行健康检查\nopenclaw doctor\n\n# 快速修复常见问题\nopenclaw doctor --fix\n\n# 检查特定组件\nopenclaw doctor --check gateway\nopenclaw doctor --check channels\n```\n\n### 12.2 系统状态\n\n```text\n# 查看频道健康状态\nopenclaw status\n\n# 查看系统事件\nopenclaw system events\n\n# 查看心跳状态\nopenclaw system heartbeat\n```\n\n### 12.3 安全检查\n\n```text\n# 运行安全检查\nopenclaw security audit\n\n# 检查权限配置\nopenclaw security check-permissions\n\n# 检查 API Key 有效性\nopenclaw security verify-keys\n```\n\n## 十三、[插件管理](https://zhida.zhihu.com/search?content_id=270757527&content_type=Article&match_order=1&q=%E6%8F%92%E4%BB%B6%E7%AE%A1%E7%90%86&zd_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ6aGlkYV9zZXJ2ZXIiLCJleHAiOjE3NzYwNjczNzAsInEiOiLmj5Lku7bnrqHnkIYiLCJ6aGlkYV9zb3VyY2UiOiJlbnRpdHkiLCJjb250ZW50X2lkIjoyNzA3NTc1MjcsImNvbnRlbnRfdHlwZSI6IkFydGljbGUiLCJtYXRjaF9vcmRlciI6MSwiemRfdG9rZW4iOm51bGx9.MkbxuNj7wfuBGHZ9dXZpCwwhMaQ-xQkgsKrrWwD7eto&zhida_source=entity)命令\n\n### 13.1 查看插件\n\n```text\n# 查看已安装插件\nopenclaw plugins list\n\n# 查看插件详情\nopenclaw plugins show <plugin-name>\n\n# 检查插件状态\nopenclaw plugins status\n```\n\n### 13.2 插件操作\n\n```text\n# 启用插件\nopenclaw plugins enable feishu\n\n# 禁用插件\nopenclaw plugins disable feishu\n\n# 重启插件\nopenclaw plugins restart feishu\n\n# 更新插件\nopenclaw plugins update\n```\n\n## 十四、浏览器控制命令\n\n### 14.1 启动/停止浏览器\n\n```text\n# 启动浏览器\nopenclaw browser start\n\n# 停止浏览器\nopenclaw browser stop\n\n# 切换配置\nopenclaw browser start --profile chrome\nopenclaw browser start --profile openclaw\n```\n\n### 14.2 浏览器操作\n\n```text\n# 打开网页\nopenclaw browser open https://example.com\n\n# 截图\nopenclaw browser screenshot --output /tmp/screenshot.png\n\n# 获取快照\nopenclaw browser snapshot --target main\n\n# 查看标签页\nopenclaw browser tabs\n```\n\n## 十五、更新与维护命令\n\n### 15.1 更新 OpenClaw\n\n```bash\n# 查看更新\nopenclaw update --dry-run\n\n# 执行更新\nopenclaw update\n\n# 更新到特定版本\nopenclaw update --tag 2026.2.22\n\n# 更新到 Beta 版\nopenclaw update --channel beta\n```\n\n### 15.2 配置文件管理\n\n```bash\n# 备份配置\ncp ~/.openclaw/config.json ~/.openclaw/config.json.backup\n\n# 重置配置（保留 CLI）\nopenclaw reset\n\n# 完全卸载（包括数据）\nopenclaw uninstall\n```\n\n### 15.3 Shell 自动补全\n\n```text\n# 生成 Bash 补全脚本\nopenclaw completion bash > ~/.openclaw-completion\n\n# 生成 Zsh 补全脚本\nopenclaw completion zsh > ~/.zsh-completion\n\n# 应用补全（Bash）\necho \"source ~/.openclaw-completion\" >> ~/.bashrc\nsource ~/.bashrc\n```\n\n## 十六、常用组合命令\n\n### 16.1 快速部署新技能\n\n```text\n# 一键创建、开发、测试技能\nopenclaw skills create my-new-skill \\\n  && cd ~/.openclaw/workspace/skills/my-new-skill \\\n  && vim SKILL.md\n```\n\n### 16.2 批量发送通知\n\n```text\n# 发送到多个目标\nopenclaw message send --target @user1 --message \"通知内容\"\nopenclaw message send --target @user2 --message \"通知内容\"\nopenclaw message send --target @user3 --message \"通知内容\"\n\n# 使用循环批量发送（需要脚本配合）\nfor target in user1 user2 user3; do\n  openclaw message send --target @$target --message \"通知内容\"\ndone\n```\n\n### 16.3 Gateway 重启 + 验证\n\n```text\n# 重启并验证\nopenclaw gateway restart \\\n  && sleep 5 \\\n  && openclaw gateway status \\\n  && openclaw health\n```\n\n### 16.4 配置备份 + 更新\n\n```text\n# 安全更新流程\ncp ~/.openclaw/config.json ~/.openclaw/config.json.backup \\\n  && openclaw update --dry-run \\\n  && openclaw update \\\n  && openclaw gateway restart\n```\n\n### 16.5 每日报告生成\n\n```text\n# 生成每日报告（Cron 脚本）\n0 9 * * * openclaw cron add \\\n  --name daily-report \\\n  --schedule \"0 9 * * *\" \\\n  --text \"生成昨日数据分析报告\"\n```\n\n## 十七、故障排除命令\n\n### 17.1 Gateway 无法启动\n\n```text\n# 检查端口占用\nsudo lsof -i :18789\n\n# 强制重启\nopenclaw gateway start --force\n\n# 查看错误日志\nopenclaw logs --filter error\n\n# 运行健康检查\nopenclaw doctor\n```\n\n### 17.2 消息发送失败\n\n```text\n# 检查频道状态\nopenclaw channels status\n\n# 测试频道\nopenclaw channels test --channel telegram\n\n# 重新登录频道\nopenclaw channels login --channel telegram\n\n# 查看详细日志\nopenclaw message send --target @mychat --message \"Test\" --verbose\n```\n\n### 17.3 模型调用失败\n\n```text\n# 检查 API Key\nopenclaw config get providers.openai.apiKey\n\n# 测试模型\nopenclaw models test --model mistral:mixtral-8x7b --prompt \"test\"\n\n# 检查网络连接\nping api.openai.com\n\n# 运行诊断\nopenclaw doctor --check models\n```\n\n### 17.4 技能加载失败\n\n```text\n# 检查技能列表\nopenclaw skills list\n\n# 验证技能\nopenclaw skills validate my-skill\n\n# 重新安装技能\nopenclaw skills uninstall my-skill\nopenclaw skills install my-skill\n\n# 查看错误日志\nopenclaw logs --filter skill\n```\n\n## 十八、生产环境最佳实践\n\n### 18.1 使用环境变量\n\n```text\n# 推荐方式：使用环境变量存储敏感信息\nexport OPENAI_API_KEY=\"sk-xxx\"\nexport MISTRAL_API_KEY=\"xxx\"\nexport TELEGRAM_BOT_TOKEN=\"xxx\"\n\n# 然后启动 Gateway\nopenclaw gateway start\n```\n\n### 18.2 配置文件权限\n\n```text\n# 限制配置文件权限\nchmod 600 ~/.openclaw/config.json\n\n# 检查权限\nls -la ~/.openclaw/config.json\n```\n\n### 18.3 日志管理\n\n```text\n# 配置日志轮转\nsudo tee /etc/logrotate.d/openclaw <<EOF\n~/.openclaw/logs/*.log {\n    daily\n    rotate 7\n    compress\n    missingok\n    notifempty\n}\nEOF\n```\n\n### 18.4 监控与告警\n\n```text\n# 检查 Gateway 状态（Cron 脚本）\n*/5 * * * * openclaw health || \\\n  openclaw message send --target @admin --message \"Gateway 宕机！\"\n```\n\n## 十九、快捷别名配置\n\n### 19.1 创建常用别名\n\n```text\n# 添加到 ~/.bashrc 或 ~/.zshrc\nalias oc=\'openclaw\'\nalias ocg=\'openclaw gateway\'\nalias ocgl=\'openclaw logs --follow\'\nalias ocs=\'openclaw sessions\'\nalias ocm=\'openclaw message send\'\nalias occ=\'openclaw channels\'\n\n# 应用别名\nsource ~/.bashrc  # Bash\nsource ~/.zshrc   # Zsh\n```\n\n### 19.2 使用别名\n\n```text\n# 启动 Gateway\nocg start\n\n# 查看实时日志\nocgl\n\n# 发送消息\nocm --target @mychat --message \"Hello\"\n\n# 查看会话\nocs\n```\n', '[openclaw官网](https://openclaw.ai/) 安装nvm 然后下载指定node版本：[[Linux 安装 nvm]] npm全局下载龙虾：npm i  g openclaw 初始化：openclaw onboard', '', 1, 0, 5, '2026-04-12 16:09:01.088', '2026-04-12 17:42:01.725', NULL, 'oss', '教程/OpenClaw 安装', 'blog/教程/OpenClaw 安装.md', '62e5fbfeedc1c24e2396e298de3e49011adf92971cdd193c4d32abd9f693057e', '2026-04-12 17:42:01.719');
INSERT INTO `articles` VALUES (9, 'github配置秘钥', '- **查本地是否存在密钥：** 查看 `~/.ssh` 文件夹下是否有 `id_rsa.pub` 或 `id_ed25519.pub` 文件。\n    \n- **生成新密钥：** 如果没有，运行 `ssh-keygen -t ed25519 -C \"你的邮箱\"`。\n    \n- **添加到 GitHub：** 复制公钥内容，前往 **GitHub Settings -> SSH and GPG keys -> New SSH key**，将其粘贴进去。\n\n有时候你生成了密钥，但系统当前的 SSH 代理（ssh-agent）并没有加载它，导致连接时没有发送密钥。\n解决方法：在终端运行以下命令手动添加\neval \"$(ssh-agent -s)\" ssh-add ~/.ssh/id_ed25519     # 如果你的文件名是 id_rsa，请相应修改 ', '查本地是否存在密钥： 查看 ~/.ssh 文件夹下是否有 id_rsa.pub 或 id_ed25519.pub 文件。 生成新密钥： 如果没有，运行 ssh keygen  t ed25519  C \"你的邮箱\"。', '', 1, 0, 5, '2026-04-12 16:09:01.122', '2026-04-12 17:42:01.775', NULL, 'oss', '教程/github配置秘钥', 'blog/教程/github配置秘钥.md', '79801e1e837751b2a1dab401f0ed9f3153827da669229fd3c93d1d0fe0408535', '2026-04-12 17:42:01.769');

-- ----------------------------
-- Table structure for categories
-- ----------------------------
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_categories_name`(`name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of categories
-- ----------------------------
INSERT INTO `categories` VALUES (1, 'Go', '2026-04-06 19:28:02.956', '2026-04-12 17:42:01.625');
INSERT INTO `categories` VALUES (2, '????', '2026-04-06 19:42:56.814', '2026-04-06 19:42:56.814');
INSERT INTO `categories` VALUES (3, 'temp', '2026-04-06 19:43:12.313', '2026-04-06 19:43:12.313');
INSERT INTO `categories` VALUES (4, 'PHP', '2026-04-06 20:43:46.916', '2026-04-12 17:42:01.644');
INSERT INTO `categories` VALUES (5, '教程', '2026-04-06 20:43:46.935', '2026-04-12 17:42:01.775');

-- ----------------------------
-- Table structure for comments
-- ----------------------------
DROP TABLE IF EXISTS `comments`;
CREATE TABLE `comments`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `article_id` bigint NULL DEFAULT NULL,
  `author` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'approved',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_comments_article_id`(`article_id` ASC) USING BTREE,
  CONSTRAINT `fk_comments_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of comments
-- ----------------------------
INSERT INTO `comments` VALUES (1, 1, '小白读者', '原来 router 和 controller 的职责真的不一样，这里终于看明白了。', 'approved', '2026-04-06 19:28:02.964', '2026-04-06 19:28:02.964');
INSERT INTO `comments` VALUES (2, 1, 'reader-test', 'reader-side-comment', 'approved', '2026-04-06 19:55:08.859', '2026-04-06 19:55:08.859');

-- ----------------------------
-- Table structure for daily_briefings
-- ----------------------------
DROP TABLE IF EXISTS `daily_briefings`;
CREATE TABLE `daily_briefings`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `briefing_date` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `source_name` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_url` varchar(800) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'manual',
  `status` tinyint NULL DEFAULT 1,
  `sort_order` bigint NULL DEFAULT 0,
  `region` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'global',
  `language` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'en',
  `origin_feed` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `source_published_at` datetime(3) NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_briefing_unique`(`source_hash` ASC) USING BTREE,
  INDEX `idx_briefing_date_status_sort`(`briefing_date` ASC, `status` ASC, `sort_order` ASC) USING BTREE,
  INDEX `idx_daily_briefings_source_type`(`source_type` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 81 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of daily_briefings
-- ----------------------------
INSERT INTO `daily_briefings` VALUES (71, '2026-04-12', 'MiniMax Just Open Sourced MiniMax M2.7: A Self-Evolving Agent Model that Scores 56.22% on SWE-Pro and 57.0% on Terminal Bench 2', 'MiniMax has officially open-sourced MiniMax M2.7, making the model weights publicly available on Hugging Face. Originally announced on March 18, 2026, MiniMax M2.7 is the MiniMax’s...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/12/minimax-just-open-sourced-minimax-m2-7-a-self-evolving-agent-model-that-scores-56-22-on-swe-pro-and-57-0-on-terminal-bench-2/', '5fdd9c7c57d2cb59600ee4c6890c2ca7753d89bf', 'auto', 1, 1, 'global', 'en', 'MarkTechPost', '2026-04-12 17:20:15.000', '2026-04-12 17:49:12.717', '2026-04-12 17:49:12.717');
INSERT INTO `daily_briefings` VALUES (72, '2026-04-12', 'AI Will Be Met with Violence, and Nothing Good Will Come of It', 'Article URL: https://www.thealgorithmicbridge.com/p/ai-will-be-met-with-violence-and Comments URL: https://news.ycombinator.com/item?id=47737563 Points: 24 # Comments: 7', 'Hacker News - Front Page: \"AI\"', 'https://www.thealgorithmicbridge.com/p/ai-will-be-met-with-violence-and', '30356f2f7556f81473d72310d31a8e884e62b996', 'auto', 1, 2, 'global', 'en', 'Hacker News - Front Page: \"AI\"', '2026-04-12 17:16:35.000', '2026-04-12 17:49:12.719', '2026-04-12 17:49:12.719');
INSERT INTO `daily_briefings` VALUES (73, '2026-04-12', 'Liquid AI Releases LFM2.5-VL-450M: a 450M-Parameter Vision-Language Model with Bounding Box Prediction, Multilingual Support, and Sub-250ms Edge Inference', 'Liquid AI just released LFM2.5-VL-450M, an updated version of its earlier LFM2-VL-450M vision-language model. The new release introduces bounding box prediction, improved instructi...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/11/liquid-ai-releases-lfm2-5-vl-450m-a-450m-parameter-vision-language-model-with-bounding-box-prediction-multilingual-support-and-sub-250ms-edge-inference/', 'fb12dd752630a5a706d4e73278cdec2d4c97ffc2', 'auto', 1, 3, 'global', 'en', 'MarkTechPost', '2026-04-12 10:41:10.000', '2026-04-12 17:49:12.720', '2026-04-12 17:49:12.720');
INSERT INTO `daily_briefings` VALUES (74, '2026-04-12', 'Researchers from MIT, NVIDIA, and Zhejiang University Propose TriAttention: A KV Cache Compression Method That Matches Full Attention at 2.5× Higher Throughput', 'Long-chain reasoning is one of the most compute-intensive tasks in modern large language models. When a model like DeepSeek-R1 or Qwen3 works through a complex math problem, it can...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/11/researchers-from-mit-nvidia-and-zhejiang-university-propose-triattention-a-kv-cache-compression-method-that-matches-full-attention-at-2-5x-higher-throughput/', '320a2176264a382f86ca5ad9bde861205d506522', 'auto', 1, 4, 'global', 'en', 'MarkTechPost', '2026-04-12 04:10:41.000', '2026-04-12 17:49:12.721', '2026-04-12 17:49:12.721');
INSERT INTO `daily_briefings` VALUES (75, '2026-04-12', 'How We Broke Top AI Agent Benchmarks: And What Comes Next', 'Article URL: https://rdi.berkeley.edu/blog/trustworthy-benchmarks-cont/ Comments URL: https://news.ycombinator.com/item?id=47733217 Points: 375 # Comments: 94', 'Hacker News - Front Page: \"AI\"', 'https://rdi.berkeley.edu/blog/trustworthy-benchmarks-cont/', 'c990ea66879d008555dcc5d25bf7d84576ffe404', 'auto', 1, 5, 'global', 'en', 'Hacker News - Front Page: \"AI\"', '2026-04-12 03:15:56.000', '2026-04-12 17:49:12.723', '2026-04-12 17:49:12.723');
INSERT INTO `daily_briefings` VALUES (76, '2026-04-12', 'How to Build a Secure Local-First Agent Runtime with OpenClaw Gateway, Skills, and Controlled Tool Execution', 'In this tutorial, we build and operate a fully local, schema-valid OpenClaw runtime. We configure the OpenClaw gateway with strict loopback binding, set up authenticated model acce...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/11/how-to-build-a-secure-local-first-agent-runtime-with-openclaw-gateway-skills-and-controlled-tool-execution/', 'c2f8b076d5ff06ccc237847a6d8c2356f453f3f1', 'auto', 1, 6, 'global', 'en', 'MarkTechPost', '2026-04-12 02:10:59.000', '2026-04-12 17:49:12.724', '2026-04-12 17:49:12.724');
INSERT INTO `daily_briefings` VALUES (77, '2026-04-12', 'How Knowledge Distillation Compresses Ensemble Intelligence into a Single Deployable AI Model', 'Complex prediction problems often lead to ensembles because combining multiple models improves accuracy by reducing variance and capturing diverse patterns. However, these ensemble...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/11/how-knowledge-distillation-compresses-ensemble-intelligence-into-a-single-deployable-ai-model/', '19690b7b38beff1b4e6e2ad2f52641b368f4419c', 'auto', 1, 7, 'global', 'en', 'MarkTechPost', '2026-04-11 15:33:41.000', '2026-04-12 17:49:12.725', '2026-04-12 17:49:12.725');
INSERT INTO `daily_briefings` VALUES (78, '2026-04-12', 'Alibaba’s Tongyi Lab Releases VimRAG: a Multimodal RAG Framework that Uses a Memory Graph to Navigate Massive Visual Contexts', 'Retrieval-Augmented Generation (RAG) has become a standard technique for grounding large language models in external knowledge — but the moment you move beyond plain text and start...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/10/alibabas-tongyi-lab-releases-vimrag-a-multimodal-rag-framework-that-uses-a-memory-graph-to-navigate-massive-visual-contexts/', '048091bc4156e62f9801824eec45b257b1509c07', 'auto', 1, 8, 'global', 'en', 'MarkTechPost', '2026-04-11 07:06:41.000', '2026-04-12 17:49:12.727', '2026-04-12 17:49:12.727');
INSERT INTO `daily_briefings` VALUES (79, '2026-04-12', 'A Coding Guide to Markerless 3D Human Kinematics with Pose2Sim, RTMPose, and OpenSim', 'In this tutorial, we build and run a complete Pose2Sim pipeline on Colab to understand how markerless 3D kinematics works in practice. We begin with environment setup, configure th...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/10/a-coding-guide-to-markerless-3d-human-kinematics-with-pose2sim-rtmpose-and-opensim/', 'af3711049cd52b800dd4e00c63b24e4199d9be8e', 'auto', 1, 9, 'global', 'en', 'MarkTechPost', '2026-04-11 04:14:41.000', '2026-04-12 17:49:12.728', '2026-04-12 17:49:12.728');
INSERT INTO `daily_briefings` VALUES (80, '2026-04-12', 'NVIDIA Releases AITune: An Open-Source Inference Toolkit That Automatically Finds the Fastest Inference Backend for Any PyTorch Model', 'Deploying a deep learning model into production has always involved a painful gap between the model a researcher trains and the model that actually runs efficiently at scale. Tenso...', 'MarkTechPost', 'https://www.marktechpost.com/2026/04/10/nvidia-releases-aitune-an-open-source-inference-toolkit-that-automatically-finds-the-fastest-inference-backend-for-any-pytorch-model/', '8f121861436ec7d1bdf5ff3863ec7c566a4cf236', 'auto', 1, 10, 'global', 'en', 'MarkTechPost', '2026-04-11 01:43:00.000', '2026-04-12 17:49:12.730', '2026-04-12 17:49:12.730');

-- ----------------------------
-- Table structure for tags
-- ----------------------------
DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_tags_name`(`name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tags
-- ----------------------------
INSERT INTO `tags` VALUES (1, 'AI', '2026-04-06 19:28:02.958', '2026-04-06 19:28:02.958');
INSERT INTO `tags` VALUES (2, 'Gin', '2026-04-06 19:28:02.960', '2026-04-06 19:28:02.960');
INSERT INTO `tags` VALUES (3, '??', '2026-04-06 19:42:56.816', '2026-04-06 19:42:56.816');
INSERT INTO `tags` VALUES (4, 'temp', '2026-04-06 19:43:12.325', '2026-04-06 19:43:12.325');
INSERT INTO `tags` VALUES (5, 'search', '2026-04-06 19:43:12.327', '2026-04-06 19:43:12.327');

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '',
  `role` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'user',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_users_username`(`username` ASC) USING BTREE,
  INDEX `idx_users_role`(`role` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO `users` VALUES (1, 'admin', '$2a$10$uXCpF7EkA0vgwLELa5BaDO/IZAAjqIZvcxUHeYx4qur.OhsH3XafK', 'https://placehold.co/120x120?text=Admin', 'admin', '2026-04-06 19:28:02.932', '2026-04-06 22:53:47.438');

SET FOREIGN_KEY_CHECKS = 1;
