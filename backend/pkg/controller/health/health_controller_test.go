package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	res dto.HealthResponse
	err error
}

func (m *mockService) Check(_ context.Context) (dto.HealthResponse, error) {
	return m.res, m.err
}

func TestGetHealthOK(t *testing.T) {
	app := fiber.New()
	controller := NewController(&mockService{res: dto.HealthResponse{Status: string(enum.HealthOK), Database: string(enum.DatabaseUp), Timestamp: "2026-04-25T00:00:00Z"}})
	app.Get("/api/health", controller.GetHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body dto.HealthResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, string(enum.HealthOK), body.Status)
	assert.Equal(t, string(enum.DatabaseUp), body.Database)
}

func TestGetHealthDegraded(t *testing.T) {
	app := fiber.New()
	controller := NewController(&mockService{res: dto.HealthResponse{Status: string(enum.HealthDegraded), Database: string(enum.DatabaseDown), Timestamp: "2026-04-25T00:00:00Z"}, err: errors.New("db unavailable")})
	app.Get("/api/health", controller.GetHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var body dto.HealthResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, string(enum.HealthDegraded), body.Status)
	assert.Equal(t, string(enum.DatabaseDown), body.Database)
}
