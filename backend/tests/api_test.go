package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

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

type AuthWorkflowTestSuite struct {
	suite.Suite
	tenantSlug  string
	email       string
	password    string
	token       string
	userID      uuid.UUID
	tenantID    uuid.UUID
	workflowIDs []uuid.UUID
}

func (s *AuthWorkflowTestSuite) SetupSuite() {
	s.workflowIDs = make([]uuid.UUID, 0)
	s.tenantSlug = "test-tenant-" + uuid.New().String()[:8]
	s.email = "test@" + s.tenantSlug + ".com"
	s.password = "Password123!"
}

func (s *AuthWorkflowTestSuite) TestAuthRegister() {
	s.Run("Positive Case: Should register new tenant and user successfully", func() {
		req := dto.RegisterRequest{
			TenantName: "Test Company",
			TenantSlug: s.tenantSlug,
			Email:      s.email,
			Password:   s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res dto.RegisterResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		s.Require().NoError(err)

		s.NotEmpty(res.Token)
		s.NotEmpty(res.UserID)
		s.NotEmpty(res.TenantID)
		s.Equal(s.email, res.Email)
		s.Equal("admin", res.Role)

		s.token = res.Token
		s.userID = res.UserID
		s.tenantID = res.TenantID
	})

	s.Run("Negative Case: Should return 400 for invalid request body", func() {
		invalidBody := map[string]string{"tenant_name": "", "tenant_slug": "", "email": "", "password": ""}
		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: invalidBody}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 409 if tenant slug already exists", func() {
		req := dto.RegisterRequest{
			TenantName: "Duplicate Tenant",
			TenantSlug: s.tenantSlug,
			Email:      "another@email.com",
			Password:   s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 409 if email already registered", func() {
		req := dto.RegisterRequest{
			TenantName: "Another Tenant",
			TenantSlug: "another-slug-" + uuid.New().String()[:8],
			Email:      s.email,
			Password:   s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for invalid email format", func() {
		req := dto.RegisterRequest{
			TenantName: "Test Tenant",
			TenantSlug: "test-slug",
			Email:      "invalid-email",
			Password:   s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for short password", func() {
		req := dto.RegisterRequest{
			TenantName: "Test Tenant",
			TenantSlug: "test-slug",
			Email:      "test@example.com",
			Password:   "123",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/register", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *AuthWorkflowTestSuite) TestAuthLogin() {
	s.Run("Positive Case: Should login successfully", func() {
		req := dto.LoginRequest{
			Email:    s.email,
			Password: s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var res dto.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		s.Require().NoError(err)

		s.NotEmpty(res.Token)
		s.NotZero(res.ExpiresAt)
		s.Equal(s.userID, res.UserID)
		s.Equal(s.email, res.Email)
		s.Equal("admin", res.Role)

		s.token = res.Token
	})

	s.Run("Negative Case: Should return 401 for invalid password", func() {
		req := dto.LoginRequest{
			Email:    s.email,
			Password: "WrongPassword123!",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 401 for non-existent email", func() {
		req := dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for invalid request body", func() {
		req := dto.LoginRequest{
			Email:    "",
			Password: s.password,
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/auth/login", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *AuthWorkflowTestSuite) TestWorkflowCreate() {
	s.Run("Positive Case: Should create workflow successfully", func() {
		definition := domain.JSONB{
			"nodes": []map[string]interface{}{
				{"id": "start", "type": "trigger"},
				{"id": "step1", "type": "action"},
			},
			"edges": []map[string]interface{}{},
		}

		req := dto.CreateWorkflowRequest{
			Name:        "Test Workflow 1",
			Description: "This is a test workflow for API testing",
			Definition:  definition,
			TriggerType: "manual",
			IsActive:    boolPtr(true),
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.NotEmpty(workflow.ID)
		s.Equal("Test Workflow 1", workflow.Name)
		s.Equal("manual", workflow.TriggerType)
		s.True(workflow.IsActive)

		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})

	s.Run("Positive Case: Should create workflow with cron trigger", func() {
		definition := domain.JSONB{"task": "daily job"}

		req := dto.CreateWorkflowRequest{
			Name:           "Cron Workflow",
			Description:    "Cron triggered workflow",
			Definition:     definition,
			TriggerType:    "cron",
			CronExpression: "0 9 * * *",
			IsActive:       boolPtr(false),
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.Equal("cron", workflow.TriggerType)
		s.Equal("0 9 * * *", workflow.CronExpression)
		s.False(workflow.IsActive)

		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})

	s.Run("Positive Case: Should create workflow with webhook trigger", func() {
		definition := domain.JSONB{"endpoint": "/webhook/test"}

		req := dto.CreateWorkflowRequest{
			Name:        "Webhook Workflow",
			Description: "Webhook triggered workflow",
			Definition:  definition,
			TriggerType: "webhook",
			IsActive:    boolPtr(true),
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.Equal("webhook", workflow.TriggerType)

		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})

	s.Run("Negative Case: Should return 401 for missing auth token", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Unauthorized Workflow",
			Definition:  domain.JSONB{},
			TriggerType: "manual",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for missing name", func() {
		req := dto.CreateWorkflowRequest{
			Description: "Missing name",
			Definition:  domain.JSONB{},
			TriggerType: "manual",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for missing definition", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Missing Definition",
			Description: "No definition provided",
			TriggerType: "manual",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for invalid trigger type", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Invalid Trigger",
			Definition:  domain.JSONB{},
			TriggerType: "invalid_type",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for cron trigger without expression", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Cron Without Expression",
			Definition:  domain.JSONB{},
			TriggerType: "cron",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *AuthWorkflowTestSuite) TestWorkflowList() {
	s.Run("Positive Case: Should list workflows with default pagination", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows", Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var result struct {
			Content []domain.Workflow `json:"content"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		s.Require().NoError(err)

		s.GreaterOrEqual(len(result.Content), 1)
	})

	s.Run("Positive Case: Should list workflows with pagination parameters", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows?page=1&page_size=2", Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var result struct {
			Content []domain.Workflow `json:"content"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		s.Require().NoError(err)

		s.LessOrEqual(len(result.Content), 2)
	})

	s.Run("Positive Case: Should handle page 2 pagination", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows?page=2&page_size=2", Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Positive Case: Should handle negative page by defaulting to 1", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows?page=-1&page_size=10", Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 401 without auth token", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})
}

func (s *AuthWorkflowTestSuite) TestWorkflowGetByID() {
	s.Require().NotEmpty(s.workflowIDs, "No workflow IDs available for testing")

	validWorkflowID := s.workflowIDs[0]

	s.Run("Positive Case: Should get workflow by ID successfully", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + validWorkflowID.String(), Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.Equal(validWorkflowID, workflow.ID)
	})

	s.Run("Negative Case: Should return 404 for non-existent workflow ID", func() {
		nonExistentID := uuid.New()
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + nonExistentID.String(), Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 400 for invalid UUID format", func() {
		invalidID := "not-a-valid-uuid"
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + invalidID, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("Negative Case: Should return 401 without auth token", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/workflows/" + validWorkflowID.String()}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})
}

func (s *AuthWorkflowTestSuite) TestWorkflowCreateWithDifferentStatuses() {
	s.Run("Should create inactive workflow", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Inactive Workflow",
			Description: "This workflow is inactive",
			Definition:  domain.JSONB{"test": true},
			TriggerType: "manual",
			IsActive:    boolPtr(false),
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.False(workflow.IsActive)
		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})

	s.Run("Should create workflow with default active status when omitted", func() {
		req := dto.CreateWorkflowRequest{
			Name:        "Default Active Workflow",
			Description: "No is_active field",
			Definition:  domain.JSONB{"test": true},
			TriggerType: "manual",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.True(workflow.IsActive)
		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})
}

func (s *AuthWorkflowTestSuite) TestWorkflowWithLongNamesAndDescriptions() {
	s.Run("Should create workflow with long name and description", func() {
		longName := "This is a very long workflow name that exceeds typical length requirements for testing purposes"
		longDescription := "This is an extremely long description that goes into great detail about what this workflow does. " +
			"It describes the purpose, the steps involved, the expected outcomes, and any potential side effects. " +
			"This helps to test that the database can handle large text fields without any issues."

		req := dto.CreateWorkflowRequest{
			Name:        longName,
			Description: longDescription,
			Definition:  domain.JSONB{"complex": "data", "nested": map[string]interface{}{"key": "value"}},
			TriggerType: "manual",
		}

		resp, err := APIRequest{Method: http.MethodPost, Path: "/workflows", Body: req, Token: s.token}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		bodyBytes, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		err = json.Unmarshal(bodyBytes, &res)
		s.Require().NoError(err)

		var workflow domain.Workflow
		if workflowData, ok := res["workflow"]; ok {
			workflowBytes, _ := json.Marshal(workflowData)
			json.Unmarshal(workflowBytes, &workflow)
		} else {
			json.Unmarshal(bodyBytes, &workflow)
		}

		s.Equal(longName, workflow.Name)
		s.Equal(longDescription, workflow.Description)
		s.workflowIDs = append(s.workflowIDs, workflow.ID)
	})
}

func (s *AuthWorkflowTestSuite) TestHealthEndpoint() {
	s.Run("Should return health status", func() {
		resp, err := APIRequest{Method: http.MethodGet, Path: "/health"}.Send()
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var healthResp dto.HealthResponse
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		s.Require().NoError(err)

		s.Equal("ok", healthResp.Status)
		s.NotEmpty(healthResp.Timestamp)
	})
}

func (s *AuthWorkflowTestSuite) TearDownSuite() {
	s.T().Logf("Test suite completed. Created %d workflows", len(s.workflowIDs))
}

func boolPtr(b bool) *bool {
	return &b
}

func TestAuthWorkflowSuite(t *testing.T) {
	suite.Run(t, new(AuthWorkflowTestSuite))
}
