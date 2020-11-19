# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [2.0.1] - 2020-10-08
### Fixed:
- Trying to unmarshal a JSON `null` value into `lduser.User` now returns an error of type `*json.UnmarshalTypeError`, rather than misleadingly returning `lduser.ErrMissingKey()`.
- Trying to unmarshal a JSON value of the wrong type into `ldvalue.OptionalBool`, `ldvalue.OptionalInt`, or `ldvalue.OptionalString` now returns an error of type `*json.UnmarshalTypeError`, rather than a generic error string from `errors.New()`.
- The error `lduser.ErrMissingKey()` had an empty error string.

## [2.0.0] - 2020-09-18
Initial release of the newer types that will be used in Go SDK 5.0 and above.

### Added:
- Package `ldlog`, which was formerly a subpackage of `go-server-sdk`.
- Package `ldlogtest`, containing test helpers for use with `ldlog`.
- Package `ldreason`, containing `EvaluationReason` and related types that were formerly in `go-server-sdk`.
- Package `ldtime`, containing `UnixMillisecondTime`.
- Package `lduser`, containing `User` and related types that were formerly in `go-server-sdk`.
- Package `jsonstream`, a fast JSON encoding tool that is used internally by the SDK.
- `ldvalue.OptionalString` now implements `encoding.TextMarshaler` and `encoding.TextUnmarshaler`. This is not used by the Go SDK, but can be helpful when using `OptionalString` in other contexts.
- `ldvalue.OptionalBool` and `ldvalue.OptionalInt` are analogous to `ldvalue.OptionalString`, representing values that may be undefined without using pointers. These are used in the Go SDK.

### Changed:
- The minimum Go version is now 1.14.
- The `User` type is now opaque and immutable; there is no direct access to its fields.
- The `User` type no longer uses pointers or `interface{}` internally, decreasing the need for heap allocations.
- Reading a `User` from JSON with `json.Unmarshal` now returns an error if the `key` property is missing or null.
- `EvaluationDetail.VariationIndex` is now an `OptionalInt` rather than an `int`.
- `EvaluationReason` is now a struct.
- This project is now a Go module, although it can still be used from non-module code.

### Removed:
- In `ldvalue`, there are no longer methods for wrapping an existing `interface{}` value in a `Value`.
- All deprecated members of types that were moved here from `go-server-sdk` have been removed.

## [1.0.0] - 2020-02-03
Initial release. This will be used in versions 4.16.0 and above of the LaunchDarkly Server-Side SDK for Go.
