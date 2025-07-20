package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFindIssues(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedCount int
		expectError   bool
	}{
		{
			name:         "single issue found",
			responseCode: http.StatusOK,
			responseBody: `[{"number": 123, "state": "open", "title": "Test Issue"}]`,
			expectedCount: 1,
			expectError:  false,
		},
		{
			name:         "no issues found",
			responseCode: http.StatusOK,
			responseBody: `[]`,
			expectedCount: 0,
			expectError:  false,
		},
		{
			name:         "multiple issues but only one matches title",
			responseCode: http.StatusOK,
			responseBody: `[
				{"number": 123, "state": "open", "title": "Test Issue"},
				{"number": 456, "state": "open", "title": "Different Issue"}
			]`,
			expectedCount: 1,
			expectError:  false,
		},
		{
			name:         "API error",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"message": "Bad credentials"}`,
			expectedCount: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				// Verify request headers
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					t.Errorf("Expected Accept header to be 'application/vnd.github+json', got '%s'", 
						r.Header.Get("Accept"))
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
			
			// Test findIssues
			issues, err := findIssues("test/repo", "Test Issue", "test-token")
			
			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check issue count
			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}
		})
	}
}

// TestMain is omitted because testing functions that call os.Exit is complex
// In production code, we'd refactor main() to return an error instead of exiting
