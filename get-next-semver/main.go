package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
	"github.com/half-ogre/go-kit/versionkit"
)

// Config holds the configuration for the get-next-semver action
type Config struct {
	CurrentVersion   string
	IncrementMajor   bool
	IncrementMinor   bool
	Prefix          string
}

// Result holds the result of the get-next-semver action
type Result struct {
	Version       string
	VersionCore   string
	Major         int
	Minor         int
	Patch         int
	IncrementType string
	Success       bool
	Error         error
}

func main() {
	config, err := getConfigFromEnvironment()
	if err != nil {
		actionskit.Error(err.Error())
		os.Exit(1)
	}

	result := run(config)
	if result.Error != nil {
		actionskit.Error(result.Error.Error())
		os.Exit(1)
	}

	// Output results
	actionskit.Info(fmt.Sprintf("Next version: %s (%s increment)", result.Version, result.IncrementType))

	// Set outputs for GitHub Actions
	err = actionskit.SetOutput("version", result.Version)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set version output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("version-core", result.VersionCore)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set version-core output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("major", fmt.Sprintf("%d", result.Major))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set major output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("minor", fmt.Sprintf("%d", result.Minor))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set minor output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("patch", fmt.Sprintf("%d", result.Patch))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set patch output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("increment-type", result.IncrementType)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set increment-type output: %v", err))
		os.Exit(1)
	}
}

// getConfigFromEnvironment reads configuration from environment variables and GitHub Actions inputs
func getConfigFromEnvironment() (*Config, error) {
	currentVersion := actionskit.GetInput("current-version")
	if currentVersion == "" {
		return nil, fmt.Errorf("current-version input is required")
	}

	incrementMajor := actionskit.GetInput("increment-major") == "true"
	incrementMinor := actionskit.GetInput("increment-minor") == "true"

	// Validate that only one increment type is specified
	if incrementMajor && incrementMinor {
		return nil, fmt.Errorf("cannot increment both major and minor versions simultaneously")
	}

	prefix := actionskit.GetInput("prefix")
	// Check if the input was explicitly provided (even if empty)
	_, prefixExplicitlySet := os.LookupEnv("INPUT_PREFIX")
	if prefix == "" && !prefixExplicitlySet {
		prefix = "v"
	}

	return &Config{
		CurrentVersion: currentVersion,
		IncrementMajor: incrementMajor,
		IncrementMinor: incrementMinor,
		Prefix:        prefix,
	}, nil
}

// run executes the get-next-semver action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Remove prefix from current version for parsing
	versionWithoutPrefix := strings.TrimPrefix(config.CurrentVersion, config.Prefix)
	
	// Parse current version using versionkit
	currentSemver, err := versionkit.ParseSemanticVersion(versionWithoutPrefix)
	if err != nil {
		result.Error = fmt.Errorf("error parsing current version %s: %v", config.CurrentVersion, err)
		return result
	}

	// Calculate next version
	nextMajor := currentSemver.MajorVersion
	nextMinor := currentSemver.MinorVersion
	nextPatch := currentSemver.PatchVersion
	incrementType := "patch" // default

	if config.IncrementMajor {
		nextMajor++
		nextMinor = 0
		nextPatch = 0
		incrementType = "major"
	} else if config.IncrementMinor {
		nextMinor++
		nextPatch = 0
		incrementType = "minor"
	} else {
		nextPatch++
		incrementType = "patch"
	}

	// Create next version (always remove pre-release and build metadata for releases)
	nextVersion := versionkit.SemanticVersion{
		MajorVersion:      nextMajor,
		MinorVersion:      nextMinor,
		PatchVersion:      nextPatch,
		PreReleaseVersion: "", // Always clear for release versions
		BuildMetadata:     "", // Always clear for release versions
	}

	nextVersionCore := nextVersion.String()
	nextVersionWithPrefix := config.Prefix + nextVersionCore

	result.Version = nextVersionWithPrefix
	result.VersionCore = nextVersionCore
	result.Major = int(nextMajor)
	result.Minor = int(nextMinor)
	result.Patch = int(nextPatch)
	result.IncrementType = incrementType
	result.Success = true
	return result
}