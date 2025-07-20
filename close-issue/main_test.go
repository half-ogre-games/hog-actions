package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		expectError bool
		errorMsg    string
		expected    *Config
	}{
		{
			name: "valid configuration",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_NUMBER", "123")
				os.Setenv("INPUT_COMMENT_BODY", "Test comment")
				os.Setenv("INPUT_STATE_REASON", "completed")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_NUMBER")
				os.Unsetenv("INPUT_COMMENT_BODY")
				os.Unsetenv("INPUT_STATE_REASON")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Repository:  "test/repo",
				IssueNumber: "123",
				CommentBody: "Test comment",
				StateReason: "completed",
				Token:       "test-token",
			},
		},
		{
			name: "minimal configuration with defaults",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_NUMBER", "456")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
				// No comment body or state reason
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_NUMBER")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Repository:  "test/repo",
				IssueNumber: "456",
				CommentBody: "",
				StateReason: "closed", // Default value
				Token:       "test-token",
			},
		},
		{
			name: "missing repository",
			setupEnv: func() {
				os.Setenv("INPUT_ISSUE_NUMBER", "123")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_ISSUE_NUMBER")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "GITHUB_REPOSITORY environment variable is required",
		},
		{
			name: "missing issue number",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "issue-number input is required",
		},
		{
			name: "missing token",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_NUMBER", "123")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_NUMBER")
			},
			expectError: true,
			errorMsg:    "github-token input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			config, err := getConfigFromEnvironment()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config.Repository != tt.expected.Repository {
				t.Errorf("Repository = %q, want %q", config.Repository, tt.expected.Repository)
			}
			if config.IssueNumber != tt.expected.IssueNumber {
				t.Errorf("IssueNumber = %q, want %q", config.IssueNumber, tt.expected.IssueNumber)
			}
			if config.CommentBody != tt.expected.CommentBody {
				t.Errorf("CommentBody = %q, want %q", config.CommentBody, tt.expected.CommentBody)
			}
			if config.StateReason != tt.expected.StateReason {
				t.Errorf("StateReason = %q, want %q", config.StateReason, tt.expected.StateReason)
			}
			if config.Token != tt.expected.Token {
				t.Errorf("Token = %q, want %q", config.Token, tt.expected.Token)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name               string
		config             *Config
		commentResponse    string
		commentStatusCode  int
		closeResponse      string
		closeStatusCode    int
		expectError        bool
		expectedCommentID  int
		expectedSuccess    bool
	}{
		{
			name: "successful close without comment",
			config: &Config{
				Repository:  "test/repo",
				IssueNumber: "123",
				CommentBody: "",
				StateReason: "completed",
				Token:       "test-token",
			},
			closeResponse:     `{"number": 123, "state": "closed"}`,
			closeStatusCode:   http.StatusOK,
			expectError:       false,
			expectedCommentID: 0,
			expectedSuccess:   true,
		},
		{
			name: "successful close with comment",
			config: &Config{
				Repository:  "test/repo",
				IssueNumber: "456",
				CommentBody: "Closing this issue",
				StateReason: "completed",
				Token:       "test-token",
			},
			commentResponse:   `{"id": 789012, "body": "Closing this issue"}`,
			commentStatusCode: http.StatusCreated,
			closeResponse:     `{"number": 456, "state": "closed"}`,
			closeStatusCode:   http.StatusOK,
			expectError:       false,
			expectedCommentID: 789012,
			expectedSuccess:   true,
		},
		{
			name: "comment fails",
			config: &Config{
				Repository:  "test/repo",
				IssueNumber: "789",
				CommentBody: "Closing this issue",
				StateReason: "completed",
				Token:       "test-token",
			},
			commentResponse:   `{"message": "Unauthorized"}`,
			commentStatusCode: http.StatusUnauthorized,
			expectError:       true,
			expectedCommentID: 0,
			expectedSuccess:   false,
		},
		{
			name: "close fails",
			config: &Config{
				Repository:  "test/repo",
				IssueNumber: "999",
				CommentBody: "",
				StateReason: "completed",
				Token:       "test-token",
			},
			closeResponse:     `{"message": "Not Found"}`,
			closeStatusCode:   http.StatusNotFound,
			expectError:       true,
			expectedCommentID: 0,
			expectedSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Route based on URL path
				if strings.Contains(r.URL.Path, "/comments") {
					// Comment endpoint
					if r.Method != "POST" {
						t.Errorf("Expected POST for comment, got %s", r.Method)
					}
					w.WriteHeader(tt.commentStatusCode)
					fmt.Fprint(w, tt.commentResponse)
				} else {
					// Close issue endpoint
					if r.Method != "PATCH" {
						t.Errorf("Expected PATCH for close, got %s", r.Method)
					}
					w.WriteHeader(tt.closeStatusCode)
					fmt.Fprint(w, tt.closeResponse)
				}
			}))
			defer server.Close()

			// Override API URL for testing
			os.Setenv("GITHUB_API_URL", server.URL)
			defer os.Unsetenv("GITHUB_API_URL")

			// Test run function
			result := run(tt.config)

			// Check error expectation
			if tt.expectError && result.Error == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
			}

			// Check success
			if result.Success != tt.expectedSuccess {
				t.Errorf("Success = %v, want %v", result.Success, tt.expectedSuccess)
			}

			// Check comment ID
			if result.CommentID != tt.expectedCommentID {
				t.Errorf("CommentID = %d, want %d", result.CommentID, tt.expectedCommentID)
			}
		})
	}
}

func TestAddComment(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		commentBody  string
		expectedID   int
		expectError  bool
	}{
		{
			name:         "successful comment creation",
			responseCode: http.StatusCreated,
			responseBody: `{"id": 123456, "body": "Test comment"}`,
			commentBody:  "Test comment",
			expectedID:   123456,
			expectError:  false,
		},
		{
			name:         "API error - unauthorized",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"message": "Bad credentials"}`,
			commentBody:  "Test comment",
			expectedID:   0,
			expectError:  true,
		},
		{
			name:         "API error - not found",
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "Not Found"}`,
			commentBody:  "Test comment",
			expectedID:   0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				// Verify URL path
				expectedPath := "/repos/test/repo/issues/123/comments"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify request headers
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					t.Errorf("Expected Accept header to be 'application/vnd.github+json', got '%s'", 
						r.Header.Get("Accept"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", 
						r.Header.Get("Content-Type"))
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header to be 'Bearer test-token', got '%s'", 
						r.Header.Get("Authorization"))
				}

				// Send response
				w.WriteHeader(tt.responseCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			// Override API URL for testing
			os.Setenv("GITHUB_API_URL", server.URL)
			defer os.Unsetenv("GITHUB_API_URL")
			
			// Test addComment
			commentID, err := addComment("test/repo", "123", tt.commentBody, "test-token")
			
			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check comment ID
			if commentID != tt.expectedID {
				t.Errorf("Expected comment ID %d, got %d", tt.expectedID, commentID)
			}
		})
	}
}

func TestCloseIssue(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		stateReason  string
		expectError  bool
	}{
		{
			name:         "successful issue close",
			responseCode: http.StatusOK,
			responseBody: `{"number": 123, "state": "closed"}`,
			stateReason:  "completed",
			expectError:  false,
		},
		{
			name:         "API error - unauthorized",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"message": "Bad credentials"}`,
			stateReason:  "completed",
			expectError:  true,
		},
		{
			name:         "API error - not found",
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "Not Found"}`,
			stateReason:  "completed",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "PATCH" {
					t.Errorf("Expected PATCH method, got %s", r.Method)
				}

				// Verify URL path
				expectedPath := "/repos/test/repo/issues/123"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify request headers
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					t.Errorf("Expected Accept header to be 'application/vnd.github+json', got '%s'", 
						r.Header.Get("Accept"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", 
						r.Header.Get("Content-Type"))
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header to be 'Bearer test-token', got '%s'", 
						r.Header.Get("Authorization"))
				}

				// Verify request body contains expected state reason
				body := make([]byte, r.ContentLength)
				r.Body.Read(body)
				bodyStr := string(body)
				if !strings.Contains(bodyStr, tt.stateReason) {
					t.Errorf("Expected request body to contain state reason '%s', got '%s'", 
						tt.stateReason, bodyStr)
				}

				// Send response
				w.WriteHeader(tt.responseCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			// Override API URL for testing
			os.Setenv("GITHUB_API_URL", server.URL)
			defer os.Unsetenv("GITHUB_API_URL")
			
			// Test closeIssue
			err := closeIssue("test/repo", "123", tt.stateReason, "test-token")
			
			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}