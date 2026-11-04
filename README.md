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

Create a new checkly `Client` by calling `checkly.NewClient()` (you will need to set your Checkly API Key and Account ID)

```go
baseUrl := "https://api.checklyhq.com"
apiKey := os.Getenv("CHECKLY_API_KEY")
accountId := os.Getenv("CHECKLY_ACCOUNT_ID")

client := checkly.NewClient(
	baseUrl,
	apiKey,
	nil, //custom http client, defaults to http.DefaultClient
	nil, //io.Writer to output debug messages
)

client.SetAccountId(accountId)
```

> Note: if you don't have an API key, you can create one at [here](https://app.checklyhq.com/account/api-keys)

> Note: the debug writer receives dumps of the requests and responses. Only the API key in the `Authorization` header is redacted; the full request and response bodies are kept, and can hold secrets such as secret environment variable values and alert channel webhook URLs or keys. Do not share debug output.

### Create your first checks

Once you have a client, you can create a check. See here how to create your first API & Browser checks.

```go
apiCheck := checkly.Check{
	Name:                 "My API Check",
	Type:                 checkly.TypeAPI,
	Frequency:            5,
	Activated:            true,
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

browserCheck := checkly.Check{
	Name:          "My Browser Check",
	Type:          checkly.TypeBrowser,
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
client.CreateCheck(ctx, apiCheck)
client.CreateCheck(ctx, browserCheck)
```

>  A complete example program! You can see an example program which creates a Checkly check in the [demo](demo/main.go) folder.

### Handling errors

When the API answers with a status a method does not accept, the method returns a `*checkly.APIError`. Its text names the request, the status and the backend's message, for example `GET /v1/checks/abc: 404 Not Found`. Match it with `errors.Is` and the sentinel errors such as `checkly.ErrNotFound`, `checkly.ErrForbidden` or `checkly.ErrRateLimited`, or inspect it with `errors.As`, rather than matching the text:

```go
_, err := client.GetCheck(ctx, id)
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
