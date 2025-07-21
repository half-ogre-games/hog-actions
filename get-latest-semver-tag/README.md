# Get Latest Semver Tag Action

A GitHub Action that retrieves the latest semantic version tag from the current repository using local git commands.

## Features

- Uses local git commands (no API tokens required)
- Filters tags by semantic version format
- Supports custom tag prefixes
- Returns parsed version components
- Handles repositories with no tags gracefully

## Usage

```yaml
- name: Get latest semver tag
  id: tag
  uses: half-ogre-games/hog-actions/get-latest-semver-tag@v1
  with:
    prefix: 'v'               # Optional: tag prefix (default: 'v')
    default-version: 'v0.0.0' # Optional: default if no tags found

- name: Use tag information
  run: |
    echo "Latest tag: ${{ steps.tag.outputs.tag }}"
    echo "Version: ${{ steps.tag.outputs.version }}"
    echo "Major: ${{ steps.tag.outputs.major }}"
    echo "Minor: ${{ steps.tag.outputs.minor }}"
    echo "Patch: ${{ steps.tag.outputs.patch }}"
    echo "Pre-release: ${{ steps.tag.outputs.prerelease }}"
    echo "Build: ${{ steps.tag.outputs.build }}"
    echo "Found: ${{ steps.tag.outputs.found }}"
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `prefix` | Tag prefix to filter by (e.g., "v" for "v1.0.0") | No | `v` |
| `default-version` | Default version to return if no tags are found | No | `v0.0.0` |

## Outputs

| Output | Description | Example |
|--------|-------------|---------|
| `tag` | The latest semver tag | `v1.2.3-alpha.1+build.456` |
| `version` | The version without prefix | `1.2.3-alpha.1+build.456` |
| `major` | The major version number | `1` |
| `minor` | The minor version number | `2` |
| `patch` | The patch version number | `3` |
| `prerelease` | The pre-release version | `alpha.1` |
| `build` | The build metadata | `build.456` |
| `found` | Whether a tag was found | `true` |

## Examples

### Basic Usage
```yaml
- uses: half-ogre-games/hog-actions/get-latest-semver-tag@v1
  id: version
```

### Custom Prefix
```yaml
- uses: half-ogre-games/hog-actions/get-latest-semver-tag@v1
  id: version
  with:
    prefix: 'release-'
    default-version: 'release-1.0.0'
```

### No Prefix
```yaml
- uses: half-ogre-games/hog-actions/get-latest-semver-tag@v1
  id: version
  with:
    prefix: ''
    default-version: '0.1.0'
```

## Behavior

- **Tag Detection**: Finds all tags matching semantic versioning format `{prefix}X.Y.Z[-prerelease][+build]`
- **Full Semver Support**: Handles pre-release versions and build metadata according to [semver.org](https://semver.org) specification
- **Proper Sorting**: Uses semantic version precedence rules (pre-releases have lower precedence than releases)
- **No Tags Found**: Returns the `default-version` and sets `found` to `false`
- **Version Parsing**: Extracts major, minor, patch, pre-release, and build metadata components

### Supported Version Formats

- `1.2.3` - Basic version
- `v1.2.3` - With prefix
- `1.2.3-alpha` - With pre-release
- `1.2.3+build.123` - With build metadata  
- `1.2.3-beta.2+build.456` - With both pre-release and build metadata

### Version Precedence Examples

- `1.0.0-alpha < 1.0.0-beta < 1.0.0-rc < 1.0.0`
- `1.0.0-alpha.1 < 1.0.0-alpha.2`
- `1.0.0-alpha.1 < 1.0.0-alpha.beta`
- `1.0.0+build1 == 1.0.0+build2` (build metadata ignored for precedence)

## Requirements

- Repository must be checked out with `actions/checkout`
- Git must be available (standard on GitHub runners)
- No authentication required (uses local git repository)

## License

MIT License - see [LICENSE.md](../LICENSE.md)