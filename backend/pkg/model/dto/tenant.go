package dto

type CreateTenantRequest struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
}

type UpdateTenantRequest struct {
	Name string `json:"name"`
}

type TenantDetailResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	UserCount int64  `json:"user_count"`
	RunCount  int64  `json:"run_count"`
	CreatedAt string `json:"created_at"`
}
