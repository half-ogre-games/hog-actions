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

func TestAcceptanceCommentIssueSuccess(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/issues/123/comments") {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id": 456789, "body": "This is a test comment"}`)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "123",
		"INPUT_COMMENT_BODY": "This is a test comment",
		"INPUT_GITHUB_TOKEN": "test-token",
		"GITHUB_API_URL":     server.URL,
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

	if !strings.Contains(stdout, "Comment added successfully (ID: 456789)") {
		t.Errorf("Expected stdout to contain 'Comment added successfully (ID: 456789)', got: %s", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestAcceptanceCommentIssueLongComment(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/issues/999/comments") {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id": 111222, "body": "This is a very long comment..."}`)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	longComment := "This is a very long comment that contains multiple sentences. " +
		"It has various punctuation marks and special characters like @user, #123, and [links](http://example.com). " +
		"The comment also includes line breaks and formatting."

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "999",
		"INPUT_COMMENT_BODY": longComment,
		"INPUT_GITHUB_TOKEN": "test-token",
		"GITHUB_API_URL":     server.URL,
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

	if !strings.Contains(stdout, "Comment added successfully (ID: 111222)") {
		t.Errorf("Expected stdout to contain 'Comment added successfully (ID: 111222)', got: %s", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestAcceptanceCommentIssueMissingInput(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup environment with missing comment body
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "123",
		"INPUT_GITHUB_TOKEN": "test-token",
		// Missing INPUT_COMMENT_BODY
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

	if !strings.Contains(stderr, "comment-body input is required") {
		t.Errorf("Expected stderr to contain 'comment-body input is required', got: %s", stderr)
	}
}

func TestAcceptanceCommentIssueMissingIssueNumber(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup environment with missing issue number
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY": "test/repo",
		"INPUT_COMMENT_BODY": "Test comment",
		"INPUT_GITHUB_TOKEN": "test-token",
		// Missing INPUT_ISSUE_NUMBER
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

	if !strings.Contains(stderr, "issue-number input is required") {
		t.Errorf("Expected stderr to contain 'issue-number input is required', got: %s", stderr)
	}
}

func TestAcceptanceCommentIssueAPIError(t *testing.T) {
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
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "456",
		"INPUT_COMMENT_BODY": "Test comment",
		"INPUT_GITHUB_TOKEN": "invalid-token",
		"GITHUB_API_URL":     server.URL,
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

	expectedStderr := []string{"error adding comment", "API request failed with status 401"}
	for _, expected := range expectedStderr {
		if !strings.Contains(stderr, expected) {
			t.Errorf("Expected stderr to contain %q, got: %s", expected, stderr)
		}
	}
}

func TestAcceptanceCommentIssueForbidden(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server that returns forbidden (no permission to comment)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message": "You do not have permission to create comments on this issue"}`)
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "789",
		"INPUT_COMMENT_BODY": "Test comment",
		"INPUT_GITHUB_TOKEN": "limited-token",
		"GITHUB_API_URL":     server.URL,
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

	expectedStderr := []string{"error adding comment", "API request failed with status 403"}
	for _, expected := range expectedStderr {
		if !strings.Contains(stderr, expected) {
			t.Errorf("Expected stderr to contain %q, got: %s", expected, stderr)
		}
	}
}

func TestAcceptanceCommentIssueNotFound(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test server that returns not found (issue doesn't exist)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message": "Not Found"}`)
	}))
	defer server.Close()

	// Setup environment
	oldEnv := setupEnv(map[string]string{
		"GITHUB_REPOSITORY":  "test/repo",
		"INPUT_ISSUE_NUMBER": "99999",
		"INPUT_COMMENT_BODY": "Test comment",
		"INPUT_GITHUB_TOKEN": "test-token",
		"GITHUB_API_URL":     server.URL,
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

	expectedStderr := []string{"error adding comment", "API request failed with status 404"}
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

// buildBinary builds the comment-issue binary and returns its path
func buildBinary(t *testing.T) string {
	t.Helper()
	
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "comment-issue")
	
	cmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	cmd.Dir = "." // Current directory should be comment-issue/
	
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