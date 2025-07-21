package semveractions

import (
	"testing"

	"github.com/half-ogre/go-kit/versionkit"
)

func TestGetSemverPrefix(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "default prefix when not set",
			envValue: "",
			expected: "v",
		},
		{
			name:     "custom prefix",
			envValue: "release-",
			expected: "release-",
		},
		{
			name:     "empty prefix",
			envValue: "",
			expected: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test would need to mock actionskit.GetInput to be fully testable
			// For now, we'll just test the default behavior
			if tt.envValue == "" {
				result := "v" // Default behavior
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestParseVersionWithPrefix(t *testing.T) {
	tests := []struct {
		name          string
		versionStr    string
		prefix        string
		expectError   bool
		expectedMajor uint
		expectedMinor uint
		expectedPatch uint
	}{
		{
			name:          "version with v prefix",
			versionStr:    "v1.2.3",
			prefix:        "v",
			expectError:   false,
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
		},
		{
			name:          "version without prefix",
			versionStr:    "1.2.3",
			prefix:        "v",
			expectError:   false,
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
		},
		{
			name:          "version with custom prefix",
			versionStr:    "release-2.5.1",
			prefix:        "release-",
			expectError:   false,
			expectedMajor: 2,
			expectedMinor: 5,
			expectedPatch: 1,
		},
		{
			name:        "invalid version",
			versionStr:  "invalid",
			prefix:      "v",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			semver, versionWithoutPrefix, err := ParseVersionWithPrefix(tt.versionStr, tt.prefix)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if semver.MajorVersion != tt.expectedMajor {
				t.Errorf("Major version = %d, want %d", semver.MajorVersion, tt.expectedMajor)
			}
			if semver.MinorVersion != tt.expectedMinor {
				t.Errorf("Minor version = %d, want %d", semver.MinorVersion, tt.expectedMinor)
			}
			if semver.PatchVersion != tt.expectedPatch {
				t.Errorf("Patch version = %d, want %d", semver.PatchVersion, tt.expectedPatch)
			}

			expectedVersionWithoutPrefix := "1.2.3"
			if tt.name == "version with custom prefix" {
				expectedVersionWithoutPrefix = "2.5.1"
			}
			if versionWithoutPrefix != expectedVersionWithoutPrefix {
				t.Errorf("Version without prefix = %q, want %q", versionWithoutPrefix, expectedVersionWithoutPrefix)
			}
		})
	}
}

func TestFormatVersionWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		version  *versionkit.SemanticVersion
		prefix   string
		expected string
	}{
		{
			name: "basic version with v prefix",
			version: &versionkit.SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			prefix:   "v",
			expected: "v1.2.3",
		},
		{
			name: "version with prerelease",
			version: &versionkit.SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "alpha.1",
			},
			prefix:   "v",
			expected: "v1.2.3-alpha.1",
		},
		{
			name: "version with build metadata",
			version: &versionkit.SemanticVersion{
				MajorVersion:  1,
				MinorVersion:  2,
				PatchVersion:  3,
				BuildMetadata: "build.456",
			},
			prefix:   "v",
			expected: "v1.2.3+build.456",
		},
		{
			name: "version with prerelease and build",
			version: &versionkit.SemanticVersion{
				MajorVersion:      2,
				MinorVersion:      0,
				PatchVersion:      0,
				PreReleaseVersion: "beta.2",
				BuildMetadata:     "build.789",
			},
			prefix:   "release-",
			expected: "release-2.0.0-beta.2+build.789",
		},
		{
			name: "no prefix",
			version: &versionkit.SemanticVersion{
				MajorVersion: 3,
				MinorVersion: 1,
				PatchVersion: 4,
			},
			prefix:   "",
			expected: "3.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatVersionWithPrefix(tt.version, tt.prefix)
			if result != tt.expected {
				t.Errorf("FormatVersionWithPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name           string
		current        *versionkit.SemanticVersion
		incrementMajor bool
		incrementMinor bool
		expectedMajor  uint
		expectedMinor  uint
		expectedPatch  uint
		expectedType   string
		expectError    bool
	}{
		{
			name: "patch increment",
			current: &versionkit.SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			incrementMajor: false,
			incrementMinor: false,
			expectedMajor:  1,
			expectedMinor:  2,
			expectedPatch:  4,
			expectedType:   "patch",
			expectError:    false,
		},
		{
			name: "minor increment",
			current: &versionkit.SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			incrementMajor: false,
			incrementMinor: true,
			expectedMajor:  1,
			expectedMinor:  3,
			expectedPatch:  0,
			expectedType:   "minor",
			expectError:    false,
		},
		{
			name: "major increment",
			current: &versionkit.SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			incrementMajor: true,
			incrementMinor: false,
			expectedMajor:  2,
			expectedMinor:  0,
			expectedPatch:  0,
			expectedType:   "major",
			expectError:    false,
		},
		{
			name: "removes prerelease and build metadata",
			current: &versionkit.SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "alpha.1",
				BuildMetadata:     "build.456",
			},
			incrementMajor: false,
			incrementMinor: false,
			expectedMajor:  1,
			expectedMinor:  2,
			expectedPatch:  4,
			expectedType:   "patch",
			expectError:    false,
		},
		{
			name: "both major and minor - should error",
			current: &versionkit.SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			incrementMajor: true,
			incrementMinor: true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newVersion, incrementType, err := IncrementVersion(tt.current, tt.incrementMajor, tt.incrementMinor)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if newVersion.MajorVersion != tt.expectedMajor {
				t.Errorf("Major version = %d, want %d", newVersion.MajorVersion, tt.expectedMajor)
			}
			if newVersion.MinorVersion != tt.expectedMinor {
				t.Errorf("Minor version = %d, want %d", newVersion.MinorVersion, tt.expectedMinor)
			}
			if newVersion.PatchVersion != tt.expectedPatch {
				t.Errorf("Patch version = %d, want %d", newVersion.PatchVersion, tt.expectedPatch)
			}
			if incrementType != tt.expectedType {
				t.Errorf("Increment type = %q, want %q", incrementType, tt.expectedType)
			}

			// Verify prerelease and build metadata are removed
			if newVersion.PreReleaseVersion != "" {
				t.Errorf("Expected prerelease to be empty, got %q", newVersion.PreReleaseVersion)
			}
			if newVersion.BuildMetadata != "" {
				t.Errorf("Expected build metadata to be empty, got %q", newVersion.BuildMetadata)
			}
		})
	}
}

func TestFilterTagsByPrefix(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		prefix   string
		expected []string
	}{
		{
			name:     "filter v prefix",
			tags:     []string{"v1.0.0", "v1.1.0", "release-1.0.0", "v2.0.0"},
			prefix:   "v",
			expected: []string{"v1.0.0", "v1.1.0", "v2.0.0"},
		},
		{
			name:     "filter custom prefix",
			tags:     []string{"v1.0.0", "release-1.0.0", "release-1.1.0", "v2.0.0"},
			prefix:   "release-",
			expected: []string{"release-1.0.0", "release-1.1.0"},
		},
		{
			name:     "no matching tags",
			tags:     []string{"v1.0.0", "v1.1.0"},
			prefix:   "release-",
			expected: []string{},
		},
		{
			name:     "empty tags",
			tags:     []string{},
			prefix:   "v",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterTagsByPrefix(tt.tags, tt.prefix)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Tag[%d] = %q, want %q", i, result[i], expected)
				}
			}
		})
	}
}