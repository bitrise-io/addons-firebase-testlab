package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/bitrise-io/go-utils/log"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
)

// TestReportsListHandler ...
func TestReportsListHandler(c buffalo.Context) error {
	buildSlug := c.Param("build_slug")

	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		log.Errorf("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	testReportRecords := []models.TestReport{}
	err := database.GetTestReports(&testReportRecords, appSlug, buildSlug)
	if err != nil {
		log.Errorf("Failed to find test reports in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		log.Errorf("Failed to create Firebase API model, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}
	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportsWithTestSuites, err := testReportFiller.Fill(testReportRecords, fAPI, parser, &http.Client{})
	if err != nil {
		log.Errorf("Failed to enrich test reports with JUNIT results, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(http.StatusOK, r.JSON(testReportsWithTestSuites))
}
