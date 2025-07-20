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

type Comment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// Config holds the configuration for the comment-issue action
type Config struct {
	Repository  string
	IssueNumber string
	CommentBody string
	Token       string
}

// Result holds the result of the comment-issue action
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

	actionskit.Info(fmt.Sprintf("Comment added successfully (ID: %d)", result.CommentID))

	// Set output for GitHub Actions
	err = actionskit.SetOutput("comment-id", fmt.Sprintf("%d", result.CommentID))
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

	issueNumber := actionskit.GetInput("issue-number")
	if issueNumber == "" {
		return nil, fmt.Errorf("issue-number input is required")
	}

	commentBody := actionskit.GetInput("comment-body")
	if commentBody == "" {
		return nil, fmt.Errorf("comment-body input is required")
	}

	token := actionskit.GetInput("github-token")
	if token == "" {
		return nil, fmt.Errorf("github-token input is required")
	}

	return &Config{
		Repository:  repository,
		IssueNumber: issueNumber,
		CommentBody: commentBody,
		Token:       token,
	}, nil
}

// run executes the comment-issue action with the given configuration
func run(config *Config) *Result {
	result := &Result{Success: false}

	// Add the comment
	commentID, err := addComment(config.Repository, config.IssueNumber, config.CommentBody, config.Token)
	if err != nil {
		result.Error = fmt.Errorf("error adding comment: %v", err)
		return result
	}

	result.CommentID = commentID
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
	var comment Comment
	if err := json.Unmarshal(body_bytes, &comment); err != nil {
		return 0, err
	}

	return comment.ID, nil
}