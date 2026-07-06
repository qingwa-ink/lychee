package service

import (
	"context"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// PhraseService 常用语业务。
type PhraseService struct {
	repo *repository.PhraseRepository
}

func NewPhraseService(repo *repository.PhraseRepository) *PhraseService {
	return &PhraseService{repo: repo}
}

// Create 新增常用语。
func (s *PhraseService) Create(ctx context.Context, userID uint, content string) (*model.Phrase, error) {
	p := &model.Phrase{UserID: userID, Content: content}
	if err := s.repo.Create(p); err != nil {
		return nil, bizerr.ErrInternal
	}
	return p, nil
}

// List 分页查询当前用户常用语。
func (s *PhraseService) List(ctx context.Context, userID uint, page, pageSize int) ([]model.Phrase, int64, error) {
	return s.repo.ListByUser(userID, page, pageSize)
}

// Update 修改常用语（须归属当前用户）。
func (s *PhraseService) Update(ctx context.Context, userID, id uint, content string) (*model.Phrase, error) {
	p, err := s.repo.FindByIDAndUser(id, userID)
	if err != nil {
		return nil, bizerr.ErrNotFound
	}
	p.Content = content
	// 仅更新 content 字段
	if err := s.repo.UpdateContent(p.ID, content); err != nil {
		return nil, bizerr.ErrInternal
	}
	return p, nil
}

// Delete 删除常用语（须归属当前用户）。
func (s *PhraseService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.repo.FindByIDAndUser(id, userID); err != nil {
		return bizerr.ErrNotFound
	}
	if err := s.repo.Delete(id); err != nil {
		return bizerr.ErrInternal
	}
	return nil
}
