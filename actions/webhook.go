package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/bitrise-io/addons-firebase-testlab/analytics"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
)

const (
	abortedBuildStatus      int    = 3
	buildTriggeredEventType string = "build/triggered"
	buildFinishedEventType  string = "build/finished"
)

// GitData ...
type GitData struct {
	Provider      string `json:"provider"`
	SrcBranch     string `json:"src_branch"`
	DstBranch     string `json:"dst_branch"`
	PullRequestID int    `json:"pull_request_id"`
}

// AppData ...
type AppData struct {
	AppSlug                string  `json:"app_slug"`
	BuildSlug              string  `json:"build_slug"`
	BuildNumber            int     `json:"build_number"`
	BuildStatus            int     `json:"build_status"`
	BuildTriggeredWorkflow string  `json:"build_triggered_workflow"`
	Git                    GitData `json:"git"`
}

// WebhookHandler ...
func WebhookHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildType := c.Request().Header.Get("Bitrise-Event-Type")

	if buildType != buildTriggeredEventType && buildType != buildFinishedEventType {
		logger.Error("Invalid Bitrise event type")
		return c.Render(http.StatusInternalServerError, r.String("Invalid Bitrise event type"))
	}

	appData := &AppData{}
	if err := json.NewDecoder(c.Request().Body).Decode(appData); err != nil {
		return c.Render(http.StatusBadRequest, r.String("Request body has invalid format"))
	}

	app := &models.App{AppSlug: appData.AppSlug}
	app, err := database.GetApp(app)
	if err != nil {
		logger.Error("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	switch buildType {
	case buildFinishedEventType:
		build := (*models.Build)(nil)
		if appData.BuildStatus == abortedBuildStatus {
			var err error
			build, err = database.GetBuild(app.AppSlug, appData.BuildSlug)
			if err != nil {
				return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
			}
			if build.TestExecutionID != "" {
				_, err := firebaseutils.CancelTestMatrix(build.TestMatrixID)
				if err != nil {
					return fmt.Errorf("Failed to cancel test matrix(id: %s), error: %+v", build.TestMatrixID, err)
				}
			}
		}

		ac := analytics.GetClient(logger)
		totals, err := GetTotals(app.AppSlug, appData.BuildSlug, logger)
		if err != nil {
			logger.Warn("Failed to get totals of test", zap.Any("app_data", appData), zap.Error(err))
			return c.Render(200, r.JSON(app))
		}

		switch {
		case totals.Failed > 0 || totals.Inconclusive > 0:
			ac.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "fail", totals.Tests, time.Now())
		case totals != (Totals{}):
			ac.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "success", totals.Tests, time.Now())
		case totals == (Totals{}):
			ac.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "empty", totals.Tests, time.Now())
		default:
			ac.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "null", totals.Tests, time.Now())
		}

		testReportRecords := []models.TestReport{}
		err = database.GetTestReports(&testReportRecords, app.AppSlug, appData.BuildSlug)
		if err != nil {
			return errors.Wrap(err, "Failed to find test reports in DB")
		}

		ac.NumberOfTestReports(app.AppSlug, appData.BuildSlug, len(testReportRecords), time.Now())

		fAPI, err := firebaseutils.New()
		if err != nil {
			return errors.Wrap(err, "Failed to create Firebase API model")
		}
		parser := &junit.Client{}
		testReportFiller := testreportfiller.Filler{}

		testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, fAPI, parser, &http.Client{}, "")
		if err != nil {
			return errors.Wrap(err, "Failed to enrich test reports with JUNIT results")
		}
		for _, tr := range testReportsWithTestSuites {
			result := "success"
			for _, ts := range tr.TestSuites {
				if ts.Totals.Failed > 0 || totals.Inconclusive > 0 {
					result = "fail"
					break
				}
			}
			ac.TestReportResult(app.AppSlug, appData.BuildSlug, result, "unit", tr.ID, time.Now())
		}

		if build != nil && build.TestHistoryID != "" && build.TestExecutionID != "" {
			details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, app.AppSlug, appData.BuildSlug)
			if err != nil {
				logger.Error("Failed to get test details", zap.Any("error", errors.WithStack(err)))
				return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
			}

			testDetails, err := fillTestDetails(details, fAPI, logger)
			if err != nil {
				logger.Error("Failed to prepare test details data structure", zap.Any("error", errors.WithStack(err)))
				return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
			}
			result := "success"
			for _, detail := range testDetails {
				outcome := detail.Outcome
				if outcome == "failure" {
					result = "failed"
				}
				if result != "failed" {
					if outcome == "skipped" || outcome == "inconclusive" {
						result = outcome
					}
				}
			}

			ac.TestReportResult(app.AppSlug, appData.BuildSlug, result, "ui", uuid.UUID{}, time.Now())
		}
	case buildTriggeredEventType:
		// Don't care
	default:
		logger.Error("Invalid build type", zap.String("build_event_type", buildType))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(200, r.JSON(app))
}
