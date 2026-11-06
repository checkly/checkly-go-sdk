# Checkly Go SDK

![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)
[![Tests](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml/badge.svg)](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/checkly/checkly-go-sdk?status.png)](http://godoc.org/github.com/checkly/checkly-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/checkly/checkly-go-sdk)](https://goreportcard.com/report/github.com/checkly/checkly-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/checkly/checkly-go-sdk)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/checkly/checkly-go-sdk?label=Version)


This project is a Go SDK for [Checkly](https://checklyhq.com/?utm_source=github&lmref=1374) monitoring service. It allows you to handle your checks, check groups, snippets, environments variables and everything you can do with our [REST API](https://www.checklyhq.com/docs/api).

## Installation

To use the client library with your Checkly account, you will need an API Key for the account. Go to the [Account Settings: API Keys page](https://app.checklyhq.com/account/api-keys) and click 'Create API Key'.

Make sure your project is using Go Modules (it will have a go.mod file in its root if it already is):

```bash
$ go mod init
```

Then, add the reference of checkly-go-sdk in a Go program using `import`:
```go
import checkly "github.com/checkly/checkly-go-sdk"
```

Run any of the normal go commands (`build/install/test`) and the  Go toolchain will resolve and fetch the  checkly-go-sdk module automatically.

Alternatively, you can also explicitly go get the package into a project:

```bash
$ go get -u github.com/checkly/checkly-go-sdk
```

## Getting Started

Create a new checkly `Client` by calling `checkly.NewClient()` with options. Every option is optional: without any, the client talks to `checkly.DefaultBaseURL` through `http.DefaultClient` and has no auth, so only the public methods (`GetStaticIPs`, `GetRuntime`) work. Most methods need auth: an API key together with the ID of the account it acts on, a UUID.

`WithAuthFromEnv()` reads the API key and account ID from `CHECKLY_API_KEY` and `CHECKLY_ACCOUNT_ID`, the variables the Checkly CLI uses. Both are required, and `NewClient` fails if either is not set:

```go
client, err := checkly.NewClient(checkly.WithAuthFromEnv())
if err != nil {
	log.Fatal(err)
}
```

Or pass the credentials with `WithAuth(apiKey, accountID)`, which also requires both; for a client without auth, pass no auth option. `NewClient` returns an error for an invalid option, such as an empty value or an account ID that is not a UUID:

```go
client, err := checkly.NewClient(
	checkly.WithAuth(apiKey, accountID),
	// Optional: another API root, and a custom HTTP client, for example with
	// a timeout.
	checkly.WithBaseURL("https://api.checklyhq.com"),
	checkly.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
if err != nil {
	log.Fatal(err)
}
```

> Note: if you don't have an API key, you can create one at [here](https://app.checklyhq.com/account/api-keys)

### Create your first checks

Once you have a client, you can create a check. See here how to create your first API & Browser checks.

```go
apiCheck := checkly.APICheck{
	Name:                 "My API Check",
	Frequency:            5,
	Activated:            true,
	DegradedResponseTime: 5000,
	MaxResponseTime:      20000,
	Locations: []string{
		"eu-west-1",
		"ap-northeast-2",
	},
	Tags: []string{ "production" },
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "https://api.checklyhq.com/v1",
	},
}

browserCheck := checkly.BrowserCheck{
	Name:          "My Browser Check",
	Frequency:     5,
	Activated:     true,
	Locations:     []string{"eu-west-1"},
	Script: `const assert = require("chai").assert;
	const puppeteer = require("puppeteer");

	const browser = await puppeteer.launch();
	const page = await browser.newPage();
	await page.goto("https://example.com");
	const title = await page.title();

	assert.equal(title, "Example Site");
	await browser.close();`,
}

ctx := context.WithTimeout(context.Background(), time.Second * 5)
client.CreateAPICheck(ctx, apiCheck)
client.CreateBrowserCheck(ctx, browserCheck)
```

> Complete, runnable examples, such as creating an API check or a URL monitor, are in the [package documentation on pkg.go.dev](https://pkg.go.dev/github.com/checkly/checkly-go-sdk#pkg-examples).

### Handling errors

When the API answers with a status a method does not accept, the method returns a `*checkly.APIError`. Its text names the request, the status and the backend's message, for example `GET /v1/checks/abc: 404 Not Found`. Match it with `errors.Is` and the sentinel errors such as `checkly.ErrNotFound`, `checkly.ErrForbidden` or `checkly.ErrRateLimited`, or inspect it with `errors.As`, rather than matching the text:

```go
_, err := client.GetAPICheck(ctx, id)
if errors.Is(err, checkly.ErrNotFound) { // or checkly.IsNotFound(err)
	// The check no longer exists.
}
var apiErr *checkly.APIError
if errors.As(err, &apiErr) {
	log.Printf("Checkly API answered %d: %s", apiErr.StatusCode, apiErr.Message)
}
```

Errors for requests that got no response wrap their cause, so `errors.Is(err, context.DeadlineExceeded)` works too.

The SDK never adds the API key or the request body to an error. It does keep the backend's error message, and the start of an error body that is not JSON, cut at 500 bytes. A validation message can quote a submitted value, and so can the text of such a body, so take care when logging errors from requests that carry secrets.

## Questions
For questions and support please open a new  [discussion](https://github.com/checkly/checkly-go-sdk/discussions). The issue list of this repo is exclusively for bug reports and feature/docs requests.

## Issues
Please make sure to respect issue requirements and choose the proper [issue template](https://github.com/checkly/checkly-go-sdk/issues/new/choose) when opening an issue. Issues not conforming to the guidelines may be closed.

## Contribution
Please make sure to read the [Contributing Guide](https://github.com/checkly/checkly-go-sdk/blob/main/CONTRIBUTING.md) before making a pull request.

## License

[MIT](https://github.com/checkly/checkly-go-sdk/blob/main/LICENSE)
