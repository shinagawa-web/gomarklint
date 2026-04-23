# Contributing to gomarklint

Thank you for your interest in contributing! This guide explains how to set up the development environment, run tests, and submit changes.

## Prerequisites

- **Go 1.23+** — [install](https://go.dev/dl/)
- **Make** — standard on macOS/Linux; on Windows use WSL or Git Bash
- **Git**

## Getting started

```sh
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/<your-username>/gomarklint.git
cd gomarklint

# Install pre-push hooks (runs lint + unit tests before every push)
make install-hooks
```

## Development workflow

### Build & test

| Task | Command |
|---|---|
| Build binary | `make build` |
| Unit tests | `make test` |
| E2E tests | `make test-e2e` |
| All tests | `make test-all` |
| Benchmarks | `make bench` |
| Static analysis | `make static-lint` |

Always run `make test-all` before opening a pull request.

### Making changes

1. **Open an issue first.** For anything beyond a trivial typo fix, open (or find) an issue and discuss the approach before writing code. This avoids wasted effort on work that won't be merged.
2. **Create a feature branch** from `main`:
   ```sh
   git checkout -b your-branch-name
   ```
3. **Keep your branch up to date** with rebase, never merge:
   ```sh
   git fetch origin main && git rebase origin/main
   ```
4. **Write tests** for any new behavior. New lint rules require both unit tests and an E2E test.
5. **Run `make test-all`** to confirm everything passes before pushing.

## Commit messages

- Use the imperative mood: _"Add rule MD013"_, not _"Added rule MD013"_
- Reference the related issue number: `feat: add max-line-length rule (MD013) (#145)`
- Keep the subject line under 72 characters

## Pull requests

- Title should match the conventional-commits style used in this repo (`feat:`, `fix:`, `docs:`, `chore:`, etc.)
- Fill in the PR template completely
- Link the PR to the issue it resolves (`Closes #NNN`)
- A maintainer will review; please address feedback promptly

## Adding a new lint rule

1. Create the rule implementation under `internal/rules/`.
2. Register the rule in the linter.
3. Add unit tests in `internal/rules/`.
4. Add a testdata fixture under `testdata/`.
5. Add an E2E test.
6. Document the rule in `docs/`.
7. Update the config example if applicable.

See the issue for rule [#106](https://github.com/shinagawa-web/gomarklint/issues/106) as a canonical example of the expected issue format and implementation plan.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Please be respectful and inclusive in all interactions.
