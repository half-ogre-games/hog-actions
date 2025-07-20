package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
)

type Issue struct {
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
}

// Config holds the configuration for the find-issue action
type Config struct {
	Repository string
	Title      string
	Token      string
}

// Result holds the result of the find-issue action
type Result struct {
	IssueNumber int
	IssueExists bool
	Success     bool
	Error       error
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
	if result.IssueExists {
		actionskit.Info(fmt.Sprintf("Found existing issue #%d", result.IssueNumber))
	} else {
		actionskit.Info("No existing issue found")
	}

	// Set outputs for GitHub Actions
	err = actionskit.SetOutput("issue-number", fmt.Sprintf("%d", result.IssueNumber))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set issue-number output: %v", err))
		os.Exit(1)
	}

	err = actionskit.SetOutput("issue-exists", fmt.Sprintf("%t", result.IssueExists))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set issue-exists output: %v", err))
		os.Exit(1)
	}
}

// getConfigFromEnvironment reads configuration from environment variables and GitHub Actions inputs
func getConfigFromEnvironment() (*Config, error) {
	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	title := actionskit.GetInput("issue-title")
	if title == "" {
		return nil, fmt.Errorf("issue-title input is required")
	}

	token := actionskit.GetInput("github-token")
	if token == "" {
		return nil, fmt.Errorf("github-token input is required")
	}

	return &Config{
		Repository: repository,
		Title:      title,
		Token:      token,
	}, nil
}

// run executes the find-issue action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Search for open issues with the title
	issues, err := findIssues(config.Repository, config.Title, config.Token)
	if err != nil {
		result.Error = fmt.Errorf("error finding issues: %v", err)
		return result
	}

	// Process results
	if len(issues) > 0 {
		issue := issues[0]
		result.IssueNumber = issue.Number
		result.IssueExists = true
	} else {
		result.IssueNumber = 0
		result.IssueExists = false
	}

	result.Success = true
	return result
}

func findIssues(repository, title, token string) ([]Issue, error) {
	// Build API URL - get all open issues and filter by title locally
	apiBase := os.Getenv("GITHUB_API_URL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/issues?state=open", 
		apiBase, repository)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", 
			resp.StatusCode, string(body))
	}

	// Parse response
	var issues []Issue
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&issues); err != nil {
		return nil, err
	}

	// Filter for open issues with exact title match
	var filtered []Issue
	for _, issue := range issues {
		if strings.EqualFold(issue.State, "open") && strings.EqualFold(issue.Title, title) {
			filtered = append(filtered, issue)
		}
	}

	return filtered, nil
}