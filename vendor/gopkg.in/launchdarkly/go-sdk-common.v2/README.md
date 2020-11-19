# LaunchDarkly Go SDK Core Packages

[![Circle CI](https://circleci.com/gh/launchdarkly/go-sdk-common.svg?style=shield)](https://circleci.com/gh/launchdarkly/go-sdk-common) [![Documentation](https://img.shields.io/static/v1?label=go.dev&message=reference&color=00add8)](https://pkg.go.dev/gopkg.in/launchdarkly/go-sdk-common.v2)

## Overview

This repository contains packages and types that are shared between the [LaunchDarkly Go SDK](https://github.com/launchdarkly/go-server-sdk) and other LaunchDarkly Go components.

Applications using the LaunchDarkly Go SDK will generally use the `lduser` subpackage, which contains the `User` type, and may also use the `ldvalue` package, which contains the `Value` type that represents arbitrary JSON values. Other packages are less frequently used.

Note that the base import path is `gopkg.in/launchdarkly/go-sdk-common.v2`, not `github.com/launchdarkly/go-sdk-common`. This ensures that the package can be referenced not only as a Go module, but also by projects that use older tools like `dep` and `govendor`, because the 5.x release of the Go SDK supports either module or non-module usage. Future releases of this package, and of the Go SDK, may drop support for non-module usage.

Also, unlike `go-server-sdk` this does not have `server` in the name, because nothing in this repository is specific to the LaunchDarkly server-side model; it could be used in a client-side context.

## Supported Go versions

This version of the project has been tested with Go 1.14 and higher.

## Learn more

Check out our [documentation](http://docs.launchdarkly.com) for in-depth instructions on configuring and using LaunchDarkly. You can also head straight to the [complete reference guide for the Go SDK](http://docs.launchdarkly.com/docs/go-sdk-reference), or the [generated API documentation](https://godoc.org/gopkg.in/launchdarkly/go-sdk-common.v2) for this project.

## Contributing

We encourage pull requests and other contributions from the community. Check out our [contributing guidelines](CONTRIBUTING.md) for instructions on how to contribute to this SDK.

## About LaunchDarkly

* LaunchDarkly is a continuous delivery platform that provides feature flags as a service and allows developers to iterate quickly and safely. We allow you to easily flag your features and manage them from the LaunchDarkly dashboard.  With LaunchDarkly, you can:
    * Roll out a new feature to a subset of your users (like a group of users who opt-in to a beta tester group), gathering feedback and bug reports from real-world use cases.
    * Gradually roll out a feature to an increasing percentage of users, and track the effect that the feature has on key metrics (for instance, how likely is a user to complete a purchase if they have feature A versus feature B?).
    * Turn off a feature that you realize is causing performance problems in production, without needing to re-deploy, or even restart the application with a changed configuration file.
    * Grant access to certain features based on user attributes, like payment plan (eg: users on the ‘gold’ plan get access to more features than users in the ‘silver’ plan). Disable parts of your application to facilitate maintenance, without taking everything offline.
* LaunchDarkly provides feature flag SDKs for a wide variety of languages and technologies. Check out [our documentation](https://docs.launchdarkly.com/docs) for a complete list.
* Explore LaunchDarkly
    * [launchdarkly.com](https://www.launchdarkly.com/ "LaunchDarkly Main Website") for more information
    * [docs.launchdarkly.com](https://docs.launchdarkly.com/  "LaunchDarkly Documentation") for our documentation and SDK reference guides
    * [apidocs.launchdarkly.com](https://apidocs.launchdarkly.com/  "LaunchDarkly API Documentation") for our API documentation
    * [blog.launchdarkly.com](https://blog.launchdarkly.com/  "LaunchDarkly Blog Documentation") for the latest product updates
    * [Feature Flagging Guide](https://github.com/launchdarkly/featureflags/  "Feature Flagging Guide") for best practices and strategies
