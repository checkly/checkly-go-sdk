<p>
  <img height="128" src="https://www.checklyhq.com/images/footer-logo.svg" align="right" />
  <h1>Checkly GO SDK</h1>
</p>

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

## Questions
For questions and support please open a new  [discussion](https://github.com/checkly/checkly-go-sdk/discussions). The issue list of this repo is exclusively for bug reports and feature/docs requests.

## Issues
Please make sure to respect issue requirements and choose the proper [issue template](https://github.com/checkly/checkly-go-sdk/issues/new/choose) when opening an issue. Issues not conforming to the guidelines may be closed.

## Contribution
Please make sure to read the [Contributing Guide](https://github.com/checkly/checkly-go-sdk/blob/main/CONTRIBUTING.md) before making a pull request.

## License

[MIT](https://github.com/checkly/checkly-go-sdk/blob/main/LICENSE)

<br>

<p align="center">
  <a href="https://checklyhq.com?utm_source=github&utm_medium=sponsor-logo-github&utm_campaign=headless-recorder" target="_blank">
  <img width="100px" src="https://www.checklyhq.com/images/text_racoon_logo.svg" alt="Checkly" />
  </a>
  <br />
  <i><sub>Delightful Active Monitoring for Developers</sub></i>
  <br>
  <b><sub>From Checkly with ♥️</sub></b>
<p>
