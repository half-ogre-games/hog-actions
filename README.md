# Half-Ogre Games (HOG) GitHub Actions

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Actions](https://img.shields.io/badge/GitHub-Actions-blue.svg)](https://github.com/features/actions)

A collection of reusable GitHub Actions for Half-Ogre Games (HOG) repositories.

> **Note:** This repository is public and primarily designed for internal HOG development workflows. However, pull requests and issues are welcome.

## Actions

- **create-issue** - Create GitHub issues with standardized formatting and labels
- **find-issue** - Search for existing open issues by title to prevent duplicates  
- **close-issue** - Close issues with optional comments and proper state reasons
- **comment-issue** - Add automated comments to existing issues

## Use

### Using Actions in Your Repository

Reference actions from this repository using the standard GitHub Actions syntax:

```yaml
- name: Create approval issue
  uses: half-ogre-games/hog-actions/create-issue@main
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    issue-title: "Deployment Approval Required"
    issue-label: "deployment-approval"
    issue-body: |
      Please review and approve this deployment.
```

## Reference

### create-issue

Creates a new GitHub issue with standardized formatting.

**Inputs:**
- `github-token` (required) - GitHub token for API access
- `issue-title` (required) - Title for the issue
- `issue-label` (required) - Primary label to apply
- `issue-body` (required) - Issue body content
- `additional-labels` (optional) - Comma-separated additional labels

**Outputs:**
- `issue-number` - Number of the created issue

### find-issue

Searches for existing open issues by title.

**Inputs:**
- `github-token` (required) - GitHub token for API access
- `issue-title` (required) - Title to search for

**Outputs:**
- `issue-number` - Issue number if found
- `issue-exists` - Boolean indicating if issue exists

### close-issue

Closes a GitHub issue with optional comment.

**Inputs:**
- `github-token` (required) - GitHub token for API access
- `issue-number` (required) - Issue number to close
- `comment-body` (optional) - Comment to add before closing
- `state-reason` (optional) - Reason for closing (completed, not_planned, closed)

### comment-issue

Adds a comment to an existing GitHub issue.

**Inputs:**
- `github-token` (required) - GitHub token for API access  
- `issue-number` (required) - Issue number to comment on
- `comment-body` (required) - Comment content

**Outputs:**
- `comment-id` - ID of the created comment

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

**No Support Guarantee:** This repository is provided as-is without any warranty or support guarantee, as outlined in the [LICENSE](LICENSE.md). Half-Ogre Games will review issues and pull requests as capacity allows, but response times are not guaranteed.

For questions, issues, or feature requests:

1. Check existing [Issues](../../issues)
2. Create a new issue with detailed description
3. Understand that review and response depend on team availability
