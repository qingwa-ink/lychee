package service

import (
	"context"
	"time"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

const (
	LogCategoryLogin     = "login"
	LogCategoryOperation = "operation"
)

// LogService 操作日志业务。
type LogService struct {
	repo *repository.OperationLogRepository
}

func NewLogService(repo *repository.OperationLogRepository) *LogService {
	return &LogService{repo: repo}
}

// List 操作历史（分页、时间范围、类别筛选）。
func (s *LogService) List(ctx context.Context, userID uint, f repository.LogFilter) ([]model.OperationLog, int64, error) {
	return s.repo.List(userID, f)
}

// ListLogins 登录日志（category=login）。
func (s *LogService) ListLogins(ctx context.Context, userID uint, f repository.LogFilter) ([]model.OperationLog, int64, error) {
	f.Category = LogCategoryLogin
	return s.repo.List(userID, f)
}

// Report 操作次数报告（按 day/hour 聚合）。start/end 为空时默认最近 7 天。
func (s *LogService) Report(ctx context.Context, userID uint, dimension, start, end string) ([]repository.BucketCount, error) {
	if dimension != "day" && dimension != "hour" {
		dimension = "day"
	}
	now := time.Now()
	st := now.AddDate(0, 0, -6) // 最近 7 天的起点
	et := now.Add(time.Second)
	if start != "" {
		if t, ok := parseTime(start); ok {
			st = t
		} else {
			return nil, bizerr.New(bizerr.CodeBadRequest, "start 时间格式应为 RFC3339 或 YYYY-MM-DD")
		}
	}
	if end != "" {
		if t, ok := parseTime(end); ok {
			et = t
		} else {
			return nil, bizerr.New(bizerr.CodeBadRequest, "end 时间格式应为 RFC3339 或 YYYY-MM-DD")
		}
	}
	return s.repo.Report(userID, dimension, st, et)
}

// parseTime 兼容 RFC3339 与 YYYY-MM-DD（日期按当日 00:00 解析）。
func parseTime(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, true
	}
	return time.Time{}, false
}
