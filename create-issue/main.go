package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/half-ogre/go-kit/actionskit"
)

type CreateIssueRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

// Config holds the configuration for the create-issue action
type Config struct {
	Repository       string
	Title            string
	Body             string
	PrimaryLabel     string
	AdditionalLabels string
	Token            string
}

// Result holds the result of the create-issue action
type Result struct {
	IssueNumber int
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

	actionskit.Info(fmt.Sprintf("Created new issue #%d", result.IssueNumber))

	// Set output for GitHub Actions
	err = actionskit.SetOutput("issue-number", fmt.Sprintf("%d", result.IssueNumber))
	if err != nil {
		actionskit.Error(fmt.Sprintf("Failed to set output: %v", err))
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

	body := actionskit.GetInput("issue-body")
	primaryLabel := actionskit.GetInput("issue-label")
	if primaryLabel == "" {
		return nil, fmt.Errorf("issue-label input is required")
	}

	additionalLabels := actionskit.GetInput("additional-labels")
	token := actionskit.GetInput("github-token")
	if token == "" {
		return nil, fmt.Errorf("github-token input is required")
	}

	return &Config{
		Repository:       repository,
		Title:            title,
		Body:             body,
		PrimaryLabel:     primaryLabel,
		AdditionalLabels: additionalLabels,
		Token:            token,
	}, nil
}

// run executes the create-issue action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Build labels array
	labels := buildLabels(config.PrimaryLabel, config.AdditionalLabels)

	// Create the issue
	issueNumber, err := createIssue(config.Repository, config.Title, config.Body, labels, config.Token)
	if err != nil {
		result.Error = fmt.Errorf("error creating issue: %v", err)
		return result
	}

	result.IssueNumber = issueNumber
	result.Success = true
	return result
}

// buildLabels constructs the labels array from primary and additional labels
func buildLabels(primaryLabel, additionalLabels string) []string {
	var labels []string
	labels = append(labels, primaryLabel)
	
	if additionalLabels != "" {
		additional := strings.Split(additionalLabels, ",")
		for _, label := range additional {
			trimmed := strings.TrimSpace(label)
			if trimmed != "" {
				labels = append(labels, trimmed)
			}
		}
	}
	
	return labels
}

func createIssue(repository, title, body string, labels []string, token string) (int, error) {
	// Build API URL
	apiBase := os.Getenv("GITHUB_API_URL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/issues", apiBase, repository)

	// Create request body
	request := CreateIssueRequest{
		Title:  title,
		Body:   body,
		Labels: labels,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read response
	body_bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Check status
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("API request failed with status %d: %s",
			resp.StatusCode, string(body_bytes))
	}

	// Parse response
	var issue Issue
	if err := json.Unmarshal(body_bytes, &issue); err != nil {
		return 0, err
	}

	return issue.Number, nil
}