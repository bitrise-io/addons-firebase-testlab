package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// TestReportResponseItem ...
type TestReportResponseItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func testReportResponseItemFromTestReport(tr models.TestReport) TestReportResponseItem {
	return TestReportResponseItem{
		ID:   tr.ID.String(),
		Name: tr.Name,
	}
}

func ftlReportItem() TestReportResponseItem {
	return TestReportResponseItem{
		ID:   "ftl",
		Name: "Firebase TestLab",
	}
}

// TestReportsListHandler ...
func TestReportsListHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildSlug := c.Param("build_slug")
	appSlug, ok := c.Session().Get("app_slug").(string)
	if !ok {
		logger.Error("Failed to get session data(app_slug)")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	testReports := []models.TestReport{}
	err := database.GetTestReports(&testReports, appSlug, buildSlug)
	if err != nil {
		logger.Error("Failed to find test reports in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	testReportsResponse := []TestReportResponseItem{}
	for _, tr := range testReports {
		testReportsResponse = append(testReportsResponse, testReportResponseItemFromTestReport(tr))
	}

	build, err := database.GetBuild(appSlug, buildSlug)
	if err != nil {
		return c.Render(http.StatusOK, r.JSON(testReportsResponse))
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		return c.Render(http.StatusOK, r.JSON(testReportsResponse))
	}

	testReportsResponse = append(testReportsResponse, ftlReportItem())

	return c.Render(http.StatusOK, r.JSON(testReportsResponse))
}
