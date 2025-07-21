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
			name: "valid configuration - patch increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectError: false,
			expected: &Config{
				CurrentVersion: "v1.2.3",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "v",
			},
		},
		{
			name: "major increment with empty prefix",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "1.2.3")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_PREFIX", "")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_PREFIX")
			},
			expectError: false,
			expected: &Config{
				CurrentVersion: "1.2.3",
				IncrementMajor: true,
				IncrementMinor: false,
				Prefix:        "",
			},
		},
		{
			name: "minor increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v2.5.8")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
			},
			expectError: false,
			expected: &Config{
				CurrentVersion: "v2.5.8",
				IncrementMajor: false,
				IncrementMinor: true,
				Prefix:        "v",
			},
		},
		{
			name: "missing current version",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "v")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
			},
			expectError: true,
			errorMsg:   "current-version input is required",
		},
		{
			name: "both major and minor increment",
			setupEnv: func() {
				os.Setenv("INPUT_CURRENT_VERSION", "v1.2.3")
				os.Setenv("INPUT_INCREMENT_MAJOR", "true")
				os.Setenv("INPUT_INCREMENT_MINOR", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_CURRENT_VERSION")
				os.Unsetenv("INPUT_INCREMENT_MAJOR")
				os.Unsetenv("INPUT_INCREMENT_MINOR")
			},
			expectError: true,
			errorMsg:   "cannot increment both major and minor versions simultaneously",
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

			if config.CurrentVersion != tt.expected.CurrentVersion {
				t.Errorf("CurrentVersion = %q, want %q", config.CurrentVersion, tt.expected.CurrentVersion)
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
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedResult *Result
		expectError    bool
	}{
		{
			name: "patch increment basic",
			config: &Config{
				CurrentVersion: "v1.2.3",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v1.2.4",
				VersionCore:   "1.2.4",
				Major:         1,
				Minor:         2,
				Patch:         4,
				IncrementType: "patch",
				Success:       true,
			},
		},
		{
			name: "minor increment",
			config: &Config{
				CurrentVersion: "v1.2.3",
				IncrementMajor: false,
				IncrementMinor: true,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v1.3.0",
				VersionCore:   "1.3.0",
				Major:         1,
				Minor:         3,
				Patch:         0,
				IncrementType: "minor",
				Success:       true,
			},
		},
		{
			name: "major increment",
			config: &Config{
				CurrentVersion: "v1.2.3",
				IncrementMajor: true,
				IncrementMinor: false,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v2.0.0",
				VersionCore:   "2.0.0",
				Major:         2,
				Minor:         0,
				Patch:         0,
				IncrementType: "major",
				Success:       true,
			},
		},
		{
			name: "version with prerelease - patch increment removes prerelease",
			config: &Config{
				CurrentVersion: "v1.2.3-alpha.1",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v1.2.4",
				VersionCore:   "1.2.4",
				Major:         1,
				Minor:         2,
				Patch:         4,
				IncrementType: "patch",
				Success:       true,
			},
		},
		{
			name: "version with build metadata - removed in result",
			config: &Config{
				CurrentVersion: "v1.2.3+build.456",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v1.2.4",
				VersionCore:   "1.2.4",
				Major:         1,
				Minor:         2,
				Patch:         4,
				IncrementType: "patch",
				Success:       true,
			},
		},
		{
			name: "no prefix",
			config: &Config{
				CurrentVersion: "1.2.3",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "",
			},
			expectedResult: &Result{
				Version:       "1.2.4",
				VersionCore:   "1.2.4",
				Major:         1,
				Minor:         2,
				Patch:         4,
				IncrementType: "patch",
				Success:       true,
			},
		},
		{
			name: "zero version major increment",
			config: &Config{
				CurrentVersion: "v0.1.0",
				IncrementMajor: true,
				IncrementMinor: false,
				Prefix:        "v",
			},
			expectedResult: &Result{
				Version:       "v1.0.0",
				VersionCore:   "1.0.0",
				Major:         1,
				Minor:         0,
				Patch:         0,
				IncrementType: "major",
				Success:       true,
			},
		},
		{
			name: "invalid version format",
			config: &Config{
				CurrentVersion: "invalid-version",
				IncrementMajor: false,
				IncrementMinor: false,
				Prefix:        "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(tt.config)

			if tt.expectError {
				if result.Error == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
				return
			}

			if result.Version != tt.expectedResult.Version {
				t.Errorf("Version = %q, want %q", result.Version, tt.expectedResult.Version)
			}
			if result.VersionCore != tt.expectedResult.VersionCore {
				t.Errorf("VersionCore = %q, want %q", result.VersionCore, tt.expectedResult.VersionCore)
			}
			if result.Major != tt.expectedResult.Major {
				t.Errorf("Major = %d, want %d", result.Major, tt.expectedResult.Major)
			}
			if result.Minor != tt.expectedResult.Minor {
				t.Errorf("Minor = %d, want %d", result.Minor, tt.expectedResult.Minor)
			}
			if result.Patch != tt.expectedResult.Patch {
				t.Errorf("Patch = %d, want %d", result.Patch, tt.expectedResult.Patch)
			}
			if result.IncrementType != tt.expectedResult.IncrementType {
				t.Errorf("IncrementType = %q, want %q", result.IncrementType, tt.expectedResult.IncrementType)
			}
			if result.Success != tt.expectedResult.Success {
				t.Errorf("Success = %v, want %v", result.Success, tt.expectedResult.Success)
			}
		})
	}
}

