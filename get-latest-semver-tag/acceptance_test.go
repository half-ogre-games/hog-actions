package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance tests that build and run the actual binary
func TestAcceptanceGetLatestSemverTag(t *testing.T) {
	// Create temporary directory for building
	tempBuildDir, err := os.MkdirTemp("", "build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp build dir: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Build the binary in temp directory
	binaryPath := filepath.Join(tempBuildDir, "get-latest-semver-tag")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		setupGit       func(t *testing.T, tempDir string)
		expectedOutput string
		expectedError  bool
		expectStderr   bool
	}{
		{
			name: "no git repository",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.0.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Don't initialize git - should cause error
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "git repository with no tags",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize git repository but don't create any tags
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}
				
				// Configure git user for commits
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				cmd.Run()
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				cmd.Run()
			},
			expectedOutput: "No tags found, using default: v0.1.0",
			expectedError:  false,
		},
		{
			name: "git repository with valid semver tags",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.0.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize git repository and create tags
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}
				
				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to configure git email: %v", err)
				}
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to configure git name: %v", err)
				}
				
				// Create an initial commit (required for tags)
				readmeFile := filepath.Join(tempDir, "README.md")
				if err := os.WriteFile(readmeFile, []byte("# Test Repo"), 0644); err != nil {
					t.Fatalf("Failed to create README.md: %v", err)
				}
				
				cmd = exec.Command("git", "add", "README.md")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to add README.md: %v", err)
				}
				
				cmd = exec.Command("git", "commit", "-m", "initial commit")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to create initial commit: %v", err)
				}
				
				// Create some tags
				tags := []string{"v1.0.0", "v1.1.0", "v1.0.1", "v2.0.0-alpha.1"}
				for _, tag := range tags {
					cmd = exec.Command("git", "tag", tag)
					cmd.Dir = tempDir
					if err := cmd.Run(); err != nil {
						t.Fatalf("Failed to create tag %s: %v", tag, err)
					}
				}
			},
			expectedOutput: "Found latest tag: v2.0.0-alpha.1",
			expectedError:  false,
		},
		{
			name: "custom prefix",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "release-")
				os.Setenv("INPUT_DEFAULT_VERSION", "release-0.0.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize git repository and create tags with custom prefix
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}
				
				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to configure git email: %v", err)
				}
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to configure git name: %v", err)
				}
				
				// Create an initial commit
				readmeFile := filepath.Join(tempDir, "README.md")
				if err := os.WriteFile(readmeFile, []byte("# Test Repo"), 0644); err != nil {
					t.Fatalf("Failed to create README.md: %v", err)
				}
				
				cmd = exec.Command("git", "add", "README.md")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to add README.md: %v", err)
				}
				
				cmd = exec.Command("git", "commit", "-m", "initial commit")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to create initial commit: %v", err)
				}
				
				// Create tags with different prefixes
				tags := []string{"release-1.0.0", "v1.0.0", "release-1.1.0"}
				for _, tag := range tags {
					cmd = exec.Command("git", "tag", tag)
					cmd.Dir = tempDir
					if err := cmd.Run(); err != nil {
						t.Fatalf("Failed to create tag %s: %v", tag, err)
					}
				}
			},
			expectedOutput: "Found latest tag: release-1.1.0",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for git repo
			tempDir, err := os.MkdirTemp("", "git-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Set up environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Set up git repository
			tt.setupGit(t, tempDir)

			// Run the binary in the temp directory
			cmd := exec.Command(binaryPath)
			cmd.Dir = tempDir
			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but command succeeded")
				}
				if tt.expectStderr && !tt.expectedError {
					t.Error("Expected stderr output but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
				return
			}

			outputStr := string(output)
			if !contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectedOutput, outputStr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}