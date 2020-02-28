package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/analyticsutils"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/renderers"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop/nulls"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	testing "google.golang.org/api/testing/v1"
)

const (
	// Android test run uses virtual devices, while iOS uses phisical devices.
	// https://cloud.google.com/sdk/gcloud/reference/firebase/test/android/run#--timeout
	androidMaxTimeoutSecs = 60 * 60
	iosMaxTimeoutSecs     = 30 * 60
)

// TestGet ...
func TestGet(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	appSlug := c.Param("app_slug")

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		logger.Error("[!] Exception: Failed to get build from DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("[!] Exception: Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		matrix, err := fAPI.GetHistoryAndExecutionIDByMatrixID(build.TestMatrixID)
		if err != nil {
			matrix, err = fAPI.GetHistoryAndExecutionIDByMatrixID(build.TestMatrixID)
			if err != nil {
				logger.Error("[!] Exception: failed to get test matrix\nwith retry...",
					zap.String("test_matrix_id", build.TestMatrixID),
					zap.String("app_slug", build.AppSlug),
					zap.String("build_slug", build.BuildSlug),
					zap.Any("error", errors.WithStack(err)),
				)
				return c.Render(http.StatusInternalServerError, r.String("Failed to get test status"))
			}
		}

		if isMessageAnError(matrix.State) {
			return c.Render(http.StatusInternalServerError, r.String("Failed to get test status: %s(%s)", matrix.State, matrix.InvalidMatrixDetails))
		}

		if len(matrix.TestExecutions) == 0 {
			build.LastRequest = nulls.NewTime(time.Now())

			err = database.UpdateBuild(build)
			if err != nil {
				logger.Error("[!] Exception: Failed to update last request timestamp", zap.Any("build_details", build), zap.Any("error", errors.WithStack(err)))
				return c.Render(http.StatusInternalServerError, r.String("Failed to get test status"))
			}
			return c.Render(http.StatusOK, r.JSON(map[string]string{"state": matrix.State}))
		}

		if matrix.TestExecutions[0].ToolResultsStep == nil {
			build.LastRequest = nulls.NewTime(time.Now())

			err = database.UpdateBuild(build)
			if err != nil {
				logger.Error("[!] Exception: Failed to update last request timestamp", zap.Any("build_details", build), zap.Any("error", errors.WithStack(err)))
				return c.Render(http.StatusInternalServerError, r.String("Failed to get test status"))
			}
			return c.Render(http.StatusOK, r.JSON(map[string]string{"state": matrix.State}))
		}

		build.TestHistoryID = matrix.TestExecutions[0].ToolResultsStep.HistoryId
		build.TestExecutionID = matrix.TestExecutions[0].ToolResultsStep.ExecutionId
	}

	steps, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug, "steps(state,name,outcome,dimensionValue,testExecutionStep)")
	if err != nil {
		steps, err = fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug, "steps(state,name,outcome,dimensionValue,testExecutionStep)")
		if err != nil {
			logger.Error("[!] Exception: failed to get test by HistoryID and ExecutionID\nwith retry failed...",
				zap.String("test_history_id", build.TestHistoryID),
				zap.String("test_execution_id", build.TestExecutionID),
				zap.String("test_matrix_id", build.TestMatrixID),
				zap.String("app_slug", build.AppSlug),
				zap.String("build_slug", build.BuildSlug),
				zap.Any("error", errors.WithStack(err)),
			)
			return c.Render(http.StatusInternalServerError, r.String("Failed to get test status"))
		}
	}

	if len(steps.Steps) > 0 {
		isIOS := false

		completed := true
		for _, step := range steps.Steps {
			if step.State != "complete" {
				completed = false
			}
			if strings.Contains(strings.ToLower(step.Name), "ios") {
				isIOS = true
			}
		}
		if build.BuildSessionEnabled && completed {
			build.BuildSessionEnabled = false

			testType := "instrumentation"

			if !strings.Contains(strings.ToLower(steps.Steps[0].Name), "instrumentation") {
				testType = "robo"
			}

			result := "success"
			for _, step := range steps.Steps {
				if step.Outcome.Summary != "success" {
					result = "failed"
				}

				if !isIOS {
					device := &testing.AndroidDevice{}
					for _, dim := range step.DimensionValue {
						if dim != nil {
							switch dim.Key {
							case "Model":
								device.AndroidModelId = dim.Value
							case "Version":
								device.AndroidVersionId = dim.Value
							case "Locale":
								device.Locale = dim.Value
							case "Orientation":
								device.Orientation = dim.Value
							}
						}
					}
					analyticsutils.SendTestingEventDevices(analyticsutils.EventTestingTestFinishedOnDevice,
						appSlug,
						buildSlug,
						testType,
						[]*testing.AndroidDevice{device},
						map[string]interface{}{
							"test_result": step.Outcome.Summary,
						})
				} else {
					device := &testing.IosDevice{}
					for _, dim := range step.DimensionValue {
						if dim != nil {
							switch dim.Key {
							case "Model":
								device.IosModelId = dim.Value
							case "Version":
								device.IosVersionId = dim.Value
							case "Locale":
								device.Locale = dim.Value
							case "Orientation":
								device.Orientation = dim.Value
							}
						}
					}
					analyticsutils.SendIOSTestingEventDevices(analyticsutils.EventIOSTestingTestFinishedOnDevice,
						appSlug,
						buildSlug,
						"",
						[]*testing.IosDevice{device},
						map[string]interface{}{
							"test_result": step.Outcome.Summary,
						})
				}
			}
			if !isIOS {
				analyticsutils.SendTestingEvent(analyticsutils.EventTestingTestFinished,
					appSlug,
					buildSlug,
					testType,
					map[string]interface{}{
						"test_result": result,
					})
			} else {
				analyticsutils.SendTestingEvent(analyticsutils.EventIOSTestingTestFinished,
					appSlug,
					buildSlug,
					"",
					map[string]interface{}{
						"test_result": result,
					})
			}
		}
	}

	build.LastRequest = nulls.NewTime(time.Now())

	err = database.UpdateBuild(build)
	if err != nil {
		logger.Error("[!] Exception: Failed to update last request timestamp", zap.Any("build_details", build), zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Failed to get test status"))
	}

	return c.Render(http.StatusOK, r.JSON(steps))
}

func maxTimeoutSecs(spec testing.TestSpecification) int {
	switch {
	case spec.AndroidInstrumentationTest != nil,
		spec.AndroidRoboTest != nil,
		spec.AndroidTestLoop != nil:
		return androidMaxTimeoutSecs
	case spec.IosXcTest != nil:
		return iosMaxTimeoutSecs
	}
	return 0
}

// TestPost ...
func TestPost(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	appSlug := c.Param("app_slug")

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		logger.Error("[!] Exception: Failed to get build from DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	if build.TestMatrixID != "" {
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "A Test Matrix has already been started for this build."}))
	}
	postTestrequestModel := &testing.TestMatrix{}
	if err := json.NewDecoder(c.Request().Body).Decode(postTestrequestModel); err != nil {
		logger.Error("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	if postTestrequestModel.EnvironmentMatrix.AndroidDeviceList != nil {
		if err := firebaseutils.ValidateAndroidDevices(postTestrequestModel.EnvironmentMatrix.AndroidDeviceList.AndroidDevices); err != nil {
			return c.Render(http.StatusNotAcceptable, r.String("Invalid device configuration: %s", err))
		}
	}

	if postTestrequestModel.EnvironmentMatrix.IosDeviceList != nil {
		if err := firebaseutils.ValidateIosDevices(postTestrequestModel.EnvironmentMatrix.IosDeviceList.IosDevices); err != nil {
			return c.Render(http.StatusNotAcceptable, r.String("Invalid device configuration: %s", err))
		}
	}

	if timeout := postTestrequestModel.TestSpecification.TestTimeout; timeout != "" {
		secs, err := strconv.ParseFloat(strings.TrimSuffix(timeout, "s"), 32)
		if err == nil {
			maxSecs := maxTimeoutSecs(*postTestrequestModel.TestSpecification)
			if maxSecs > 0 && secs > float64(maxSecs) {
				logger.Warn(fmt.Sprintf("Incoming TestSpecification.TestTimeout '%s' from build '%s' exceeds limit of '%ds', overriding it to '%ds'", timeout, appSlug, maxSecs, maxSecs))
				postTestrequestModel.TestSpecification.TestTimeout = fmt.Sprintf("%ds", maxSecs)
			}
		}
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("[!] Exception: Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	startResp, err := fAPI.StartTestMatrix(appSlug, buildSlug, postTestrequestModel)
	if err != nil {
		logger.Error("[!] Exception: Failed to start Test Matrix", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("%s", err))
	}

	build.TestMatrixID = startResp.TestMatrixId

	startTime, err := time.Parse("2006-01-02T15:04:05.999Z", startResp.Timestamp)
	if err != nil {
		logger.Error("[!] Exception: Failed to parse startTime", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	build.TestStartTime = nulls.NewTime(startTime)
	build.TestEndTime = nulls.NewTime(startTime)
	build.LastRequest = nulls.NewTime(time.Now())

	err = database.UpdateBuild(build)
	if err != nil {
		logger.Error("[!] Exception: Failed to update DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	if postTestrequestModel.TestSpecification.IosXcTest == nil {
		testType := "robo"
		if postTestrequestModel.TestSpecification.AndroidInstrumentationTest != nil {
			testType = "instrumentation"
		}

		analyticsutils.SendTestingEvent(analyticsutils.EventTestingTestStarted,
			appSlug,
			buildSlug,
			testType,
			nil)
		analyticsutils.SendTestingEventDevices(analyticsutils.EventTestingTestStartedOnDevice,
			appSlug,
			buildSlug,
			testType,
			postTestrequestModel.EnvironmentMatrix.AndroidDeviceList.AndroidDevices,
			nil)
	} else {
		analyticsutils.SendTestingEvent(analyticsutils.EventIOSTestingTestStarted,
			appSlug,
			buildSlug,
			"",
			nil)
		analyticsutils.SendIOSTestingEventDevices(analyticsutils.EventIOSTestingTestStartedOnDevice,
			appSlug,
			buildSlug,
			"",
			postTestrequestModel.EnvironmentMatrix.IosDeviceList.IosDevices,
			nil)
	}

	return c.Render(http.StatusOK, r.String(""))
}

// TestAssetsGet ...
func TestAssetsGet(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	downloadUrlsModel, err := fAPI.DownloadTestAssets(buildSlug)
	if err != nil {
		logger.Error("Failed to get asset download urls", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	return c.Render(http.StatusOK, renderers.JSON(downloadUrlsModel))
}

// TestAssetUploadURLsAndroid handles request for Android test assets upload URLs
func TestAssetUploadURLsAndroid(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	appSlug := c.Param("app_slug")

	buildExists, err := database.IsBuildExists(appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to check if build exists", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	if buildExists {
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Build already exists"}))
	}

	var testAssetRequest firebaseutils.TestAssetsAndroid
	if err := json.NewDecoder(c.Request().Body).Decode(&testAssetRequest); err != nil {
		logger.Error("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	const maxObbFiles = 10
	if len(testAssetRequest.ObbFiles) > maxObbFiles {
		logger.Error(fmt.Sprintf("Number of obb fields requested is more than %d", maxObbFiles), zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": fmt.Sprintf("More than %d obb files requested", maxObbFiles)}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	resp, err := fAPI.TestAssetsUploadURLsAndroid(buildSlug, testAssetRequest)
	if err != nil {
		logger.Error("Failed to get Android upload URLs", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	err = database.AddBuild(&models.Build{BuildSlug: buildSlug, AppSlug: appSlug, LastRequest: nulls.NewTime(time.Now()), BuildSessionEnabled: true})
	if err != nil {
		logger.Error("Failed to save build", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	analyticsutils.SendUploadEvent(analyticsutils.EventUploadFileUploadRequested,
		appSlug,
		buildSlug)

	return c.Render(http.StatusOK, r.JSON(resp))
}

// TestAssetsPost ...
func TestAssetsPost(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	appSlug := c.Param("app_slug")

	buildExists, err := database.IsBuildExists(appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to check if build exists", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	if buildExists {
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Build already exists"}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	resp, err := fAPI.UploadTestAssets(buildSlug)
	if err != nil {
		logger.Error("Failed to get upload urls", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	err = database.AddBuild(&models.Build{BuildSlug: buildSlug, AppSlug: appSlug, LastRequest: nulls.NewTime(time.Now()), BuildSessionEnabled: true})
	if err != nil {
		logger.Error("Failed to save build", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	analyticsutils.SendUploadEvent(analyticsutils.EventUploadFileUploadRequested,
		appSlug,
		buildSlug)

	return c.Render(http.StatusOK, r.JSON(resp))
}

func isMessageAnError(message string) bool {
	errorMessages := []string{
		//"TEST_STATE_UNSPECIFIED",
		//"VALIDATING",
		//"PENDING",
		//"RUNNING",
		//"FINISHED",
		"ERROR",
		"UNSUPPORTED_ENVIRONMENT",
		"INCOMPATIBLE_ENVIRONMENT",
		"INCOMPATIBLE_ARCHITECTURE",
		"CANCELLED",
		"INVALID",
	}
	return sliceutil.IsStringInSlice(message, errorMessages)
}
