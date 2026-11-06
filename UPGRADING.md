# Upgrading from v1 to v2

v2.0.0 is a major release. It changes how a client is built, returns typed errors, adds dedicated
types and methods for API, browser and multi-step checks, removes APIs that were deprecated or no
longer backed by the public API, and changes what some create and update requests send. This guide
lists every change from v1.27.0 that can affect your code, most disruptive first.

With v2.0.0 the module path becomes `github.com/checkly/checkly-go-sdk/v2`:

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
  | `ErrNotFound` (also `IsNotFound(err)`) | 404 |
  | `ErrUnauthorized` | 401: the server rejected the API key |
  | `ErrForbidden` | any 403 |
  | `ErrAccountNotAccessible` | 403 with code `ERR_ACCOUNT_NOT_ACCESSIBLE`; also matches `ErrForbidden` |
  | `ErrPlanLimit` | 402 |
  | `ErrConflict` | 409 |
  | `ErrRateLimited` | 429 still rate limited after the client's retries |
  | `ErrCheckTypeMismatch` | a typed getter got the ID of a check of another type (not an `*APIError`) |
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
backed by the per-type endpoints. The generic `Check` type and `CreateCheck`, `GetCheck`,
`UpdateCheck` and `DeleteCheck` still work but are deprecated: `UpdateCheck` uses the generic
`PUT /v1/checks/{id}` endpoint, which the API has deprecated.

| Check type | Type | Methods | Create and update endpoint |
|---|---|---|---|
| API | `APICheck` | `CreateAPICheck`, `GetAPICheck`, `UpdateAPICheck`, `DeleteAPICheck` | `/v1/checks/api` |
| Browser | `BrowserCheck` | `CreateBrowserCheck`, `GetBrowserCheck`, `UpdateBrowserCheck`, `DeleteBrowserCheck` | `/v1/checks/browser` |
| Multi-step | `MultiStepCheck` | `CreateMultiStepCheck`, `GetMultiStepCheck`, `UpdateMultiStepCheck`, `DeleteMultiStepCheck` | `/v1/checks/multistep` |
| Heartbeat | `HeartbeatMonitor` | `CreateHeartbeatMonitor`, `GetHeartbeatMonitor`, `UpdateHeartbeatMonitor`, `DeleteHeartbeatMonitor` | `/v1/checks/heartbeat` |

The monitor types (`URLMonitor`, `TCPMonitor`, `DNSMonitor`, `ICMPMonitor`, `GRPCMonitor`,
`TracerouteMonitor`, `SSLMonitor`) and `PlaywrightCheck` keep their methods.

The `Client` interface gains the twelve methods in the first three rows, so your own types that
implement `Client`, such as mocks or wrappers, must add them.

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
- `MultiStepCheck` no longer has `Type` or `FrequencyOffset`.
- New on `APICheck`, `BrowserCheck`, `MultiStepCheck` and `PlaywrightCheck`: `Intent`
  (`*CheckIntent`) and `AIAutoRepairEnabled` (`*bool`), for automatic AI repair.

**Typed getters reject other check types.** `GetAPICheck`, `GetBrowserCheck`, `GetMultiStepCheck`,
`GetPlaywrightCheck`, `GetHeartbeatMonitor` and the URL, TCP, DNS, ICMP, gRPC, traceroute and SSL
monitor getters return an error matching `ErrCheckTypeMismatch` when the ID belongs to a check of
another type. In v1, the Playwright, heartbeat and monitor getters decoded the check anyway and
dropped its type-specific fields.

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
| `Check.SSLCheck`, `AlertSettings.SSLCertificates`, type `SSLCertificates` | Nothing: the public API no longer has them |
| `Assertion.Edit`, `Assertion.Order`, `Assertion.ArrayIndex`, `Assertion.ArraySelector` | Nothing: the API discarded them, and requests no longer send them |

`Create` posted every check to `/v1/checks`. Its replacements post to the endpoint of the check's
type, and the deprecated `CreateCheck` rejects monitor types, which have their own create methods.

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
| `Assertion.Source` | `string` | `AssertionSource` |
| `Assertion.Comparison` | `string` | `AssertionComparison` |
| `AlertSettings.EscalationType` | `string` | `EscalationType` |
| `AlertChannel.Type` | `string` | `AlertChannelType` |
| `Check.Type` | `string` | `CheckType` |
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

## Behaviour changes on create and update

Most update endpoints keep the stored value of a key the request leaves out. v1 left out many empty
fields, so an update could not clear them. v2 sends them, so **an update replaces the stored
values with the struct you pass**: a field you leave empty is cleared or reset. To keep a value set
elsewhere, such as in the web app, read the resource, change it and send it back.

Fields now always sent, as `[]`, `null`, `""` or `0` when empty:

- **Heartbeat, TCP, URL, DNS, ICMP, gRPC, traceroute and SSL monitors and `Check`:** `tags` and
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
- **`Check`:** `setupSnippetId`, `tearDownSnippetId`, `localSetupScript` and `localTearDownScript`
  as `null` when unset.
- **Check groups (`Group`):** `setupSnippetId`, `tearDownSnippetId`, `localSetupScript` and
  `localTearDownScript` as `null`; `tags` and `alertChannelSubscriptions` as `[]` (also on
  `GroupV2`).
- **`Request` (API checks, `Check`, URL monitors):** `BasicAuth` as `null` when nil, so removing
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
`URLRequest.Assertions`, `DNSRequest.Assertions`, `ICMPRequest.Assertions`, `Group.Tags`,
`GroupV2.Tags`, `StatusPage.Cards` and `StatusPageCard.Services`.

Other new fields: `Assertion.Regex` (for `HEADERS` assertions; always sent, so updates no longer
drop a regex set elsewhere) and the maintenance window types `MaintenanceWindowComponentConfig` and
`MaintenanceWindowComponentImpact`.

**`GetCheckResults`** sends any non-empty `CheckResultsFilter.CheckType`. v1 sent only `API` and
`BROWSER` and silently dropped other types, returning unfiltered results; the API now answers 404
when the check is not of the given type.

## Decoding fixes

- `CreatedAt` and `UpdatedAt` on `Check`, `HeartbeatMonitor`, `Group`, `GroupV2`, `Snippet` and
  `MultiStepCheck` are now populated: they decode from the `created_at` and `updated_at` keys the API
  returns. If you marshal these structs yourself, their JSON keys change accordingly.
- Alert channels returned by `CreateAlertChannel`, `GetAlertChannel` and `UpdateAlertChannel` have
  `CreatedAt` and `UpdatedAt` set; they were always empty.
- Alert channel responses with a `null` or wrongly typed field no longer panic. `null` leaves the
  field unset. A wrongly typed top-level field, or a fractional `id` or `sslExpiryThreshold`,
  returns an error naming the field; a wrongly typed field inside the config is skipped; only a
  config that is not a JSON object, or one for an unknown channel type, returns an error.
- `CreateDashboard`, `CreateMaintenanceWindow` and `CreatePrivateLocation` check the status before
  decoding, so an error page that is not JSON returns an `*APIError` instead of a JSON syntax error.
