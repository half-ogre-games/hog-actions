# Get Next Semver Action

A GitHub Action that calculates the next semantic version based on increment type and current version, following semantic versioning specifications.

## Features

- **Smart Increment Logic**: Handles major, minor, and patch increments properly
- **Semantic Versioning**: Uses versionkit for proper semver parsing and generation
- **Pre-release Cleanup**: Removes pre-release and build metadata for release versions
- **Flexible Prefixes**: Supports custom version prefixes or no prefix
- **Version Reset**: Correctly resets lower version components (minor/patch to 0 on major increment)

## Usage

```yaml
- name: Calculate next version
  id: next_version
  uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: 'v1.2.3'
    increment-minor: true
    prefix: 'v'

- name: Use next version
  run: |
    echo "Next version: ${{ steps.next_version.outputs.version }}"
    echo "Version core: ${{ steps.next_version.outputs.version-core }}"
    echo "Increment type: ${{ steps.next_version.outputs.increment-type }}"
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `current-version` | Current semantic version (e.g., "1.2.3" or "v1.2.3-alpha.1") | Yes | - |
| `increment-major` | Increment major version (resets minor and patch to 0) | No | `false` |
| `increment-minor` | Increment minor version (resets patch to 0) | No | `false` |
| `prefix` | Version prefix to preserve (e.g., "v" for "v1.2.3") | No | `v` |

## Outputs

| Output | Description | Example |
|--------|-------------|---------|
| `version` | The next semantic version with prefix | `v1.3.0` |
| `version-core` | The next semantic version without prefix | `1.3.0` |
| `major` | The major version number | `1` |
| `minor` | The minor version number | `3` |
| `patch` | The patch version number | `0` |
| `increment-type` | The type of increment performed | `minor` |

## Examples

### Patch Increment (Default)
```yaml
- uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: 'v1.2.3'
# Output: v1.2.4
```

### Minor Increment
```yaml
- uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: 'v1.2.3'
    increment-minor: true
# Output: v1.3.0
```

### Major Increment
```yaml
- uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: 'v1.2.3'
    increment-major: true
# Output: v2.0.0
```

### No Prefix
```yaml
- uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: '1.2.3'
    prefix: ''
    increment-minor: true
# Output: 1.3.0
```

### With Pre-release (Cleaned)
```yaml
- uses: half-ogre-games/hog-actions/get-next-semver@v1
  with:
    current-version: 'v1.2.3-beta.1+build.456'
    increment-patch: true
# Output: v1.2.4 (pre-release and build metadata removed)
```

## Behavior

### Increment Types
- **Patch** (default): Increments patch version (`1.2.3` → `1.2.4`)
- **Minor**: Increments minor, resets patch to 0 (`1.2.3` → `1.3.0`)
- **Major**: Increments major, resets minor and patch to 0 (`1.2.3` → `2.0.0`)

### Version Processing
- **Pre-release Removal**: Always removes pre-release identifiers for clean releases
- **Build Metadata Removal**: Always removes build metadata for clean releases
- **Prefix Preservation**: Maintains the specified prefix in output
- **Component Reset**: Lower components reset to 0 when higher components increment

### Input Validation
- Only one increment type can be specified
- Current version must be valid semantic version format
- Supports versions with or without prefixes
- Handles complex versions like `v1.2.3-alpha.1+build.456`

## Error Handling

The action will fail with descriptive error messages for:
- Invalid semantic version format
- Missing required current-version input
- Conflicting increment flags (both major and minor set to true)

## Requirements

- No external dependencies (uses local Git repository context)
- Works with any semantic versioning scheme
- Compatible with GitHub Actions environment

## License

MIT License - see [LICENSE.md](../LICENSE.md)