package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const baseURL = "http://localhost:5000/api"

type APIRequest struct {
	Method      string
	Path        string
	Body        interface{}
	Token       string
	ContentType string
}

func (r APIRequest) Send() (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if r.Body != nil {
		bodyBytes, err = json.Marshal(r.Body)
		if err != nil {
			return nil, err
		}
	}

	contentType := r.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	req, err := http.NewRequest(r.Method, baseURL+r.Path, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}

	return http.DefaultClient.Do(req)
}

type APITestSuite struct {
	suite.Suite
	tenantSlug      string
	adminEmail      string
	adminPassword   string
	adminToken      string
	adminUserID     uuid.UUID
	tenantID        uuid.UUID
	superAdminToken string
	workflowIDs     []uuid.UUID
	runIDs          []uuid.UUID
	createdUserIDs  []uuid.UUID
	createdTenantID uuid.UUID
}

func validDAGDefinition() domain.JSONB {
	return domain.JSONB{
		"steps": []map[string]interface{}{
			{
				"id":   "step_1",
				"name": "Start",
				"type": "delay",
				"config": map[string]interface{}{
					"duration_ms": 100,
				},
			},
		},
	}
}

func (s *APITestSuite) decodeJSON(resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	err = json.Unmarshal(bodyBytes, target)
	s.Require().NoError(err)
}

func (s *APITestSuite) registerAdminTenant() {
	req := dto.RegisterRequest{
		TenantName: "Test Tenant",
		TenantSlug: s.tenantSlug,
		Email:      s.adminEmail,
		Password:   s.adminPassword,
	}

	resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var res dto.RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	s.Require().NoError(err)
	s.Require().NotEmpty(res.Token)
	s.Require().NotEqual(uuid.Nil, res.UserID)
	s.Require().NotEqual(uuid.Nil, res.TenantID)

	s.adminToken = res.Token
	s.adminUserID = res.UserID
	s.tenantID = res.TenantID
}

func (s *APITestSuite) loginAdmin() {
	req := dto.LoginRequest{Email: s.adminEmail, Password: s.adminPassword}
	resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var res dto.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	s.Require().NoError(err)
	s.Require().NotEmpty(res.Token)
	s.adminToken = res.Token
}

func (s *APITestSuite) tryLoginSuperAdmin() {
	req := dto.LoginRequest{Email: "superadmin@flowforge.com", Password: "password"}
	resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
	s.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.superAdminToken = ""
		return
	}

	var res dto.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	s.Require().NoError(err)
	s.superAdminToken = res.Token
}

func (s *APITestSuite) createWorkflow(name, triggerType string, isActive bool) uuid.UUID {
	req := dto.CreateWorkflowRequest{
		Name:        name,
		Description: "integration test workflow",
		Definition:  validDAGDefinition(),
		TriggerType: triggerType,
		IsActive:    boolPtr(isActive),
	}
	if triggerType == "cron" {
		req.CronExpression = "0 9 * * *"
	}

	resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.adminToken}.Send()
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var wf domain.Workflow
	err = json.NewDecoder(resp.Body).Decode(&wf)
	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, wf.ID)
	s.workflowIDs = append(s.workflowIDs, wf.ID)
	return wf.ID
}

func (s *APITestSuite) SetupSuite() {
	s.workflowIDs = make([]uuid.UUID, 0)
	s.runIDs = make([]uuid.UUID, 0)
	s.createdUserIDs = make([]uuid.UUID, 0)
	s.tenantSlug = "test-tenant-" + uuid.New().String()[:8]
	s.adminEmail = "admin-" + uuid.New().String()[:8] + "@example.com"
	s.adminPassword = "Password123!"

	s.registerAdminTenant()
	s.loginAdmin()
	s.tryLoginSuperAdmin()
}

func (s *APITestSuite) TestHealthAPI() {
	resp, err := APIRequest{Method: http.MethodGet, Path: "/health"}.Send()
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	s.Require().NoError(err)
	s.Equal("ok", body["status"])
}

func (s *APITestSuite) TestEventsAPI() {
	s.Run("Events unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/events"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Events authorized", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		transport := &http.Transport{DisableKeepAlives: true}
		defer transport.CloseIdleConnections()
		client := &http.Client{Transport: transport, Timeout: 35 * time.Second}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/events", nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+s.adminToken)
		req.Close = true

		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, resp.StatusCode)

		cancel()
		_ = resp.Body.Close()
	})
}

func (s *APITestSuite) TestAuthAPIs() {
	s.Run("Login positive", func() {
		req := dto.LoginRequest{Email: s.adminEmail, Password: s.adminPassword}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Login invalid credentials", func() {
		req := dto.LoginRequest{Email: s.adminEmail, Password: "WrongPassword123"}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Login invalid request", func() {
		req := dto.LoginRequest{Email: "", Password: ""}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Register duplicate slug", func() {
		req := dto.RegisterRequest{
			TenantName: "Duplicate Slug",
			TenantSlug: s.tenantSlug,
			Email:      "other-" + uuid.New().String()[:8] + "@example.com",
			Password:   s.adminPassword,
		}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("Register duplicate email", func() {
		req := dto.RegisterRequest{
			TenantName: "Duplicate Email",
			TenantSlug: "tenant-" + uuid.New().String()[:8],
			Email:      s.adminEmail,
			Password:   s.adminPassword,
		}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusConflict, resp.StatusCode)
	})
}

func (s *APITestSuite) TestWorkflowAndVersionAPIs() {
	workflowID := s.createWorkflow("WF Main", "manual", true)

	s.Run("List workflows positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Get workflow positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + workflowID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Get workflow invalid id", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/invalid-id", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Update workflow positive", func() {
		req := dto.UpdateWorkflowRequest{
			Name:        "WF Main Updated",
			Description: "updated",
			Definition:  validDAGDefinition(),
			TriggerType: "manual",
			IsActive:    boolPtr(false),
		}
		resp, err := APIRequest{Method: http.MethodPut, Path: "/workflows/" + workflowID.String(), Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("List versions positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + workflowID.String() + "/versions?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)

		var list struct {
			Content []map[string]interface{} `json:"content"`
		}
		err = json.NewDecoder(resp.Body).Decode(&list)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(list.Content), 1)
	})

	s.Run("Create version import from previous positive", func() {
		req := dto.CreateWorkflowVersionRequest{ImportFromVersion: intPtr(1)}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows/" + workflowID.String() + "/versions", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("Create version invalid workflow", func() {
		req := dto.CreateWorkflowVersionRequest{ImportFromVersion: intPtr(99)}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows/" + uuid.New().String() + "/versions", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("Activate version positive", func() {
		resp, err := APIRequest{Method: http.MethodPut, Path: "/workflows/" + workflowID.String() + "/versions/1/activate", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Activate version not found", func() {
		resp, err := APIRequest{Method: http.MethodPut, Path: "/workflows/" + workflowID.String() + "/versions/999/activate", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("Rollback version positive", func() {
		resp, err := APIRequest{Method: http.MethodPut, Path: "/workflows/" + workflowID.String() + "/rollback/1", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Workflow unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Delete workflow invalid id", func() {
		resp, err := APIRequest{Method: http.MethodDelete, Path: "/workflows/invalid-id", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Delete workflow not found", func() {
		resp, err := APIRequest{Method: http.MethodDelete, Path: "/workflows/" + uuid.New().String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("Delete workflow unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodDelete, Path: "/workflows/" + workflowID.String()}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Delete workflow positive", func() {
		deleteWorkflowID := s.createWorkflow("WF Delete", "manual", true)
		resp, err := APIRequest{Method: http.MethodDelete, Path: "/workflows/" + deleteWorkflowID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNoContent, resp.StatusCode)

		getResp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + deleteWorkflowID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer getResp.Body.Close()
		s.Equal(http.StatusNotFound, getResp.StatusCode)
	})
}

func (s *APITestSuite) TestRunAndWebhookAPIs() {
	manualWorkflowID := s.createWorkflow("WF Run Manual", "manual", true)
	webhookWorkflowID := s.createWorkflow("WF Run Webhook", "webhook", true)

	s.Run("Trigger run positive", func() {
		req := dto.TriggerRunRequest{Inputs: map[string]interface{}{"source": "test"}}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows/" + manualWorkflowID.String() + "/runs", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusCreated, resp.StatusCode)

		var run domain.WorkflowRun
		err = json.NewDecoder(resp.Body).Decode(&run)
		s.Require().NoError(err)
		s.NotEqual(uuid.Nil, run.ID)
		s.runIDs = append(s.runIDs, run.ID)
	})

	s.Run("Trigger run invalid workflow", func() {
		req := dto.TriggerRunRequest{}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows/" + uuid.New().String() + "/runs", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("List runs by workflow positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + manualWorkflowID.String() + "/runs?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("List runs by tenant positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("List runs unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs?page=1&page_size=20"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Get run positive", func() {
		s.Require().NotEmpty(s.runIDs)
		runID := s.runIDs[0]
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs/" + runID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Get run invalid id", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs/invalid-id", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Get stats positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs/stats?hours=24", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Get stats unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/runs/stats?hours=24"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Cancel run positive or conflict", func() {
		s.Require().NotEmpty(s.runIDs)
		runID := s.runIDs[0]
		resp, err := APIRequest{Method: http.MethodPost, Path: "/runs/" + runID.String() + "/cancel", Body: map[string]interface{}{}, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.True(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict)
	})

	s.Run("Webhook trigger positive", func() {
		resp, err := APIRequest{Method: http.MethodPost, Path: "/webhooks/" + webhookWorkflowID.String(), Body: map[string]interface{}{"hello": "world"}}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusAccepted, resp.StatusCode)
	})

	s.Run("Webhook invalid workflow id", func() {
		resp, err := APIRequest{Method: http.MethodPost, Path: "/webhooks/invalid-id", Body: map[string]interface{}{}}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Webhook wrong trigger type", func() {
		resp, err := APIRequest{Method: http.MethodPost, Path: "/webhooks/" + manualWorkflowID.String(), Body: map[string]interface{}{}}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *APITestSuite) TestUserAPIs() {
	var createdUserID uuid.UUID
	newUserEmail := "user-" + uuid.New().String()[:8] + "@example.com"

	s.Run("Create user positive", func() {
		req := dto.CreateUserRequest{Email: newUserEmail, Password: "Password123!", Role: "viewer"}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/users", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusCreated, resp.StatusCode)

		var user domain.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		s.Require().NoError(err)
		createdUserID = user.ID
		s.createdUserIDs = append(s.createdUserIDs, user.ID)
	})

	s.Run("Create user duplicate email", func() {
		req := dto.CreateUserRequest{Email: newUserEmail, Password: "Password123!", Role: "viewer"}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/users", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("Create user invalid role", func() {
		req := dto.CreateUserRequest{Email: "user-" + uuid.New().String()[:8] + "@example.com", Password: "Password123!", Role: "invalid-role"}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/users", Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("List users positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/users?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("List users unauthorized", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/users?page=1&page_size=20"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Get user by id positive", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/users/" + createdUserID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Update user positive", func() {
		req := dto.UpdateUserRequest{Role: "editor", IsActive: boolPtr(false)}
		resp, err := APIRequest{Method: http.MethodPut, Path: "/users/" + createdUserID.String(), Body: req, Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Delete user positive", func() {
		resp, err := APIRequest{Method: http.MethodDelete, Path: "/users/" + createdUserID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNoContent, resp.StatusCode)
	})

	s.Run("Get deleted user should 404", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/users/" + createdUserID.String(), Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})
}

func (s *APITestSuite) TestTenantAdminAPIs() {
	s.Run("Admin tenants unauthorized no token", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants?page=1&page_size=20"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Admin tenants forbidden for tenant admin", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants?page=1&page_size=20", Token: s.adminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusForbidden, resp.StatusCode)
	})

	s.Run("Admin tenants invalid id with super-admin token", func() {
		if s.superAdminToken == "" {
			s.T().Skip("super admin token unavailable")
		}
		resp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants/invalid-id", Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	if s.superAdminToken == "" {
		s.T().Skip("super admin token unavailable; skipping super-admin positive tests")
	}

	s.Run("Super-admin list tenants", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants?page=1&page_size=20", Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Super-admin create/get/update/delete tenant", func() {
		slug := "tenant-" + uuid.New().String()[:8]
		adminEmail := "tenant-admin-" + uuid.New().String()[:8] + "@example.com"
		createReq := dto.CreateTenantRequest{
			Name:          "Tenant API Test",
			Slug:          slug,
			AdminEmail:    adminEmail,
			AdminPassword: "Password123!",
		}

		createResp, err := APIRequest{Method: http.MethodPost, Path: "/admin/tenants", Body: createReq, Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer createResp.Body.Close()
		s.Equal(http.StatusCreated, createResp.StatusCode)

		var payload map[string]interface{}
		err = json.NewDecoder(createResp.Body).Decode(&payload)
		s.Require().NoError(err)

		tenantObj := payload["tenant"].(map[string]interface{})
		tenantIDStr := tenantObj["id"].(string)
		tenantID, err := uuid.Parse(tenantIDStr)
		s.Require().NoError(err)
		s.createdTenantID = tenantID

		getResp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants/" + tenantID.String(), Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer getResp.Body.Close()
		s.Equal(http.StatusOK, getResp.StatusCode)

		updateReq := dto.UpdateTenantRequest{Name: "Tenant API Test Updated"}
		updateResp, err := APIRequest{Method: http.MethodPut, Path: "/admin/tenants/" + tenantID.String(), Body: updateReq, Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer updateResp.Body.Close()
		s.Equal(http.StatusOK, updateResp.StatusCode)

		deleteResp, err := APIRequest{Method: http.MethodDelete, Path: "/admin/tenants/" + tenantID.String(), Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer deleteResp.Body.Close()
		s.Equal(http.StatusNoContent, deleteResp.StatusCode)

		getDeletedResp, err := APIRequest{Method: http.MethodGet, Path: "/admin/tenants/" + tenantID.String(), Token: s.superAdminToken}.Send()
		s.Require().NoError(err)
		defer getDeletedResp.Body.Close()
		s.Equal(http.StatusNotFound, getDeletedResp.StatusCode)
	})
}

func (s *APITestSuite) TestZZRateLimit() {
	tooMany := 0
	total := 140

	for i := 0; i < total; i++ {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/health"}.Send()
		s.Require().NoError(err)
		if resp.StatusCode == http.StatusTooManyRequests {
			tooMany++
		}
		_ = resp.Body.Close()
	}

	s.Greater(tooMany, 0, fmt.Sprintf("expected some 429 responses from rate limiter, got %d", tooMany))
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
