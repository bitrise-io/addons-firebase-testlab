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

// Totals ...
type Totals struct {
	Tests        int `json:"tests"`
	Passed       int `json:"passed"`
	Skipped      int `json:"skipped"`
	Failed       int `json:"failed"`
	Inconclusive int `json:"inconclusive"`
}

// TestSummaryResponseModel ...
type TestSummaryResponseModel struct {
	Totals Totals `json:"totals"`
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

	totals, err := GetTotals(appSlug, buildSlug, logger)
	if err != nil {
		logger.Error("Failed to get totals", zap.Error(err))
		return c.Render(http.StatusInternalServerError, r.String("Invalid request"))
	}

	return c.Render(http.StatusOK, r.JSON(TestSummaryResponseModel{
		Totals: totals,
	}))
}

// GetTotals ...
func GetTotals(appSlug, buildSlug string, logger *zap.Logger) (Totals, error) {
	testReportRecords := []models.TestReport{}
	err := database.GetTestReports(&testReportRecords, appSlug, buildSlug)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to find test reports in DB")
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to create Firebase API model")
	}
	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, fAPI, parser, &http.Client{}, "")
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to enrich test reports with JUNIT results")
	}

	var totals Totals

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
		return totals, nil
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		// no Firebase tests, it's fine, we can return
		return totals, nil
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to get test details")
	}

	testDetails, err := fillTestDetails(details, fAPI, logger)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to prepare test details data structure")
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
	return totals, nil
}
