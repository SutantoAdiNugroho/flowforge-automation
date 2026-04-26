package domain

type DAGStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config,omitempty"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
}

type RetryPolicy struct {
	MaxRetries      int `json:"max_retries"`
	BackoffMultiplier float64 `json:"backoff_multiplier"`
	InitialDelayMs  int `json:"initial_delay_ms"`
	MaxDelayMs      int `json:"max_delay_ms"`
}

type DAGDefinition struct {
	Steps      []DAGStep `json:"steps"`
	Timeout    int       `json:"timeout,omitempty"`
	Version    string    `json:"version"`
}

type DAGValidationError struct {
	StepID  string
	Field   string
	Message string
}

type DAGValidationResult struct {
	Valid  bool
	Errors []DAGValidationError
}
