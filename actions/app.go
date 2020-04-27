package actions

import (
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	testing "google.golang.org/api/testing/v1"

	"github.com/bitrise-io/go-utils/log"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/markbates/pop/nulls"
	"github.com/pkg/errors"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/tasks"

	"github.com/bitrise-io/addons-firebase-testlab/configs"
)

const (
	contextUserUsage                    = "user_usage"
	contextUserAllUsages                = "user_usages"
	contextMatrixHistoryAndExecutionIDs = "matrix_history_and_execution_ids"
)

var app *buffalo.App
var r *render.Engine

func initApp() error {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	err := configs.Setup()
	if err != nil {
		return errors.Wrap(err, "Failed to init configs")
	}

	err = database.InitDB()
	if err != nil {
		return errors.Wrap(err, "Failed to init DB")
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		return errors.Wrap(err, "Failed to create Firebase API model")
	}

	// init devices catalog
	firebaseutils.DevicesCatalog, err = fAPI.GetDeviceCatalog()
	if err != nil {
		return errors.Wrap(err, "Failed to get devices catalog")
	}

	return nil
}

// App ...
func App() *buffalo.App {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if app == nil {
		r = render.New(render.Options{TemplateEngine: render.GoTemplateEngine})

		if err := initApp(); err != nil {
			logger.Error("[!] Exception: Failed to init app", zap.Any("error", errors.WithStack(err)))
		}

		app = buffalo.Automatic(buffalo.Options{
			Env:         configs.GetENV(),
			SessionName: "_addons-firebase-testlab_session",
		})

		if len(os.Args) > 1 {
			switch os.Args[1] {
			case "clean-builds":
				err := tasks.CleanBuilds()
				if err != nil {
					log.Errorf("Failed to run task: clean-builds, error: %s", err)
					os.Exit(1)
				}
			case "builds-count":
				err := tasks.GetBuildsCount()
				if err != nil {
					log.Errorf("Failed to run task: builds-count, error: %s", err)
					os.Exit(1)
				}
			case "add-backend-user":
				//
				// DEVELOPMENT
				if configs.GetENV() == "development" {
					prefillAppSlug := os.Getenv("DEV_PREFILL_APPSLUG")
					prefillBuildSlug := os.Getenv("DEV_PREFILL_BUILDSLUG")
					buildExists, err := database.IsBuildExists(prefillAppSlug, prefillBuildSlug)
					if err != nil {
						log.Errorf("[DEV] Failed to get build from DB, error: %s", err)
					}
					if !buildExists {
						err = database.AddBuild(&models.Build{TestMatrixID: "matrix-1lpev7dwzfs6l", TestHistoryID: "bh.ac1fd45f1d1a66a5", TestExecutionID: "4928682413877019352", BuildSlug: prefillBuildSlug, AppSlug: prefillAppSlug, LastRequest: nulls.NewTime(time.Now()), BuildSessionEnabled: true})
						if err != nil {
							log.Errorf("[DEV] Failed to save build, error: %+v", errors.WithStack(err))
						}
					}
				}
			case "add-test-user":
				if len(os.Args) != 3 {
					log.Errorf("No appSlug specified")
					os.Exit(1)
				}
				err := tasks.AddDefaultUser(os.Args[2])
				if err != nil {
					log.Errorf("Failed to run task: builds-count, error: %s", err)
					os.Exit(1)
				}
			case "test-devices":
				device := &testing.AndroidDevice{
					AndroidModelId:   "Nexus4",
					AndroidVersionId: "21",
				}
				log.Infof("Checking device: %s with version: %s", device.AndroidModelId, device.AndroidVersionId)
				if err := firebaseutils.ValidateAndroidDevices([]*testing.AndroidDevice{device}); err != nil {
					log.Errorf("Device: %s should be available with version: %s, error: %s", device.AndroidModelId, device.AndroidVersionId, err)
				} else {
					log.Donef("OK")
				}
				device = &testing.AndroidDevice{
					AndroidModelId:   "Nexus45",
					AndroidVersionId: "21",
				}
				log.Infof("Checking device: %s with version: %s", device.AndroidModelId, device.AndroidVersionId)
				if err := firebaseutils.ValidateAndroidDevices([]*testing.AndroidDevice{device}); err == nil {
					log.Errorf("Device: %s shouldn't be available with version: %s, error: %s", device.AndroidModelId, device.AndroidVersionId, err)
				} else {
					log.Donef("OK - %s", err)
				}
				device = &testing.AndroidDevice{
					AndroidModelId:   "Nexus4",
					AndroidVersionId: "10",
				}
				log.Infof("Checking device: %s with version: %s", device.AndroidModelId, device.AndroidVersionId)
				if err := firebaseutils.ValidateAndroidDevices([]*testing.AndroidDevice{device}); err == nil {
					log.Errorf("Device: %s shouldn't be available with version: %s, error: %s", device.AndroidModelId, device.AndroidVersionId, err)
				} else {
					log.Donef("OK - %s", err)
				}
				device = &testing.AndroidDevice{
					AndroidModelId:   "hammerhead",
					AndroidVersionId: "21",
				}
				log.Infof("Checking device: %s with version: %s", device.AndroidModelId, device.AndroidVersionId)
				if err := firebaseutils.ValidateAndroidDevices([]*testing.AndroidDevice{device}); err == nil {
					log.Errorf("Device: %s shouldn't be available with version: %s, error: %s", device.AndroidModelId, device.AndroidVersionId, err)
				} else {
					log.Donef("OK - %s", err)
				}
			}
			os.Exit(0)
		}

		//
		// ENDPOINTS

		//
		// ROOT
		app.Use(addLogger)
		app.GET("/", RootGetHandler)

		// WEBHOOK
		app.POST("/webhook", verifySignature(WebhookHandler))

		//
		// PROVISIONING
		provision := app.Group("/provision")
		provision.Use(authenticateWithAccessToken)              // only accessible with access token
		provision.POST("/", ProvisionPostHandler)               // set new provision
		provision.PUT("/{app_slug}", ProvisionPutHandler)       // update provision
		provision.DELETE("/{app_slug}", ProvisionDeleteHandler) // delete provision

		//
		// TESTING
		test := app.Group("/test") // main addon test functionality
		test.Use(authenticateRequestWithToken)
		test.GET("/{app_slug}/{build_slug}/{token}", authorizeForBuild(TestGet))                                          // get test matrix data
		test.POST("/{app_slug}/{build_slug}/{token}", authorizeForRunningBuildViaBitriseAPI(authorizeForBuild(TestPost))) // start test matrix

		test.POST("/assets/android/{app_slug}/{build_slug}/{token}", authorizeForRunningBuildViaBitriseAPI(TestAssetUploadURLsAndroid)) // get signed upload urls for Android assets
		test.POST("/assets/{app_slug}/{build_slug}/{token}", authorizeForRunningBuildViaBitriseAPI(TestAssetsPost))                     // get signed upload urls for assets
		test.GET("/assets/{app_slug}/{build_slug}/{token}", authorizeForBuild(TestAssetsGet))                                           // get signed download urls for assets

		//
		// TEST REPORTS
		test.POST("/apps/{app_slug}/builds/{build_slug}/test_reports/{token}", authorizeForRunningBuildViaBitriseAPI(TestReportsPostHandler))
		test.PATCH("/apps/{app_slug}/builds/{build_slug}/test_reports/{test_report_id}/{token}", authorizeForRunningBuildViaBitriseAPI(authorizeForTestReport(TestReportPatchHandler)))

		//
		// API
		api := app.Group("/api")                                           // api group
		api.Use(validateUserLoginStatus)                                   // check if signature is valid
		api.GET("/app", DashboardAppGetHandler)                            // return app info
		api.GET("/builds/{build_slug}", DashboardAPIGetHandler)            // return dashboard resources
		api.GET("/builds/{build_slug}/steps/{step_id}", StepAPIGetHandler) // return step resources
		api.GET("/builds/{build_slug}/test_reports", TestReportsListHandler)
		api.GET("/builds/{build_slug}/test_summary", TestSummaryHandler)
		api.GET("/builds/{build_slug}/test_reports/ftl", DashboardAPIGetHandler) // Alternative route for FTL reports keeping the legacy route intact
		api.GET("/builds/{build_slug}/test_reports/{test_report_id}", TestReportShowHandler)

		//
		// DASHBOARD
		//app.Use(validateUserLoginStatus)                                          // check if signature is valid - this is not required
		app.Use(serveSVGs)                                                        // serve svgs content in the template
		app.GET("/builds/{build_slug}", DashboardIndexGetHandler)                 // dashboard index page
		app.GET("/builds/{build_slug}/steps/{step_id}", DashboardIndexGetHandler) // dashboard index page
		app.GET("/templates/dashboard", DashboardGetHandler)                      // dashboard main page
		app.GET("/templates/details", DashboardDetailsGetHandler)                 // dashboard details page
		app.POST("/login", DashboardLoginPostHandler)                             // sso login handler
		app.ServeFiles("/assets", http.Dir("./frontend/assets/compiled"))         // serve assets for dashboard

	}

	return app
}
