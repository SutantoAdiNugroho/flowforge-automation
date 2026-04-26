package execution

import (
	"context"
	"math"
	"time"

	"flowforge-automation-backend/pkg/model/domain"
)

type retryExecutor struct{}

func newRetryExecutor() *retryExecutor {
	return &retryExecutor{}
}

func (r *retryExecutor) executeWithRetry(
	ctx context.Context,
	policy *domain.RetryPolicy,
	fn func(ctx context.Context) (map[string]interface{}, error),
) (map[string]interface{}, int, error) {
	maxRetries := 0
	initialDelay := 1000
	maxDelay := 30000
	multiplier := 2.0

	if policy != nil {
		maxRetries = policy.MaxRetries
		if policy.InitialDelayMs > 0 {
			initialDelay = policy.InitialDelayMs
		}
		if policy.MaxDelayMs > 0 {
			maxDelay = policy.MaxDelayMs
		}
		if policy.BackoffMultiplier > 0 {
			multiplier = policy.BackoffMultiplier
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn(ctx)
		if err == nil {
			return result, attempt, nil
		}
		lastErr = err

		if attempt < maxRetries {
			delay := float64(initialDelay) * math.Pow(multiplier, float64(attempt))
			if int(delay) > maxDelay {
				delay = float64(maxDelay)
			}
			select {
			case <-ctx.Done():
				return nil, attempt, ctx.Err()
			case <-time.After(time.Duration(delay) * time.Millisecond):
			}
		}
	}
	return nil, maxRetries, lastErr
}
