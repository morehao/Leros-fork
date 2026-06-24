package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/types"
)

type skillService struct {
	db *gorm.DB
}

// NewSkillService creates a new SkillService.
func NewSkillService(db *gorm.DB) contract.SkillService {
	return &skillService{db: db}
}

func (s *skillService) ToggleSkillStatus(ctx context.Context, code string, req *contract.ToggleSkillStatusRequest) (*contract.ToggleSkillStatusResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Status != string(types.SkillStatusActive) && req.Status != string(types.SkillStatusInactive) {
		return nil, fmt.Errorf("invalid status: %s (must be 'active' or 'inactive')", req.Status)
	}

	var skill types.Skill
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&skill).Error; err != nil {
		return nil, fmt.Errorf("skill not found: %s", code)
	}

	if skill.Status == req.Status {
		return &contract.ToggleSkillStatusResponse{Code: code, Status: req.Status}, nil
	}

	if err := s.db.WithContext(ctx).Model(&skill).Update("status", req.Status).Error; err != nil {
		return nil, fmt.Errorf("failed to update skill status: %w", err)
	}

	return &contract.ToggleSkillStatusResponse{Code: code, Status: req.Status}, nil
}
