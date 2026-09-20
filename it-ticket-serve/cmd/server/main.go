package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/config"
	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/router"
	"it-ticket-api/internal/store"
)

func main() {
	logger.Init() // 先定日志格式，后面 AccessLog 才能按 key=value 打到这个终端
	gin.SetMode(gin.ReleaseMode)

	cfg := config.Load()

	db, err := store.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("mysql 连不上: %v（先执行 brew services start mysql）", err)
	}
	defer db.Close()

	r := router.New(db, cfg.JWTSecret)
	// r.Run 会卡住听端口，成功时自己不打字。先打印地址，避免终端看起来像没启动。
	// HTTPAddr 是 ":8080" 这种形式，拼上本机回环地址方便复制到浏览器或 curl。
	log.Printf("服务已启动  http://127.0.0.1%s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}
