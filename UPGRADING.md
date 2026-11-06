# Upgrading from v1 to v2

v2.0.0 is a major release. It changes how a client is built, returns typed errors, adds dedicated
types and methods for API, browser and multi-step checks, removes APIs that were deprecated or no
longer backed by the public API, and changes what some create and update requests send. This guide
lists every change from v1.27.0 that can affect your code, most disruptive first.

The module path is now `github.com/checkly/checkly-go-sdk/v2`:

```sh
go get github.com/checkly/checkly-go-sdk/v2@v2.0.0
```

```go
// v1
import "github.com/checkly/checkly-go-sdk"

// v2
import "github.com/checkly/checkly-go-sdk/v2"
```

The package name stays `checkly`, so code that refers to `checkly.X` only needs the new import path.

v2 requires Go 1.26 or later.

## Creating the client

`NewClient(baseURL, apiKey, httpClient, debug) Client` becomes
`NewClient(opts ...ClientOption) (Client, error)`, and `SetAccountId` and `SetChecklySource` are
removed. A client is not modified after construction and is safe for concurrent use.

```go
// v1
client := checkly.NewClient("https://api.checklyhq.com", apiKey, nil, nil)
client.SetAccountId(accountID)
client.SetChecklySource("my-tool")
```

```go
// v2
client, err := checkly.NewClient(
	checkly.WithAuth(apiKey, accountID),
	checkly.WithSource("my-tool"),
)
if err != nil {
	return err
}
```

| Option | Sets | Default |
|---|---|---|
| `WithBaseURL(url)` | API root; an http or https URL with a host and no query or fragment | `DefaultBaseURL` (`https://api.checklyhq.com`) |
| `WithHTTPClient(c)` | HTTP client that sends requests | `http.DefaultClient` |
| `WithSource(s)` | `x-checkly-source` header | `go-sdk` |
| `WithAuth(apiKey, accountID)` | API key and the ID of the account it acts on | no auth |
| `WithAuthFromEnv()` | API key and account ID from `CHECKLY_API_KEY` and `CHECKLY_ACCOUNT_ID` | no auth |

- Every option is optional. Options apply in order and a later one wins; an empty or nil value
  restores the default of `WithBaseURL`, `WithHTTPClient` and `WithSource`.
- `NewClient` returns a nil `Client` and the error of the first invalid option. The error starts
  with `checkly: ` and names the option, and never includes the API key or account ID.
- **Every API key now needs its account ID.** v1 only required it for user keys. `WithAuth` with an
  empty key or account ID, `WithAuthFromEnv` with either variable unset, and an account ID that is
  not a UUID (for example the API key passed in its place) all fail in `NewClient`:

  | Error matches | When |
  |---|---|
  | `ErrMissingAPIKey` | empty API key, or `CHECKLY_API_KEY` not set |
  | `ErrMissingAccountID` | empty account ID, or `CHECKLY_ACCOUNT_ID` not set; replaces v1's `missing Checkly Account ID` request error |
  | `ErrInvalidAccountID` | account ID is not a UUID |

- `NewClient` never reads the environment unless you pass `WithAuthFromEnv`.
- A client without auth can call `GetStaticIPs` and `GetRuntime`, whose endpoints are public; they
  no longer send credentials even when the client has them. Every other method fails with an error
  matching `ErrMissingAPIKey` without sending the request.
- The `debug` argument is gone, and the client no longer dumps requests and responses (the dumps
  contained the API key and secret values). To inspect API traffic, route it through an
  intercepting HTTPS proxy, for example with the `HTTPS_PROXY` environment variable.

## Errors

The text of API and transport errors has changed. Code that matched error strings,
such as the `unexpected response status 404:` prefix, must switch to `errors.Is` and `errors.As`.

```go
// v1
monitor, err := client.GetURLMonitor(ctx, id)
if err != nil {
	if strings.Contains(err.Error(), "unexpected response status 404") {
		return nil // already deleted
	}
	return err
}
log.Println(monitor.Name)
```

```go
// v2
monitor, err := client.GetURLMonitor(ctx, id)
if err != nil {
	if errors.Is(err, checkly.ErrNotFound) {
		return nil // already deleted
	}
	var apiErr *checkly.APIError
	if errors.As(err, &apiErr) {
		log.Printf("status %d, code %q: %s", apiErr.StatusCode, apiErr.Code, apiErr.Message)
	}
	return err
}
log.Println(monitor.Name)
```

- **`*APIError`** is returned for every response with a status the method does not accept. Its
  fields are `Method`, `Path`, `StatusCode`, `Code` (the backend's error code), `Message`,
  `Details` (the fields a validation error names) and `RetryAfter` (for a 429). Its text reads
  `GET /v1/checks/abc: 404 Not Found`, followed by the error code, message and keys when present.
- **Sentinels**, matched with `errors.Is`:

  | Sentinel | Matches |
  |---|---|
  | `ErrBadRequest` | any 400 |
  | `ErrValidation` | 400 with the API's validation block, sent when a request fails schema validation; also matches `ErrBadRequest` |
  | `ErrNotFound` (also `IsNotFound(err)`) | 404 |
  | `ErrUnauthorized` | 401: the server rejected the API key |
  | `ErrForbidden` | any 403 |
  | `ErrAccountNotAccessible` | 403 with code `ERR_ACCOUNT_NOT_ACCESSIBLE`; also matches `ErrForbidden` |
  | `ErrPlanLimit` | 402 |
  | `ErrConflict` | 409 |
  | `ErrRateLimited` | 429 still rate limited after the client's retries |
  | `ErrCheckTypeMismatch` | a typed getter got the ID of a check of another type (not an `*APIError`) |
  | `ErrUnsupportedCheckType` | `GetCheck` got the ID of a check of a type the SDK has no type for (not an `*APIError`) |
  | `ErrMissingAPIKey`, `ErrMissingAccountID`, `ErrInvalidAccountID` | client construction and missing auth (not `*APIError`s, and they do not match `ErrUnauthorized`) |

- **`ErrCodeBundleNotFound`** is now returned wrapped together with the `*APIError`; replace
  `err == checkly.ErrCodeBundleNotFound` with `errors.Is(err, checkly.ErrCodeBundleNotFound)`.
  Only a 404, or the 403 the backend returns for another account's key, counts as not found.
- **Transport errors** wrap their cause, so `errors.Is(err, context.DeadlineExceeded)` and
  `errors.As` with a `*url.Error` work.
- **Cancelling the context during a rate-limit wait** returns an error that matches both
  `ErrRateLimited` and the context error; compare with `errors.Is`, not `==`.
- Errors never quote request payloads or success response bodies; a decode error names the
  request, status and body size instead. An error response body is kept as `Message`: its JSON
  `message`, or the raw body when it is not JSON, cut at 500 bytes.

## Checks

v2 adds `APICheck` and `BrowserCheck`, redefines `MultiStepCheck`, and gives each its own methods
backed by the per-type endpoints, which are now the only way to create and update a check.
`GetCheck` and `DeleteCheck` remain for an ID whose check type is not known.

| Check type | Type | Methods | Create and update endpoint |
|---|---|---|---|
| API | `APICheck` | `CreateAPICheck`, `GetAPICheck`, `UpdateAPICheck`, `DeleteAPICheck` | `/v1/checks/api` |
| Browser | `BrowserCheck` | `CreateBrowserCheck`, `GetBrowserCheck`, `UpdateBrowserCheck`, `DeleteBrowserCheck` | `/v1/checks/browser` |
| Multi-step | `MultiStepCheck` | `CreateMultiStepCheck`, `GetMultiStepCheck`, `UpdateMultiStepCheck`, `DeleteMultiStepCheck` | `/v1/checks/multistep` |
| Heartbeat | `HeartbeatMonitor` | `CreateHeartbeatMonitor`, `GetHeartbeatMonitor`, `UpdateHeartbeatMonitor`, `DeleteHeartbeatMonitor` | `/v1/checks/heartbeat` |

The monitor types (`URLMonitor`, `TCPMonitor`, `DNSMonitor`, `ICMPMonitor`, `GRPCMonitor`,
`TracerouteMonitor`, `SSLMonitor`) and `PlaywrightCheck` keep their methods.

The `Client` interface gains the twelve methods in the first three rows, so your own types that
implement `Client`, such as mocks or wrappers, must add them. They must also change `GetCheck`,
which now returns `(Check, error)`.

```go
// v1
check, err := client.CreateCheck(ctx, checkly.Check{
	Name:                 "API health",
	Type:                 checkly.TypeAPI,
	Frequency:            5,
	Activated:            true,
	Locations:            []string{"eu-west-1"},
	DegradedResponseTime: 5000,
	MaxResponseTime:      10000,
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/health",
	},
})
if err != nil {
	return err
}
log.Println(check.ID)
```

```go
// v2
check, err := client.CreateAPICheck(ctx, checkly.APICheck{
	Name:                 "API health",
	Frequency:            5,
	Activated:            true,
	Locations:            []string{"eu-west-1"},
	DegradedResponseTime: 5000,
	MaxResponseTime:      10000,
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "https://example.com/health",
	},
})
if err != nil {
	return err
}
log.Println(check.ID)
```

Moving from `Check` to a per-type struct:

- There is no `Type` field; the method sets the check type.
- `AlertSettings` is a `*AlertSettings`. `LocalSetupScript` and `LocalTearDownScript` on `APICheck`
  are `*string`.
- `APICheck` has no `Script` or `EnvironmentVariables`: the API check endpoint ignores them.
  `BrowserCheck` and `MultiStepCheck` have `Script`, `ScriptPath`, `Dependencies` and
  `EnvironmentVariables`; `BrowserCheck` has no `Request` or response-time limits.
- There is no `DoubleCheck`: the per-type payloads always send `false`. Use `RetryStrategy`.
- Per-type updates send what the struct holds. `Tags`, `AlertChannelSubscriptions`, the snippet IDs
  and the local scripts, which v1's generic update left out when empty so the API kept them, are
  now cleared. Read the check, change it and send it back: a sparse struct can drop the alert
  channel subscriptions and stop the check's alerts with no error.
- `MultiStepCheck` no longer has `Type` or `FrequencyOffset`.
- New on `APICheck`, `BrowserCheck`, `MultiStepCheck` and `PlaywrightCheck`: `Intent`
  (`*CheckIntent`) and `AIAutoRepairEnabled` (`*bool`), for automatic AI repair.

**Typed getters reject other check types.** `GetAPICheck`, `GetBrowserCheck`, `GetMultiStepCheck`,
`GetPlaywrightCheck`, `GetHeartbeatMonitor` and the URL, TCP, DNS, ICMP, gRPC, traceroute and SSL
monitor getters return an error matching `ErrCheckTypeMismatch` when the ID belongs to a check of
another type. In v1, the Playwright, heartbeat and monitor getters decoded the check anyway and
dropped its type-specific fields.

**`GetCheck` returns the check's own type.** Use it for an ID whose check type you don't know.
`Check` is now an interface instead of a struct, holding a pointer to one of the twelve check and
monitor types, so read it with a type switch. `Check` exposes no fields, so even shared fields such
as `Name` need a case per type, and a `default` case covers types added in later versions. A check
of a type the SDK has no type for, such as
`AGENTIC`, is an error matching `ErrUnsupportedCheckType`; `errors.As` with
`*UnsupportedCheckTypeError` gives its `ID` and `Type`.

```go
// v1
check, err := client.GetCheck(ctx, id)
if err != nil {
	return err
}
if check.Type == checkly.TypeAPI {
	log.Println(check.Request.URL)
}
```

```go
// v2
check, err := client.GetCheck(ctx, id)
if err != nil {
	return err
}
switch c := check.(type) {
case *checkly.APICheck:
	log.Println(c.Request.URL)
case *checkly.URLMonitor:
	log.Println(c.Request.URL)
default:
	log.Printf("unhandled check type %T", check)
}
```

## Check groups

`Group`, `CreateGroup`, `GetGroup`, `UpdateGroup` and `DeleteGroup` are removed, because the API
deprecated the v1 create and update endpoints they used. Use `GroupV2`, `CreateGroupV2`, `GetGroupV2`,
`UpdateGroupV2` and `DeleteGroupV2`.

- **If you already use `GroupV2`,** the type and its methods keep their names and field types, and
  the rest of this section doesn't apply. Changes elsewhere in this guide still reach it, such as
  [renamed constants](#renamed-identifiers-and-changed-types) in its alert settings and assertions.
  `Tags` and `AlertChannelSubscriptions` are now sent as `[]` when nil, and
  [the timestamp JSON keys change](#decoding-fixes).
- **If you use `Group`,** move to `GroupV2`, which behaves differently, as this section describes.
  It is `Group` without `DoubleCheck`, with the fields in
  [the table below](#renamed-identifiers-and-changed-types) changed to pointers.

```go
// v1
group, err := client.CreateGroup(ctx, checkly.Group{
	Name:             "API checks",
	Concurrency:      3,
	Locations:        []string{"us-east-1"},
	RunParallel:      true,
	DoubleCheck:      true,
	AlertSettings:    settings,
	LocalSetupScript: "setup()",
})
```

```go
// v2
group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
	Name:                   "API checks",
	Concurrency:            3,
	Locations:              []string{"us-east-1"},
	RunParallel:            new(true),
	RetryStrategy:          &checkly.RetryStrategy{Type: checkly.RetryStrategyTypeSingleRetry},
	UseGlobalAlertSettings: new(false),
	AlertSettings:          &settings,
	LocalSetupScript:       new("setup()"),
})
```

**A group that leaves these settings unset behaves differently.** A `Group` literal that never set
`RunParallel`, `UseGlobalAlertSettings` or `AlertSettings` needs only the rename to compile as a
`GroupV2`, but then sends `null` for each, so each check keeps its own setting. `Group` sent `false`
for the first two and, on update, an empty `AlertSettings`, which replaced the settings of every
check in the group.
To keep overriding the checks, set the fields explicitly: a non-nil value replaces the setting of
every check in the group. `AlertSettings` must be complete: the API answers
`&checkly.AlertSettings{}` with a 400, on create and update, so set at least the threshold of its
escalation type, such as `RunBasedEscalation.FailedRunThreshold`.

**A group that sets `AlertSettings` needs `UseGlobalAlertSettings` too.** `Group` sent `false`
implicitly, which took every check in the group off the account's alert settings. With
`AlertSettings` alone, the key is `null`, so a check that uses global alert settings keeps them and ignores the group's. To
override those checks as well, set `UseGlobalAlertSettings` to `new(false)`, as in the example above.

**Group alert channels need one of these settings.** The API applies a group's
`AlertChannelSubscriptions` only when `UseGlobalAlertSettings` is non-nil or `AlertSettings` is set.
With both nil, each check uses its own subscriptions, so a group that only has channels of its own
silently stops alerting them. If your group has alert channels, set `UseGlobalAlertSettings`:
`new(false)` with `AlertSettings`, or `new(true)` to use the account's alert settings.

With `GroupV2`, creating a group without locations succeeds, and each check then keeps its own
locations. `Group` answered a 400.

**Snippet IDs and local scripts are sent as given.** In `Group`, `0` and `""` meant none. A pointer
to `0` or `""` is sent literally, so leave `SetupSnippetID`, `TearDownSnippetID`, `LocalSetupScript` and
`LocalTearDownScript` nil for none.

**A nil `RetryStrategy` still means no retries.** The API's default of `FALLBACK` applies only to
a request that leaves the key out, and the SDK always sends it: nil is sent as `null`, and the group
reads back with no retry strategy, so its checks don't retry. To let each check use its own retry
strategy, set `&checkly.RetryStrategy{Type: checkly.RetryStrategyTypeFallback}`.

**`DoubleCheck` is gone.** Replace `DoubleCheck: true` with
`&checkly.RetryStrategy{Type: checkly.RetryStrategyTypeSingleRetry}`, leaving the other fields zero,
which every plan can use. Other retry strategies need the retry feature on the account's plan.

## Removed APIs

| Removed | Use instead |
|---|---|
| `Create`, `Get`, `Update`, `Delete` | The per-type methods, such as `CreateAPICheck` |
| `CreateHeartbeat`, `UpdateHeartbeat`, `GetHeartbeatCheck` | `CreateHeartbeatMonitor`, `UpdateHeartbeatMonitor`, `GetHeartbeatMonitor` |
| `CreateTCPCheck`, `UpdateTCPCheck`, `GetTCPCheck` | `CreateTCPMonitor`, `UpdateTCPMonitor`, `GetTCPMonitor` |
| Type aliases `HeartbeatCheck`, `TCPCheck` | `HeartbeatMonitor`, `TCPMonitor` |
| `CreateTriggerCheck`, `GetTriggerCheck`, `DeleteTriggerCheck`, `CreateTriggerGroup`, `GetTriggerGroup`, `DeleteTriggerGroup`, types `TriggerCheck`, `TriggerGroup` | To run checks on demand, the Checkly CLI (`checkly trigger`) or `POST /v1/check-sessions/trigger` |
| `SetAccountId`, `SetChecklySource`, `NewClient`'s `debug` argument | See [Creating the client](#creating-the-client) |
| `AlertChannel.GetConfig`, `AlertChannel.SetConfig`, `AlertChannelConfigFromJSON` | The config field that matches `Type`, such as `Email` for `AlertChannelTypeEmail` |
| `CreateCheck`, `UpdateCheck` | The per-type methods, such as `CreateAPICheck` and `UpdateAPICheck` |
| `Check.DoubleCheck` | `RetryStrategy` |
| `Group`, `CreateGroup`, `GetGroup`, `UpdateGroup`, `DeleteGroup` | `GroupV2`, `CreateGroupV2`, `GetGroupV2`, `UpdateGroupV2`, `DeleteGroupV2`. See [Check groups](#check-groups) |
| `Group.DoubleCheck` | `RetryStrategy`. See [Check groups](#check-groups) |
| `Check.SSLCheck`, `AlertSettings.SSLCertificates`, type `SSLCertificates` | Nothing: the public API no longer has them |
| `Assertion.Edit`, `Assertion.Order`, `Assertion.ArrayIndex`, `Assertion.ArraySelector` | Nothing: the API discarded them, and requests no longer send them |

`Create` posted every check to `/v1/checks`. Its replacements post to the endpoint of the check's
type.

```go
// v1
channel := checkly.AlertChannel{Type: checkly.AlertTypeEmail}
channel.SetConfig(&checkly.AlertChannelEmail{Address: "ops@example.com"})
created, err := client.CreateAlertChannel(ctx, channel)
```

```go
// v2
channel := checkly.AlertChannel{
	Type:  checkly.AlertChannelTypeEmail,
	Email: &checkly.AlertChannelEmail{Address: "ops@example.com"},
}
created, err := client.CreateAlertChannel(ctx, channel)
```

Read a channel's config the same way, from `created.Email`, instead of `GetConfig`.

## Renamed identifiers and changed types

Exported names now use Go's initialism casing. JSON tags are unchanged.

| v1 | v2 |
|---|---|
| `SSLRequest.SSLClientCertificateId` | `SSLRequest.SSLClientCertificateID` |
| `Dashboard.CustomUrl` | `Dashboard.CustomURL` |
| `PrivateLocationKey.Id` | `PrivateLocationKey.ID` |
| `DashboardKey.Id` | `DashboardKey.ID` |
| type `ApiCheckResult`, field `CheckResult.ApiCheckResult` | type `APICheckResult`, field `CheckResult.APICheckResult` |

Constants for a fixed set of API values are typed, and named after their type, like the existing
`IncidentSeverityMinor`. Their values are unchanged.

| v1 | v2 |
|---|---|
| `StatusCode`, `JSONBody`, `TextBody`, `Headers`, `ResponseTime`, `ResponseData`, `Certificate`, `Connection`, `JSONResponse`, `TextResponse` | `AssertionSourceStatusCode`, `AssertionSourceJSONBody`, `AssertionSourceTextBody`, `AssertionSourceHeaders`, `AssertionSourceResponseTime`, `AssertionSourceResponseData`, `AssertionSourceCertificate`, `AssertionSourceConnection`, `AssertionSourceJSONResponse`, `AssertionSourceTextResponse` |
| `Equals`, `NotEquals`, `IsEmpty`, `NotEmpty`, `GreaterThan`, `LessThan`, `Contains`, `NotContains`, `IsNull`, `NotNull` | `AssertionComparisonEquals`, `AssertionComparisonNotEquals`, `AssertionComparisonIsEmpty`, `AssertionComparisonNotEmpty`, `AssertionComparisonGreaterThan`, `AssertionComparisonLessThan`, `AssertionComparisonContains`, `AssertionComparisonNotContains`, `AssertionComparisonIsNull`, `AssertionComparisonNotNull` |
| `RunBased`, `TimeBased` | `EscalationTypeRunBased`, `EscalationTypeTimeBased` |
| `AlertTypeEmail`, `AlertTypeSlack`, `AlertTypeSlackApp`, `AlertTypeWebhook`, `AlertTypeSMS`, `AlertTypePagerduty`, `AlertTypeOpsgenie`, `AlertTypeCall` | `AlertChannelTypeEmail`, `AlertChannelTypeSlack`, `AlertChannelTypeSlackApp`, `AlertChannelTypeWebhook`, `AlertChannelTypeSMS`, `AlertChannelTypePagerduty`, `AlertChannelTypeOpsgenie`, `AlertChannelTypeCall` |
| `TypeAPI`, `TypeBrowser`, `TypeHeartbeat` | `CheckTypeAPI`, `CheckTypeBrowser`, `CheckTypeHeartbeat` |

| Field | v1 type | v2 type |
|---|---|---|
| `Dashboard.Paginate` | `bool` | `*bool`; nil gets the API default, `true` |
| `MultiStepCheck.AlertSettings` | `AlertSettings` | `*AlertSettings` |
| `Group.RunParallel` | `bool` | `*bool` on `GroupV2`; nil leaves each check its own setting |
| `Group.UseGlobalAlertSettings` | `bool` | `*bool` on `GroupV2`; nil leaves each check its own setting |
| `Group.AlertSettings` | `AlertSettings` | `*AlertSettings` on `GroupV2`; nil leaves each check its own |
| `Group.SetupSnippetID`, `Group.TearDownSnippetID` | `int64` | `*int64` on `GroupV2`; nil means none |
| `Group.LocalSetupScript`, `Group.LocalTearDownScript` | `string` | `*string` on `GroupV2`; nil means none |
| `Assertion.Source` | `string` | `AssertionSource` |
| `Assertion.Comparison` | `string` | `AssertionComparison` |
| `AlertSettings.EscalationType` | `string` | `EscalationType` |
| `AlertChannel.Type` | `string` | `AlertChannelType` |
| `Request.BodyType` | `string` | `BodyType` |
| `IPFamily` of `Request`, `URLRequest`, `TCPRequest`, `ICMPRequest`, `TracerouteRequest`, `GRPCRequest` and `SSLConfig` | `string` | `IPFamily` |
| `DNSRequest.RecordType` | `string` | `DNSRecordType` |
| `DNSRequest.Protocol` | `string` | `DNSProtocol` |
| `TracerouteRequest.Protocol` | `string` | `TracerouteProtocol` |
| `GRPCConfig.Mode` | `string` | `GRPCMode` |
| `GRPCConfig.Encoding` | `string` | `GRPCEncoding` |
| `GRPCConfig.ServiceDefinition` | `string` | `GRPCServiceDefinition` |
| `Heartbeat.PeriodUnit`, `Heartbeat.GraceUnit` | `string` | `HeartbeatUnit` |
| `SSLConfig.ClientCertificateMode` | `string` | `SSLClientCertificateMode` |
| `SSLBaselineTLSRule.Value` | `string` | `TLSVersion` |
| `Severity` of `SSLBaselineTLSRule`, `SSLBaselineKeySizeRule` and `SSLBaselineSeverityRule` | `string` | `SSLBaselineSeverity` |
| `PlaywrightCheck.Engine` | `*string` | `*PlaywrightEngine` |
| `AlertChannelOpsgenie.Region` | `string` | `OpsgenieRegion` |
| `AlertChannelOpsgenie.Priority` | `string` | `OpsgeniePriority` |
| `RetryStrategy.Type` | `string` | `RetryStrategyType` |
| `RetryStrategy.OnlyOn` | `[]string` | `[]RetryOnlyOn` |
| `MaintenanceWindow.RepeatUnit` | `string` | `MaintenanceWindowRepeatUnit` |
| `MaintenanceWindowStatusPageVisibility.Severity` | `string` | `IncidentSeverity` |
| `Dashboard.Width` | `string` | `DashboardWidth` |
| `Runtime.Stage` | `string` | `RuntimeStage` |

Untyped constants and string literals, such as `"IPv4"` or `checkly.IPFamilyIPv4`, still compile in these
fields. A `string` variable needs a conversion, such as `checkly.IPFamily(s)`, and so does
passing one of these fields to a `string` parameter or comparing it with a `string` variable. A
`*string` engine converts with `(*checkly.PlaywrightEngine)(p)`; a `[]string` cannot be converted,
so build `OnlyOn` as a `[]checkly.RetryOnlyOn`. Check result fields, such as
`GRPCCheckResponse.GRPCMode` and `SSLCheckResult.TLSVersion`, stay `string`.

New constants cover the rest of each value set: the DNS, ICMP, traceroute and gRPC assertion
sources (such as `AssertionSourceResponseCode` and `AssertionSourceGRPCStatusCode`), the comparisons
`AssertionComparisonHasKey`, `AssertionComparisonNotHasKey`, `AssertionComparisonHasValue` and
`AssertionComparisonNotHasValue`, and a `CheckType` constant for every check type, including
`CheckTypeAgentic`. The new `MaintenanceWindowComponentImpact.Status` is a
`StatusPageComponentStatusV3`.

## Check results

`GetCheckResults` moves to the v2 endpoint, which pages with a cursor. It returns a
`*CheckResultsPage` instead of `[]CheckResult`, and `CheckResultsFilter.Page` is replaced by
`CheckResultsFilter.NextID`. The number of results is `len(page.Entries)`. The signature of
`GetCheckResult` is unchanged.

```go
// v1
for page := int64(1); ; page++ {
	results, err := client.GetCheckResults(ctx, checkID, &checkly.CheckResultsFilter{
		Limit: 100,
		Page:  page,
	})
	if err != nil {
		return err
	}
	if len(results) == 0 {
		break
	}
	process(results)
}
```

```go
// v2
filter := checkly.CheckResultsFilter{
	Limit: 100,
	From:  time.Now().Add(-6 * time.Hour).Unix(),
}
for {
	page, err := client.GetCheckResults(ctx, checkID, &filter)
	if err != nil {
		return err
	}
	process(page.Entries)
	if page.NextID == "" {
		break
	}
	filter.NextID = page.NextID
}
```

- There is no longer a default window, or a six-hour limit between `From` and `To`. Where v1 read
  the last six hours, a nil filter now spans all the raw results the account retains, which depends
  on its plan. The v2 loop above sets `From` to keep reading six hours.
- A filter with only `From` or only `To` no longer gets its counterpart six hours away: the other
  end stays open. Set both for a bounded window.
- On the free plan, a `From` or `To` older than the plan's retention no longer answers 402. On every
  plan, the API returns only the results it retains.
- A wide query can take too long, in which case the API answers 408. Narrow `From` and `To`,
  especially with `HasFailures`. To detect it, use `errors.As` to get the `*APIError` and compare
  its `StatusCode` with `http.StatusRequestTimeout`.
- The endpoint allows 60 requests per 60 seconds, and a request that sets `HasFailures` counts as
  five. A loop over many pages can reach that limit; set `Limit` to get more results per request.
- `CheckType` is sent for every type. v1 sent only `API` and `BROWSER` and dropped other types,
  returning unfiltered results. The API answers 404 when the check is not of the given type.
- `HasFailures` now works. v1 sent `hasFailures=1`, which the API rejects with a 400.

## Behaviour changes on create and update

Most update endpoints keep the stored value of a key the request leaves out. v1 left out many empty
fields, so an update could not clear them. v2 sends them, so **an update replaces the stored
values with the struct you pass**: a field you leave empty is cleared or reset. To keep a value set
elsewhere, such as in the web app, read the resource, change it and send it back.

Fields now always sent, as `[]`, `null`, `""` or `0` when empty:

- **Heartbeat, TCP, URL, DNS, ICMP, gRPC, traceroute and SSL monitors:** `tags` and
  `alertChannelSubscriptions` as `[]`; an unset `groupOrder` as `null` (as `0` for TCP monitors;
  heartbeat monitors have no group order).
- **TCP, URL and DNS monitors:** `degradedResponseTime` and `maxResponseTime`. **ICMP monitors:**
  `degradedPacketLossThreshold` and `maxPacketLossThreshold`. A zero value is no longer replaced by
  the API's default (and the API rejects a packet loss threshold of 0), so set these explicitly.
- **TCP monitors:** `request.assertions` as `[]`.
- **gRPC monitors:** an empty `GRPCConfig.Mode` is sent as `BEHAVIOR`. On a stored `HEALTH` monitor,
  an update with an empty mode now switches it to `BEHAVIOR` (or fails with a 400 when `Method` is
  empty); set `Mode` to keep `HEALTH`. With `ServiceDefinition` `REFLECTION`, an empty
  `ProtoContent` is sent as `""`, which clears stored proto content. On update only, the request
  timeout is sent as `null` when zero.
- **Traceroute monitors:** on update only, the request assertions as `[]`, and the protocol, port,
  max hops, max unknown hops and timeout as `null` when zero.
- **SSL monitors:** the request assertions as `[]`, and `SSLClientCertificateID` and
  `SSLConfig.ServerName` as `null` when nil.
- **Check groups:** `tags` and `alertChannelSubscriptions` as `[]`. Code moving from `Group` to
  `GroupV2` also sends `setupSnippetId`, `tearDownSnippetId`, `localSetupScript` and
  `localTearDownScript` as `null`.
- **`Request` (API checks, URL monitors):** `BasicAuth` as `null` when nil, so removing
  credentials clears them.
- **`AlertSettings`:** `reminders.amount` and `parallelRunFailureThreshold.enabled`, on create too.
- **Status pages (v1):** `customDomain` as `null`, and `logo`, `redirectTo` and `favicon` as `""`.
- **Dashboards:** `UpdateDashboard` sends empty `CustomDomain`, `Logo`, `Link`, `Description` and
  `Favicon` as `null` and an empty `Header` as `""`. `CreateDashboard` still leaves them out.
- **`APICheck`, `BrowserCheck`, `MultiStepCheck`, `PlaywrightCheck`:** `intent` and
  `aiAutoRepairEnabled` as `null` when nil, so an update clears an intent or override set elsewhere.
  Nil `CheckIntent.RequiredOutcomes` and `MustPreserve` are sent as `[]`.
- **`PrivateLocation`:** `proxyUrl` as `null` when `ProxyURL` is nil or empty, which removes the
  proxy.
- **`DNSMonitor`, `ICMPMonitor`, `TracerouteMonitor`:** the new `PrivateLocations` field is sent as
  `[]` when nil, as on the other check types, so an update clears the stored private locations.

Fields now left out so the API default applies:

- **`PlaywrightCheck.Browsers`:** nil or empty is left out, so the API's default browser applies.
  v1 sent `[]`, which the API rejected. On update this resets the browsers to the default.
- **Dashboards:** `RefreshRate` and `PaginationRate` are left out when zero (the API defaults both
  to 60 seconds; v1 sent 0, which the API rejected). `Paginate` and the new `ShowGroupNames` are
  left out when nil (default `true`), and the new `SortByStatus` when `false`. Because omitted keys
  get their default on update too, an update without them resets them.
- **`AlertChannel.AutoSubscribe`** (new) is left out when nil; the API then sets it to `false`, on
  update too.
- **`MaintenanceWindowStatusPageVisibility.ComponentConfigs`** (new) is left out when nil, so an
  update keeps the stored configs; an empty list removes them.

Lists that the API rejected as `null` are now sent as `[]`, so leaving them unset no longer fails
with a 400: `Request.Headers`, `Request.QueryParameters` and `Request.Assertions`,
`URLRequest.Assertions`, `DNSRequest.Assertions`, `ICMPRequest.Assertions`, `GroupV2.Tags`,
`StatusPage.Cards` and `StatusPageCard.Services`.

Other new fields: `Assertion.Regex` (for `HEADERS` assertions; always sent, so updates no longer
drop a regex set elsewhere) and the maintenance window types `MaintenanceWindowComponentConfig` and
`MaintenanceWindowComponentImpact`.

## Decoding fixes

- `CheckResult.ResponseTime` is a `float64` instead of an `int64`, in `GetCheckResult` and
  `GetCheckResults`. ICMP and traceroute results report a fractional average latency, which failed to
  decode in v1. Code using it as an `int64` needs a conversion, which truncates sub-millisecond values
  to 0, and `%d` no longer formats it.
- `CreatedAt` and `UpdatedAt` on `HeartbeatMonitor`, `GroupV2`, `Snippet` and `MultiStepCheck` are
  now populated: they decode from the `created_at` and `updated_at` keys the API returns. If you marshal these structs yourself, their JSON keys change accordingly.
- Alert channels returned by `CreateAlertChannel`, `GetAlertChannel` and `UpdateAlertChannel` have
  `CreatedAt` and `UpdatedAt` set; they were always empty.
- Alert channel responses with a `null` or wrongly typed field no longer panic. `null` leaves the
  field unset. A wrongly typed top-level field, or a fractional `id` or `sslExpiryThreshold`,
  returns an error naming the field; a wrongly typed field inside the config is skipped; only a
  config that is not a JSON object, or one for an unknown channel type, returns an error.
- `CreateDashboard`, `CreateMaintenanceWindow` and `CreatePrivateLocation` check the status before
  decoding, so an error page that is not JSON returns an `*APIError` instead of a JSON syntax error.

## Deprecated APIs

`CreateStatusPage`, `GetStatusPage`, `UpdateStatusPage`, `DeleteStatusPage`,
`CreateStatusPageService`, `GetStatusPageService`, `UpdateStatusPageService`,
`DeleteStatusPageService` and the types `StatusPage`, `StatusPageCard` and `StatusPageService` are
deprecated. Use the V3 methods, such as `CreateStatusPageV3` and `CreateStatusPageComponentV3`, and
the types `StatusPageV3` and `StatusPageComponentV3` instead. The public API cannot migrate a v1
status page to v3, so the deprecated identifiers remain for existing v1 pages.
`IncidentTrigger.ServiceID` takes only v1 services, not component IDs: v3 pages automate incidents with `StatusPageAutomationRuleV3` and `CreateStatusPageAutomationRuleV3`.
