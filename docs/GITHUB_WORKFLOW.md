# GitHub Workflow Guide

## Overview
This project follows a **strict PR-based workflow**. All changes must go through Pull Requests - direct pushes to `main` are not allowed.

## Quick Start

### 1. Pick an Issue
Browse open issues: https://github.com/tushardhara/docs-ai/issues

Issues are labeled by priority and sprint week:
- `week-3` - Extension endpoint refinements
- `week-4` - Browser extension UI
- `browser-extension` - Chrome extension related
- `production` - Production readiness tasks

### 2. Create Feature Branch
```bash
# Fetch latest changes
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/issue-N-short-description

# Example: git checkout -b feature/issue-3-extension-directory
```

### 3. Make Changes
- Implement the feature according to the issue description
- Follow project conventions (see copilot-instructions.md)
- Write tests for new functionality
- Ensure code compiles: `go build ./...`

### 4. Test Your Changes
```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Test specific package
go test ./internal/media/...

# Check for linting issues (if golangci-lint installed)
golangci-lint run
```

### 5. Commit Your Changes
Follow conventional commits:
```bash
# Feature addition
git commit -m "feat: implement DOM capture content script (#5)"

# Bug fix
git commit -m "fix: handle nil pointer in OCR handler (#7)"

# Documentation
git commit -m "docs: update API endpoint documentation (#8)"

# Tests
git commit -m "test: add integration tests for video handler (#9)"

# Refactoring
git commit -m "refactor: simplify hybrid search logic (#10)"

# Chores
git commit -m "chore: update dependencies (#11)"
```

### 6. Push Branch
```bash
git push origin feature/issue-N-short-description
```

### 7. Create Pull Request
1. Go to: https://github.com/tushardhara/docs-ai/pulls
2. Click "New Pull Request"
3. Select your feature branch
4. Fill out the PR template:
   - Link to issue: `Fixes #N`
   - Describe changes
   - Explain testing approach
5. Request review from team members

### 8. Code Review
- Address review comments
- Push additional commits to the same branch
- Discuss changes in PR comments
- Get approval from at least one reviewer

### 9. Merge
Once approved:
1. Squash and merge (recommended for clean history)
2. Delete the feature branch
3. Issue will auto-close via "Fixes #N" in PR description

## Commit Message Types

| Type | Usage | Example |
|------|-------|---------|
| `feat` | New feature | `feat: add element highlighting (#9)` |
| `fix` | Bug fix | `fix: resolve memory leak in orchestrator (#7)` |
| `docs` | Documentation | `docs: update Week 4 checklist (#8)` |
| `test` | Tests | `test: add coverage for YouTube handler (#10)` |
| `refactor` | Code restructuring | `refactor: simplify DOM parsing logic (#11)` |
| `perf` | Performance improvement | `perf: optimize search query execution (#12)` |
| `chore` | Maintenance | `chore: bump dependencies (#13)` |
| `style` | Code style (formatting) | `style: fix indentation in handlers.go (#14)` |

## Branch Naming Convention

```
feature/issue-N-short-description   # New features
fix/issue-N-short-description       # Bug fixes
docs/issue-N-short-description      # Documentation
test/issue-N-short-description      # Test improvements
refactor/issue-N-short-description  # Code refactoring
```

**Examples:**
- `feature/issue-3-extension-directory`
- `feature/issue-5-dom-capture-script`
- `fix/issue-7-nil-pointer-ocr`
- `docs/issue-8-api-documentation`

## PR Review Guidelines

### For Authors
- Keep PRs focused (one issue per PR)
- Write clear descriptions
- Include tests
- Ensure CI passes
- Respond to feedback promptly

### For Reviewers
- Check code follows project patterns
- Verify tests exist and pass
- Look for potential bugs
- Ensure documentation is updated
- Test the changes locally if needed

## CI/CD Pipeline

GitHub Actions runs on every PR:
1. **Build Check**: `go build ./...`
2. **Unit Tests**: `go test ./...`
3. **Race Detector**: `go test -race ./...`
4. **Linting**: `golangci-lint run` (if configured)

All checks must pass before merge.

## Troubleshooting

### "Branch is out of date"
```bash
git checkout main
git pull origin main
git checkout feature/issue-N-description
git rebase main
git push --force-with-lease
```

### "Merge conflicts"
```bash
git checkout feature/issue-N-description
git rebase main
# Resolve conflicts in your editor
git add .
git rebase --continue
git push --force-with-lease
```

### "CI failing"
1. Check GitHub Actions logs
2. Run tests locally: `go test ./...`
3. Fix issues
4. Commit and push fixes

## Questions?
- Check existing issues and PRs for examples
- Review `docs/START_HERE.md` for project setup
- Read `.github/copilot-instructions.md` for development patterns
