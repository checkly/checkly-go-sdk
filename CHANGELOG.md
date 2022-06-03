# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<br>

## v1.5.7 - 2022-05-26
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

## v1.5.6 - 2022-01-26
### Fixed
- Add `skipSSL` request property [#67](https://github.com/checkly/checkly-go-sdk/pull/67)

## v1.5.5 - 2022-01-26
### Fixed
- Fix `SetChecklySource` type definition [#66](https://github.com/checkly/checkly-go-sdk/pull/66)

## v1.5.4 - Skipped
N/A
## v1.5.3 - 2022-01-26
### Added
- Add `x-checkly-source` header [#65](https://github.com/checkly/checkly-go-sdk/pull/65)

## v1.5.2 - 2022-01-06
### Changed
- Add new URL property to Command Line Triggers [#73](https://github.com/checkly/checkly-go-sdk/issues/73)


## v1.5.1 - 2022-01-05
### Changed
- Update default branch to main [#60](https://github.com/checkly/checkly-go-sdk/issues/60)


## v1.5.0 - 2022-01-05
### Added
- Support for Command Line Triggers [#58](https://github.com/checkly/checkly-go-sdk/issues/58)

## v1.4.0 - 2021-11-09
### Added
- Support for Public Dashboards [#40](https://github.com/checkly/checkly-go-sdk/issues/40)
- Support for Maintenance Windows [#48](https://github.com/checkly/checkly-go-sdk/issues/48)
- Support for User API Keys [#51](https://github.com/checkly/checkly-go-sdk/issues/51)

## v1.3.1 - 2021-10-15
### Changed
- Fix optional and required values in webhooks alert channels [#41](https://github.com/checkly/checkly-go-sdk/issues/41)

## 1.3.0 - 2021-07-14
### Added
- Support for versioned runtimes [#31](https://github.com/checkly/checkly-go-sdk/issues/31)

## v1.2.0 - 2021-06-02
### Added
- Support for PagerDuty alert channels integration [#29](https://github.com/checkly/checkly-go-sdk/pull/29)

## v1.1.0 - 2021-05-20
### Added
- Support for API high frequency checks [#26](https://github.com/checkly/checkly-go-sdk/issues/26).

## v1.0.1 - 2021-02-25

### Changed
- Set `autoAssignAlerts` as false for all request [#23](https://github.com/checkly/checkly-go-sdk/issues/23)
