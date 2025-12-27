---
inclusion: always
---

# Git Commit Requirements

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

## Git Workflow Steps

1. **Create topic branch** - `git checkout -b <type>/task-<id>-<description> origin/main && git pull`
2. **Complete the task** - Implement all code changes for the current task
3. **Run tests** - Ensure all tests pass before committing
4. **Stage changes** - `git add .` or selectively stage relevant files
5. **Commit changes** - `git commit -m "<type>(scope): Task X.Y - Description"`
6. **Verify commit** - `git log --oneline -1` to confirm commit was created

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

## Benefits

- **Change Tracking**: Clear history of what was implemented in each task with standardized conventional commit format
- **Automated Tooling**: Conventional commits enable automated changelog generation and semantic versioning
- **Rollback Capability**: Ability to revert to any previous task completion state
- **Progress Visibility**: Easy to see implementation progress through git history with clear commit types
- **Debugging Aid**: Isolate issues to specific task implementations using commit scopes
- **Collaboration**: Clear change history for team members or future reference with standardized format
- **CI/CD Integration**: Conventional commits can trigger automated workflows based on commit types