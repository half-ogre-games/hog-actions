package main

import (
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
			name: "valid configuration with defaults",
			setupEnv: func() {
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Branch:         "main", // fallback
				Commit:         "HEAD",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:         "v",
				DefaultVersion: "v0.1.0",
				GitHubToken:    "test-token",
				DefaultBranch:  "main",
			},
		},
		{
			name: "custom configuration",
			setupEnv: func() {
				os.Setenv("INPUT_BRANCH", "develop")
				os.Setenv("INPUT_COMMIT", "abc123")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_PREFIX", "release-")
				os.Setenv("INPUT_DEFAULT_VERSION", "release-1.0.0")
				os.Setenv("INPUT_GITHUB_TOKEN", "custom-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_BRANCH")
				os.Unsetenv("INPUT_COMMIT")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Branch:         "develop",
				Commit:         "abc123",
				IncrementMajor: true,
				IncrementMinor: false,
				Prefix:         "release-",
				DefaultVersion: "release-1.0.0",
				GitHubToken:    "custom-token",
				DefaultBranch:  "develop",
			},
		},
		{
			name: "minor increment",
			setupEnv: func() {
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: false,
			expected: &Config{
				Branch:         "main",
				Commit:         "HEAD",
				IncrementMajor: false,
				IncrementMinor: true,
				Prefix:         "v",
				DefaultVersion: "v0.1.0",
				GitHubToken:    "test-token",
				DefaultBranch:  "main",
			},
		},
		{
			name: "missing github token",
			setupEnv: func() {
				// Don't set github token
			},
			cleanupEnv: func() {},
			expectError: true,
			errorMsg:    "input required and not supplied: github-token",
		},
		{
			name: "both major and minor increment - should error",
			setupEnv: func() {
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			expectError: true,
			errorMsg:    "cannot increment both major and minor versions simultaneously",
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

			if config.Branch != tt.expected.Branch {
				t.Errorf("Branch = %q, want %q", config.Branch, tt.expected.Branch)
			}
			if config.Commit != tt.expected.Commit {
				t.Errorf("Commit = %q, want %q", config.Commit, tt.expected.Commit)
			}
			if config.IncrementMajor != tt.expected.IncrementMajor {
				t.Errorf("IncrementMajor = %v, want %v", config.IncrementMajor, tt.expected.IncrementMajor)
			}
			if config.IncrementMinor != tt.expected.IncrementMinor {
				t.Errorf("IncrementMinor = %v, want %v", config.IncrementMinor, tt.expected.IncrementMinor)
			}
			if config.Prefix != tt.expected.Prefix {
				t.Errorf("Prefix = %q, want %q", config.Prefix, tt.expected.Prefix)
			}
			if config.DefaultVersion != tt.expected.DefaultVersion {
				t.Errorf("DefaultVersion = %q, want %q", config.DefaultVersion, tt.expected.DefaultVersion)
			}
			if config.GitHubToken != tt.expected.GitHubToken {
				t.Errorf("GitHubToken = %q, want %q", config.GitHubToken, tt.expected.GitHubToken)
			}
		})
	}
}

func TestGetTargetCommitSHA(t *testing.T) {
	tests := []struct {
		name        string
		commit      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "HEAD commit",
			commit:      "HEAD",
			expectError: false,
		},
		{
			name:        "empty commit defaults to HEAD",
			commit:      "",
			expectError: false,
		},
		{
			name:        "invalid commit SHA",
			commit:      "invalid-sha-that-does-not-exist",
			expectError: true,
			errorMsg:    "commit invalid-sha-that-does-not-exist does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getTargetCommitSHA(tt.commit)

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

			// For HEAD or empty, result should be a valid SHA (40 characters)
			if len(result) != 40 {
				t.Errorf("Expected SHA to be 40 characters, got %d: %s", len(result), result)
			}
		})
	}
}

func TestTagExists(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "non-existent tag",
			tag:      "v999.999.999",
			expected: false,
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagExists(tt.tag)
			if result != tt.expected {
				t.Errorf("tagExists(%q) = %v, want %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestSetOutputs(t *testing.T) {
	tests := []struct {
		name   string
		result *Result
	}{
		{
			name: "complete result",
			result: &Result{
				PreviousVersion: "v1.0.0",
				NewVersion:      "v1.1.0",
				IncrementType:   "minor",
				ReleaseURL:      "https://github.com/repo/releases/tag/v1.1.0",
				TargetCommit:    "abc123",
				Success:         true,
			},
		},
		{
			name: "first release",
			result: &Result{
				PreviousVersion: "none",
				NewVersion:      "v0.1.0",
				IncrementType:   "patch",
				ReleaseURL:      "https://github.com/repo/releases/tag/v0.1.0",
				TargetCommit:    "def456",
				Success:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just ensures the function doesn't panic
			// In a real environment, it would test GitHub Actions output setting
			err := setOutputs(tt.result)
			if err != nil {
				t.Errorf("setOutputs() error = %v", err)
			}
		})
	}
}