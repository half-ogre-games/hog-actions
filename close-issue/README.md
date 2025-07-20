# Close GitHub Issue Action

A Go-based GitHub Action that closes a GitHub issue with an optional comment.

## Local Testing

### Build and run locally:

```bash
# Build the binary
go build -o close-issue main.go

# Run the action with command-line arguments
./close-issue owner/repo issue-number "optional comment" "state-reason" "github_token"
```

### Examples:
```bash
# Close with comment
./close-issue half-ogre-games/rpgish-claude 123 "Fixed in PR #456" "completed" "ghp_xxxxxxxxxxxx"

# Close without comment (use empty string)
./close-issue half-ogre-games/rpgish-claude 123 "" "not_planned" "ghp_xxxxxxxxxxxx"
```

## GitHub Actions Usage

The action is configured in `action.yml` to build and run the Go binary directly:

```yaml
- uses: ./.github/actions/close-issue
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    issue-number: 123
    comment-body: "Optional comment before closing"
    state-reason: "completed"
```

## Inputs

- `github-token`: GitHub token for API access (required)
- `issue-number`: Issue number to close (required)
- `comment-body`: Optional comment to add before closing (optional, default: empty)
- `state-reason`: Reason for closing - "completed", "not_planned", or "closed" (optional, default: "closed")

## Outputs

None - the action will exit with an error code if the close operation fails.