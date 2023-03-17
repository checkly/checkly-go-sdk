# Contributing to Checkly Go SDK

> 🤝 Learn about our [Commitment to Open Source](https://checklyhq.com/oss).

Hi! We are really excited that you are interested in contributing to Checkly Go SDK, and we really appreciate your commitment. Before submitting your contribution, please make sure to take a moment and read through the following guidelines:

- [Code of Conduct](./CODE_OF_CONDUCT.md)
- [Issue Reporting Guidelines](#issue-reporting-guidelines)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Development Setup](#development-setup)
- [Scripts](#scripts)
- [Project Structure](#project-structure)
- [Release Process](#releases-process)

## Issue Reporting Guidelines

- Always use [GitHub issues](https://github.com/checkly/checkly-go-sdk/issues/new/choose) to create new issues and select the corresponding issue template.

## Pull Request Guidelines

- Checkout a topic branch from a base branch, e.g. `main`, and merge back against that branch.

- [Make sure to tick the "Allow edits from maintainers" box](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork). This allows us to directly make minor edits / refactors and saves a lot of time.

- Add accompanying documentation, usage samples & test cases
- Add/update demo files to showcase your changes.
- Use existing resources as templates and ensure that each property has a corresponding `description` field.
- Each PR should be linked with an issue, use [GitHub keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests) for that.
- Be sure to follow up project code styles (`$ make fmt`)

- If adding a new feature:
  - Provide a convincing reason to add this feature. Ideally, you should open a "feature request" issue first and have it approved before working on it (it should has the label "state: confirmed")
  - Each new feature should be linked to

- If fixing a bug:
  - Provide a detailed description of the bug in the PR. A working demo would be great!

- It's OK to have multiple small commits as you work on the PR - GitHub can automatically squash them before merging.

- Make sure tests pass!

- Commit messages must follow the [semantic commit messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716) so that changelogs can be automatically generated.

## Development Setup

The development branch is `main`. This is the branch that all pull requests should be made against.

> ⚠️ Before you start, is recommended to have a good understanding on how the provider works, the resources it has and its different configurations. Here are the ["getting started"](https://github.com/checkly/checkly-go-sdk#getting-started) guides.

### Pre-requirements
- [Go](https://go.dev/doc/install) >= 1.18.2

After you have installed  Go, you can clone the repository and start working on the `main` branch.

1. [Fork](https://help.github.com/articles/fork-a-repo/) this repository to your own GitHub account and then [clone](https://help.github.com/articles/cloning-a-repository/) it to your local device.

  > If you don't need the whole git history, you can clone with depth 1 to reduce the download size:

  ```sh
  $ git clone --depth=1 git@github.com:checkly/checkly-go-sdk.git
  ```

2. Navigate into the project and create a new branch:
  ```sh
  cd checkly-go-sdk && git checkout -b MY_BRANCH_NAME
  ```

3. Download go packages
  ```sh
  $ go mod tidy
  ```

4. Run build process to ensure that everything is on place
  ```sh
  $ go build
  ```

## Debugging

If things aren't working as you expect, you can pass an `io.Writer` to `checkly.NewClient's fourth arg` to receive debug output. If `debug` is non-nil, then all API requests and responses will be dumped to the specified writer (for example, `os.Stderr`).

Regardless of the debug setting, if a request fails with HTTP status 400 Bad Request, the full response will be dumped (to standard error if no debug writer is set):

```go
debugOutput := os.Stderr
client.NewClient(
	"https://api.checklyhq.com",
	"your-api-key",
	nil,
	debugOutput,
)
```

## Scripts

In order to facilitate the development process, we have a `Makefile` with a few scripts that are used to build, test and run working examples.

> TBA

## Project Structure

- `./.github`: contains github (and github actions) related files.

> TBA


## Release Process
The release process is automatically handled with [goreleaser](https://goreleaser.com/) and GitHub `release` action.
To trigger a new release you need to create a new git tag, using [SemVer](https://semver.org) pattern and then push it to the `main` branch.

Remember to create releases candidates releases and spend some time testing in production before publishing a final version. You can also tag the release as "Pre-Release" on GitHub until you consider it mature enough.
