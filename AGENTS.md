# Repository Guidelines

## Project Structure & Module Organization
This repository is currently bootstrapping and contains only `AGENTS.md`. As code is added, keep a simple, predictable layout:
- `src/` for application code
- `tests/` for automated tests
- `docs/` for design notes and runbooks
- `assets/` for static files (images, fixtures)
If you adopt a different layout, update this file and keep build/test commands aligned with those paths.

## Build, Test, and Development Commands
No build or test commands are configured yet. When tooling is introduced, list the exact commands here and what they do. Example entries to include:
- `npm run dev` — start the local dev server
- `npm test` — run the test suite
- `npm run lint` — run formatter/linter checks

## Coding Style & Naming Conventions
Until a formatter is added, use 2-space indentation for JSON/YAML and 4-space indentation for Python. Prefer descriptive, kebab-case file names (e.g., `daily-digest.ts`) and clear module names. If a formatter/linter is adopted (Prettier, ESLint, Black, Ruff, etc.), document the config files and the command to run it.

## Testing Guidelines
Tests are not set up yet. When added, specify the framework (e.g., Jest, Pytest), where tests live (`tests/` or alongside code), and naming patterns such as `*.test.ts` or `test_*.py`. Include commands to run all tests and a single test file.

## Commit & Pull Request Guidelines
No Git history is available in this workspace, so no existing commit convention can be inferred. When Git is initialized, adopt a clear standard such as Conventional Commits (`feat: add rss parser`) and keep commits scoped and descriptive. Pull requests should include a summary, linked issues (if any), and screenshots or logs for user-visible changes.

## Configuration & Secrets
Never commit credentials. Use `.env` for local secrets, add it to `.gitignore`, and provide a `.env.example` template. Document required environment variables in `README.md` once the project setup is defined.
