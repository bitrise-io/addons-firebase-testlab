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

type totals struct {
	Tests        int `json:"tests"`
	Passed       int `json:"passed"`
	Skipped      int `json:"skipped"`
	Failed       int `json:"failed"`
	Inconclusive int `json:"inconclusive"`
}

// TestSummaryResponseModel ...
type TestSummaryResponseModel struct {
	Totals totals `json:"totals"`
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

	testReportRecords := []models.TestReport{}
	err := database.GetTestReports(&testReportRecords, appSlug, buildSlug)
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

	testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, fAPI, parser, &http.Client{}, "")
	if err != nil {
		logger.Error("Failed to enrich test reports with JUNIT results", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	totals := totals{}

	for _, testReport := range testReportsWithTestSuites {
		for _, testSuite := range testReport.TestSuites {
			totals.Passed = totals.Passed + testSuite.Totals.Passed
			totals.Failed = totals.Failed + testSuite.Totals.Failed + testSuite.Totals.Error
			totals.Skipped = totals.Skipped + testSuite.Totals.Skipped
			totals.Tests = totals.Tests + testSuite.Totals.Tests
		}
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		// no Firebase tests, it's fine, we can return
		return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
			Totals: totals,
		}))
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		// no Firebase tests, it's fine, we can return
		return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
			Totals: totals,
		}))
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to get test details", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	testDetails, err := fillTestDetails(details, fAPI, logger)
	if err != nil {
		logger.Error("Failed to prepare test details data structure", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
	}

	for _, testDetail := range testDetails {
		switch testDetail.Outcome {
		case "success":
			totals.Passed++
		case "failure":
			totals.Failed++
		case "skipped":
			totals.Skipped++
		case "inconclusive":
			totals.Inconclusive++
		}
	}
	return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
		Totals: totals,
	}))
}
