# Firebase Test Lab Add-on for Bitrise.io

This add-on enables Bitrise.io users to test their apps on Firebase Test Lab using the [Virtual Device Testing for Android step](https://www.bitrise.io/integrations/steps/virtual-device-testing-for-android) in their workflows.

## Development with Visual Studio Code

NOTE: originally, the project was set up with a docker-based development workflow in mind. To support debugging some of the flows using VS Code, a few
adjustments has been made later. The original development flows still work and are required to fully test the app (see later sections).

1. clone `example.env` and rename it to `.env` - this tores the env variables required for running the upp using VS Code
2. run `docker-compose up -d` to start the local development database
3. run the upp with the `Debug app` configuration

Note that not all use-cases have been tested and debuggable this way but provisioning/deprovisioning are working.

## How to contribute

### Install

- Please make sure you have [Go](https://golang.org) installed, your [Go Workspace](https://golang.org/doc/code.html#Workspaces) and your [$GOPATH](https://golang.org/doc/code.html#GOPATH) is set up properly. This project requires Go version 1.8.1 or higher.
- [Bitrise CLI](https://app.bitrise.io/cli) is needed to run the pre-defined workflows.
- An sqlite version of Buffalo is needed to set up DB and run migrations.
```bash
$ go get -u -v -tags sqlite github.com/gobuffalo/buffalo/buffalo
```
- sass is needed for the asset pipeline
```bash
$ brew install sass/sass/sass
```
- Grab the project
```bash
$ go get -u -v github.com/bitrise-io/addons-firebase-testlab
```
- Copy `.bitrise.secrets.example.yml` to `.bitrise.secrets.yml` and fill in the `SERVICE_ACCOUNT_KEY_JSON` env var with your Google credentials. Please do the same with `BUCKET` and `PROJECT_ID`.
- Fill in the `APP_APK_PATH` and `TEST_APK_PATH` env var with the local path of the app/test `.apk` files you want to request a virtual device test for.

### Setup the environment

Run the server setup flow from the root folder of the project.
```bash
$ bitrise run setup-server
```

This will:
- build/rebuild the docker images
- create, migrate and seed the development database

### Start the server and request a test

From the root folder of the project you can start the server.
```bash
$ bitrise run start-server
```

This will:
- compile the assets.
- start the server on `http://localhost:5001`.
- hot reload a server on code change.

In another session you can request a test by running:

```bash
$ bitrise run android-test
```

This will run a `virtual-device-testing-for-android` step that will use the add-on server you've just started and use the .apk files defined in `.bitrise.secrets.yml`.

The step will print out an URL where you can review the test results.

### Detailed flow

__1. Provisioning (done by website)__

To enable the add-on for an app, you have to make a provisioning request. The add-on uses a token present in the `Authentication` header to authenticate provisioning related requests from the Bitrise website. You can modify this token by setting ADDON_ACCESS_TOKEN in `bitrise.secrets.yml` The `api_token` in this request is a Bitrise API token that enables the add-on to authorize apps via the Bitrise API.

```bash
curl -X POST -H "Authentication: addon-access-token" -H "Content-Type: application/json" -d '{"app_slug":"app-slug1","api_token":"bitrise_token1","plan":"free"}' "http://localhost:5000/provision"
```

The add-on returns a JSON response. You can make a test request to the URL present in the `ADDON_VDTESTING_API_URL` key and the `ADDON_VDTESTING_API_TOKEN` value is needed for authentication.

```json
{
  "envs": [
    {"key":"ADDON_VDTESTING_API_URL", "value":"/test"},
    {"key":"ADDON_VDTESTING_API_TOKEN","value":"api-token"}
  ]
}
```

__2. Requesting an upload URL (done by VDT step)__

To request a test, first you need to upload the .apk files. You can request pre-signed upload URLs from the add-on.

```bash
curl -X POST -H "Content-Type: application/json" "http://localhost:5000/test/assets/app-slug1/build_slug1/api-token"
```

```json
{
  "appUrl":"https://storage.googleapis.com/presigned-upload-url1",
  "testAppUrl":"https://storage.googleapis.com/presigned-upload-url2"
}
```



Please note the use of the `app-slug1` (to which the add-on was provisioned before) and the `api-token` which was generated by the add-on on provisioning. The `build_slug` should be a build that belongs to the app and is running.

In production, the add-on calls the Bitrise API to make sure the build that requested a test belongs to the app the add-on was provisioned for, and to make sure the build is still running. To skip calling Bitrise API while developing/testing the add-on (and enabling us to use any build slug with the pre-populated app), we set the `SKIP_AUTH_WITH_BITRISE_API` env var to `yes` in `.bitrise.secrets.yml` (to use a non-production or a mocked version of the Bitrise API instead, set the `BITRISE_API_URL` env var to `https://your.bitrise.api`).

__3. Starting a test (done by VDT step)__

The client should upload the .apk files on its own and then request a test by sending test matrix data.

```bash
curl -X POST -H "Content-Type: application/json" -d '{"test_matrix":"data"}' "http://localhost:5000/test/app-slug1/build_slug1/api-token"
```

__4. Getting the test results (done by VDT step)__

```bash
curl -X GET -H "Content-Type: application/json" "http://localhost:5000/test/app-slug1/build_slug1/api-token"
```

__5. Getting the test assets (done by VDT step)__

```bash
curl -X GET -H "Content-Type: application/json" "http://localhost:5000/test/assets/app-slug1/build_slug1/api-token"
```

__6. Logging in to the dashboard provided by the add-on (done by website)__

To log in, you need to provide an app and a build slug, a timestamp and a hash of these data signed by an SSO token that can be set as `ADDON_SSO_TOKEN` in `.bitrise.secrets.yml`

```bash
curl -X POST -F "timestamp=2535644284" -F "app_slug=app-slug1"  -F "build_slug=build-slug1" -F "token=token"  "http://localhost:5000/login"
```

In production you cannot view the dashboard without making this request. In development we can initiate a session without this, we set the `SKIP_SESSION_AUTH` to `yes` in `.bitrise.secrets.yml`.

__7. Disabling the add-on for an app (done by website)__

```bash
curl -X DELETE -H "Authentication: addon-access-token" -H "Content-Type: application/json" "http://localhost:5000/provision/app-slug1"
```

__8. Create a test report and get an upload URL for the XML__

You can upload your test reports and the related assets through this endpoint. The request body should look like the example below:

```json
{
  "name":"My awesome file",
  "filename":"myjunit2.xml",
  "filesize":3,
  "step_info": {
    "id":"steps-virtual-device-testing-for-android",
    "version":"1.0.5",
    "title":"[BETA] Virtual Device Testing for Android",
    "number":12
  },
  "assets":[
    {"filename":"mytestasset1.png","filesize":32},
    {"filename":"mytestasset2.txt","filesize":41}
  ]
}
```

```bash
curl -X POST  -H "Content-Type: application/json" -d '{...}' "http://localhost:5001/test/apps/app_slug1/builds/build_slug1/test_reports/test-api-token"
```

__9. Upload it to GCP using the upload URL__

After generating the upload URLs for your test report file and its asset files, you have to upload the files to GCP

```bash
curl -v --upload-file "myjunit2.xml" "https://storage.googleapis.com/quick-test/builds/build_slug1/test_reports/a1b61365-899d-49b3-ae2d-b9bfbd2276f0/myjunit2.xml?Expires=...."
```

__10. Marking it as uploaded__

After the files are uploaded, confirm their uploaded state with calling this endpoint

```bash
curl -X PATCH  -H "Content-Type: application/json" -d '{"uploaded":true}' \
"http://localhost:5001/test/apps/app_slug1/builds/build_slug1/test_reports/a1b61365-899d-49b3-ae2d-b9bfbd2276f0/test-api-token"
```

__11. List the test reports of a sepcific build__

```bash
curl "http://localhost:5001/api/builds/build_slug1/test_reports"
```

__12. Show a specific test report__

```bash
curl "http://localhost:5001/api/builds/build_slug1/test_reports/a1b61365-899d-49b3-ae2d-b9bfbd2276f0"
```

### Testing

When you create a Pull Request, Bitrise CI will run the `test-with-docker-compose` workflow defined in `bitrise.yml`. You can do the same locally if you have Docker set up on your machine. Please note that this workflow does not expect a local path to the .apk files, but downloads them from the interwebs (or a local server) using `wget`, so please  set `BITRISEIO_APP_APK_URL` and `BITRISEIO_TEST_APK_URL` env vars in your `.bitrise.secrets.yml` accordingly. Then you can:

```bash
$ bitrise run test-with-docker-compose
```
