# Contributing to Checkly Go SDK

> 🤝 Learn about our [Commitment to Open Source](https://checklyhq.com/oss).

Hi! We are really excited that you are interested in contributing to Checkly Go SDK, and we really appreciate your commitment. Before submitting your contribution, please make sure to take a moment and read through the following guidelines:

- [Code of Conduct](./CODE_OF_CONDUCT.md)
- [Issue Reporting Guidelines](#issue-reporting-guidelines)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Development Setup](#development-setup)
- [Scripts](#scripts)
- [Project Structure](#project-structure)
- [Release Process](#release-process)

## Issue Reporting Guidelines

- Always use [GitHub issues](https://github.com/checkly/checkly-go-sdk/issues/new/choose) to create new issues and select the corresponding issue template.

## Pull Request Guidelines

- Checkout a topic branch from a base branch, e.g. `main`, and merge back against that branch.

- [Make sure to tick the "Allow edits from maintainers" box](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork). This allows us to directly make minor edits / refactors and saves a lot of time.

- Add accompanying documentation, usage samples & test cases
- Add or update the runnable examples in `example_test.go` to showcase your changes.
- Use existing resources as templates and ensure that each property has a corresponding `description` field.
- Each PR should be linked with an issue, use [GitHub keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests) for that.
- Be sure to follow up project code styles (`$ make fmt`)

- If adding a new feature:
  - Provide a convincing reason to add this feature. Ideally, you should open a "feature request" issue first and have it approved before working on it (it should has the label "state: confirmed")
  - Each new feature should be linked to

- If fixing a bug:
  - Provide a detailed description of the bug in the PR. A minimal program that reproduces it would be great!

- It's OK to have multiple small commits as you work on the PR - GitHub can automatically squash them before merging.

- Make sure tests pass!

- Pull request titles must follow the [semantic commit messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716) format, as GitHub's generated release notes list each pull request by its title. Pull requests are squash-merged, so keep the commit message of a single-commit pull request in the same format: GitHub uses it as the squash commit message.

- Describe in the pull request description, under "Changes for SDK Users", what changes for SDK users, and call out breaking changes with what to do instead. The release notes are written from these descriptions.

## Pull request review: Ship / Show / Ask

This repo supports the [Ship / Show / Ask](https://martinfowler.com/articles/ship-show-ask.html)
strategy. Signal your intent by ending the **PR title** with a tag (case-insensitive), e.g.
`feat: add retries [ship]`:

- **`[ship]`** — No review needed. `github-actions[bot]` auto-approves the PR; merge it yourself
  once CI is green. Use for low-risk changes you're confident in.
- **`[show]`** — Merge now, review after the fact. Also auto-approved; open it for visibility and
  merge once CI passes, inviting comments for follow-up.
- **`[ask]`** or **no tag** — Normal review. The PR waits for a human approval before it can merge.

Notes:

- **Auto-approval is limited to PRs opened from a branch in this repository** (i.e. members and
  collaborators with push access). Pull requests from forks are **never** auto-approved and always
  require manual review, regardless of the title tag.
- The tag only counts at the very end of the title. A `[ship]`/`[show]` elsewhere in the title is
  ignored — so e.g. a GitHub-generated revert title (`Revert "feat: x [ship]"`) still gets manual
  review.
- If you tag a PR `[ship]`/`[show]` and later change the title to `[ask]` (or drop the tag), the
  earlier auto-approval is dismissed so the PR returns to manual review.
- Dismissing the bot's approval by hand doesn't stick: while the title still ends with
  `[ship]`/`[show]`, the next push (or other PR update) re-approves. To durably demand review on a
  tagged PR, submit a **Request changes** review or retitle the PR to `[ask]`.
- A draft PR tagged `[ship]`/`[show]` is approved only once it's marked ready for review.
- This only adds an approving review; merging still respects branch protection (passing CI, etc.).
  It does not auto-merge.

The behavior lives in [`.github/workflows/ship-show-ask.yml`](.github/workflows/ship-show-ask.yml).
It depends on two GitHub settings outside this repo: **Allow GitHub Actions to create and approve
pull requests** must be enabled, and branch protection on `main` must require an approving review
(a review count — not a Code Owners review, which a bot can't satisfy).

## Development Setup

The development branch is `main`. This is the branch that all pull requests should be made against.

> ⚠️ Before you start, is recommended to have a good understanding on how the provider works, the resources it has and its different configurations. Here are the ["getting started"](https://github.com/checkly/checkly-go-sdk#getting-started) guides.

### Pre-requirements
- [Go](https://go.dev/doc/install) 1.26 or later

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

To inspect the requests the client sends and the responses it receives, route them through an intercepting HTTPS proxy. The default HTTP client honors the `HTTPS_PROXY` environment variable; the proxy's CA certificate must be trusted, for example through the system keychain on macOS or `SSL_CERT_FILE` on Linux. Proxied traffic includes the API key and secret values in plain text, so do not share captures.

## Scripts

In order to facilitate the development process, we have a `Makefile` with a few scripts that are used to build, test and run working examples.

- `make test`: run the unit tests and then the integration tests.
- `make test-unit`: run the unit tests, which use canned responses and need no Checkly account.
- `make test-integration`: build with the `integration` tag and run the integration tests against the real API, along with the unit tests. They need the `CHECKLY_API_KEY` and `CHECKLY_ACCOUNT_ID` environment variables, plus `CHECKLY_API_URL` (it defaults to `http://localhost:3000`), and they create and delete resources in that account.

## Project Structure

The SDK is a single package, `checkly`, in the repository root.

- `client.go`: the `Client` interface and `NewClient`. `options.go` holds the `ClientOption`s.
- `transport.go`: sending requests, authentication and response decoding. `rate_limit.go` holds the retry policy for HTTP 429 responses, and `errors.go` the error types and sentinel errors.
- One file per API resource, such as `api_check.go`, `url_monitor.go`, `alert_channel.go` or `status_page_v3.go`. Each holds the resource's types, its request payloads and its client methods, in that order, with the methods in the order of the `Client` interface. Types shared by several resources have their own files, such as `assertion.go`, `alert_settings.go` and `retry_strategy.go`.
- `<resource>_test.go` tests a resource through the exported API, and `<resource>_internal_test.go` tests unexported code, such as request payloads. Shared helpers are in `helpers_test.go` and `helpers_internal_test.go`. `status_test.go` checks the status codes, credentials and retries of every client method.
- `integration_test.go`: tests against the real API, built with the `integration` tag.
- `fixtures/`: canned API responses for the tests.
- `example_test.go`: the examples shown on pkg.go.dev. They are compiled by `go test` but not run, as they call the API.
- `.github/`: GitHub and GitHub Actions configuration.

When adding a resource, add its types, payloads and methods in a new `<resource>.go`, its methods to the `Client` interface next to the related ones, and a row per method to `statusCases` in `status_test.go`.


## Release Process
Releases are published by hand on GitHub. Publishing a release creates its git tag, which must follow the [SemVer](https://semver.org) pattern. Go modules are fetched from the tag, so there is nothing to build or upload.

The GitHub release is the changelog: there is no `CHANGELOG.md`. Generate its notes from the merged pull requests, then expand them from the pull request descriptions. The notes can be edited after the release to clarify things.

Remember to create release candidates and spend some time testing in production before publishing a final version. You can also mark the release as "Pre-release" on GitHub until you consider it mature enough.
