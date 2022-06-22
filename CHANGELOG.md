# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<br>

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
