# Contributing to CGAP

First off, thank you for considering contributing to CGAP! We appreciate your interest and effort. This document will guide you through the contribution process.

---

## Code of Conduct

Please review our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing. We are committed to providing a welcoming and inspiring community for all.

---

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Git
- Basic familiarity with Go, TypeScript/React, and PostgreSQL

### Setup Development Environment
```bash
# Clone the repository
git clone https://github.com/tushardhara/docs-ai.git
cd docs-ai

# Copy environment template
cp .env.example .env

# Update .env with your API keys (don't commit!)
# Edit .env and add your:
# - OPENAI_API_KEY or GEMINI_API_KEY
# - MEILISEARCH_KEY (change from masterKey if desired)

# Start dependencies
docker-compose up -d postgres meilisearch redis

# Run migrations
cd db/migrations && bash run.sh && cd ../..

# Start API server
go run cmd/api/main.go

# In another terminal, start worker
go run cmd/worker/main.go
```

### Verify Setup
```bash
# Check API is running
curl http://localhost:8080/health

# Check logs for errors
# Should see CGAP ASCII art banner
```

---

## How to Contribute

### 1. Pick an Issue

- Browse [GitHub Issues](https://github.com/tushardhara/docs-ai/issues)
- Look for `good-first-issue` or `help-wanted` labels
- Comment to claim the issue
- Ask questions if unclear

### 2. Create a Branch

```bash
# Create feature branch with descriptive name
git checkout -b feature/issue-N-description

# Examples:
# git checkout -b feature/issue-42-add-ocr-handler
# git checkout -b fix/issue-51-handle-null-pointer
# git checkout -b docs/issue-33-update-readme
```

### 3. Make Changes

**Follow these guidelines:**

#### Code Style
- Use Go conventions (`gofmt`, `golangci-lint`)
- Use TypeScript for extension code
- Follow existing patterns in the codebase
- Write clear, self-documenting code

#### Commits
```bash
# Write clear, descriptive commit messages
git commit -m "feat: add OCR handler for images (#42)"

# Use conventional commits:
# feat: new feature
# fix: bug fix
# docs: documentation
# test: tests
# refactor: code refactoring
# chore: build, deps, etc.
```

#### Testing
```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Check coverage
go test -cover ./...

# Run linter
golangci-lint run
```

#### Security Checklist Before Committing
- [ ] No `.env` file committed (verify in `git status`)
- [ ] No API keys or secrets in code
- [ ] No credentials in commit messages
- [ ] Run `git diff --cached` to verify changes
- [ ] Check `git log -p` for any sensitive data

**Command to verify no secrets:**
```bash
# Check staged changes for secrets
git diff --cached | grep -i -E "api.?key|secret|password|token"
# Should return nothing (exit 0)
```

### 4. Write Tests

Tests are **required** for new features:

```go
// Example test structure
func TestMyFeature(t *testing.T) {
    // Arrange
    expected := "something"
    
    // Act
    result := MyFunction()
    
    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Test requirements:**
- Unit tests for new functions
- Integration tests for API endpoints
- Mock external services (no real API calls)
- Aim for >85% coverage
- Use `*testing.T` not assertions library

### 5. Update Documentation

- [ ] Update README if new features added
- [ ] Add code comments for complex logic
- [ ] Update API docs if endpoints changed
- [ ] Add examples for new functionality
- [ ] Update CHANGELOG if applicable

### 6. Create a Pull Request

```bash
# Push your branch
git push origin feature/issue-N-description

# Go to GitHub and create PR
```

**PR Checklist:**
- [ ] Link to related issue (`Fixes #N`)
- [ ] Clear description of changes
- [ ] Screenshots if UI changes
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Commits are clean and descriptive

### 7. Respond to Review

- Be open to feedback
- Ask questions if confused
- Push additional commits (don't force push unless asked)
- Request re-review when addressing comments

### 8. Merge

Once approved:
- Maintainer will squash and merge
- Branch will be deleted automatically
- Your contribution will be in the next release!

---

## Development Workflow

### Building
```bash
# Build all packages
go build ./...

# Build specific binary
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestName ./package/...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting
```bash
# Run all linters
golangci-lint run

# Run specific linter
golangci-lint run --enable=gofmt
```

### Docker
```bash
# Build API image
docker build -f build/Dockerfile.api -t cgap-api:latest .

# Build worker image
docker build -f build/Dockerfile.worker -t cgap-worker:latest .

# Build and run everything
docker-compose up --build
```

---

## Project Structure

```
docs-ai/
├── cmd/
│   ├── api/          # API server entrypoint
│   ├── worker/       # Worker service entrypoint
│   └── migrate/      # Migration utilities
├── api/
│   ├── handlers.go   # HTTP handlers
│   ├── types.go      # Request/response types
│   └── handlers_test.go
├── internal/
│   ├── service/      # Business logic
│   ├── storage/      # Data access interfaces
│   ├── postgres/     # PostgreSQL implementation
│   ├── search/       # Hybrid search
│   ├── media/        # Media processing
│   ├── embedding/    # Embedding providers
│   ├── llm/          # LLM providers
│   └── testutil/     # Test fixtures & mocks
├── db/
│   ├── migrations/   # SQL migrations
│   └── schema.sql    # Database schema
├── extension/        # Browser extension (TypeScript/React)
├── docs/            # Documentation
└── docker-compose.yml
```

---

## Areas We Need Help With

### Backend (Go)
- [ ] Performance optimization
- [ ] Additional LLM provider support
- [ ] Rate limiting & caching improvements
- [ ] API documentation improvements
- [ ] Security enhancements

### Frontend (TypeScript/React)
- [ ] Browser extension UI polish
- [ ] Element highlighting improvements
- [ ] User experience enhancements
- [ ] Cross-browser testing

### Documentation
- [ ] API documentation
- [ ] Setup guides
- [ ] Troubleshooting guides
- [ ] Architecture diagrams

### DevOps
- [ ] CI/CD improvements
- [ ] Docker optimization
- [ ] Kubernetes manifests
- [ ] Monitoring & alerting

---

## Reporting Bugs

### Before Filing
1. Check existing issues (might be already reported)
2. Try to reproduce with latest code
3. Gather error messages and logs

### Filing a Bug Report
```markdown
**Title**: [BUG] Clear, descriptive title

**Reproduce Steps**:
1. ...
2. ...
3. ...

**Expected Behavior**: What should happen

**Actual Behavior**: What actually happens

**Environment**:
- OS: macOS/Linux/Windows
- Go version: 1.21+
- Branch: main

**Logs/Screenshots**: Attach relevant output

**Additional Context**: Any other details
```

---

## Requesting Features

```markdown
**Title**: [FEATURE] Clear description of feature

**Problem**: What problem does this solve?

**Solution**: How should it work?

**Alternatives**: Other possible approaches?

**Examples**: How would users interact with it?
```

---

## Review Process

### What Reviewers Look For
- ✅ Code follows project conventions
- ✅ Tests are included and passing
- ✅ No secrets or sensitive data
- ✅ Documentation is updated
- ✅ No breaking changes
- ✅ Commits are clean and descriptive

### Timeline
- Initial review: 1-3 days
- Feedback incorporation: 1-2 days per round
- Merge after approval: same day

---

## Questions?

- Check [GitHub Discussions](https://github.com/tushardhara/docs-ai/discussions)
- Review [README.md](README.md) and [docs/](docs/)
- Email: dev@example.com
- Open an issue with `question` label

---

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Thanked in release notes
- Recognized in community updates

---

## Thank You! 🙏

We appreciate all contributions, from code to documentation to bug reports. You're helping build something amazing!

---

**Happy Contributing!** 🚀

Last Updated: December 23, 2025
