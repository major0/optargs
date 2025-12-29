---
inclusion: always
---

# Git Workflow Requirements

## Branch Management Strategy

### Main Branch Protection
- **main** branch is protected and requires pull request reviews
- **develop** branch is used for integration and testing
- Direct pushes to main/develop branches are prohibited

### Topic Branch Workflow

#### Branch Creation Requirements
- **MUST** create topic branches from `origin/main` for proper isolation
- **MUST** use descriptive branch names following the established convention
- **MUST** pull latest changes after branch creation

#### Branch Naming Convention
Topic branches must follow the format: `<type>/task-<id>-<description>`

**Branch Types**:
- `feat/`: New feature implementation
- `fix/`: Bug fixes
- `test/`: Adding or modifying tests
- `docs/`: Documentation changes
- `refactor/`: Code refactoring without changing functionality
- `perf/`: Performance improvements
- `chore/`: Maintenance tasks, tooling changes

**Examples**:
```bash
feat/task-1-implement-posix-getopt
fix/task-2-handle-edge-case-parsing
test/task-3-add-property-based-tests
docs/task-4-update-api-documentation
```

#### Branch Creation Commands
```bash
# Create and switch to new topic branch
git checkout -b <type>/task-<id>-<description> origin/main

# Pull latest changes to ensure up-to-date state
git pull
```

## Pull Request Workflow

### Pull Request Requirements
- **MUST** create pull request from topic branch to main
- **MUST** pass all CI/CD workflow checks before merge
- **MUST** include descriptive title and detailed description
- **MUST** reference related issues or tasks
- **MUST** request appropriate reviewers

### CI/CD Integration
All pull requests trigger automated workflows:

#### Pre-commit Workflow
- Code quality validation
- Formatting and linting checks
- Security vulnerability scanning
- Commit message format validation

#### Build Workflow
- Multi-version Go compatibility (1.21, 1.22, 1.23)
- Cross-platform build verification (Linux, macOS, Windows, FreeBSD)
- Dependency verification

#### Coverage Workflow
- Test coverage validation
- Coverage regression detection
- Comprehensive coverage reporting

### Merge Requirements
- **MUST** have all CI workflow checks passing (green status)
- **MUST** have required reviewer approvals
- **MUST** resolve all merge conflicts
- **MUST** have up-to-date branch with main

## Automated Workflow Triggers

### Push Triggers
- **main branch**: Build and Coverage workflows
- **develop branch**: Build and Coverage workflows
- **topic branches**: No automatic triggers (only on PR)

### Pull Request Triggers
- **All PRs to main/develop**: Pre-commit, Build, and Coverage workflows
- **Draft PRs**: Limited workflow execution
- **Ready for review**: Full workflow execution

## Workflow Status Requirements

### Required Status Checks
All pull requests must pass these status checks:

#### Pre-commit Workflow
- ✅ Pre-commit hooks validation
- ✅ Code formatting (go fmt, goimports)
- ✅ Static analysis (go vet, golangci-lint when available)
- ✅ Security scanning (detect-secrets)
- ✅ Commit message format (commitlint)

#### Build Workflow
- ✅ Go 1.21 build (Linux, macOS, Windows, FreeBSD)
- ✅ Go 1.22 build (Linux, macOS, Windows, FreeBSD)
- ✅ Go 1.23 build (Linux, macOS, Windows, FreeBSD)
- ✅ Dependency verification

#### Coverage Workflow
- ✅ Test coverage validation
- ✅ Coverage regression check
- ✅ Coverage report generation

### Workflow Failure Handling
- **Pre-commit failures**: Automated PR comments with fix instructions
- **Build failures**: Detailed error reporting and platform-specific guidance
- **Coverage failures**: Coverage analysis and gap identification

## Branch Cleanup

### After Merge
- **MUST** delete topic branch after successful merge
- **MUST** update local main branch with merged changes
- **MUST** clean up local topic branch references

### Cleanup Commands
```bash
# After PR merge, update local main and clean up
git checkout main
git pull origin main
git branch -d <topic-branch-name>
git remote prune origin
```

## Git Configuration Requirements

### Local Git Setup
```bash
# Configure user information
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Configure commit message template (optional)
git config commit.template .gitmessage

# Configure pull behavior
git config pull.rebase false
```

### Pre-commit Hook Installation
```bash
# Install pre-commit hooks locally
pip install pre-commit
pre-commit install
pre-commit install --hook-type commit-msg
```

## Workflow Best Practices

### Development Workflow
1. **Create topic branch** from latest main
2. **Implement changes** with frequent commits
3. **Run local validation** using Makefile targets
4. **Push topic branch** and create pull request
5. **Address review feedback** and CI failures
6. **Merge after approval** and clean up branch

### Commit Practices
- Make atomic commits that represent complete, working changes
- Use descriptive commit messages following Conventional Commits
- Test changes locally before pushing
- Avoid force-pushing to shared branches

### Collaboration Guidelines
- Keep pull requests focused and reasonably sized
- Provide clear descriptions and context
- Respond promptly to review feedback
- Help maintain code quality standards

## Emergency Procedures

### Hotfix Workflow
For critical production issues:
1. Create hotfix branch from main: `git checkout -b hotfix/critical-issue main`
2. Implement minimal fix with tests
3. Create emergency PR with expedited review
4. Merge after essential checks pass
5. Backport to develop if needed

### Rollback Procedures
If issues are discovered after merge:
1. Identify problematic commit using git log
2. Create revert PR: `git revert <commit-hash>`
3. Follow standard PR workflow for revert
4. Investigate and fix underlying issue in separate PR
