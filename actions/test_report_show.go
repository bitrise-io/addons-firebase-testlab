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

// TestReportShowHandler ...
func TestReportShowHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	testReportID := c.Param("test_report_id")
	status := c.Param("status")
	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		logger.Error("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	// Authorize for access via appSlug and buildSlug
	ok, err := database.TestReportExistsForAppAndBuild(testReportID, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to find test report in DB", zap.Any("error", errors.WithStack(err)))
	}
	if !ok {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
	}

	testReport := models.TestReport{}
	if err := database.FindTestReport(&testReport, testReportID); err != nil {
		logger.Error("Failed to find test report in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}
	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportWithTestSuite, err := testReportFiller.FillOne(testReport, fAPI, parser, &http.Client{}, status)
	if err != nil {
		logger.Error("Failed to enrich test report with JUNIT results", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(http.StatusOK, r.JSON(testReportWithTestSuite))
}
