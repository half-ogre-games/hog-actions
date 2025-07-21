package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
	"github.com/half-ogre-games/hog-actions/internal/semveractions"
)

// Config holds the configuration for the tag-and-create-semver-release action
type Config struct {
	Branch           string
	Commit           string
	IncrementMajor   bool
	IncrementMinor   bool
	Prefix           string
	DefaultVersion   string
	GitHubToken      string
	DefaultBranch    string // Will be populated from GitHub context
}

// Result holds the result of the tag-and-create-semver-release action
type Result struct {
	PreviousVersion string
	NewVersion      string
	IncrementType   string
	ReleaseURL      string
	TargetCommit    string
	Success         bool
	Error           error
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
	actionskit.Info(fmt.Sprintf("✅ Created tag %s for commit %s", result.NewVersion, result.TargetCommit))
	if result.ReleaseURL != "" {
		actionskit.Info(fmt.Sprintf("✅ Created release: %s", result.ReleaseURL))
	}

	// Set outputs for GitHub Actions
	if err := setOutputs(result); err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set outputs: %v", err))
		os.Exit(1)
	}
}

// getConfigFromEnvironment reads configuration from environment variables and GitHub Actions inputs
func getConfigFromEnvironment() (*Config, error) {
	branch := actionskit.GetInput("branch")
	if branch == "" {
		// Get default branch from GitHub context
		branch = os.Getenv("GITHUB_REF_NAME")
		if branch == "" {
			branch = "main" // fallback
		}
	}

	commit := actionskit.GetInput("commit")
	if commit == "" {
		commit = "HEAD"
	}

	incrementMajorStr := actionskit.GetInput("increment-major")
	incrementMajor := incrementMajorStr == "true"

	incrementMinorStr := actionskit.GetInput("increment-minor")
	incrementMinor := incrementMinorStr == "true"

	// Validate increment flags
	if err := semveractions.ValidateIncrementFlags(incrementMajor, incrementMinor); err != nil {
		return nil, err
	}

	prefix := actionskit.GetInput("prefix")
	if prefix == "" {
		prefix = "v"
	}

	defaultVersion := actionskit.GetInput("default-version")
	if defaultVersion == "" {
		defaultVersion = "v0.1.0"
	}

	githubToken, err := actionskit.GetInputRequired("github-token")
	if err != nil {
		return nil, err
	}

	return &Config{
		Branch:         branch,
		Commit:         commit,
		IncrementMajor: incrementMajor,
		IncrementMinor: incrementMinor,
		Prefix:         prefix,
		DefaultVersion: defaultVersion,
		GitHubToken:    githubToken,
		DefaultBranch:  branch,
	}, nil
}

// run executes the tag-and-create-semver-release action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Step 1: Get target commit SHA
	targetCommit, err := getTargetCommitSHA(config.Commit)
	if err != nil {
		result.Error = fmt.Errorf("error getting target commit: %v", err)
		return result
	}
	result.TargetCommit = targetCommit

	// Step 2: Get latest version tag
	tags, err := semveractions.GetAllTags()
	if err != nil {
		result.Error = fmt.Errorf("error getting tags: %v", err)
		return result
	}

	latestTag, found, err := semveractions.FindLatestSemverTag(tags, config.Prefix)
	if err != nil {
		result.Error = fmt.Errorf("error finding latest tag: %v", err)
		return result
	}

	var currentVersion string
	if !found {
		currentVersion = config.DefaultVersion
		result.PreviousVersion = "none"
		actionskit.Info(fmt.Sprintf("No tags found, using default: %s", currentVersion))
	} else {
		currentVersion = latestTag
		result.PreviousVersion = latestTag
		actionskit.Info(fmt.Sprintf("Found latest tag: %s", latestTag))
	}

	// Step 3: Calculate new version
	semver, _, err := semveractions.ParseVersionWithPrefix(currentVersion, config.Prefix)
	if err != nil {
		result.Error = fmt.Errorf("error parsing current version: %v", err)
		return result
	}

	newSemver, incrementType, err := semveractions.IncrementVersion(semver, config.IncrementMajor, config.IncrementMinor)
	if err != nil {
		result.Error = fmt.Errorf("error incrementing version: %v", err)
		return result
	}

	newVersionTag := semveractions.FormatVersionWithPrefix(newSemver, config.Prefix)
	result.NewVersion = newVersionTag
	result.IncrementType = incrementType

	actionskit.Info(fmt.Sprintf("Next version: %s (%s increment)", newVersionTag, incrementType))

	// Step 4: Check if tag already exists
	if tagExists(newVersionTag) {
		result.Error = fmt.Errorf("tag %s already exists", newVersionTag)
		return result
	}

	// Step 5: Create and push tags
	if err := createAndPushTags(newVersionTag, targetCommit, int(newSemver.MajorVersion)); err != nil {
		result.Error = fmt.Errorf("error creating tags: %v", err)
		return result
	}

	// Step 6: Create GitHub release
	releaseURL, err := createGitHubRelease(config, result)
	if err != nil {
		result.Error = fmt.Errorf("error creating release: %v", err)
		return result
	}

	result.ReleaseURL = releaseURL
	result.Success = true
	return result
}

// getTargetCommitSHA resolves the target commit SHA
func getTargetCommitSHA(commit string) (string, error) {
	if commit == "HEAD" || commit == "" {
		cmd := exec.Command("git", "rev-parse", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get HEAD commit: %v", err)
		}
		return strings.TrimSpace(string(output)), nil
	}

	// Verify the commit exists
	cmd := exec.Command("git", "cat-file", "-e", commit+"^{commit}")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("commit %s does not exist", commit)
	}

	return commit, nil
}

// tagExists checks if a git tag already exists
func tagExists(tag string) bool {
	if tag == "" {
		return false
	}
	cmd := exec.Command("git", "tag", "-l", tag)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == tag
}

// createAndPushTags creates and pushes the semver tag and major version tag
func createAndPushTags(newVersionTag, targetCommit string, majorVersion int) error {
	// Configure git user
	if err := exec.Command("git", "config", "user.name", "github-actions[bot]").Run(); err != nil {
		return fmt.Errorf("failed to configure git user name: %v", err)
	}
	if err := exec.Command("git", "config", "user.email", "github-actions[bot]@users.noreply.github.com").Run(); err != nil {
		return fmt.Errorf("failed to configure git user email: %v", err)
	}

	// Create and push the semver tag
	cmd := exec.Command("git", "tag", "-a", newVersionTag, targetCommit, "-m", fmt.Sprintf("Release %s", newVersionTag))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tag %s: %v", newVersionTag, err)
	}

	cmd = exec.Command("git", "push", "origin", newVersionTag)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push tag %s: %v", newVersionTag, err)
	}

	actionskit.Info(fmt.Sprintf("✅ Created and pushed tag %s for commit %s", newVersionTag, targetCommit))

	// Create major version tag if major version is 1 or more
	if majorVersion >= 1 {
		majorTag := fmt.Sprintf("v%d", majorVersion)
		actionskit.Info(fmt.Sprintf("Creating major version tag: %s", majorTag))

		// Delete existing major tag if it exists (force update)
		cmd = exec.Command("git", "tag", "-d", majorTag)
		cmd.Run() // Ignore error if tag doesn't exist locally

		cmd = exec.Command("git", "push", "origin", ":refs/tags/"+majorTag)
		cmd.Run() // Ignore error if tag doesn't exist remotely

		// Create and push new major tag
		cmd = exec.Command("git", "tag", "-a", majorTag, targetCommit, "-m", fmt.Sprintf("Major version %s (latest: %s)", majorTag, newVersionTag))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create major tag %s: %v", majorTag, err)
		}

		cmd = exec.Command("git", "push", "origin", majorTag)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push major tag %s: %v", majorTag, err)
		}

		actionskit.Info(fmt.Sprintf("✅ Created and pushed major version tag %s", majorTag))
	} else {
		actionskit.Info("Major version is 0, skipping major version tag creation")
	}

	return nil
}

// createGitHubRelease creates a GitHub release with auto-generated notes
func createGitHubRelease(config *Config, result *Result) (string, error) {
	// Create release notes file with metadata
	releaseNotes := fmt.Sprintf(`## What's Changed

This release was created from commit %s on branch %s.

### Version Details
- **Previous version**: %s
- **New version**: %s
- **Version type**: %s release
`, result.TargetCommit, config.Branch, result.PreviousVersion, result.NewVersion, result.IncrementType)

	// Write release notes to temporary file
	releaseNotesFile := "release_notes.md"
	if err := os.WriteFile(releaseNotesFile, []byte(releaseNotes), 0644); err != nil {
		return "", fmt.Errorf("failed to write release notes file: %v", err)
	}
	defer os.Remove(releaseNotesFile)

	// Create release with auto-generated notes
	cmd := exec.Command("gh", "release", "create", result.NewVersion,
		"--title", result.NewVersion,
		"--notes-file", releaseNotesFile,
		"--generate-notes",
		"--latest",
		"--target", result.TargetCommit)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create release: %v\nOutput: %s", err, string(output))
	}

	// Extract release URL from output
	releaseURL := strings.TrimSpace(string(output))
	actionskit.Info(fmt.Sprintf("✅ Created release %s with auto-generated notes", result.NewVersion))

	return releaseURL, nil
}

// setOutputs sets the GitHub Actions outputs
func setOutputs(result *Result) error {
	outputs := map[string]string{
		"previous-version": result.PreviousVersion,
		"new-version":      result.NewVersion,
		"increment-type":   result.IncrementType,
		"release-url":      result.ReleaseURL,
		"target-commit":    result.TargetCommit,
	}

	for name, value := range outputs {
		if err := actionskit.SetOutput(name, value); err != nil {
			return fmt.Errorf("failed to set %s output: %v", name, err)
		}
	}

	return nil
}