# Project Guidelines & Rules

## Git Workflow Automation & Standards
- **Conventional Commits**: After completing any code changes, bug fixes, or feature implementations, always follow the `git-workflow` skill (`.agent/skills/git-workflow/SKILL.md`).
- **Commit Cadence**: Stage relevant files and create atomic commits using Conventional Commit syntax:
  - `feat(<scope>): ...` for new features
  - `fix(<scope>): ...` for bug fixes
  - `refactor(<scope>): ...` for refactoring without behavior change
  - `test(<scope>): ...` for adding or updating tests
  - `chore(<scope>): ...` for maintenance, config, dependencies
  - `docs(<scope>): ...` for documentation changes
- **Verification**: Ensure tests (`go test ./...`) pass before committing code changes.
