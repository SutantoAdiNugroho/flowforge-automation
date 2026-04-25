package constant

type HealthStatus string
type DatabaseStatus string

const (
	HealthStatusOK       HealthStatus = "ok"
	HealthStatusDegraded HealthStatus = "degraded"
	DatabaseStatusUp     DatabaseStatus = "up"
	DatabaseStatusDown   DatabaseStatus = "down"
)
