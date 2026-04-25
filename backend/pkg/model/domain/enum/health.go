package enum

type HealthStatus string
type DatabaseStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthDegraded HealthStatus = "degraded"
	DatabaseUp     DatabaseStatus = "up"
	DatabaseDown   DatabaseStatus = "down"
)
