package service

import (
	"context"
	"time"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// 打卡类型 → 默认单位。类型可扩展：新增一行即可。
var checkInTypes = map[string]string{
	"water":   "ml",  // 喝水
	"exercise": "min", // 起身运动
	"nap":     "min", // 午睡
}

const dateLayout = "2006-01-02"

// CreateRecordInput 新增打卡记录入参。
type CreateRecordInput struct {
	Type       string
	Value      float64
	Unit       string
	RecordDate string // 可空，默认今天
}

// UpsertGoalInput 设置每日目标入参。
type UpsertGoalInput struct {
	Type        string
	DailyTarget float64
	Unit        string
}

// TypeReport 每日报告中单个类型的统计。
type TypeReport struct {
	Type            string  `json:"type"`
	Total           float64 `json:"total"`
	Unit            string  `json:"unit"`
	DailyTarget     float64 `json:"daily_target"`
	AchievementRate float64 `json:"achievement_rate"` // 0~1+，无目标则为 0
	HasGoal         bool    `json:"has_goal"`
}

// DailyReport 每日报告。
type DailyReport struct {
	Date  string       `json:"date"`
	Types []TypeReport `json:"types"`
}

// CheckInService 打卡健康业务。
type CheckInService struct {
	recordRepo *repository.CheckInRecordRepository
	goalRepo   *repository.CheckInGoalRepository
}

func NewCheckInService(recordRepo *repository.CheckInRecordRepository, goalRepo *repository.CheckInGoalRepository) *CheckInService {
	return &CheckInService{recordRepo: recordRepo, goalRepo: goalRepo}
}

// ListRecords 查询打卡记录（按类型/日期筛选、分页）。
func (s *CheckInService) ListRecords(ctx context.Context, userID uint, f repository.RecordFilter) ([]model.CheckInRecord, int64, error) {
	if f.Type != "" && !isValidCheckInType(f.Type) {
		return nil, 0, bizerr.New(bizerr.CodeBadRequest, "打卡类型非法")
	}
	return s.recordRepo.List(userID, f)
}

// CreateRecord 新增打卡记录，校验类型/数值/日期。
func (s *CheckInService) CreateRecord(ctx context.Context, userID uint, in CreateRecordInput) (*model.CheckInRecord, error) {
	if !isValidCheckInType(in.Type) {
		return nil, bizerr.New(bizerr.CodeBadRequest, "打卡类型非法")
	}
	if in.Value <= 0 {
		return nil, bizerr.New(bizerr.CodeBadRequest, "数值需大于 0")
	}
	date := in.RecordDate
	if date == "" {
		date = time.Now().Format(dateLayout)
	} else if !isValidDate(date) {
		return nil, bizerr.New(bizerr.CodeBadRequest, "日期格式应为 YYYY-MM-DD")
	}
	unit := in.Unit
	if unit == "" {
		unit = checkInTypes[in.Type]
	}
	rec := &model.CheckInRecord{
		UserID:     userID,
		Type:       in.Type,
		Value:      in.Value,
		Unit:       unit,
		RecordDate: date,
	}
	if err := s.recordRepo.Create(rec); err != nil {
		return nil, bizerr.ErrInternal
	}
	return rec, nil
}

// DailyReport 生成某日期（默认今天）的每日报告：各类型累计值与目标完成度。
func (s *CheckInService) DailyReport(ctx context.Context, userID uint, date string) (*DailyReport, error) {
	if date == "" {
		date = time.Now().Format(dateLayout)
	} else if !isValidDate(date) {
		return nil, bizerr.New(bizerr.CodeBadRequest, "日期格式应为 YYYY-MM-DD")
	}
	sums, err := s.recordRepo.SumsByDate(userID, date)
	if err != nil {
		return nil, bizerr.ErrInternal
	}
	goals, err := s.goalRepo.ListByUser(userID)
	if err != nil {
		return nil, bizerr.ErrInternal
	}

	// goalByType: 类型 → 目标
	goalByType := make(map[string]model.CheckInGoal, len(goals))
	for _, g := range goals {
		goalByType[g.Type] = g
	}
	// totalByType: 类型 → 当日累计
	totalByType := make(map[string]float64, len(sums))
	for _, s2 := range sums {
		totalByType[s2.Type] = s2.Total
	}

	// 合并类型集合：有目标或有记录的类型都展示
	seen := make(map[string]bool)
	var types []TypeReport
	addType := func(typ string) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		total := totalByType[typ]
		tr := TypeReport{Type: typ, Total: total}
		if g, ok := goalByType[typ]; ok {
			tr.DailyTarget = g.DailyTarget
			tr.Unit = g.Unit
			tr.HasGoal = true
			if g.DailyTarget > 0 {
				tr.AchievementRate = round2(total / g.DailyTarget)
			}
		} else {
			tr.Unit = checkInTypes[typ]
		}
		types = append(types, tr)
	}
	for _, g := range goals {
		addType(g.Type)
	}
	for _, s2 := range sums {
		addType(s2.Type)
	}

	return &DailyReport{Date: date, Types: types}, nil
}

// ListGoals 查询某用户的全部每日目标。
func (s *CheckInService) ListGoals(ctx context.Context, userID uint) ([]model.CheckInGoal, error) {
	return s.goalRepo.ListByUser(userID)
}

// UpsertGoal 新增或更新某类型的每日目标。
func (s *CheckInService) UpsertGoal(ctx context.Context, userID uint, in UpsertGoalInput) (*model.CheckInGoal, error) {
	if !isValidCheckInType(in.Type) {
		return nil, bizerr.New(bizerr.CodeBadRequest, "打卡类型非法")
	}
	if in.DailyTarget <= 0 {
		return nil, bizerr.New(bizerr.CodeBadRequest, "目标值需大于 0")
	}
	unit := in.Unit
	if unit == "" {
		unit = checkInTypes[in.Type]
	}
	g := &model.CheckInGoal{
		UserID:      userID,
		Type:        in.Type,
		DailyTarget: in.DailyTarget,
		Unit:        unit,
	}
	if err := s.goalRepo.Upsert(g); err != nil {
		return nil, bizerr.ErrInternal
	}
	stored, err := s.goalRepo.FindByUserType(userID, in.Type)
	if err != nil {
		return nil, bizerr.ErrInternal
	}
	return stored, nil
}

func isValidCheckInType(t string) bool {
	_, ok := checkInTypes[t]
	return ok
}

func isValidDate(s string) bool {
	_, err := time.Parse(dateLayout, s)
	return err == nil
}

// round2 保留两位小数（完成度展示用）。
func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
