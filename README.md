# Checkly Go SDK

[![Tests](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml/badge.svg)](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/checkly/checkly-go-sdk/v2.svg)](https://pkg.go.dev/github.com/checkly/checkly-go-sdk/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/checkly/checkly-go-sdk)](https://goreportcard.com/report/github.com/checkly/checkly-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/checkly/checkly-go-sdk)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/checkly/checkly-go-sdk?label=Version)

The Checkly Go SDK is a Go client for the [Checkly](https://www.checklyhq.com) [public REST API](https://www.checklyhq.com/docs/api). It manages checks and monitors, check groups, alert channels, dashboards, status pages, maintenance windows, environment variables, snippets, private locations and more. The [Checkly Terraform provider](https://github.com/checkly/terraform-provider-checkly) is built on it.

## Installation

```bash
go get github.com/checkly/checkly-go-sdk/v2
```

```go
import "github.com/checkly/checkly-go-sdk/v2"
```

## Getting Started

Most methods need an API key and the ID of the Checkly account it acts on, a UUID. You can create an API key [in the Checkly web app](https://app.checklyhq.com/account/api-keys).

### Creating a client

`NewClient` builds a client, configured by options. `WithAuthFromEnv` reads the API key and account ID from the `CHECKLY_API_KEY` and `CHECKLY_ACCOUNT_ID` environment variables, the same ones the Checkly CLI and the Terraform provider use. `NewClient` fails if either is not set:

```go
client, err := checkly.NewClient(checkly.WithAuthFromEnv())
if err != nil {
	log.Fatal(err)
}
```

`WithAuth` takes the credentials as arguments instead, and `WithHTTPClient` sets the HTTP client, for example to bound how long a single request may take. `NewClient` returns an error for an invalid option, such as an empty value or an account ID that is not a UUID:

```go
client, err := checkly.NewClient(
	checkly.WithAuth(apiKey, accountID),
	checkly.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
)
if err != nil {
	log.Fatal(err)
}
```

Without an auth option, only the public methods `GetStaticIPs` and `GetRuntime` work; every other method fails with `checkly.ErrMissingAPIKey` without sending a request.

### Creating an API check

An API check sends an HTTP request every `Frequency` minutes from each of its locations and fails when an assertion does not hold. Set both response time limits: they are always sent, so leaving them at 0 sets a limit of 0 milliseconds.

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

check, err := client.CreateAPICheck(ctx, checkly.APICheck{
	Name:                   "Example API",
	Frequency:              10,
	Activated:              true,
	Locations:              []string{"eu-west-1", "us-east-1"},
	Tags:                   []string{"example"},
	DegradedResponseTime:   5000,
	MaxResponseTime:        20000,
	UseGlobalAlertSettings: true,
	Request: checkly.Request{
		Method:          http.MethodGet,
		URL:             "https://example.com/api/health",
		FollowRedirects: true,
		Headers: []checkly.KeyValue{
			{Key: "Accept", Value: "application/json"},
		},
		Assertions: []checkly.Assertion{
			{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
			{Source: checkly.AssertionSourceResponseTime, Comparison: checkly.AssertionComparisonLessThan, Target: "2000"},
		},
	},
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("created API check %q with ID %s\n", check.Name, check.ID)
```

Each check and monitor type has its own type and methods, such as `GetAPICheck`, `UpdateAPICheck` and `DeleteAPICheck` for API checks. `GetCheck` reads a check whose type is not known and returns a pointer to its type, and `DeleteCheck` deletes one.

### Creating a URL monitor

A URL monitor requests a URL with GET and checks the response status, which is cheaper than an API check when that is all you need:

```go
monitor, err := client.CreateURLMonitor(ctx, checkly.URLMonitor{
	Name:                   "Example homepage",
	Frequency:              5,
	Activated:              true,
	Locations:              []string{"eu-west-1", "us-east-1"},
	Tags:                   []string{"example"},
	DegradedResponseTime:   3000,
	MaxResponseTime:        10000,
	UseGlobalAlertSettings: true,
	Request: checkly.URLRequest{
		URL:             "https://example.com/",
		FollowRedirects: true,
		Assertions: []checkly.Assertion{
			{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
		},
	},
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("created URL monitor %q with ID %s\n", monitor.Name, monitor.ID)
```

### Handling errors

When the API answers with a status a method does not accept, the method returns a `*checkly.APIError`. Its text names the request, the status and the backend's message, for example `GET /v1/checks/abc: 404 Not Found`. Match it with `errors.Is` and the sentinel errors such as `checkly.ErrNotFound`, `checkly.ErrForbidden` or `checkly.ErrRateLimited`, or inspect it with `errors.As`, rather than matching the text:

```go
_, err := client.GetAPICheck(ctx, id)
if errors.Is(err, checkly.ErrNotFound) {
	// The check does not exist.
}
var apiErr *checkly.APIError
if errors.As(err, &apiErr) {
	log.Printf("Checkly API answered %d: %s", apiErr.StatusCode, apiErr.Message)
}
```

A request the API rejects as invalid matches `checkly.ErrBadRequest`. If it failed schema validation it also matches `checkly.ErrValidation`, and `apiErr.Details` lists the fields it names, if any.

Creating or updating a Slack App alert channel fails with a `*checkly.SlackIntegrationNotConnectedError`, matching `checkly.ErrSlackIntegrationNotConnected`, when the account has no connected Slack workspace; its text only says how to connect one. It fails with a `*checkly.UnresolvableSlackHandlesError`, matching `checkly.ErrUnresolvableSlackHandles`, when the workspace cannot resolve some of its handles, which `Handles` lists. `errors.As` still finds the `*checkly.APIError` that both wrap.

Errors for requests that got no response wrap their cause, so `errors.Is(err, context.DeadlineExceeded)` works too. A request answered with HTTP 429 is retried a few times, waiting as the API asks while the context's deadline allows. A request that can safely be re-sent, such as a GET, PUT or DELETE, is also retried twice when it gets no response, for example because its connection was reset, unless the cause is one a retry would only repeat: a timeout, a DNS failure, a rejected certificate or a failure to connect. Errors returned by a custom transport or redirect policy passed to `WithHTTPClient` are retried too.

The SDK never adds the API key or the request body to an error. It does keep the backend's error message, and the start of an error body that is not JSON, cut at 500 bytes. A validation message can quote a submitted value, and so can the text of such a body, so take care when logging errors from requests that carry secrets.

## Documentation

- The [package reference](https://pkg.go.dev/github.com/checkly/checkly-go-sdk/v2) on pkg.go.dev, with [examples](https://pkg.go.dev/github.com/checkly/checkly-go-sdk/v2#pkg-examples).
- [Upgrading from v1](UPGRADING.md): the changes in v2 and how to update your code.
- The [Checkly API reference](https://www.checklyhq.com/docs/api) for the resources and their fields.

## Questions

For questions and support, please open a [discussion](https://github.com/checkly/checkly-go-sdk/discussions). The issue list of this repository is for bug reports and feature or documentation requests only.

## Issues

Please choose the matching [issue template](https://github.com/checkly/checkly-go-sdk/issues/new/choose) when opening an issue. Issues that do not follow the templates may be closed.

## Contributing

Please read the [Contributing Guide](CONTRIBUTING.md) before opening a pull request.

## License

[MIT](LICENSE)
