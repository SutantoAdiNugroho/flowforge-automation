package health

import (
	"context"
	"flowforge-automation-backend/pkg/model/dto"
)

type Service interface {
	Check(ctx context.Context) (dto.HealthResponse, error)
}
