package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptanceCreateIssueSuccess(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/issues") {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"number": 123, "title": "Bug Report"}`)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":       "test/repo",
		"INPUT_ISSUE_TITLE":       "Bug Report",
		"INPUT_ISSUE_BODY":        "Something is broken",
		"INPUT_ISSUE_LABEL":       "bug",
		"INPUT_ADDITIONAL_LABELS": "urgent",
		"INPUT_GITHUB_TOKEN":      "test-token",
		"GITHUB_API_URL":          server.URL,
	})
	defer restoreEnv(oldEnv)

	// Execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	
	stdout, stderr, exitCode := runCommand(cmd)

	// Assertions
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	if !strings.Contains(stdout, "Created new issue #123") {
		t.Errorf("Expected stdout to contain 'Created new issue #123', got: %s", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestAcceptanceCreateIssueMinimal(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/issues") {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"number": 456, "title": "Feature Request"}`)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Setup environment - minimal required fields
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY": "test/repo",
		"INPUT_ISSUE_TITLE": "Feature Request",
		"INPUT_ISSUE_LABEL": "enhancement",
		"INPUT_GITHUB_TOKEN": "test-token",
		"GITHUB_API_URL":    server.URL,
		// No body or additional labels
	})
	defer restoreEnv(oldEnv)

	// Execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	
	stdout, stderr, exitCode := runCommand(cmd)

	// Assertions
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	if !strings.Contains(stdout, "Created new issue #456") {
		t.Errorf("Expected stdout to contain 'Created new issue #456', got: %s", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestAcceptanceCreateIssueMissingInput(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup environment with missing title
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY": "test/repo",
		"INPUT_ISSUE_LABEL": "bug",
		"INPUT_GITHUB_TOKEN": "test-token",
		// Missing INPUT_ISSUE_TITLE
	})
	defer restoreEnv(oldEnv)

	// Execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	
	stdout, stderr, exitCode := runCommand(cmd)

	// Assertions
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	if !strings.Contains(stderr, "issue-title input is required") {
		t.Errorf("Expected stderr to contain 'issue-title input is required', got: %s", stderr)
	}
}

func TestAcceptanceCreateIssueAPIError(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server that returns unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message": "Bad credentials"}`)
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY": "test/repo",
		"INPUT_ISSUE_TITLE": "Test Issue",
		"INPUT_ISSUE_LABEL": "bug",
		"INPUT_GITHUB_TOKEN": "invalid-token",
		"GITHUB_API_URL":    server.URL,
	})
	defer restoreEnv(oldEnv)

	// Execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	
	stdout, stderr, exitCode := runCommand(cmd)

	// Assertions
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	expectedStderr := []string{"error creating issue", "API request failed with status 401"}
	for _, expected := range expectedStderr {
		if !strings.Contains(stderr, expected) {
			t.Errorf("Expected stderr to contain %q, got: %s", expected, stderr)
		}
	}
}

func TestAcceptanceCreateIssueValidationError(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server that returns validation error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"message": "Validation Failed", "errors": [{"field": "title", "code": "missing"}]}`)
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY": "test/repo",
		"INPUT_ISSUE_TITLE": "Test Issue",
		"INPUT_ISSUE_LABEL": "bug",
		"INPUT_GITHUB_TOKEN": "test-token",
		"GITHUB_API_URL":    server.URL,
	})
	defer restoreEnv(oldEnv)

	// Execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	
	stdout, stderr, exitCode := runCommand(cmd)

	// Assertions
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	expectedStderr := []string{"error creating issue", "API request failed with status 422"}
	for _, expected := range expectedStderr {
		if !strings.Contains(stderr, expected) {
			t.Errorf("Expected stderr to contain %q, got: %s", expected, stderr)
		}
	}
}

// setupEnv sets environment variables and returns the old values for restoration
func setupEnv(envVars map[string]string) map[string]string {
	oldEnv := make(map[string]string)
	for key, value := range envVars {
		oldEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	return oldEnv
}

// restoreEnv restores environment variables to their previous values
func restoreEnv(oldEnv map[string]string) {
	for key, value := range oldEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// buildBinary builds the create-issue binary and returns its path
func buildBinary(t *testing.T) string {
	t.Helper()
	
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "create-issue")
	
	cmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	cmd.Dir = "." // Current directory should be create-issue/
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	
	return binaryPath
}

// runCommand executes a command and returns stdout, stderr, and exit code
func runCommand(cmd *exec.Cmd) (stdout, stderr string, exitCode int) {
	stdoutBytes, stderrBytes, err := runCommandBytes(cmd)
	stdout = string(stdoutBytes)
	stderr = string(stderrBytes)
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1 // Some other error
		}
	} else {
		exitCode = 0
	}
	
	return stdout, stderr, exitCode
}

// runCommandBytes executes a command and returns stdout and stderr as bytes
func runCommandBytes(cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	stdoutBuf := &strings.Builder{}
	stderrBuf := &strings.Builder{}
	
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	
	err = cmd.Run()
	stdout = []byte(stdoutBuf.String())
	stderr = []byte(stderrBuf.String())
	
	return stdout, stderr, err
}