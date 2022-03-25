<p align="center">
  <img width="300px" src="https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/1200px-Go_Logo_Blue.svg.png" alt="Golang" />
</p>


<p>
  <img height="128" src="https://www.checklyhq.com/images/footer-logo.svg" align="right" />
  <h1>Checkly GO SDK</h1>
</p>


[![Tests](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml/badge.svg)](https://github.com/checkly/checkly-go-sdk/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/checkly/checkly-go-sdk?status.png)](http://godoc.org/github.com/checkly/checkly-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/checkly/checkly-go-sdk)](https://goreportcard.com/report/github.com/checkly/checkly-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/checkly/checkly-go-sdk)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/checkly/checkly-go-sdk?label=Version)


> 🦦 Go SDK library for use with the Checkly API

<br>

## 👀 Overview

This project is a Go SDK for [Checkly](https://checklyhq.com/?utm_source=github&lmref=1374) monitoring service. It allows you to handle your checks, check groups, snippets, environments variables and everything you can do with our [REST API](https://www.checklyhq.com/docs/api).

While you can manage your Checkly account entirely in Go code, using this library, you may prefer to use Terraform. In that case, you can use the Checkly [Terraform provider](https://github.com/checkly/terraform-provider-checkly) (which is built on top of this library):

<br>

## 🔧 How to use?

To use the client library with your Checkly account, you will need an API Key for the account. Go to the [Account Settings: API Keys page](https://app.checklyhq.com/account/api-keys) and click 'Create API Key'.


### Import the SDK

```go
import checkly "github.com/checkly/checkly-go-sdk"
```

### Create a client

Create a new `Client` by calling `checkly.NewClient()` with your API key:

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

> ⚠️ Account ID is only required if you are using new User API keys. If you are using legacy Account API keys you can omit it.

> Note: if you don't have an API key, you can create one at [here](https://app.checklyhq.com/account/api-keys)

### Create a check

Once you have a client, you can create a check. First, populate a Check struct with the parameters you want:

```go
check := checkly.Check{
	Name:                 "My API Check",
	Type:                 checkly.TypeAPI,
	Frequency:            5,
	DegradedResponseTime: 5000,
	MaxResponseTime:      15000,
	Activated:            true,
	Muted:                false,
	ShouldFail:           false,
	DoubleCheck:          false,
	SSLCheck:             true,
	LocalSetupScript:     "",
	LocalTearDownScript:  "",
	Locations: []string{
		"eu-west-1",
		"ap-northeast-2",
	},
	Tags: []string{
		"foo",
		"bar",
	},
	AlertSettings:          alertSettings,
	UseGlobalAlertSettings: false,
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
	},
}
```

Now you can pass it to `client.Create()` to create a check. This returns the newly-created Check object, or an error if there was a problem:

```go
ctx := context.WithTimeout(context.Background(), time.Second * 5)
check, err := client.Create(ctx, check)
```

For browser checks, the options are slightly different:

```go
check := checkly.Check{
	Name:          "My Browser Check",
	Type:          checkly.TypeBrowser,
	Frequency:     5,
	Activated:     true,
	Muted:         false,
	ShouldFail:    false,
	DoubleCheck:   false,
	SSLCheck:      true,
	Locations:     []string{"eu-west-1"},
	AlertSettings: alertSettings,
	Script: `const assert = require("chai").assert;
	const puppeteer = require("puppeteer");

	const browser = await puppeteer.launch();
	const page = await browser.newPage();
	await page.goto("https://example.com");
	const title = await page.title();

	assert.equal(title, "Example Site");
	await browser.close();`,
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "HELLO",
			Value: "Hello world",
		},
	},
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
	},
}
```

### Retrieve a check

`client.Get(ctx, ID)` finds an existing check by ID and returns a Check struct containing its details:

```go
check, err := client.Get(ctx, "87dd7a8d-f6fd-46c0-b73c-b35712f56d72")
fmt.Println(check.Name)
// Output: My Awesome Check

```

### Update a check

`client.Update(ctx, ID, check)` updates an existing check with the specified details. For example, to change the name of a check:

```go
ID := "87dd7a8d-f6fd-46c0-b73c-b35712f56d72"
check, err := client.Get(ctx, ID)
check.Name = "My updated check name"
updatedCheck, err = client.Update(ctx, ID, check)
```

### Delete a check

Use `client.Delete(ctx, ID)` to delete a check by ID.

```go
err := client.Delete(ctx, "73d29ea2-6540-4bb5-967e-e07fa2c9465e")
```

### Create a check group

Checkly checks can be combined into a group, so that you can configure default values for all the checks within it:

```go
var wantGroup = checkly.Group{
	Name:        "test",
	Activated:   true,
	Muted:       false,
	Tags:        []string{"auto"},
	Locations:   []string{"eu-west-1"},
	Concurrency: 3,
	APICheckDefaults: checkly.APICheckDefaults{
		BaseURL: "example.com/api/test",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		BasicAuth: checkly.BasicAuth{
			Username: "user",
			Password: "pass",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	DoubleCheck:            true,
	UseGlobalAlertSettings: false,
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Amount:   0,
			Interval: 5,
		},
	},
	AlertChannelSubscriptions: []checkly.Subscription{
		{
			Activated: true,
		},
	},
	LocalSetupScript:    "setup-test",
	LocalTearDownScript: "teardown-test",
}
group, err := client.CreateGroup(ctx, wantGroup)
```

<br>

## 👌 A complete example program!

You can see an example program which creates a Checkly check in the [examples/demo](examples/demo/main.go) folder.

<br>

## 🧪 Testing

There are two different set of tests: unit test and integration tests. Both can be run with the `go test` command.

```bash
$ go test ./... # unit tests
$ go test ./... -tags=integration # integration tests
```

<br>

## 🐛 Debugging

If things aren't working as you expect, you can pass an `io.Writer` to `checkly.NewClient's fourth arg` to receive debug output. If `debug` is non-nil, then all API requests and responses will be dumped to the specified writer (for example, `os.Stderr`).

Regardless of the debug setting, if a request fails with HTTP status 400 Bad Request), the full response will be dumped (to standard error if no debug writer is set):

```go
debugOutput := os.Stderr
client.NewClient(
	"https://api.checklyhq.com",
	"your-api-key",
	nil,
	debugOutput,
)
```

Example request and response dump:

```
POST /v1/checks HTTP/1.1
Host: api-test.checklyhq.com
User-Agent: Go-http-client/1.1
Content-Length: 1078
Authorization: Bearer XXX
Content-Type: application/json
Accept-Encoding: gzip

{"id":"","name":"test","checkType":"API","frequency":10,"activated":true,
"muted":false,"shouldFail":false,"locations":["eu-west-1"],
"degradedResponseTime":15000,"maxResponseTime":30000,"script":"foo",
"environmentVariables":[{"key":"ENVTEST","value":"Hello world","locked":false}],
"doubleCheck":true,"tags":["foo","bar"],"sslCheck":true,
"localSetupScript":"setitup","localTearDownScript":"tearitdown","alertSettings":
{"escalationType":"RUN_BASED","runBasedEscalation":{"failedRunThreshold":1},
"timeBasedEscalation":{"minutesFailingThreshold":5},"reminders":{"interval":5},
"useGlobalAlertSettings":false,"request":{"method":"GET","url":"https://example.
com","followRedirects":false,"body":"","bodyType":"NONE","headers":[
{"key":"X-Test","value":"foo","locked":false}],"queryParameters":[
{"key":"query","value":"foo","locked":false}],"assertions":[
{"edit":false,"order":0,"arrayIndex":0,"arraySelector":0,
"source":"STATUS_CODE","property":"","comparison":"EQUALS",
"target":"200"}],"basicAuth":{"username":"","password":""}}}

HTTP/1.1 201 Created
Transfer-Encoding: chunked
Cache-Control: no-cache
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Thu, 28 May 2020 11:18:31 GMT
Server: Cowboy
Vary: origin,accept-encoding
Via: 1.1 vegur

4ea
{"name":"test","checkType":"API","frequency":10,"activated":true,"muted":false,
"shouldFail":false,"locations":["eu-west-1"],"degradedResponseTime":15000,
"maxResponseTime":30000,"script":"foo","environmentVariables":[{"key":"ENVTEST",
"value":"Hello world","locked":false}],"doubleCheck":true,"tags":["foo","bar"],
"sslCheck":true,"localSetupScript":"setitup","localTearDownScript":"tearitdown",
"alertSettings":{"escalationType":"RUN_BASED","runBasedEscalation":
{"failedRunThreshold":1},"timeBasedEscalation":{"minutesFailingThreshold":5},
"reminders":{"interval":5,"amount":0},"useGlobalAlertSettings":false,"request":{"method":"GET",
"url":"https://example.com","followRedirects":false,"body":"","bodyType":"NONE",
"headers":[{"key":"X-Test","value":"foo","locked":false}],"queryParameters":[
{"key":"query","value":"foo","locked":false}],"assertions":[
{"source":"STATUS_CODE","property":"","comparison":"EQUALS","target":"200"}],
"basicAuth":{"username":"","password":""}},"setupSnippetId":null,
"tearDownSnippetId":null,"groupId":null,"groupOrder":null,
"alertChannelSubscriptions":[{"activated":true,"alertChannelId":35}],
"created_at":"2020-05-28T11:18:31.280Z",
"id":"29815146-8ab5-492d-a092-9912c1ab8333"}
0
```

<br>

##  🚀 Release

Release process is automatically handled using tags and the `release` GitHub Action. To create a new release, you have to create and push a new version tag: `vX.X.X`

>  🔢 When creating a new tag, be sure to follow [SemVer](https://semver.org/).

<br>

## 📝 Bugs and feature requests

If you find a bug in the `checkly` client or library, please [open an issue](https://github.com/checkly/checkly-go-sdk/issues). Similarly, if you'd like a feature added or improved, let me know via an issue.

Not all the functionality of the Checkly API is implemented yet.

Pull requests welcome!

<br>

## 📄 License

[MIT](https://github.com/checkly/checkly-go-sdk/blob/main/LICENSE)

<p align="center">
  <a href="https://checklyhq.com?utm_source=github&utm_medium=sponsor-logo-github&utm_campaign=headless-recorder" target="_blank">
  <img width="100px" src="https://www.checklyhq.com/images/text_racoon_logo.svg" alt="Checkly" />
  </a>
  <br />
  <i><sub>Delightful Active Monitoring for Developers</sub></i>
  <br>
  <b><sub>From Checkly with ♥️</sub></b>
<p>
