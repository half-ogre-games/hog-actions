package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
	"github.com/half-ogre/go-kit/versionkit"
)

// Config holds the configuration for the get-latest-tag action
type Config struct {
	Prefix         string
	DefaultVersion string
}

// Result holds the result of the get-latest-tag action
type Result struct {
	Tag        string
	Version    string
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Found      bool
	Success    bool
	Error      error
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
	if result.Found {
		actionskit.Info(fmt.Sprintf("Found latest tag: %s", result.Tag))
	} else {
		actionskit.Info(fmt.Sprintf("No tags found, using default: %s", result.Tag))
	}

	// Set outputs for GitHub Actions
	err = actionskit.SetOutput("tag", result.Tag)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set tag output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("version", result.Version)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set version output: %v", err))
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

	err = actionskit.SetOutput("found", fmt.Sprintf("%t", result.Found))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set found output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("prerelease", result.Prerelease)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set prerelease output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("build", result.Build)
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set build output: %v", err))
		os.Exit(1)
	}
}

// getConfigFromEnvironment reads configuration from environment variables and GitHub Actions inputs
func getConfigFromEnvironment() (*Config, error) {
	prefix := actionskit.GetInput("prefix")
	if prefix == "" {
		prefix = "v"
	}

	defaultVersion := actionskit.GetInput("default-version")
	if defaultVersion == "" {
		defaultVersion = "v0.0.0"
	}

	return &Config{
		Prefix:         prefix,
		DefaultVersion: defaultVersion,
	}, nil
}

// run executes the get-latest-tag action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Get all tags from the local git repository
	tags, err := getTags()
	if err != nil {
		result.Error = fmt.Errorf("error getting tags: %v", err)
		return result
	}

	// Filter and find the latest semver tag
	latestTag := findLatestSemverTag(tags, config.Prefix)

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
		result.Error = fmt.Errorf("error parsing version %s: %v", result.Tag, err)
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

func getTags() ([]string, error) {
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

func findLatestSemverTag(tags []string, prefix string) string {
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
		return ""
	}

	// Sort by semantic version using versionkit's Compare method
	sort.Slice(validVersions, func(i, j int) bool {
		return validVersions[i].Version.Compare(validVersions[j].Version) < 0
	})

	// Return the latest (last in sorted order)
	return validVersions[len(validVersions)-1].Tag
}

type TagWithVersion struct {
	Tag     string
	Version versionkit.SemanticVersion
}

