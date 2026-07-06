// Package model 定义全部 GORM 模型，对应数据库表结构（见 doc/项目设计文档.md §4）。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Base 包含各业务表通用的主键与时间字段，启用软删除。
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 用户表。
type User struct {
	Base
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"` // bcrypt，禁止外泄
	Locale       string `gorm:"default:'zh'" json:"locale"`
	Status       int    `gorm:"default:1" json:"status"` // 1=正常
}

// EmailVerificationCode 邮箱验证码。
type EmailVerificationCode struct {
	Base
	Email     string    `gorm:"index;not null" json:"email"`
	Code      string    `gorm:"not null" json:"-"`
	Type      string    `gorm:"not null" json:"type"` // register / forgot / change_password
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
	Used      bool      `gorm:"default:false" json:"used"`
}

// Phrase 常用语。
type Phrase struct {
	Base
	UserID  uint   `gorm:"index;not null" json:"user_id"`
	Content string `gorm:"not null" json:"content"`
}

// TaskGroup 任务分组，ParentID 为空表示顶层分组，支持任意层级嵌套。
type TaskGroup struct {
	Base
	UserID    uint   `gorm:"index;not null" json:"user_id"`
	ParentID  *uint  `gorm:"index" json:"parent_id"`
	Name      string `gorm:"not null" json:"name"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
}

// Task 任务。
type Task struct {
	Base
	UserID   uint       `gorm:"index;not null" json:"user_id"`
	GroupID  uint       `gorm:"index;not null" json:"group_id"`
	Content  string     `gorm:"not null" json:"content"`
	Priority int        `gorm:"default:3" json:"priority"`            // 0-5，P0 最高
	Status   string     `gorm:"default:'editing'" json:"status"`      // editing / pending / completed
	DueDate  *time.Time `gorm:"index" json:"due_date"`
}

// CheckInRecord 打卡记录。
type CheckInRecord struct {
	Base
	UserID     uint    `gorm:"index;not null" json:"user_id"`
	Type       string  `gorm:"index;not null" json:"type"` // water / exercise / nap ...
	Value      float64 `gorm:"not null" json:"value"`
	Unit       string  `json:"unit"`                       // ml / min
	RecordDate string  `gorm:"index;not null" json:"record_date"` // YYYY-MM-DD
}

// CheckInGoal 打卡每日目标，(UserID, Type) 唯一。
type CheckInGoal struct {
	Base
	UserID      uint    `gorm:"uniqueIndex:idx_goal_user_type;not null" json:"user_id"`
	Type        string  `gorm:"uniqueIndex:idx_goal_user_type;not null" json:"type"`
	DailyTarget float64 `gorm:"not null" json:"daily_target"`
	Unit        string  `json:"unit"`
}

// OperationLog 操作日志（含登录），不启用软删除。
type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UserID    *uint     `gorm:"index" json:"user_id"`
	Category  string    `gorm:"index;not null" json:"category"` // login / operation
	Action    string    `json:"action"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `gorm:"index" json:"ip"`
	UA        string    `gorm:"column:ua" json:"ua"`
	Params    string    `json:"params"`
}
