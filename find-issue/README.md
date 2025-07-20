# Find GitHub Issue Action

A Go-based GitHub Action that finds an existing open issue by title.

## Local Testing

### Build and run locally:

```bash
# Build the binary
go build -o find-issue main.go

# Run the action with command-line arguments
./find-issue owner/repo "Issue Title" "your_github_token"
```

### Example:
```bash
./find-issue half-ogre-games/rpgish-claude "🚨 Terraform Drift Detected in shared Environment" "ghp_xxxxxxxxxxxx"
```

### Using Makefile:
```bash
# Build and run with make
make run ARGS='owner/repo "Issue Title" "your_github_token"'

# Or run tests
make test
```

## GitHub Actions Usage

The action is configured in `action.yml` to build and run the Go binary directly:

```yaml
- uses: ./.github/actions/find-issue
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    issue-title: "🚨 Terraform Drift Detected in shared Environment"
```

## Inputs

- `github-token`: GitHub token for API access (required)
- `issue-title`: Title to search for (required)

## Outputs

- `issue-number`: Issue number if found, empty if not found
- `issue-exists`: Whether an open issue with the title exists (true/false)
