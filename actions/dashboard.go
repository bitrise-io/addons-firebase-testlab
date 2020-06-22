package actions

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/analytics"
	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/metrics"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/renderers"
	"github.com/bitrise-io/addons-firebase-testlab/trackables"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	toolresults "google.golang.org/api/toolresults/v1beta3"
)

// Test ...
type Test struct {
	DeviceName   string         `json:"device_name,omitempty"`
	APILevel     string         `json:"api_level,omitempty"`
	Status       string         `json:"status,omitempty"` //pending,inProgress,complete
	TestResults  []TestResults  `json:"test_results,omitempty"`
	Outcome      string         `json:"outcome,omitempty"` //failure,inconclusive,success,skipped?
	Orientation  string         `json:"orientation,omitempty"`
	Locale       string         `json:"locale,omitempty"`
	StepID       string         `json:"step_id,omitempty"`
	OutputURLs   OutputURLModel `json:"output_urls,omitempty"`
	TestType     string         `json:"test_type,omitempty"`
	TestIssues   []TestIssue    `json:"test_issues,omitempty"`
	StepDuration int            `json:"step_duration_in_seconds,omitempty"`
}

// TestIssue ...
type TestIssue struct {
	Name       string `json:"name,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Stacktrace string `json:"stacktrace,omitempty"`
}

// OutputURLModel ...
type OutputURLModel struct {
	ScreenshotURLs  []string          `json:"screenshot_urls,omitempty"`
	VideoURL        string            `json:"video_url,omitempty"`
	ActivityMapURL  string            `json:"activity_map_url,omitempty"`
	TestSuiteXMLURL string            `json:"test_suite_xml_url,omitempty"`
	LogURLs         []string          `json:"log_urls,omitempty"`
	AssetURLs       map[string]string `json:"asset_urls,omitempty"`
}

// TestResults ...
type TestResults struct {
	Skipped int `json:"in_progress,omitempty"`
	Failed  int `json:"failed,omitempty"`
	Total   int `json:"total,omitempty"`
}

// RootGetHandler ...
func RootGetHandler(c buffalo.Context) error {
	client := metrics.NewDogStatsDMetrics("")
	defer client.Close()
	client.Track(trackables.Root{}, "rootPathOpened")

	return c.Render(http.StatusOK, r.String("Welcome to bitrise!"))
}

// DashboardLoginPostHandler ...
func DashboardLoginPostHandler(c buffalo.Context) error {
	timestamp := c.Request().FormValue("timestamp")
	token := c.Request().FormValue("token")
	appSlug := c.Request().FormValue("app_slug")
	buildSlug := c.Param("build_slug")
	appTitle := c.Param("app_title")
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	logger.Info("Login form data",
		zap.String("timestamp", timestamp),
		zap.String("token", token),
		zap.String("app_slug", appSlug),
		zap.String("build_slug", buildSlug),
	)

	ac := analytics.GetClient(logger)
	ac.SendAddonSSOLoginEvent(appSlug, "", "")

	appSlugStored, ok := c.Session().Get("app_slug").(string)
	if ok {
		if appSlug == appSlugStored {
			if buildSlug == "" {
				var err error
				buildSlug, err = fetchBuildSlug(appSlug)
				if err != nil {
					logger.Error("Failed to fetch latest build slug for app", zap.Error(err))
					return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
				}
			}
			return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/builds/%s", buildSlug))
		}
	}

	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		logger.Error("Failed to parse timestamp int", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	tm := time.Unix(i, 0)

	if time.Now().After(tm.Add(5 * time.Minute)) {
		logger.Error("Token expired", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Token expired"}))
	}

	hashPrefix := "sha256-"
	var hash hash.Hash
	if strings.HasPrefix(token, hashPrefix) {
		token = strings.TrimPrefix(token, hashPrefix)
		hash = sha256.New()
	} else {
		hash = sha1.New()
	}

	_, err = hash.Write([]byte(fmt.Sprintf("%s:%s:%s", appSlug, configs.GetAddonSSOToken(), timestamp)))
	if err != nil {
		logger.Error("Failed to write into sha1 buffer", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	refToken := fmt.Sprintf("%x", hash.Sum(nil))

	if token != refToken {
		logger.Error("Token mismatch")
		c.Session().Clear()
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Forbidden, invalid credentials"}))
	}

	c.Session().Set("app_slug", appSlug)
	c.Session().Set("app_title", appTitle)

	err = c.Session().Save()
	if err != nil {
		logger.Error("Failed to save session", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	if buildSlug == "" {
		var err error
		buildSlug, err = fetchBuildSlug(appSlug)
		if err != nil {
			logger.Error("Failed to fetch latest build slug for app", zap.Error(err))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}
	}

	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/builds/%s", buildSlug))
}

//
// API endpoints

// DashboardAppGetHandler ...
func DashboardAppGetHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		logger.Error("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}
	appTitle, ok := c.Session().Get("app_title").(string)
	if !ok {
		logger.Error("Failed to get session data(app_title)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}
	return c.Render(http.StatusOK, r.JSON(map[string]string{"app_slug": appSlug, "app_title": appTitle}))
}

// StepAPIGetHandler ...
func StepAPIGetHandler(c buffalo.Context) error {
	stepID := c.Param("step_id")
	buildSlug := c.Param("build_slug")
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		logger.Error("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to get build from DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	samples, err := fAPI.GetTestMetricSamples(build.TestHistoryID, build.TestExecutionID, stepID, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to get sample data", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	return c.Render(http.StatusOK, renderers.JSON(samples))
}

// DashboardAPIGetHandler ...
func DashboardAPIGetHandler(c buffalo.Context) error {
	buildSlug := c.Param("build_slug")
	status := c.Param("status")

	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		logger.Error("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		logger.Error("No TestHistoryID or TestExecutionID found for build", zap.String("build_slug", build.BuildSlug))
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to get test details", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	//
	// prepare data structure
	testDetails, err := fillTestDetails(details, fAPI, logger)
	if err != nil {
		logger.Error("Failed to prepare test details data structure", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	if status != "" {
		testDetails = filterTestsByStatus(testDetails, status)
	}

	return c.Render(200, renderers.JSON(testDetails))
}

//
// STATIC endpoints

// DashboardGetHandler ...
func DashboardGetHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("frontend/dashboard.html"))
}

// DashboardDetailsGetHandler ...
func DashboardDetailsGetHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("frontend/details.html"))
}

// DashboardIndexGetHandler ...
func DashboardIndexGetHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("frontend/index.html"))
}

func filterTestsByStatus(tests []*Test, status string) []*Test {
	filteredTests := []*Test{}

	for _, test := range tests {
		if statusMatch(test.Outcome, status) || test.Status == "inProgress" { // include currently running tests too
			filteredTests = append(filteredTests, test)
		}
	}

	return filteredTests
}

func statusMatch(testStatus string, expected string) bool {
	if testStatus == expected {
		return true
	}

	if testStatus == "success" && expected == "passed" {
		return true
	}

	if testStatus == "failure" && expected == "failed" {
		return true
	}

	return false
}

func fillTestDetails(details *toolresults.ListStepsResponse, fAPI *firebaseutils.APIModel, logger *zap.Logger) ([]*Test, error) {
	testDetails := make([]*Test, len(details.Steps))

	var wg sync.WaitGroup
	wg.Add(len(details.Steps))
	errChannel := make(chan error, 1)

	for index, d := range details.Steps {
		go func(detail *toolresults.Step, i int) {
			defer func() {
				wg.Done()
			}()
			test := &Test{}
			for _, dimension := range detail.DimensionValue {
				switch dimension.Key {
				case "Model":
					test.DeviceName = firebaseutils.GetDeviceNameByID(dimension.Value)
				case "Version":
					prefixByPlatform := "API Level"
					if strings.Contains(strings.ToLower(detail.Name), "ios") {
						prefixByPlatform = "iOS"
					}
					test.APILevel = fmt.Sprintf("%s %s", prefixByPlatform, dimension.Value)
				case "Locale":
					test.Locale = firebaseutils.GetLangByCountryCode(dimension.Value)
				case "Orientation":
					test.Orientation = dimension.Value
				}
			}

			if detail.Outcome != nil {
				test.Outcome = detail.Outcome.Summary
			}
			test.Status = detail.State
			test.StepID = detail.StepId

			if detail.TestExecutionStep != nil {
				if len(detail.TestExecutionStep.TestIssues) > 0 {
					test.TestIssues = []TestIssue{}
					for _, issue := range detail.TestExecutionStep.TestIssues {
						testIssue := TestIssue{Name: issue.ErrorMessage}
						if issue.StackTrace != nil {
							testIssue.Stacktrace = issue.StackTrace.Exception
						}
						test.TestIssues = append(test.TestIssues, testIssue)
					}
				}
				outputURLs := OutputURLModel{}
				outputURLs.ScreenshotURLs = []string{}
				outputURLs.AssetURLs = map[string]string{}
				if detail.TestExecutionStep.TestTiming != nil {
					if detail.TestExecutionStep.TestTiming.TestProcessDuration != nil {
						test.StepDuration = int(detail.TestExecutionStep.TestTiming.TestProcessDuration.Seconds)
					}
				}

				test.TestResults = []TestResults{}
				for _, overview := range detail.TestExecutionStep.TestSuiteOverviews {
					testResult := TestResults{Total: int(overview.TotalCount), Failed: int(overview.FailureCount), Skipped: int(overview.SkippedCount)}
					test.TestResults = append(test.TestResults, testResult)
				}

				if detail.TestExecutionStep.ToolExecution != nil {
					//get logcat
					for _, testlog := range detail.TestExecutionStep.ToolExecution.ToolLogs {
						//create signed url for assets
						signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(testlog.FileUri)
						if err != nil {
							logger.Error("Failed to get signed url",
								zap.String("file_uri", testlog.FileUri),
								zap.Any("error", errors.WithStack(err)),
							)
							if len(errChannel) == 0 {
								errChannel <- err
							}
							return
						}

						outputURLs.LogURLs = append(outputURLs.LogURLs, signedURL)
					}

					// parse output files by type
					for _, output := range detail.TestExecutionStep.ToolExecution.ToolOutputs {
						{
							if strings.Contains(output.Output.FileUri, "results/") {
								//create signed url for asset
								signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
								if err != nil {
									logger.Error("Failed to get signed url",
										zap.String("output_file_uri", output.Output.FileUri),
										zap.Any("error", errors.WithStack(err)),
									)
									if len(errChannel) == 0 {
										errChannel <- err
									}
									return
								}
								resultAbsPath := strings.Join(strings.Split(strings.Split(output.Output.FileUri, "results/")[1], "/")[1:], "/")
								outputURLs.AssetURLs[resultAbsPath] = signedURL
							}
						}

						if strings.HasSuffix(output.Output.FileUri, "video.mp4") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.VideoURL = signedURL
						}

						if strings.HasSuffix(output.Output.FileUri, "sitemap.png") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.ActivityMapURL = signedURL
						}

						if strings.HasSuffix(output.Output.FileUri, ".png") && !strings.HasSuffix(output.Output.FileUri, "sitemap.png") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.ScreenshotURLs = append(outputURLs.ScreenshotURLs, signedURL)
						}
					}
				}
				if detail.TestExecutionStep.TestSuiteOverviews != nil {
					//get xmls
					for _, overview := range detail.TestExecutionStep.TestSuiteOverviews {
						//create signed url for assets
						signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(overview.XmlSource.FileUri)
						if err != nil {
							logger.Error("Failed to get signed url",
								zap.String("xml_source_file_uri", overview.XmlSource.FileUri),
								zap.Any("error", errors.WithStack(err)),
							)
							if len(errChannel) == 0 {
								errChannel <- err
							}
							return
						}

						outputURLs.TestSuiteXMLURL = signedURL
					}
				}
				test.OutputURLs = outputURLs
			}

			if test.OutputURLs.ActivityMapURL != "" {
				test.TestType = "robo"
			}
			if test.OutputURLs.TestSuiteXMLURL != "" {
				test.TestType = "instrumentation"
			}

			testDetails[i] = test
		}(d, index)
	}
	wg.Wait()
	close(errChannel)

	var err error
	err = <-errChannel
	return testDetails, err
}

func fetchBuildSlug(appSlug string) (string, error) {
	app, err := database.GetApp(&models.App{AppSlug: appSlug})
	if err != nil {
		return "", errors.WithStack(err)
	}
	bc := bitrise.NewClient(app.APIToken)
	build, err := bc.GetLatestBuildOfApp(appSlug)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return build.Slug, nil
}
