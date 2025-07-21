package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance tests that build and run the actual binary
func TestAcceptanceGetNextSemver(t *testing.T) {
	// Create temporary directory for building
	tempBuildDir, err := os.MkdirTemp("", "build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp build dir: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Build the binary in temp directory
	binaryPath := filepath.Join(tempBuildDir, "get-next-semver")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		expectedOutput string
		expectedError  bool
		expectStderr   bool
	}{
		{
			name: "patch increment with v prefix",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.2.4 (patch increment)",
			expectedError:  false,
		},
		{
			name: "minor increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.3.0 (minor increment)",
			expectedError:  false,
		},
		{
			name: "major increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v2.0.0 (major increment)",
			expectedError:  false,
		},
		{
			name: "version with prerelease - removes prerelease",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3-alpha.1")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.2.4 (patch increment)",
			expectedError:  false,
		},
		{
			name: "version with build metadata - removes build metadata",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3+build.456")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.2.4 (patch increment)",
			expectedError:  false,
		},
		{
			name: "version with both prerelease and build metadata",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3-beta.1+build.789")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.3.0 (minor increment)",
			expectedError:  false,
		},
		{
			name: "no prefix",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "1.2.3")
				os.Setenv("INPUT_PREFIX", "")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: 1.2.4 (patch increment)",
			expectedError:  false,
		},
		{
			name: "custom prefix",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "release-1.2.3")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_PREFIX", "release-")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: release-2.0.0 (major increment)",
			expectedError:  false,
		},
		{
			name: "zero version major increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v0.1.0")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v1.0.0 (major increment)",
			expectedError:  false,
		},
		{
			name: "missing current version",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "invalid current version",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "invalid-version")
				os.Setenv("INPUT_PREFIX", "")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "both major and minor increment - should error",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedError: true,
			expectStderr:  true,
		},
		{
			name: "default prefix when not specified",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				// Don't set INPUT_PREFIX - should default to "v"
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
			},
			expectedOutput: "Next version: v1.2.4 (patch increment)",
			expectedError:  false,
		},
		{
			name: "large version numbers",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v999.888.777")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectedOutput: "Next version: v999.889.0 (minor increment)",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Run the binary
			cmd := exec.Command(binaryPath)
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

			// Verify that GitHub Actions outputs are set correctly
			if !tt.expectedError {
				checkOutputsInAcceptanceTest(t, outputStr, tt.name)
			}
		})
	}
}

// Helper function to check that GitHub Actions outputs are properly formatted
func checkOutputsInAcceptanceTest(t *testing.T, output, testName string) {
	expectedOutputs := []string{
		"::set-output name=version::",
		"::set-output name=version-core::",
		"::set-output name=major::",
		"::set-output name=minor::",
		"::set-output name=patch::",
		"::set-output name=increment-type::",
	}

	for _, expectedOutput := range expectedOutputs {
		if !contains(output, expectedOutput) {
			t.Errorf("Test %s: Expected output to contain %q, but it was missing from: %s", testName, expectedOutput, output)
		}
	}
}

// Test that specific outputs match expected values for key scenarios
func TestAcceptanceGetNextSemverOutputs(t *testing.T) {
	// Create temporary directory for building
	tempBuildDir, err := os.MkdirTemp("", "build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp build dir: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Build the binary in temp directory
	binaryPath := filepath.Join(tempBuildDir, "get-next-semver")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name           string
		currentVersion string
		incrementMajor string
		incrementMinor string
		prefix         string
		expectedOutputs map[string]string
	}{
		{
			name:           "patch increment outputs",
			currentVersion: "v1.2.3",
			prefix:         "v",
			expectedOutputs: map[string]string{
				"version":        "v1.2.4",
				"version-core":   "1.2.4",
				"major":          "1",
				"minor":          "2",
				"patch":          "4",
				"increment-type": "patch",
			},
		},
		{
			name:           "minor increment outputs",
			currentVersion: "v1.2.3",
			incrementMinor: "true",
			prefix:         "v",
			expectedOutputs: map[string]string{
				"version":        "v1.3.0",
				"version-core":   "1.3.0",
				"major":          "1",
				"minor":          "3",
				"patch":          "0",
				"increment-type": "minor",
			},
		},
		{
			name:           "major increment outputs",
			currentVersion: "v1.2.3",
			incrementMajor: "true",
			prefix:         "v",
			expectedOutputs: map[string]string{
				"version":        "v2.0.0",
				"version-core":   "2.0.0",
				"major":          "2",
				"minor":          "0",
				"patch":          "0",
				"increment-type": "major",
			},
		},
		{
			name:           "no prefix outputs",
			currentVersion: "1.2.3",
			prefix:         "",
			expectedOutputs: map[string]string{
				"version":        "1.2.4",
				"version-core":   "1.2.4",
				"major":          "1",
				"minor":          "2",
				"patch":          "4",
				"increment-type": "patch",
			},
		},
		{
			name:           "custom prefix outputs",
			currentVersion: "release-5.10.15",
			incrementMinor: "true",
			prefix:         "release-",
			expectedOutputs: map[string]string{
				"version":        "release-5.11.0",
				"version-core":   "5.11.0",
				"major":          "5",
				"minor":          "11",
				"patch":          "0",
				"increment-type": "minor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("INPUT_CURRENT_VERSION", tt.currentVersion)
			if tt.incrementMajor != "" {
				os.Setenv("INPUT_INCREMENT_MAJOR", tt.incrementMajor)
			}
			if tt.incrementMinor != "" {
				os.Setenv("INPUT_INCREMENT_MINOR", tt.incrementMinor)
			}
			os.Setenv("INPUT_PREFIX", tt.prefix)

			defer func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
				os.Unsetenv("INPUT_PREFIX")
			}()

			// Run the binary
			cmd := exec.Command(binaryPath)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
				return
			}

			outputStr := string(output)

			// Check each expected output
			for outputName, expectedValue := range tt.expectedOutputs {
				expectedLine := "::set-output name=" + outputName + "::" + expectedValue
				if !contains(outputStr, expectedLine) {
					t.Errorf("Expected output line %q not found in output:\n%s", expectedLine, outputStr)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}