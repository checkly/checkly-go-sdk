[![GoDoc](https://godoc.org/github.com/bitfield/checkly?status.png)](http://godoc.org/github.com/bitfield/checkly)[![Go Report Card](https://goreportcard.com/badge/github.com/bitfield/checkly)](https://goreportcard.com/report/github.com/bitfield/checkly)[![CircleCI](https://circleci.com/gh/bitfield/checkly.svg?style=svg)](https://circleci.com/gh/bitfield/checkly)

# checkly

`checkly` is a Go library for the [Checkly](https://checklyhq.com/?utm_source=github&lmref=1374) website monitoring service. It allows you to create new checks, get data on existing checks, and delete checks.

While you can manage your Checkly checks entirely in Go code, using this library, you may prefer to use Terraform. In that case, you can use the Checkly Terraform provider (which in turn uses this library):

https://github.com/bitfield/terraform-provider-checkly

## Setting your API key

To use the client library with your Checkly account, you will need an API Key for the account. Go to the [Account Settings: API Keys page](https://app.checklyhq.com/account/api-keys) and click 'Create API Key'.

## Using the Go library

Import the library using:

```go
import "github.com/bitfield/checkly"
```

## Creating a client

Create a new `Client` object by calling `checkly.NewClient()` with your API key:

```go
apiKey := "3a4405dfb5894f4580785b40e48e6e10"
client := checkly.NewClient(apiKey)
```

Or read the key from an environment variable:

```go
client := checkly.NewClient(os.Getenv("CHECKLY_API_KEY"))
```

## Creating a new check

Once you have a client, you can create a check. First, populate a Check struct with the required parameters:

```go
check := checkly.Check{
		Name:      "My Awesome Check",
		Type:      checkly.TypeAPI,
		Frequency: 5,
		Activated: true,
		Locations: []string{"eu-west-1"},
		Request: checkly.Request{
			    Method: http.MethodGet,
			    URL:    "http://example.com",
		},
}
```

Now you can pass it to `client.Create()` to create a check. This returns the ID string of the newly-created check:

```go
ID, err := client.Create(check)
```

## Retrieving a check

`client.Get(ID)` finds an existing check by ID and returns a Check struct containing its details:

```go
check, err := client.Get("87dd7a8d-f6fd-46c0-b73c-b35712f56d72")
fmt.Println(check.Name)
// Output: My Awesome Check

```

## Deleting a check

Use `client.Delete(ID)` to delete a check by ID.

```go
err := client.Delete("73d29ea2-6540-4bb5-967e-e07fa2c9465e")
```

## A complete example program

You can see an example program which creates a Checkly check in the [examples/demo](examples/demo/main.go) folder.

## Debugging

If things aren't working as you expect, you can assign an `io.Writer` to `client.Debug` to receive debug output. If `client.Debug` is non-nil, then all API requests and responses will be dumped to the specified writer (for example, `os.Stderr`).

Regardless of the debug setting, if a request fails with HTTP status 400 Bad Request), the full response will be dumped (to standard error if no debug writer is set):

```go
client.Debug = os.Stderr
```

Example request and response dump:

```
POST /v1/checks HTTP/1.1
Host: api.checklyhq.com
User-Agent: Go-http-client/1.1
Content-Length: 852
Authorization: Bearer XXX
Content-Type: application/json
Accept-Encoding: gzip

{"id":"","name":"integrationTestCreate","checkType":"API","frequency":5,"activated":true,"muted":false,"shouldFail":false,"locations":["eu-west-1"],"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","environment_variables":null,"doubleCheck":false,"sslCheck":false,"sslCheckDomain":"example.com","alertChannels":{},"alertSettings":{"runBasedEscalation":{"failedRunThreshold":1},"timeBasedEscalation":{"minutesFailingThreshold":5},"reminders":{"interval":5},"sslCertificates":{"enabled":false,"alertThreshold":3}},"useGlobalAlertSettings":false,"request":{"method":"GET","url":"http://example.com","followRedirects":false,"body":"","bodyType":"NONE","headers":[],"queryParameters":[],"assertions":[{"edit":false,"order":0,"arrayIndex":0,"arraySelector":0,"source":"STATUS_CODE","property":"","comparison":"EQUALS","target":"200"}]}}

HTTP/1.1 201 Created
Transfer-Encoding: chunked
Cache-Control: no-cache
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Thu, 18 Jul 2019 15:48:21 GMT
Server: Cowboy
Vary: origin,accept-encoding
Via: 1.1 vegur

3c8
{"name":"integrationTestCreate","checkType":"API","frequency":5,"activated":true,"muted":false,"shouldFail":false,"locations":["eu-west-1"],"doubleCheck":false,"sslCheck":false,"sslCheckDomain":"example.com","alertSettings":{"runBasedEscalation":{"failedRunThreshold":1},"timeBasedEscalation":{"minutesFailingThreshold":5},"reminders":{"interval":5,"amount":0},"sslCertificates":{"enabled":false,"alertThreshold":3}},"useGlobalAlertSettings":false,"request":{"method":"GET","url":"http://example.com","followRedirects":false,"body":"","bodyType":"NONE","headers":[],"queryParameters":[],"assertions":[{"source":"STATUS_CODE","property":"","comparison":"EQUALS","target":"200"}],"basicAuth":{"username":"","password":"f15c6e9c867c529b74b9dd2f9585ba76:1c97b45322f1cd139122666eb13c7562"}},"setupSnippetId":null,"tearDownSnippetId":null,"localSetupScript":null,"localTearDownScript":null,"created_at":"2019-07-18T15:48:21.844Z","id":"763fa73d-1d14-4046-88e6-14f883ceddc9"}
0
```

## Bugs and feature requests

If you find a bug in the `checkly` client or library, please [open an issue](https://github.com/bitfield/checkly/issues). Similarly, if you'd like a feature added or improved, let me know via an issue.

Not all the functionality of the Checkly API is implemented yet.

Pull requests welcome!
