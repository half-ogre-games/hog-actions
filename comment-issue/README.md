# Comment on GitHub Issue Action

A Go-based GitHub Action that adds a comment to an existing GitHub issue.

## Local Testing

### Build and run locally:

```bash
# Build the binary
go build -o comment-issue main.go

# Run the action with command-line arguments
./comment-issue owner/repo issue-number "Comment body content" "github_token"
```

### Example:
```bash
./comment-issue half-ogre-games/rpgish-claude 123 "This is a test comment" "ghp_xxxxxxxxxxxx"
```

## GitHub Actions Usage

The action is configured in `action.yml` to build and run the Go binary directly:

```yaml
- uses: ./.github/actions/comment-issue
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    issue-number: 123
    comment-body: "Comment content"
```

## Inputs

- `github-token`: GitHub token for API access (required)
- `issue-number`: Issue number to comment on (required)
- `comment-body`: Comment content to add (required)

## Outputs

None - the action will exit with an error code if the comment fails to be added.