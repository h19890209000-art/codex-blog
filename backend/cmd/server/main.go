package main

import (
	"log"

	"ai-blog/backend/internal/bootstrap"
)

// main 是整个后端程序的入口。
// 你可以把它理解成 PHP 项目里的 public/index.php。
func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
