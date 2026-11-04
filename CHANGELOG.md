# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<br>

## [Unreleased]
### Added
- Add `APICheck` and `BrowserCheck` types and replace `MultiStepCheck`, each with `Create/Get/Update/Delete` client methods backed by the dedicated per-type endpoints (`checks/api`, `checks/browser`, `checks/multistep`). The getters return an error when the ID belongs to a check of another type
- Add `APIError`, returned for every response whose status the method does not accept, with the request's `Method` and `Path`, the `StatusCode`, the backend's error `Code`, `Message` and validation `Details`, and `RetryAfter` for a 429. Its text reads `METHOD /v1/path: 404 Not Found[ [ERROR_CODE]][: message][ (keys: …)]`, or `METHOD /v1/path: unexpected 200 OK response, want 201 Created` for a success status the method does not accept, followed by a hint for a 401 (check the API key), a 403 for an inaccessible account (check the account ID) or a 403 refused for the API key's role, and by the retries and the server's wait for a 429
- Add sentinel errors `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrAccountNotAccessible`, `ErrPlanLimit`, `ErrConflict` and `ErrRateLimited`, which an `APIError` matches with `errors.Is` by its status (and, for `ErrAccountNotAccessible`, its error code), and `IsNotFound(err)`

### Changed
- **Breaking:** `MultiStepCheck` is redefined for the `checks/multistep` endpoint: `Type` and `FrequencyOffset` are removed, `AlertSettings` is now a pointer, `ScriptPath` and `Dependencies` are added, and the timestamps decode from `created_at`/`updated_at`
- **Breaking:** the check update endpoints keep the stored value of an omitted key, so heartbeat, TCP, URL, DNS and ICMP monitor and check group create/update payloads now send clearable fields even when empty: `tags` and `alertChannelSubscriptions` as `[]` (heartbeat and TCP monitors excepted for subscriptions, whose endpoints clear them on omission), unset `groupOrder` as `null` (URL, DNS and ICMP monitors) or `0` (TCP monitors), TCP `request.assertions` as `[]`, and unset group `setupSnippetId`, `tearDownSnippetId`, `localSetupScript` and `localTearDownScript` as `null`. Removing the last value of such a field now clears it instead of leaving it unchanged
- **Breaking:** `Request.BasicAuth` is always serialized, as `null` when unset, so that removing credentials clears them
- **Breaking:** TCP, URL and DNS monitor payloads always send `degradedResponseTime` and `maxResponseTime`, and ICMP monitor payloads always send `degradedPacketLossThreshold` and `maxPacketLossThreshold`, so that update can set every value. A response time of 0 is no longer replaced by the API's default, and a packet loss threshold of 0 is rejected by the API. Callers that relied on the defaults when leaving these fields at 0 must now set them explicitly
- **Breaking:** the text of every error returned by client methods changes, including transport and decode errors. Code that matched it, such as the `unexpected response status 404:` prefix, must switch to `errors.Is` with the sentinel errors or `errors.As` with `*APIError`
- **Breaking:** `ErrCodeBundleNotFound` is now returned wrapped together with the `*APIError`, so match it with `errors.Is` rather than `==`. Only a 404, or the 403 the backend answers for a key of another account, counts as not found: a 403 for an inaccessible account, a role without permission or a plan without the entitlement is returned as a plain `*APIError`
- **Breaking:** `AlertChannelConfigFromJSON` returns `nil` for a `null` config instead of a pointer to an empty config, and an error, naming the channel type, for an unknown type or a config that is not a JSON object. It still skips fields of an unexpected type. `SetConfig(nil)` does nothing
- Errors no longer quote JSON response bodies, success bodies or request payloads (create errors for alert channels, dashboards, maintenance windows and private locations used to include the payload). The backend's error message, or the start of an error body that is not JSON, such as a proxy's HTML page, is kept with whitespace collapsed and cut at 500 bytes; a decode error names the request, the status and the body's size instead of quoting the body
- Transport errors wrap their cause, so `errors.Is(err, context.DeadlineExceeded)` and `errors.As` with a `*url.Error` work. The text names the request and the API host: `GET /v1/checks/abc: request to api.checklyhq.com failed: context deadline exceeded`
- Cancelling the context during a rate-limit wait returns an error that matches both `ErrRateLimited` and the context error, so check it with `errors.Is` rather than `==`
- `PeekCodeBundle` is retried on HTTP 429 like the other requests; `UploadCodeBundle`, which streams its body, still is not
- Code bundle upload and peek errors keep the backend's message, such as the payload size limit, instead of reporting only the status
- `CreateDashboard`, `CreateMaintenanceWindow`, `CreatePrivateLocation`, `CreateTriggerCheck` and `CreateTriggerGroup` check the status before decoding the response, so an error body that is not JSON, such as a proxy's HTML page, returns an `*APIError` instead of a JSON syntax error
- Alert channel responses with a `null` or wrongly typed field no longer panic: `null` leaves the field unset, and a wrong type returns an error that names the field; a `null` config leaves every config unset
- `GetStaticIPs` reports an invalid IPv4 address as `could not parse IP address from …` instead of `could not parse CIDR from …`
- The error for an unsupported retry strategy value starts in lowercase: `unsupported value "…" for retry strategy`

### Security
- The debug request dump replaces the `Authorization` header, which carries the API key, with `[REDACTED]`

## [v1.27.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.27.0) - 2026-10-07
### Changed
- gRPC monitor create/update payloads send empty `assertions` and `grpcConfig.metadata` as `[]`, and always send `grpcConfig.message` (`grpcConfig.service` in `HEALTH` mode), even when empty. The update endpoint keeps the stored value of an omitted field, so these empty values now clear the stored value instead of leaving it unchanged

## [v1.26.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.26.0) - 2026-10-07
### Added
- Add status pages v3 support: `StatusPageV3` (incl. `supportLink` and `themeColors`), `StatusPageComponentV3` (incl. the typed per-type `configuration`) and `StatusPageAutomationRuleV3` types + `Create/Get/Update/Delete` client methods (`/v3/status-pages`, `/v3/status-pages/{id}/components`, `/v3/status-pages/{id}/automation-rules`)

### Changed
- API requests rejected with HTTP 429 are retried up to 4 times (code bundle upload and peek excepted), waiting as long as `Retry-After` or `X-RateLimit-Reset` asks (capped at 60s, plus up to 1.5s of padding and jitter) or with exponential backoff and jitter otherwise. A wait that would not leave enough of the request context's deadline for the re-sent request is skipped and the 429 is returned at once, so callers with short per-call deadlines need to allow for the wait to benefit; cancelling the context during a wait returns the context error

## [v1.25.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.25.0) - 2026-10-07
### Changed
- gRPC monitors preserve their FlatBuffers monitor config [#172](https://github.com/checkly/checkly-go-sdk/pull/172)

## [v1.24.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.24.0) - 2026-10-06
### Changed
- `MaintenanceWindow` sends the complete window, with timezone, scope and status page settings [#170](https://github.com/checkly/checkly-go-sdk/pull/170)

## [v1.23.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.23.0) - 2026-09-03
### Added
- Add optional `Path` field to `ClientCertificate` (URL path prefix that limits the certificate to API and Multistep check requests under that path)

## [v1.22.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.22.0) - 2026-07-23
### Added
- Add SSL assertion grammar: new assertion sources `CERTIFICATE`, `CONNECTION`, `RESPONSE_TIME`, `JSON_RESPONSE` and `TEXT_RESPONSE`, plus `IS_NULL`/`NOT_NULL` comparisons
- Add `GRPCMonitor` type + `Create/Get/Update/Delete` client methods (`checks/grpc`) [#c0f841a](https://github.com/checkly/checkly-go-sdk/commit/c0f841a)
- Add `TracerouteMonitor` type + `Create/Get/Update/Delete` client methods (`checks/traceroute`) [#85860b0](https://github.com/checkly/checkly-go-sdk/commit/85860b0)
- Add `SSLMonitor` type + `Create/Get/Update/Delete` client methods (`checks/ssl`) [#e30a577](https://github.com/checkly/checkly-go-sdk/commit/e30a577)

## [v1.6.1](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.6.1) - 2022-06-22
### Added
- Allow check/group private location assigments using pointers [#90d4a59](https://github.com/checkly/checkly-go-sdk/commit/90d4a598596329a22c413b10b12b74673c337bfd)

## [v1.6.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.6.0) - 2022-06-21
### Added
- Support private locations [#92](https://github.com/checkly/checkly-go-sdk/issues/92)
- Add CoC, more docs and issue templates [#604b9a1](https://github.com/checkly/checkly-go-sdk/commit/604b9a12bdd8a913411378f3817aabe5b00f0cee)

## [v1.5.7](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.7) - 2022-05-26
### Added
- Add Dependabot GitHub action [#69b1aeb](https://github.com/checkly/checkly-go-sdk/commit/69b1aebf23a1d86d3645ef13ff20f57aafee4672)
- Add CODEOWNERS file [#02e1ce4](https://github.com/checkly/checkly-go-sdk/commit/02e1ce4fff4e00eb0bbcbfb0e849324df4664b75)

### Changed
- Use new Public API endpoints for checks creation [#72](https://github.com/checkly/checkly-go-sdk/pull/72)
- Improve code demo and update test cases [#ae0cc2e](https://github.com/checkly/checkly-go-sdk/commit/ae0cc2e4f2431f4862352a8dcfc61fbf0d5009d7)
- Upgrade actions/checkout to v3 [#80](https://github.com/checkly/checkly-go-sdk/pull/80)
- Upgrade github/codeql-action to v2  [#89](https://github.com/checkly/checkly-go-sdk/pull/83)
- Upgrade github.com/google/go-cmp to v0.5.8 [#84](https://github.com/checkly/checkly-go-sdk/pull/84)
- Deprecate `sslCertificates` from `AlertSettings` and `sslChecks` from `Checks` [#76](https://github.com/checkly/checkly-go-sdk/pull/76)
- Migrate test cases to use Checkly User Keys (#88)[https://github.com/checkly/checkly-go-sdk/pull/88]

### Fixed
- Fix typos and remove emojis from docs [#41931cc](https://github.com/checkly/checkly-go-sdk/commit/41931ccd73f18344374ef1164d67fcf09a916fb7)
- Set `omitempty` in property `RepeatInterval` of  `MaintenanceWindow` [#79](https://github.com/checkly/checkly-go-sdk/pull/79)
- Set `omitempty` in `Key` of `EnvironmentVariable` [#90](https://github.com/checkly/checkly-go-sdk/pull/90)
- Set missing `omitempty` in `Dashboard` and `MaintenanecWindow` [#87](https://github.com/checkly/checkly-go-sdk/pull/87)

## [v1.5.6](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.6) - 2022-01-26
### Fixed
- Add `skipSSL` request property [#67](https://github.com/checkly/checkly-go-sdk/pull/67)

## [v1.5.5](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.5) - 2022-01-26
### Fixed
- Fix `SetChecklySource` type definition [#66](https://github.com/checkly/checkly-go-sdk/pull/66)

## v1.5.4 - Skipped
N/A
## [v1.5.3](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.3) - 2022-01-26
### Added
- Add `x-checkly-source` header [#65](https://github.com/checkly/checkly-go-sdk/pull/65)

## [v1.5.2](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.2) - 2022-01-06
### Changed
- Add new URL property to Command Line Triggers [#73](https://github.com/checkly/checkly-go-sdk/issues/73)


## [v1.5.1](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.1) - 2022-01-05
### Changed
- Update default branch to main [#60](https://github.com/checkly/checkly-go-sdk/issues/60)


## [v1.5.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.5.0) - 2022-01-05
### Added
- Support for Command Line Triggers [#58](https://github.com/checkly/checkly-go-sdk/issues/58)

## [v1.4.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.4.0) - 2021-11-09
### Added
- Support for Public Dashboards [#40](https://github.com/checkly/checkly-go-sdk/issues/40)
- Support for Maintenance Windows [#48](https://github.com/checkly/checkly-go-sdk/issues/48)
- Support for User API Keys [#51](https://github.com/checkly/checkly-go-sdk/issues/51)

## [v1.3.1](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.3.1) - 2021-10-15
### Changed
- Fix optional and required values in webhooks alert channels [#41](https://github.com/checkly/checkly-go-sdk/issues/41)

## [v1.3.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.3.0) - 2021-07-14
### Added
- Support for versioned runtimes [#31](https://github.com/checkly/checkly-go-sdk/issues/31)

## [v1.2.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.2.0) - 2021-06-02
### Added
- Support for PagerDuty alert channels integration [#29](https://github.com/checkly/checkly-go-sdk/pull/29)

## [v1.1.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.1.0) - 2021-05-20
### Added
- Support for API high frequency checks [#26](https://github.com/checkly/checkly-go-sdk/issues/26).

## [v1.0.1](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.0.1) - 2021-02-25

### Changed
- Set `autoAssignAlerts` as false for all request [#23](https://github.com/checkly/checkly-go-sdk/issues/23)

## [v1.0.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v1.0.0) - 2021-02-22
### Added

- Introduce go `context`
- Introduce go `interface`

### Changed
- Set `autoAssignAlerts` as false for all request [#23](https://github.com/checkly/checkly-go-sdk/issues/23)

## [v0.5.6](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.4) - 2021-02-25
### Fixed
- Minor bugs fixing

## v0.5.5 - Skiped
N/A

## [v0.5.4](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.4) - 2021-02-25
### Changed
- Add `autoAssignAlerts` query to deprecate the old auto assigning (alert channels to checks)

## [v0.5.3](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.3) - 2021-01-21
### Fixed
- Minor bugs fixing

## [v0.5.2](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.2) - 2021-01-21
### Fixed
- Minor bugs fixing

## [v0.5.1](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.1) - 2021-01-21

### Added
- `Alert Channels` flag settings

## [v0.5.0](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.5.0) - 2021-01-11

### Fixed

- Minor issue when updatin `Alert Channles`

## [v0.4.9](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.9) - 2021-01-11
### Fixed
- Minor `Alert Channels` bug fixing

## [v0.4.8](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.8) - 2021-01-11
### Added
- Support `Alert Channels`

## [v0.4.7](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.7) - 2020-09-30

### Fixed
- Check results filters from/to bug fix

## [v0.4.6](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.6) - 2020-09-29

### Added
- Support for `Snippets`
- Support for `Environment Variables`

## [v0.4.5](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.5) - 2020-09-16

### Changed
- Avoid omitting tags when empty
- Made raw check results available

## [v0.4.4](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.4) - 2020-09-03

### Added
- Add `GetCheckResult`
- Add `GetCheckResults`


## [v0.4.3](https://github.com/checkly/checkly-go-sdk/releases/tag/v0.4.3) - 2020-07-29

### Changed
- Makes API Check's `request.basic_auth` optional
