# Contributing to Sample App

Thank you for your interest in contributing! This guide outlines the process for contributing to this project.

## Fork and PR Workflow

We follow a typical "fork-and-pull" workflow:

1.  **Fork** this repository to your own GitHub account.
2.  **Clone** your fork to your local machine:
    ```bash
    git clone https://github.com/your-username/sample-app.git
    ```
3.  **Create a new branch** for your changes (see [Branch Naming](#branch-naming-convention) below).
4.  **Make your changes** and commit them using the [Commit Message Style](#commit-message-style).
5.  **Push** the branch to your fork:
    ```bash
    git push origin your-username/short-description
    ```
6.  **Open a Pull Request** (PR) from your fork's branch to our `main` branch.

## Branch Naming Convention

Please use the following naming convention for your branches:

`your-username/short-description`

**Example:** `jdoe/fix-auth-bug` or `jdoe/add-metrics-endpoint`

## Commit Message Style

We use [Conventional Commits](https://www.conventionalcommits.org/) to maintain a clean history and automate versioning.

Format: `<type>: <description>`

Common types:
- `feat:` A new feature
- `fix:` A bug fix
- `docs:` Documentation changes
- `ci:` Changes to our CI configuration or scripts
- `refactor:` A code change that neither fixes a bug nor adds a feature
- `test:` Adding missing tests or correcting existing tests

**Example:** `feat: add user profile endpoint`

## Pull Request Requirements

When opening a Pull Request, please ensure:
- The title clearly describes the change (use Conventional Commits).
- The description includes a summary of the changes and links to any related issues.
- All CI checks pass.

## CI Requirements

Before submitting your PR, please ensure your changes pass the following local checks, which are also run in our CI pipeline:

### 1. Linting
We use `golangci-lint` to ensure code quality:
```bash
golangci-lint run
```

### 2. Testing
All unit tests must pass. We run tests with the race detector enabled:
```bash
go test ./... -v -race -count=1
```

### 3. Building
The Docker image must build successfully. You can test this locally:
```bash
docker build -t sample-app .
```
