package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAddComment(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		issueNumber  string
		commentBody  string
		expectedID   int
		expectError  bool
	}{
		{
			name:         "successful comment creation",
			responseCode: http.StatusCreated,
			responseBody: `{"id": 789012, "body": "This is a test comment"}`,
			issueNumber:  "456",
			commentBody:  "This is a test comment",
			expectedID:   789012,
			expectError:  false,
		},
		{
			name:         "empty comment body",
			responseCode: http.StatusCreated,
			responseBody: `{"id": 789013, "body": ""}`,
			issueNumber:  "456",
			commentBody:  "",
			expectedID:   789013,
			expectError:  false,
		},
		{
			name:         "API error - unauthorized",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"message": "Bad credentials"}`,
			issueNumber:  "456",
			commentBody:  "Test comment",
			expectedID:   0,
			expectError:  true,
		},
		{
			name:         "API error - issue not found",
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "Not Found"}`,
			issueNumber:  "999",
			commentBody:  "Test comment",
			expectedID:   0,
			expectError:  true,
		},
		{
			name:         "API error - forbidden",
			responseCode: http.StatusForbidden,
			responseBody: `{"message": "Forbidden"}`,
			issueNumber:  "456",
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
				expectedPath := fmt.Sprintf("/repos/test/repo/issues/%s/comments", tt.issueNumber)
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
				if r.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
					t.Errorf("Expected X-GitHub-Api-Version header to be '2022-11-28', got '%s'", 
						r.Header.Get("X-GitHub-Api-Version"))
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
			
			// Test addComment function
			commentID, err := addComment("test/repo", tt.issueNumber, tt.commentBody, "test-token")
			
			// Check error expectation
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

func TestAddCommentWithMultilineBody(t *testing.T) {
	multilineComment := `This is a multiline comment.

It spans multiple lines and includes:
- Lists
- **Bold text**
- Code blocks

## Headers

And other markdown content.`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id": 999999, "body": "multiline content"}`)
	}))
	defer server.Close()

	os.Setenv("GITHUB_API_URL", server.URL)
	
	commentID, err := addComment("test/repo", "123", multilineComment, "test-token")
	
	if err != nil {
		t.Errorf("Unexpected error with multiline comment: %v", err)
	}
	if commentID != 999999 {
		t.Errorf("Expected comment ID 999999, got %d", commentID)
	}
}

// TestMain is omitted because testing functions that call os.Exit is complex
// In production code, we'd refactor main() to return an error instead of exiting