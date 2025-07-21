package main

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/half-ogre/go-kit/versionkit"
)

func TestGetConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		expected   *Config
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				// No environment variables set
			},
			cleanupEnv: func() {
				// Nothing to clean up
			},
			expected: &Config{
				Prefix:         "v",
				DefaultVersion: "v0.0.0",
			},
		},
		{
			name: "custom prefix and default version",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "release-")
				os.Setenv("INPUT_DEFAULT_VERSION", "release-1.0.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			expected: &Config{
				Prefix:         "release-",
				DefaultVersion: "release-1.0.0",
			},
		},
		{
			name: "empty prefix",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "")
				os.Setenv("INPUT_DEFAULT_VERSION", "1.0.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			expected: &Config{
				Prefix:         "v",
				DefaultVersion: "1.0.0",
			},
		},
		{
			name: "no prefix",
			setupEnv: func() {
				os.Setenv("INPUT_PREFIX", "none")
				os.Setenv("INPUT_DEFAULT_VERSION", "0.1.0")
			},
			cleanupEnv: func() {
				os.Unsetenv("INPUT_PREFIX")
				os.Unsetenv("INPUT_DEFAULT_VERSION")
			},
			expected: &Config{
				Prefix:         "none",
				DefaultVersion: "0.1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			config, err := getConfigFromEnvironment()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config.Prefix != tt.expected.Prefix {
				t.Errorf("Prefix = %q, want %q", config.Prefix, tt.expected.Prefix)
			}
			if config.DefaultVersion != tt.expected.DefaultVersion {
				t.Errorf("DefaultVersion = %q, want %q", config.DefaultVersion, tt.expected.DefaultVersion)
			}
		})
	}
}

func TestFindLatestSemverTag(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		prefix   string
		expected string
	}{
		{
			name:     "no tags",
			tags:     []string{},
			prefix:   "v",
			expected: "",
		},
		{
			name:     "single valid tag",
			tags:     []string{"v1.0.0"},
			prefix:   "v",
			expected: "v1.0.0",
		},
		{
			name:     "multiple valid tags - returns latest",
			tags:     []string{"v1.0.0", "v1.1.0", "v1.0.1"},
			prefix:   "v",
			expected: "v1.1.0",
		},
		{
			name:     "mixed valid and invalid tags",
			tags:     []string{"v1.0.0", "invalid-tag", "v1.1.0", "not-a-version"},
			prefix:   "v",
			expected: "v1.1.0",
		},
		{
			name:     "tags with prerelease versions",
			tags:     []string{"v1.0.0", "v1.1.0-alpha.1", "v1.0.1", "v1.1.0-beta.1", "v1.1.0"},
			prefix:   "v",
			expected: "v1.1.0",
		},
		{
			name:     "only prerelease versions",
			tags:     []string{"v1.0.0-alpha.1", "v1.0.0-beta.1", "v1.0.0-rc.1"},
			prefix:   "v",
			expected: "v1.0.0-rc.1",
		},
		{
			name:     "tags with build metadata",
			tags:     []string{"v1.0.0+build.1", "v1.0.0+build.2", "v1.0.1"},
			prefix:   "v",
			expected: "v1.0.1",
		},
		{
			name:     "custom prefix",
			tags:     []string{"release-1.0.0", "v1.0.0", "release-1.1.0"},
			prefix:   "release-",
			expected: "release-1.1.0",
		},
		{
			name:     "no prefix",
			tags:     []string{"1.0.0", "v1.0.0", "1.1.0"},
			prefix:   "",
			expected: "1.1.0",
		},
		{
			name:     "tags don't match prefix",
			tags:     []string{"release-1.0.0", "beta-1.1.0"},
			prefix:   "v",
			expected: "",
		},
		{
			name:     "complex semver comparison",
			tags:     []string{"v2.0.0", "v2.1.0-alpha.1", "v2.0.1", "v2.1.0-beta.1", "v2.1.0"},
			prefix:   "v",
			expected: "v2.1.0",
		},
		{
			name:     "prerelease precedence",
			tags:     []string{"v1.0.0-alpha.1", "v1.0.0-alpha.2", "v1.0.0-beta.1", "v1.0.0-rc.1"},
			prefix:   "v",
			expected: "v1.0.0-rc.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findLatestSemverTag(tt.tags, tt.prefix)
			if result != tt.expected {
				t.Errorf("findLatestSemverTag() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRunWithMockTags(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		mockTags     []string
		expectedTag  string
		expectedFound bool
		expectError  bool
	}{
		{
			name: "no tags found - use default",
			config: &Config{
				Prefix:         "v",
				DefaultVersion: "v0.0.0",
			},
			mockTags:      []string{},
			expectedTag:   "v0.0.0",
			expectedFound: false,
			expectError:   false,
		},
		{
			name: "tags found - use latest",
			config: &Config{
				Prefix:         "v",
				DefaultVersion: "v0.0.0",
			},
			mockTags:      []string{"v1.0.0", "v1.1.0", "v1.0.1"},
			expectedTag:   "v1.1.0",
			expectedFound: true,
			expectError:   false,
		},
		{
			name: "custom prefix",
			config: &Config{
				Prefix:         "release-",
				DefaultVersion: "release-0.0.0",
			},
			mockTags:      []string{"release-1.0.0", "v1.0.0", "release-1.1.0"},
			expectedTag:   "release-1.1.0",
			expectedFound: true,
			expectError:   false,
		},
		{
			name: "invalid default version",
			config: &Config{
				Prefix:         "v",
				DefaultVersion: "invalid-version",
			},
			mockTags:    []string{},
			expectError: true,
		},
		{
			name: "prerelease versions",
			config: &Config{
				Prefix:         "v",
				DefaultVersion: "v0.0.0",
			},
			mockTags:      []string{"v1.0.0-alpha.1", "v1.0.0-beta.1", "v1.0.0-rc.1"},
			expectedTag:   "v1.0.0-rc.1",
			expectedFound: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock version of run that doesn't call git
			result := runWithMockTags(tt.config, tt.mockTags)

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

			if result.Tag != tt.expectedTag {
				t.Errorf("Tag = %q, want %q", result.Tag, tt.expectedTag)
			}
			if result.Found != tt.expectedFound {
				t.Errorf("Found = %v, want %v", result.Found, tt.expectedFound)
			}
			if !result.Success {
				t.Error("Expected Success = true")
			}

			// Verify version parsing
			if tt.expectedTag != "" {
				versionWithoutPrefix := tt.expectedTag
				if tt.config.Prefix != "" {
					versionWithoutPrefix = tt.expectedTag[len(tt.config.Prefix):]
				}
				
				semver, err := versionkit.ParseSemanticVersion(versionWithoutPrefix)
				if err != nil {
					t.Errorf("Failed to parse expected version %q: %v", versionWithoutPrefix, err)
					return
				}

				if result.Major != int(semver.MajorVersion) {
					t.Errorf("Major = %d, want %d", result.Major, semver.MajorVersion)
				}
				if result.Minor != int(semver.MinorVersion) {
					t.Errorf("Minor = %d, want %d", result.Minor, semver.MinorVersion)
				}
				if result.Patch != int(semver.PatchVersion) {
					t.Errorf("Patch = %d, want %d", result.Patch, semver.PatchVersion)
				}
				if result.Prerelease != semver.PreReleaseVersion {
					t.Errorf("Prerelease = %q, want %q", result.Prerelease, semver.PreReleaseVersion)
				}
				if result.Build != semver.BuildMetadata {
					t.Errorf("Build = %q, want %q", result.Build, semver.BuildMetadata)
				}
			}
		})
	}
}

// runWithMockTags is a test version of run that uses mock tags instead of calling git
func runWithMockTags(config *Config, mockTags []string) *Result {
	result := &Result{Success: false}

	// Find the latest semver tag from mock tags
	latestTag := findLatestSemverTag(mockTags, config.Prefix)

	if latestTag == "" {
		// No tags found, use default
		result.Found = false
		result.Tag = config.DefaultVersion
	} else {
		result.Found = true
		result.Tag = latestTag
	}

	// Parse version components using versionkit
	versionWithoutPrefix := strings.TrimPrefix(result.Tag, config.Prefix)
	
	semver, err := versionkit.ParseSemanticVersion(versionWithoutPrefix)
	if err != nil {
		result.Error = err
		return result
	}

	result.Version = versionWithoutPrefix
	result.Major = int(semver.MajorVersion)
	result.Minor = int(semver.MinorVersion)
	result.Patch = int(semver.PatchVersion)
	result.Prerelease = semver.PreReleaseVersion
	result.Build = semver.BuildMetadata
	result.Success = true
	return result
}

func TestTagWithVersionSorting(t *testing.T) {
	// Test the sorting logic specifically
	tags := []TagWithVersion{
		{Tag: "v1.0.0", Version: mustParseVersion("1.0.0")},
		{Tag: "v2.0.0", Version: mustParseVersion("2.0.0")},
		{Tag: "v1.1.0", Version: mustParseVersion("1.1.0")},
		{Tag: "v1.0.1", Version: mustParseVersion("1.0.1")},
		{Tag: "v1.0.0-alpha.1", Version: mustParseVersion("1.0.0-alpha.1")},
		{Tag: "v1.0.0-beta.1", Version: mustParseVersion("1.0.0-beta.1")},
	}

	// Sort using the same logic as findLatestSemverTag
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Version.Compare(tags[j].Version) < 0
	})

	expected := []string{
		"v1.0.0-alpha.1",
		"v1.0.0-beta.1", 
		"v1.0.0",
		"v1.0.1",
		"v1.1.0",
		"v2.0.0",
	}

	for i, tag := range tags {
		if tag.Tag != expected[i] {
			t.Errorf("Position %d: got %q, want %q", i, tag.Tag, expected[i])
		}
	}
}

func mustParseVersion(v string) versionkit.SemanticVersion {
	parsed, err := versionkit.ParseSemanticVersion(v)
	if err != nil {
		panic(err)
	}
	return *parsed
}

