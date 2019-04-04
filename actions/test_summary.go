package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type totalsModel struct {
	tests        int64 `json:"tests'`
	passed       int64 `json:"passed'`
	skipped      int64 `json:"skipped'`
	failure      int64 `json:"failure'`
	inconclusive int64 `json:"inconclusive'`
}

// TestSummaryResponseModel ...
type TestSummaryResponseModel struct {
	Totals totalsModel `json:"totals"`
}

// TestSummaryHandler ...
func TestSummaryHandler(c buffalo.Context) error {
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
		return c.Render(http.StatusNoContent, r.String("Invalid request"))
	}

	testReportRecords := []models.TestReport{}
	err = database.GetTestReports(&testReportRecords, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to find test reports in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}
	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportsWithTestSuites, err := testReportFiller.Fill(testReportRecords, fAPI, parser, &http.Client{})
	if err != nil {
		logger.Error("Failed to enrich test reports with JUNIT results", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	totals := totalsModel{}

	for _, testReport := range testReportsWithTestSuites {
		for _, testSuite := range testReport.TestSuites {
			totals.passed = totals.passed + testSuite.Totals.Passed
			totals.failure = totals.failure + testSuite.Totals.Failed
			totals.failure = totals.failure + testSuite.Totals.Error
			totals.skipped = totals.passed + testSuite.Totals.Skipped
			totals.tests++
		}
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		logger.Error("No TestHistoryID or TestExecutionID found for build", zap.String("build_slug", build.BuildSlug))
		return c.Render(http.StatusNoContent, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		// no Firebase tests, it's fine, we can return
		return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
			Totals: totals,
		}))
	}

	testDetails := make([]*Test, len(details.Steps))

	for _, testDetail := range testDetails {
		switch testDetail.Outcome {
		case "success":
			totals.passed++
		case "failure":
			totals.failure++
		case "skipped":
			totals.skipped++
		case "inconclusive":
			totals.inconclusive++
		}
	}
	return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
		Totals: totals,
	}))
}
