package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
				os.Setenv("INPUT_ISSUE_TITLE", "Bug Report")
				os.Setenv("INPUT_ISSUE_BODY", "Something is broken")
				os.Setenv("INPUT_ISSUE_LABEL", "bug")
				os.Setenv("INPUT_ADDITIONAL_LABELS", "urgent, high-priority")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_TITLE")
				os.Unsetenv("INPUT_ISSUE_BODY")
				os.Unsetenv("INPUT_ISSUE_LABEL")
				os.Unsetenv("INPUT_ADDITIONAL_LABELS")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Repository:       "test/repo",
				Title:            "Bug Report",
				Body:             "Something is broken",
				PrimaryLabel:     "bug",
				AdditionalLabels: "urgent, high-priority",
				Token:            "test-token",
			},
		},
		{
			name: "minimal configuration without additional labels",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_TITLE", "Feature Request")
				os.Setenv("INPUT_ISSUE_LABEL", "enhancement")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
				// No body or additional labels
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_TITLE")
				os.Unsetenv("INPUT_ISSUE_LABEL")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Repository:       "test/repo",
				Title:            "Feature Request",
				Body:             "",
				PrimaryLabel:     "enhancement",
				AdditionalLabels: "",
				Token:            "test-token",
			},
		},
		{
			name: "missing repository",
			setupEnv: func() {
				os.Setenv("INPUT_ISSUE_TITLE", "Test Issue")
				os.Setenv("INPUT_ISSUE_LABEL", "test")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_ISSUE_TITLE")
				os.Unsetenv("INPUT_ISSUE_LABEL")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "GITHUB_REPOSITORY environment variable is required",
		},
		{
			name: "missing title",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_LABEL", "test")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_LABEL")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "issue-title input is required",
		},
		{
			name: "missing primary label",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_TITLE", "Test Issue")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_TITLE")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "issue-label input is required",
		},
		{
			name: "missing token",
			setupEnv: func() {
				os.Setenv("GITHUB_REPOSITORY", "test/repo")
				os.Setenv("INPUT_ISSUE_TITLE", "Test Issue")
				os.Setenv("INPUT_ISSUE_LABEL", "test")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("INPUT_ISSUE_TITLE")
				os.Unsetenv("INPUT_ISSUE_LABEL")
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
			if config.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", config.Title, tt.expected.Title)
			}
			if config.Body != tt.expected.Body {
				t.Errorf("Body = %q, want %q", config.Body, tt.expected.Body)
			}
			if config.PrimaryLabel != tt.expected.PrimaryLabel {
				t.Errorf("PrimaryLabel = %q, want %q", config.PrimaryLabel, tt.expected.PrimaryLabel)
			}
			if config.AdditionalLabels != tt.expected.AdditionalLabels {
				t.Errorf("AdditionalLabels = %q, want %q", config.AdditionalLabels, tt.expected.AdditionalLabels)
			}
			if config.Token != tt.expected.Token {
				t.Errorf("Token = %q, want %q", config.Token, tt.expected.Token)
			}
		})
	}
}

func TestBuildLabels(t *testing.T) {
	tests := []struct {
		name             string
		primaryLabel     string
		additionalLabels string
		expected         []string
	}{
		{
			name:             "only primary label",
			primaryLabel:     "bug",
			additionalLabels: "",
			expected:         []string{"bug"},
		},
		{
			name:             "primary and additional labels",
			primaryLabel:     "bug",
			additionalLabels: "urgent, high-priority",
			expected:         []string{"bug", "urgent", "high-priority"},
		},
		{
			name:             "additional labels with extra spaces",
			primaryLabel:     "enhancement",
			additionalLabels: " feature ,  ui , design ",
			expected:         []string{"enhancement", "feature", "ui", "design"},
		},
		{
			name:             "additional labels with empty entries",
			primaryLabel:     "bug",
			additionalLabels: "urgent, , high-priority,",
			expected:         []string{"bug", "urgent", "high-priority"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildLabels(tt.primaryLabel, tt.additionalLabels)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d labels, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Label[%d] = %q, want %q", i, result[i], expected)
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name               string
		config             *Config
		responseCode       int
		responseBody       string
		expectError        bool
		expectedIssueNumber int
		expectedSuccess    bool
	}{
		{
			name: "successful issue creation",
			config: &Config{
				Repository:       "test/repo",
				Title:            "Bug Report",
				Body:             "Something is broken",
				PrimaryLabel:     "bug",
				AdditionalLabels: "urgent",
				Token:            "test-token",
			},
			responseCode:        http.StatusCreated,
			responseBody:        `{"number": 123, "title": "Bug Report"}`,
			expectError:         false,
			expectedIssueNumber: 123,
			expectedSuccess:     true,
		},
		{
			name: "successful issue creation with empty body and no additional labels",
			config: &Config{
				Repository:       "test/repo",
				Title:            "Feature Request",
				Body:             "",
				PrimaryLabel:     "enhancement",
				AdditionalLabels: "",
				Token:            "test-token",
			},
			responseCode:        http.StatusCreated,
			responseBody:        `{"number": 456, "title": "Feature Request"}`,
			expectError:         false,
			expectedIssueNumber: 456,
			expectedSuccess:     true,
		},
		{
			name: "API error - unauthorized",
			config: &Config{
				Repository:       "test/repo",
				Title:            "Test Issue",
				Body:             "Test body",
				PrimaryLabel:     "test",
				AdditionalLabels: "",
				Token:            "invalid-token",
			},
			responseCode:        http.StatusUnauthorized,
			responseBody:        `{"message": "Bad credentials"}`,
			expectError:         true,
			expectedIssueNumber: 0,
			expectedSuccess:     false,
		},
		{
			name: "API error - validation failure",
			config: &Config{
				Repository:       "test/repo",
				Title:            "",
				Body:             "Test body",
				PrimaryLabel:     "test",
				AdditionalLabels: "",
				Token:            "test-token",
			},
			responseCode:        http.StatusUnprocessableEntity,
			responseBody:        `{"message": "Validation Failed"}`,
			expectError:         true,
			expectedIssueNumber: 0,
			expectedSuccess:     false,
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
				expectedPath := "/repos/test/repo/issues"
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
				if r.Header.Get("Authorization") != "Bearer "+tt.config.Token {
					t.Errorf("Expected Authorization header to be 'Bearer %s', got '%s'", 
						tt.config.Token, r.Header.Get("Authorization"))
				}

				// Send response
				w.WriteHeader(tt.responseCode)
				fmt.Fprint(w, tt.responseBody)
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

			// Check issue number
			if result.IssueNumber != tt.expectedIssueNumber {
				t.Errorf("IssueNumber = %d, want %d", result.IssueNumber, tt.expectedIssueNumber)
			}
		})
	}
}

func TestCreateIssue(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		title          string
		body           string
		labels         []string
		expectedNumber int
		expectError    bool
	}{
		{
			name:           "successful issue creation",
			responseCode:   http.StatusCreated,
			responseBody:   `{"number": 123, "title": "Test Issue"}`,
			title:          "Test Issue",
			body:           "Test body",
			labels:         []string{"bug", "urgent"},
			expectedNumber: 123,
			expectError:    false,
		},
		{
			name:           "API error - unauthorized",
			responseCode:   http.StatusUnauthorized,
			responseBody:   `{"message": "Bad credentials"}`,
			title:          "Test Issue",
			body:           "Test body",
			labels:         []string{"bug"},
			expectedNumber: 0,
			expectError:    true,
		},
		{
			name:           "API error - validation failure",
			responseCode:   http.StatusUnprocessableEntity,
			responseBody:   `{"message": "Validation Failed"}`,
			title:          "",
			body:           "Test body",
			labels:         []string{},
			expectedNumber: 0,
			expectError:    true,
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
			
			// Test createIssue
			issueNumber, err := createIssue("test/repo", tt.title, tt.body, tt.labels, "test-token")
			
			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check issue number
			if issueNumber != tt.expectedNumber {
				t.Errorf("Expected issue number %d, got %d", tt.expectedNumber, issueNumber)
			}
		})
	}
}

func TestCreateIssueEmptyLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"number": 456, "title": "Test Issue"}`)
	}))
	defer server.Close()

	os.Setenv("GITHUB_API_URL", server.URL)
	defer os.Unsetenv("GITHUB_API_URL")
	
	// Test with empty labels
	issueNumber, err := createIssue("test/repo", "Test Issue", "Test body", []string{}, "test-token")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if issueNumber != 456 {
		t.Errorf("Expected issue number 456, got %d", issueNumber)
	}
}