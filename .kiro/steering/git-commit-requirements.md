---
inclusion: always
---

# Git Commit Requirements

## Security Notice

**CRITICAL SECURITY REQUIREMENT**: All commit messages and PR descriptions containing markdown syntax (especially backticks) MUST use temporary files instead of CLI arguments to prevent accidental shell command execution.

## Mandatory Git Workflow

All code changes must be committed to git after completing each task to ensure proper version control and change tracking.

## Branch Management Requirements

### Topic Branch Workflow
- **MUST** work on topic branches, never directly on main branch
- **MUST** create new topic branch before starting any task (see git-workflow-requirements.md)
- **MUST** create branches directly from `origin/main` for proper isolation
- **MUST** use descriptive branch names following the established convention

### Branch Creation
- Topic branches must be created using: `git checkout -b <type>/<task-id>-<description> origin/main`
- Branch names must follow: `<type>/task-<id>-<description>` format
- Always pull after branch creation to ensure up-to-date state

## Commit Requirements

### After Each Task Completion
- **MUST** commit all changes to git before moving to the next task
- **MUST** use descriptive commit messages that reference the task being completed
- **MUST** include the task number and brief description in commit message
- **MUST** add "Closes: Task X.Y" line to commit messages for traceability

### Commit Message Format

**MUST** follow Conventional Commits specification with task reference:

```
<type>(scope): Task X.Y - brief description

- Specific changes made
- Requirements addressed
- Any notable implementation details

Closes: Task X.Y
```

#### Conventional Commit Types
- **feat**: New feature implementation
- **test**: Adding or modifying tests
- **fix**: Bug fixes
- **refactor**: Code refactoring without changing functionality
- **perf**: Performance improvements
- **docs**: Documentation changes
- **chore**: Maintenance tasks, tooling changes

#### Scope Guidelines
- **core**: Changes to core parsing functionality
- **api**: Changes to public API
- **tests**: Test-related changes
- **build**: Build system or CI changes

### Examples
```
test(core): Task 2.1 - Add property test for POSIX/GNU specification compliance

- Added comprehensive property-based test for getopt compliance
- Validates Requirements 1.1, 2.1, 3.1
- Uses testing/quick with 100+ iterations

Closes: Task 2.1

test(core): Task 7.1 - Fill coverage gaps identified in baseline analysis

- Added unit tests for uncovered error handling paths
- Achieved 100% line coverage for parser.go
- Added edge case tests for malformed input

Closes: Task 7.1

feat(core): Task 10.2 - Enhance POSIXLY_CORRECT environment variable support

- Added environment variable detection and parsing
- Integrated with existing ParseMode system
- Validates Requirements 1.4

Closes: Task 10.2
```

## Security Requirements for Commit Messages

### Mandatory Temp File Usage
- **MUST NEVER** pass commit message body directly on CLI using `-m` flag with multi-line messages
- **MUST** use temporary files for commit messages to prevent shell execution of markdown backticks
- **MUST** use `git commit -F <temp-file>` instead of `git commit -m` for multi-line messages
- **MUST** use `gh pr create -F <temp-file>` instead of `gh pr create -b` for PR descriptions
- **MUST** clean up temporary files after successful commit/PR creation

### Temp File Security Pattern
```bash
# Create commit message file using direct file creation (NOT shell commands)
# Use file editor/IDE to create .commit-msg.txt with content:
# feat(specs): Add comprehensive module dependencies and testing infrastructure
#
# - Updated goarg and pflags specs with complete testing infrastructure requirements
# - Added module dependency management with local/remote configurations
# - Implemented cascading build system with intelligent change detection
# - Enhanced GitHub workflows to handle all modules without duplication
# - Added script flexibility requirements for target directory parameters
# - Ensured 100% coverage requirements for module-specific core functions
#
# Closes: Task spec-module-dependencies

# Commit using temp file
git commit -F ".commit-msg.txt"

# Clean up temp file
rm ".commit-msg.txt"
```

### Security Rationale
**Why Direct File Editing Is Required:**
- Shell commands like `echo`, `printf`, `cat` can cause shell expansion
- Direct file editing eliminates risk of command substitution or expansion
- Temporary files isolate the message content from shell interpretation
- This requirement applies to both `git commit` and `gh pr create` commands

### CRITICAL: Shell Command Prohibition
**NEVER use shell commands to write commit messages or PR descriptions:**
- `echo`, `printf`, `cat` can cause shell expansion and security issues
- Here-docs (<<EOF) can cause command execution to hang indefinitely
- Automated systems cannot detect when here-doc commands complete
- Use direct file editing tools instead of shell commands

### Prohibited Patterns
```bash
# NEVER DO THIS - Security Risk
git commit -m "feat: Add parser with \`parseArgs()\` function"

# NEVER DO THIS - Shell Expansion Risk
echo "feat: Add parser with \`parseArgs()\` function" > "$COMMIT_MSG_FILE"

# NEVER DO THIS - Causes Hanging
cat > "$COMMIT_MSG_FILE" << 'EOF'
feat: Add parser
EOF

# NEVER DO THIS - Shell Expansion Risk
printf "feat: Add parser\n- Added \`parseArgs()\` function\n" > "$COMMIT_MSG_FILE"
```

### Required Secure Patterns
```bash
# ALWAYS DO THIS - Safe and Reliable
# Create .commit-msg.txt using file editor/IDE (NOT shell commands)
# Write commit message content directly to the file
git commit -F ".commit-msg.txt"
rm ".commit-msg.txt"
```

### Pull Request Security Pattern
```bash
# Push branch to remote first
git push -u origin <branch-name>

# Check if branch has existing PR and monitor workflows after push
if gh pr view --json state 2>/dev/null; then
  echo "Branch has existing PR - monitoring workflows after push"
  gh pr checks
  gh pr status
fi

# Create PR description file using direct file creation (NOT shell commands)
# Use file editor/IDE to create .pr-desc.txt with content:
# ## Summary
# Brief description of changes
#
# ## Changes Made
# - Specific implementation details
# - Requirements addressed
#
# ## Testing
# - Tests added/modified
# - Coverage validation
#
# Closes: Task X.Y

# Create PR using temp file (if not already exists)
if ! gh pr view --json state 2>/dev/null; then
  gh pr create --title "<type>(scope): Task X.Y - Description" -F ".pr-desc.txt"
fi

# Monitor GitHub workflows after PR creation/update
gh pr checks

# Validate no workflow errors (all checks should pass)
gh pr status

# Clean up temp file
rm ".pr-desc.txt"
```

## Git Workflow Steps

1. **Create topic branch** - `git checkout -b <type>/task-<id>-<description> origin/main && git pull`
2. **Complete the task** - Implement all code changes for the current task
3. **Run tests** - Ensure all tests pass before committing
4. **Stage changes** - `git add .` or selectively stage relevant files
5. **Create commit message temp file** - Write commit message to temporary file using file editor/IDE
6. **Commit changes** - `git commit -F <temp-file>` (NEVER use `-m` for multi-line messages)
7. **Clean up temp file** - Remove temporary commit message file
8. **Verify commit** - `git log --oneline -1` to confirm commit was created
9. **Push branch** - `git push -u origin <branch-name>` to push branch to remote
10. **Monitor workflows** - If branch has PR, wait for all GitHub workflows to complete and fix any failures
11. **Create PR** - Use `gh pr create` with temp file for description (if not already created)
12. **Monitor PR workflows** - Wait for all GitHub workflows to complete and validate no errors
13. **Fix any failures** - Address any workflow failures before proceeding to next task
14. **Clean up PR temp file** - Remove temporary PR description file

## PR Workflow Integration Requirements

### After Each Commit and Push
When working on a branch with an existing PR, the following process MUST be followed after every `git push`:

1. **Automatic Workflow Monitoring**:
   ```bash
   # This script MUST be run after every push to a branch with existing PR
   if gh pr view --json state 2>/dev/null; then
     echo "Monitoring workflows for existing PR after push..."

     # Wait for workflows to complete
     while true; do
       CHECKS_STATUS=$(gh pr checks --json state --jq '.[].state' | sort -u)
       if echo "$CHECKS_STATUS" | grep -q "PENDING\|IN_PROGRESS"; then
         echo "Workflows running... waiting 30 seconds"
         sleep 30
       else
         break
       fi
     done

     # Check for failures and require fixes
     FAILED_CHECKS=$(gh pr checks --json name,conclusion --jq '.[] | select(.conclusion == "FAILURE") | .name')
     if [ -n "$FAILED_CHECKS" ]; then
       echo "❌ WORKFLOW FAILURES DETECTED - MUST FIX BEFORE PROCEEDING"
       echo "$FAILED_CHECKS"
       exit 1
     fi
   fi
   ```

2. **Mandatory Failure Resolution**:
   - **MUST** fix any workflow failures before proceeding to next task
   - **MUST** commit and push fixes
   - **MUST** wait for workflows to pass before continuing development

### After PR Creation
When creating a new PR, the following process MUST be followed:

1. **Immediate Workflow Monitoring**:
   ```bash
   # Create PR and wait for workflows
   gh pr create --title "<type>(scope): Task X.Y - Description" -F ".pr-desc.txt"

   # Wait for workflows to initialize and complete
   sleep 10
   while true; do
     CHECKS_STATUS=$(gh pr checks --json state --jq '.[].state' | sort -u)
     if echo "$CHECKS_STATUS" | grep -q "PENDING\|IN_PROGRESS"; then
       sleep 30
     else
       break
     fi
   done

   # Validate all checks passed
   FAILED_CHECKS=$(gh pr checks --json name,conclusion --jq '.[] | select(.conclusion == "FAILURE") | .name')
   if [ -n "$FAILED_CHECKS" ]; then
     echo "❌ PR CREATION FAILED - WORKFLOWS FAILING"
     echo "$FAILED_CHECKS"
     exit 1
   fi
   ```

2. **Validation Requirements**:
   - **MUST** wait for all workflows to complete before considering PR ready
   - **MUST** fix any failures detected during PR creation
   - **MUST** validate that all required workflows are triggered and pass

## Topic Branch Management

- **MUST** work on topic branches created from origin/main
- **MUST NOT** work directly on main branch
- Each commit should represent a complete, working state
- Avoid partial commits that leave the codebase in a broken state
- Topic branches should focus on a single task or closely related tasks

## Commit Verification

Before proceeding to the next task:
- Verify the commit was created successfully
- Ensure the commit message follows the required format
- Confirm all changes for the current task are included in the commit

## Workflow Validation

### Successful Implementation Verification

**Status**: ✅ **VALIDATED** - Security requirements successfully implemented and tested

**Validation Details**:
- **Date**: December 29, 2025
- **Branch**: `chore/git-security-requirements`
- **Commit**: `c0122ec` - chore(security): Add git security requirements for commit messages
- **PR**: https://github.com/major0/optargs/pull/17

**Security Workflow Verification**:
- ✅ **Topic branch creation** from `origin/main` with proper naming convention
- ✅ **Temp file usage** for commit message (no CLI `-m` flag used)
- ✅ **Secure commit pattern** using `git commit -F <temp-file>`
- ✅ **Temp file cleanup** after successful commit
- ✅ **Pre-commit hooks** passed (formatting, linting, security scans)
- ✅ **Temp file usage** for PR description (no CLI `-b` flag used)
- ✅ **Secure PR pattern** using `gh pr create -F <temp-file>`
- ✅ **Temp file cleanup** after successful PR creation

**Files Successfully Updated**:
- `.kiro/steering/git-commit-requirements.md` - Added comprehensive security requirements
- `.kiro/steering/git-workflow-requirements.md` - Enhanced with security guidelines
- `.gitignore` - Added temp file patterns and performance_baselines.json exclusion

**Security Measures Validated**:
- Markdown backticks in commit messages isolated from shell interpretation
- No accidental command execution through CLI arguments
- Proper temporary file lifecycle management
- Compliance with all security patterns and requirements

This validation confirms that the security requirements are not only documented but have been successfully implemented in practice, demonstrating the effectiveness of the secure git workflow.

## Benefits

- **Change Tracking**: Clear history of what was implemented in each task with standardized conventional commit format
- **Automated Tooling**: Conventional commits enable automated changelog generation and semantic versioning
- **Rollback Capability**: Ability to revert to any previous task completion state
- **Progress Visibility**: Easy to see implementation progress through git history with clear commit types
- **Debugging Aid**: Isolate issues to specific task implementations using commit scopes
- **Collaboration**: Clear change history for team members or future reference with standardized format
- **CI/CD Integration**: Conventional commits can trigger automated workflows based on commit types
- **Quality Assurance**: Mandatory PR workflow monitoring ensures all changes pass quality checks
- **Error Prevention**: Automated failure detection prevents proceeding with broken code
- **Workflow Reliability**: Comprehensive monitoring ensures consistent development process across all contributors
