# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [1.0.1] - 2020-10-08
### Fixed:
- When serializing flags and segments to JSON, properties with default values (such as false booleans or empty arrays) were being dropped entirely to save bandwidth. However, these representations may be consumed by SDKs other than the Go SDK, and some of the LaunchDarkly SDKs do not tolerate missing properties, so this has been fixed to remain consistent with the less efficient behavior of Go SDK 4.x.

## [1.0.0] - 2020-09-18
Initial release of this flag evaluation support code that will be used with versions 5.0.0 and above of the LaunchDarkly Server-Side SDK for Go.
