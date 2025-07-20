# Create GitHub Issue Action

A Go-based GitHub Action that creates a new GitHub issue with specified labels. Uses the actionskit library from go-kit for GitHub Actions integration.

## Local Testing

### Build and run locally:

```bash
# Build the binary
go build -o create-issue main.go

# Set required environment variables
export GITHUB_REPOSITORY="owner/repo"
export INPUT_ISSUE_TITLE="Issue Title"
export INPUT_ISSUE_BODY="Issue body content"
export INPUT_ISSUE_LABEL="primary-label"
export INPUT_ADDITIONAL_LABELS="additional,labels"
export INPUT_GITHUB_TOKEN="github_token"

# Run the action
./create-issue
```

### Example:
```bash
export GITHUB_REPOSITORY="half-ogre-games/rpgish-claude"
export INPUT_ISSUE_TITLE="Test Issue"
export INPUT_ISSUE_BODY="This is a test issue body"
export INPUT_ISSUE_LABEL="bug"
export INPUT_ADDITIONAL_LABELS="priority-high,needs-investigation"
export INPUT_GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
./create-issue
```

## GitHub Actions Usage

The action is configured in `action.yml` to build and run the Go binary directly:

```yaml
- uses: ./.github/actions/create-issue
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    issue-title: "Issue Title"
    issue-body: "Issue body content"
    issue-label: "primary-label"
    additional-labels: "label1,label2"
```

## Inputs

- `github-token`: GitHub token for API access (required)
- `issue-title`: Title for the issue (required)
- `issue-body`: Body content for the issue (required)
- `issue-label`: Primary label to apply to the issue (required)
- `additional-labels`: Additional labels to apply (comma-separated, optional)

## Outputs

- `issue-number`: Number of the created issue