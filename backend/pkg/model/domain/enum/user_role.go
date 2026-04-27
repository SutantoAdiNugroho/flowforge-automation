package enum

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super-admin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleEditor     UserRole = "editor"
	UserRoleViewer     UserRole = "viewer"
)
