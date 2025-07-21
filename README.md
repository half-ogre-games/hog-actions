# Half-Ogre Games (HOG) GitHub Actions

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Actions](https://img.shields.io/badge/GitHub-Actions-blue.svg)](https://github.com/features/actions)

A collection of reusable GitHub Actions for Half-Ogre Games (HOG) repositories.

> **Note:** This repository is public and primarily designed for internal HOG development workflows. However, pull requests and issues are welcome.

## Actions

| Action | Description | Key Inputs | Key Outputs |
|--------|-------------|------------|-------------|
| [create-issue](./create-issue) | Create GitHub issues with standardized formatting and labels | `issue-title`, `issue-label`, `github-token` | `issue-number` |
| [find-issue](./find-issue) | Search for existing open issues by title to prevent duplicates | `issue-title`, `github-token` | `issue-number`, `issue-exists` |
| [close-issue](./close-issue) | Close issues with optional comments and proper state reasons | `issue-number`, `github-token`, `comment-body` (optional) | `comment-id` |
| [comment-issue](./comment-issue) | Add automated comments to existing issues | `issue-number`, `comment-body`, `github-token` | `comment-id` |
| [get-latest-semver-tag](./get-latest-semver-tag) | Get the latest semantic version tag from the current repository (supports pre-release and build metadata) | `prefix` (optional), `default-version` (optional) | `tag`, `version`, `major`, `minor`, `patch`, `prerelease`, `build`, `found` |
| [get-next-semver](./get-next-semver) | Calculate the next semantic version based on increment type | `current-version`, `increment-major` (optional), `increment-minor` (optional), `prefix` (optional) | `version`, `version-core`, `major`, `minor`, `patch`, `increment-type` |

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

For detailed documentation on each action, click the action name in the table above to view its individual README.

## Versioning

This repository follows [Semantic Versioning (SemVer)](https://semver.org/) for all releases:

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions  
- **PATCH** version for backwards-compatible bug fixes

Each release creates two types of tags:

1. **Full semantic version** (e.g., `v1.2.3`)
2. **Major version tag** (e.g., `v1`) - automatically updated to point to the latest release within that major version

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

**No Support Guarantee:** This repository is provided as-is without any warranty or support guarantee, as outlined in the [LICENSE](LICENSE.md). Half-Ogre Games will review issues and pull requests as capacity allows, but response times are not guaranteed.

For questions, issues, or feature requests:

1. Check existing [Issues](../../issues)
2. Create a new issue with detailed description
3. Understand that review and response depend on team availability
