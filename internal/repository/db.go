// Package repository 封装数据库连接与迁移。
package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/qingwa-ink/lychee/internal/config"
	"github.com/qingwa-ink/lychee/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB 初始化 SQLite 连接并执行 AutoMigrate。
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	// 确保数据文件所在目录存在
	if dir := filepath.Dir(cfg.DB.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	// 业务中把 record not found 当作正常情况（查重、校验验证码等），因此静音该日志。
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	db, err := gorm.Open(sqlite.Open(cfg.DB.Path), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.EmailVerificationCode{},
		&model.Phrase{},
		&model.TaskGroup{},
		&model.Task{},
		&model.CheckInRecord{},
		&model.CheckInGoal{},
		&model.OperationLog{},
		&model.RefreshToken{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}
