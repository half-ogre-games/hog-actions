package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/half-ogre/go-kit/actionskit"
)

type CommentRequest struct {
	Body string `json:"body"`
}

type CloseIssueRequest struct {
	State       string `json:"state"`
	StateReason string `json:"state_reason"`
}

type Issue struct {
	Number int    `json:"number"`
	State  string `json:"state"`
}

// Config holds the configuration for the close-issue action
type Config struct {
	Repository  string
	IssueNumber string
	CommentBody string
	StateReason string
	Token       string
}

// Result holds the result of the close-issue action
type Result struct {
	CommentID int
	Success   bool
	Error     error
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

	actionskit.Info(fmt.Sprintf("Issue #%s has been closed", config.IssueNumber))
}

// getConfigFromEnvironment reads configuration from environment variables and GitHub Actions inputs
func getConfigFromEnvironment() (*Config, error) {
	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	issueNumber := actionskit.GetInput("issue-number")
	if issueNumber == "" {
		return nil, fmt.Errorf("issue-number input is required")
	}

	commentBody := actionskit.GetInput("comment-body")
	stateReason := actionskit.GetInput("state-reason")
	if stateReason == "" {
		stateReason = "closed"
	}

	token := actionskit.GetInput("github-token")
	if token == "" {
		return nil, fmt.Errorf("github-token input is required")
	}

	return &Config{
		Repository:  repository,
		IssueNumber: issueNumber,
		CommentBody: commentBody,
		StateReason: stateReason,
		Token:       token,
	}, nil
}

// run executes the close-issue action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Add comment if provided
	if config.CommentBody != "" {
		actionskit.Info(fmt.Sprintf("Adding comment before closing issue #%s", config.IssueNumber))
		commentID, err := addComment(config.Repository, config.IssueNumber, config.CommentBody, config.Token)
		if err != nil {
			result.Error = fmt.Errorf("error adding comment: %v", err)
			return result
		}
		result.CommentID = commentID
		actionskit.Info("Comment added successfully")
	}

	// Close the issue
	actionskit.Info(fmt.Sprintf("Closing issue #%s", config.IssueNumber))
	err := closeIssue(config.Repository, config.IssueNumber, config.StateReason, config.Token)
	if err != nil {
		result.Error = fmt.Errorf("error closing issue: %v", err)
		return result
	}

	result.Success = true
	return result
}

func addComment(repository, issueNumber, body, token string) (int, error) {
	// Build API URL
	apiBase := os.Getenv("GITHUB_API_URL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/issues/%s/comments", apiBase, repository, issueNumber)

	// Create request body
	request := CommentRequest{
		Body: body,
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
	var comment struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body_bytes, &comment); err != nil {
		return 0, err
	}

	return comment.ID, nil
}

func closeIssue(repository, issueNumber, stateReason, token string) error {
	// Build API URL
	apiBase := os.Getenv("GITHUB_API_URL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/issues/%s", apiBase, repository, issueNumber)

	// Create request body
	request := CloseIssueRequest{
		State:       "closed",
		StateReason: stateReason,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()

	// Read response
	body_bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", 
			resp.StatusCode, string(body_bytes))
	}

	return nil
}