package health

import (
	"context"
	"time"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	healthrepository "flowforge-automation-backend/pkg/repository/health"
)

type serviceImpl struct {
	repo healthrepository.Repository
}

func NewHealthService(repo healthrepository.Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) Check(ctx context.Context) (dto.HealthResponse, error) {
	response := dto.HealthResponse{
		Status:    string(enum.HealthOK),
		Database:  string(enum.DatabaseUp),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.Ping(ctx); err != nil {
		response.Status = string(enum.HealthDegraded)
		response.Database = string(enum.DatabaseDown)
		return response, err
	}

	return response, nil
}
