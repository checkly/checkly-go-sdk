[![GoDoc](https://godoc.org/github.com/bitfield/checkly?status.png)](http://godoc.org/github.com/bitfield/checkly)[![Go Report Card](https://goreportcard.com/badge/github.com/bitfield/checkly)](https://goreportcard.com/report/github.com/bitfield/checkly)[![CircleCI](https://circleci.com/gh/bitfield/checkly.svg?style=svg)](https://circleci.com/gh/bitfield/checkly)

# checkly

`checkly` is a Go library for the [Checkly](https://checkly.com/) website monitoring service. It allows you to create new checks, get data on existing checks, and delete checks.

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
client = checkly.New(apiKey)
```

## Creating a new check

Once you have a client, you can use `client.Create()` to create a check. It returns the ID string of the newly-created check:

```go
params = checkly.Params{
        "name": "My awesome check",
        "checkType": "BROWSER",
        "activated": "true",
}
ID, err := client.Create(params)
if err != nil {
        log.Fatal(err)
}
fmt.Println(ID)
// Output: My awesome check
```

## Deleting a check

Use `client.Delete(ID)` to delete a check.

```go
err := client.Delete("73d29ea2-6540-4bb5-967e-e07fa2c9465e")
if err != nil {
    log.Fatal(err)
}
```

## Debugging

If things aren't working as you expect, you can assign an `io.Writer` to `client.Debug` to receive debug output. If `client.Debug` is non-nil, `MakeAPICall()` will dump the HTTP request and response to it:

```go
client.Debug = os.Stdout
p := checkly.Params{
    "frogurt": "cursed",
}
status, response, err := client.MakeAPICall(http.MethodGet, "monkeyPaw", p); if err != nil {
    log.Fatal(err)
}
```

outputs:

```
POST /v1/monkeyPaw HTTP/1.1
Host: api.checkly.com
User-Agent: Go-http-client/1.1
Content-Length: 52
Content-Type: application/x-www-form-urlencoded
Accept-Encoding: gzip

api_key=XXX&format=json&frogurt=cursed
...
```

## Bugs and feature requests

If you find a bug in the `checkly` client or library, please [open an issue](https://github.com/bitfield/checkly/issues). Similarly, if you'd like a feature added or improved, let me know via an issue.

Not all the functionality of the Checkly API is implemented yet.

Pull requests welcome!
