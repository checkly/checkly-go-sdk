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

Create a new `Client` object by calling `checkly.NewClient()` with your API key:

```go
client = checkly.New(apiKey)
```

Once you have a client, you can use it to create a check:

```go
params = Params{
        "name": "My awesome check",
        "checkType": "BROWSER",
}
check, err := client.CreateCheck(params)
if err != nil {
        log.Fatal(err)
}
fmt.Println(check.Name)
// Output: My awesome check
```

## Debugging

If things aren't working as you expect, you can assign an `io.Writer` to `client.Debug` to receive debug output. If `client.Debug` is non-nil, `MakeAPICall()` will dump the HTTP request and response to it:

```go
client.Debug = os.Stdout
r := checkly.Response{}
p := checkly.Params{
    "frogurt": "cursed",
}
if err := client.MakeAPICall("monkeyPaw", &r, p); err != nil {
    log.Fatal(err)
}
```

outputs:

```
POST /v2/monkeyPaw HTTP/1.1
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

Not all the functionality of the checkly API is implemented yet.

Pull requests welcome!
