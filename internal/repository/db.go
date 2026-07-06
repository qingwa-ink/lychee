// Package repository 封装数据库连接与迁移。
package repository

import (
	"fmt"
	"os"
	"path/filepath"

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

	db, err := gorm.Open(sqlite.Open(cfg.DB.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
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
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}
