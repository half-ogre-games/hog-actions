package semveractions

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
	"github.com/half-ogre/go-kit/versionkit"
)

// SemverConfig holds common configuration for semver operations
type SemverConfig struct {
	Prefix string
}

// SemverResult represents the result of a semver operation
type SemverResult struct {
	Tag        string // Full tag with prefix
	Version    string // Version without prefix
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Found      bool   // Whether a tag was found (for get-latest operations)
	Success    bool
	Error      error
}

// TagWithVersion pairs a git tag with its parsed semantic version
type TagWithVersion struct {
	Tag     string
	Version versionkit.SemanticVersion
}

// GetSemverPrefix reads prefix from GitHub Actions input with fallback to "v"
func GetSemverPrefix() string {
	prefix := actionskit.GetInput("prefix")
	if prefix == "" {
		prefix = "v"
	}
	return prefix
}

// ParseVersionWithPrefix parses a version string, handling prefix removal
func ParseVersionWithPrefix(versionStr, prefix string) (*versionkit.SemanticVersion, string, error) {
	// Remove prefix if present
	versionWithoutPrefix := strings.TrimPrefix(versionStr, prefix)
	
	// Parse version using versionkit
	semver, err := versionkit.ParseSemanticVersion(versionWithoutPrefix)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing version %s: %v", versionStr, err)
	}
	
	return semver, versionWithoutPrefix, nil
}

// FormatVersionWithPrefix combines version and prefix
func FormatVersionWithPrefix(version *versionkit.SemanticVersion, prefix string) string {
	versionStr := fmt.Sprintf("%d.%d.%d", version.MajorVersion, version.MinorVersion, version.PatchVersion)
	
	// Add prerelease if present
	if version.PreReleaseVersion != "" {
		versionStr += "-" + version.PreReleaseVersion
	}
	
	// Add build metadata if present
	if version.BuildMetadata != "" {
		versionStr += "+" + version.BuildMetadata
	}
	
	return prefix + versionStr
}

// CreateSemverResult creates standardized result from parsed version
func CreateSemverResult(tag, versionWithoutPrefix string, semver *versionkit.SemanticVersion, found bool) *SemverResult {
	return &SemverResult{
		Tag:        tag,
		Version:    versionWithoutPrefix,
		Major:      int(semver.MajorVersion),
		Minor:      int(semver.MinorVersion),
		Patch:      int(semver.PatchVersion),
		Prerelease: semver.PreReleaseVersion,
		Build:      semver.BuildMetadata,
		Found:      found,
		Success:    true,
	}
}

// GetAllTags retrieves all git tags from current repository
func GetAllTags() ([]string, error) {
	// Use git command to list all tags
	cmd := exec.Command("git", "tag", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git tag command: %v", err)
	}

	// Parse output into tag names
	tagOutput := strings.TrimSpace(string(output))
	if tagOutput == "" {
		// No tags found
		return []string{}, nil
	}

	tags := strings.Split(tagOutput, "\n")
	
	// Filter out empty strings
	var validTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			validTags = append(validTags, tag)
		}
	}

	return validTags, nil
}

// FilterTagsByPrefix filters tags that match the given prefix
func FilterTagsByPrefix(tags []string, prefix string) []string {
	var filteredTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			filteredTags = append(filteredTags, tag)
		}
	}
	return filteredTags
}

// FindLatestSemverTag finds the latest semantic version from a list of tags
func FindLatestSemverTag(tags []string, prefix string) (string, bool, error) {
	var validVersions []TagWithVersion
	
	for _, tag := range tags {
		// Check if tag starts with prefix
		if !strings.HasPrefix(tag, prefix) {
			continue
		}
		
		// Extract version without prefix
		versionStr := strings.TrimPrefix(tag, prefix)
		
		// Try to parse as semantic version
		semver, err := versionkit.ParseSemanticVersion(versionStr)
		if err != nil {
			continue // Skip invalid versions
		}
		
		validVersions = append(validVersions, TagWithVersion{
			Tag:     tag,
			Version: *semver,
		})
	}

	if len(validVersions) == 0 {
		return "", false, nil
	}

	// Sort by semantic version using versionkit's Compare method
	sort.Slice(validVersions, func(i, j int) bool {
		return validVersions[i].Version.Compare(validVersions[j].Version) < 0
	})

	// Return the latest (last in sorted order)
	return validVersions[len(validVersions)-1].Tag, true, nil
}

// SetSemverOutputs sets all standard semver outputs for GitHub Actions
func SetSemverOutputs(result *SemverResult) error {
	outputs := map[string]string{
		"tag":        result.Tag,
		"version":    result.Version,
		"major":      fmt.Sprintf("%d", result.Major),
		"minor":      fmt.Sprintf("%d", result.Minor),
		"patch":      fmt.Sprintf("%d", result.Patch),
		"prerelease": result.Prerelease,
		"build":      result.Build,
		"found":      fmt.Sprintf("%t", result.Found),
	}

	for name, value := range outputs {
		if err := actionskit.SetOutput(name, value); err != nil {
			return fmt.Errorf("failed to set %s output: %v", name, err)
		}
	}

	return nil
}

// SetVersionOutputs sets version component outputs for get-next-semver
func SetVersionOutputs(result *SemverResult, incrementType string) error {
	outputs := map[string]string{
		"version":        result.Tag,
		"version-core":   result.Version,
		"major":          fmt.Sprintf("%d", result.Major),
		"minor":          fmt.Sprintf("%d", result.Minor),
		"patch":          fmt.Sprintf("%d", result.Patch),
		"increment-type": incrementType,
	}

	for name, value := range outputs {
		if err := actionskit.SetOutput(name, value); err != nil {
			return fmt.Errorf("failed to set %s output: %v", name, err)
		}
	}

	return nil
}

// IncrementVersion calculates next version based on increment type
func IncrementVersion(current *versionkit.SemanticVersion, incrementMajor, incrementMinor bool) (*versionkit.SemanticVersion, string, error) {
	// Validate increment flags
	if incrementMajor && incrementMinor {
		return nil, "", fmt.Errorf("cannot increment both major and minor versions simultaneously")
	}

	// Create a copy of the current version to avoid modifying the original
	newVersion := versionkit.SemanticVersion{
		MajorVersion:      current.MajorVersion,
		MinorVersion:      current.MinorVersion,
		PatchVersion:      current.PatchVersion,
		PreReleaseVersion: "", // Always remove pre-release and build metadata
		BuildMetadata:     "",
	}

	var incrementType string

	if incrementMajor {
		newVersion.MajorVersion++
		newVersion.MinorVersion = 0
		newVersion.PatchVersion = 0
		incrementType = "major"
	} else if incrementMinor {
		newVersion.MinorVersion++
		newVersion.PatchVersion = 0
		incrementType = "minor"
	} else {
		// Default to patch increment
		newVersion.PatchVersion++
		incrementType = "patch"
	}

	return &newVersion, incrementType, nil
}

// ValidateIncrementFlags ensures only one increment type is specified
func ValidateIncrementFlags(incrementMajor, incrementMinor bool) error {
	if incrementMajor && incrementMinor {
		return fmt.Errorf("cannot increment both major and minor versions simultaneously")
	}
	return nil
}