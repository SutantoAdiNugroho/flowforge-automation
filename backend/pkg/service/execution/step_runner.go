package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"flowforge-automation-backend/pkg/model/domain"
)

type StepRunner interface {
	Run(ctx context.Context, step domain.DAGStep, inputs map[string]interface{}) (map[string]interface{}, error)
}

type httpRunner struct {
	client *http.Client
}

func newHTTPRunner() *httpRunner {
	return &httpRunner{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *httpRunner) Run(ctx context.Context, step domain.DAGStep, inputs map[string]interface{}) (map[string]interface{}, error) {
	cfg := step.Config
	url, _ := cfg["url"].(string)
	method, _ := cfg["method"].(string)
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if bodyData, ok := cfg["body"]; ok {
		b, err := json.Marshal(bodyData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if headers, ok := cfg["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(respBody))
	}

	output := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
	}

	var parsed interface{}
	if json.Unmarshal(respBody, &parsed) == nil {
		output["json"] = parsed
	}

	return output, nil
}

type scriptRunner struct{}

func newScriptRunner() *scriptRunner {
	return &scriptRunner{}
}

func (r *scriptRunner) Run(ctx context.Context, step domain.DAGStep, inputs map[string]interface{}) (map[string]interface{}, error) {
	cfg := step.Config
	command, _ := cfg["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("script step requires 'command' in config")
	}

	shell := "sh"
	shellArg := "-c"
	if s, ok := cfg["shell"].(string); ok {
		shell = s
	}

	cmd := exec.CommandContext(ctx, shell, shellArg, command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := map[string]interface{}{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": cmd.ProcessState.ExitCode(),
	}

	if err != nil {
		return output, fmt.Errorf("script failed: %w, stderr: %s", err, stderr.String())
	}

	return output, nil
}

type delayRunner struct{}

func newDelayRunner() *delayRunner {
	return &delayRunner{}
}

func (r *delayRunner) Run(ctx context.Context, step domain.DAGStep, inputs map[string]interface{}) (map[string]interface{}, error) {
	durationMs := 1000
	if d, ok := step.Config["duration_ms"].(float64); ok {
		durationMs = int(d)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(durationMs) * time.Millisecond):
	}

	return map[string]interface{}{"delayed_ms": durationMs}, nil
}

type conditionalRunner struct{}

func newConditionalRunner() *conditionalRunner {
	return &conditionalRunner{}
}

func (r *conditionalRunner) Run(ctx context.Context, step domain.DAGStep, inputs map[string]interface{}) (map[string]interface{}, error) {
	cfg := step.Config
	checkStep, _ := cfg["check_step"].(string)
	checkField, _ := cfg["check_field"].(string)
	expectedValue, hasExpected := cfg["expected_value"]

	result := true

	if checkStep != "" {
		stepOutput, ok := inputs[checkStep]
		if !ok {
			result = false
		} else if outputMap, ok := stepOutput.(map[string]interface{}); ok && checkField != "" {
			actualValue, exists := outputMap[checkField]
			if !exists {
				result = false
			} else if hasExpected {
				result = fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedValue)
			}
		}
	}

	// simple expression: if "expression" field exists, evaluate as truthy
	if expr, ok := cfg["expression"].(string); ok && expr != "" {
		if checkStep == "" {
			result = expr != "" && expr != "false" && expr != "0"
		}
	}

	return map[string]interface{}{
		"condition_result": result,
	}, nil
}

func getRunner(stepType string) StepRunner {
	switch stepType {
	case "http":
		return newHTTPRunner()
	case "script":
		return newScriptRunner()
	case "delay":
		return newDelayRunner()
	case "conditional", "condition":
		return newConditionalRunner()
	default:
		return newHTTPRunner()
	}
}
