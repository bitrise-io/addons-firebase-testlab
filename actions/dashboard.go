package actions

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/analyticsutils"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/renderers"
	"github.com/bitrise-io/go-utils/log"
	"github.com/gobuffalo/buffalo"
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
	return c.Render(http.StatusOK, r.String("Welcome to bitrise!"))
}

// DashboardLoginPostHandler ...
func DashboardLoginPostHandler(c buffalo.Context) error {
	timestamp := c.Request().FormValue("timestamp")
	token := c.Request().FormValue("token")
	appSlug := c.Request().FormValue("app_slug")
	buildSlug := c.Param("build_slug")

	fmt.Printf("Login form data - timestamp: %s, token: %s, appSlug: %s, buildSlug: %s", timestamp, token, appSlug, buildSlug)

	analyticsutils.SendAddonEvent(analyticsutils.EventAddonSSOLogin, appSlug, "", "")

	appSlugStored, ok := c.Session().Get("app_slug").(string)
	if ok {
		fmt.Printf("stored appSlug: %s", appSlugStored)
		if appSlug == appSlugStored {
			fmt.Printf("appSlug already saved, redirect...")
			return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/builds/%s", buildSlug))
		}
	}

	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Errorf("Failed to parse timestamp int, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	tm := time.Unix(i, 0)

	if time.Now().After(tm.Add(5 * time.Minute)) {
		log.Errorf("Token expired, error: %s", err)
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Token expired"}))
	}
	log.Printf("Token is still not expired")

	s := sha1.New()

	_, err = s.Write([]byte(fmt.Sprintf("%s:%s:%s", appSlug, configs.GetAddonSSOToken(), timestamp)))
	if err != nil {
		log.Errorf("Failed to write into sha1 buffer, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	refToken := fmt.Sprintf("%x", s.Sum(nil))
	log.Printf("refToken: %s", refToken)

	if token != refToken {
		log.Errorf("Token mismatch")
		c.Session().Clear()
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Forbidden, invalid credentials"}))
	}

	log.Printf("token is allright, save session")
	c.Session().Set("app_slug", appSlug)

	err = c.Session().Save()
	if err != nil {
		log.Errorf("Failed to save session, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	log.Printf("redirect...")

	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/builds/%s", buildSlug))
}

//
// API endpoints

// StepAPIGetHandler ...
func StepAPIGetHandler(c buffalo.Context) error {
	stepID := c.Param("step_id")
	buildSlug := c.Param("build_slug")

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		log.Errorf("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		log.Errorf("Failed to get build from DB, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		log.Errorf("Failed to create Firebase API model, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	samples, err := fAPI.GetTestMetricSamples(build.TestHistoryID, build.TestExecutionID, stepID)
	if err != nil {
		log.Errorf("Failed to get sample data, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	return c.Render(http.StatusOK, renderers.JSON(samples))
}

// DashboardAPIGetHandler ...
func DashboardAPIGetHandler(c buffalo.Context) error {
	buildSlug := c.Param("build_slug")

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		log.Errorf("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		log.Errorf("Failed to get build from DB, error: %s", err)
		return c.Render(http.StatusNoContent, r.String("Invalid request"))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		log.Errorf("Failed to create Firebase API model, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		log.Errorf("No TestHistoryID or TestExecutionID found for build: %s", build.BuildSlug)
		return c.Render(http.StatusNoContent, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID)
	if err != nil {
		log.Errorf("Failed to get test details, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	//
	// prepare data structure
	testDetails := make([]*Test, len(details.Steps))

	// wait group
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
							log.Errorf("Failed to get signed url for: %s, error: %s", testlog.FileUri, err)
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
									log.Errorf("Failed to get signed url for: %s, error: %s", output.Output.FileUri, err)
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
								log.Errorf("Failed to get signed url for: %s, error: %s", output.Output.FileUri, err)
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
								log.Errorf("Failed to get signed url for: %s, error: %s", output.Output.FileUri, err)
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
								log.Errorf("Failed to get signed url for: %s, error: %s", output.Output.FileUri, err)
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
							log.Errorf("Failed to get signed url for: %s, error: %s", overview.XmlSource.FileUri, err)
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

	err = <-errChannel
	if err != nil {
		log.Errorf("One of the requests is failed. Error: %s", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
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
