package health

import (
	"context"
	"errors"
	"testing"

	"flowforge-automation-backend/pkg/model/domain/enum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	err error
}

func (m *mockRepository) Ping(_ context.Context) error {
	return m.err
}

func TestCheckSuccess(t *testing.T) {
	svc := NewHealthService(&mockRepository{})

	res, err := svc.Check(context.Background())
	require.NoError(t, err)
	assert.Equal(t, string(enum.HealthOK), res.Status)
	assert.Equal(t, string(enum.DatabaseUp), res.Database)
	assert.NotEmpty(t, res.Timestamp)
}

func TestCheckDBDown(t *testing.T) {
	svc := NewHealthService(&mockRepository{err: errors.New("db down")})

	res, err := svc.Check(context.Background())
	require.Error(t, err)
	assert.Equal(t, string(enum.HealthDegraded), res.Status)
	assert.Equal(t, string(enum.DatabaseDown), res.Database)
	assert.NotEmpty(t, res.Timestamp)
}
