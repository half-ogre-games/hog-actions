package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance tests that build and run the actual binary
func TestAcceptanceTagAndCreateSemverRelease(t *testing.T) {
	// Create temporary directory for building
	tempBuildDir, err := os.MkdirTemp("", "build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp build dir: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Build the binary in temp directory
	binaryPath := filepath.Join(tempBuildDir, "tag-and-create-semver-release")

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
			name: "missing github token",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
				// Don't set github token
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize minimal git repo
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}

				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				cmd.Run()
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				cmd.Run()
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "both major and minor increment",
			setupEnv: func() {
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
				os.Setenv("INPUT_GITHUB_TOKEN", "fake-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize minimal git repo
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}

				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				cmd.Run()
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				cmd.Run()
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "no git repository",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
				os.Setenv("INPUT_GITHUB_TOKEN", "fake-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Don't initialize git - should cause error when trying to get tags
			},
			expectedError: true,
			expectStderr:  true,
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
				if tt.expectStderr && err == nil {
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

// Test that the action can successfully process version calculation without actually creating tags/releases
func TestAcceptanceTagAndCreateSemverReleaseVersionCalculation(t *testing.T) {
	// Create temporary directory for building
	tempBuildDir, err := os.MkdirTemp("", "build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp build dir: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Build the binary in temp directory
	binaryPath := filepath.Join(tempBuildDir, "tag-and-create-semver-release")

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
		stopBeforeGit  bool // Stop execution before git operations to test version calculation
	}{
		{
			name: "version calculation with no existing tags",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
				os.Setenv("INPUT_GITHUB_TOKEN", "fake-token")
				// We'll let this fail at git operations, but it should calculate version first
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize git repo with no tags
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}

				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				cmd.Run()
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				cmd.Run()

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
			},
			expectedOutput: "Next version: v0.1.1 (patch increment)",
		},
		{
			name: "version calculation with existing tags",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
				os.Setenv("INPUT_DEFAULT_VERSION", "v0.1.0")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_GITHUB_TOKEN", "fake-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_GITHUB_TOKEN")
			},
			setupGit: func(t *testing.T, tempDir string) {
				// Initialize git repo with existing tags
				cmd := exec.Command("git", "init")
				cmd.Dir = tempDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to initialize git: %v", err)
				}

				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tempDir
				cmd.Run()
				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tempDir
				cmd.Run()

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

				// Create existing tags
				tags := []string{"v1.0.0", "v1.1.0", "v1.0.1"}
				for _, tag := range tags {
					cmd = exec.Command("git", "tag", tag)
					cmd.Dir = tempDir
					if err := cmd.Run(); err != nil {
						t.Fatalf("Failed to create tag %s: %v", tag, err)
					}
				}
			},
			expectedOutput: "Next version: v1.2.0 (minor increment)",
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

			// We expect this to fail at git operations (tagging/release creation)
			// but we want to verify it calculates the version correctly first
			outputStr := string(output)

			if !contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectedOutput, outputStr)
			}

			// The command should fail at git operations, but that's expected
			// We're testing the version calculation logic
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}