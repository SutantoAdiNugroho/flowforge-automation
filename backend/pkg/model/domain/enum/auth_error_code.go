package enum

// AuthErrorCode represents standardized auth-related API error codes.
type AuthErrorCode string

const (
	AuthErrorInvalidRequest    AuthErrorCode = "INVALID_REQUEST"
	AuthErrorLoginFailed       AuthErrorCode = "LOGIN_FAILED"
	AuthErrorRegisterFailed    AuthErrorCode = "REGISTER_FAILED"
	AuthErrorMissingAuth       AuthErrorCode = "MISSING_AUTH"
	AuthErrorInvalidAuthFormat AuthErrorCode = "INVALID_AUTH_FORMAT"
	AuthErrorInvalidToken      AuthErrorCode = "INVALID_TOKEN"
	AuthErrorUnauthorized      AuthErrorCode = "UNAUTHORIZED"
	AuthErrorForbidden         AuthErrorCode = "FORBIDDEN"
)
