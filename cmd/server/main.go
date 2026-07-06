// Command lychee 启动荔枝小秘书后端服务。
package main

import (
	"fmt"
	"log"

	"github.com/qingwa-ink/lychee/internal/config"
	"github.com/qingwa-ink/lychee/internal/repository"
	"github.com/qingwa-ink/lychee/internal/router"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	_ = db

	r := router.New(cfg)
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("lychee server listening on %s (env=%s)", addr, cfg.App.Env)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
