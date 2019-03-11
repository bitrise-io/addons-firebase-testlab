package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/pkg/errors"
)

type testReportsPostParams struct {
	Filename string          `json:"filename"`
	Filesize int             `json:"filesize"`
	Step     models.StepInfo `json:"step"`
}

type testReportPatchParams struct {
	Uploaded bool `json:"uploaded"`
}

type testReportWithUploadURL struct {
	models.TestReport
	UploadURL string `json:"upload_url"`
}

func newTestReportWithUploadURL(testReport models.TestReport, uploadURL string) testReportWithUploadURL {
	return testReportWithUploadURL{
		testReport,
		uploadURL,
	}
}

// TestReportsPostHandler ...
func TestReportsPostHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug := c.Param("app_slug")
	buildSlug := c.Param("build_slug")
	params := testReportsPostParams{}
	if err := json.NewDecoder(c.Request().Body).Decode(&params); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode test report data"}))
	}

	stepInfo, err := json.Marshal(params.Step)
	if err != nil {
		logger.Error("Failed to marshal step info", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	testReport := &models.TestReport{
		Filename:  params.Filename,
		Filesize:  params.Filesize,
		Step:      stepInfo,
		Uploaded:  false,
		AppSlug:   appSlug,
		BuildSlug: buildSlug,
	}

	verrs, err := database.CreateTestReport(testReport)
	if err != nil {
		logger.Error("Failed to create test report in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	fAPI, err := firebaseutils.New(nil)
	if err != nil {
		logger.Error("Failed to create Firebase API model", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	preSignedURL, err := fAPI.UploadURLforPath(testReport.PathInBucket())
	if err != nil {
		logger.Error("Failed to create upload url, error: %s", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	testReportWithUploadURL := newTestReportWithUploadURL(*testReport, preSignedURL)

	// Default JSON renderer would mess up the URL encoding
	return c.Render(201, r.Func("application/json", func(w io.Writer, d render.Data) error {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(testReportWithUploadURL); err != nil {
			return errors.Wrapf(err, "Failed to respond (encode) with JSON for response model: %#v", testReportWithUploadURL)
		}
		return nil
	}))
}

// TestReportPatchHandler ...
func TestReportPatchHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	id := c.Param("test_report_id")
	params := testReportPatchParams{}
	if err := json.NewDecoder(c.Request().Body).Decode(&params); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode test report data"}))
	}

	tr := models.TestReport{}
	if err := database.FindTestReport(&tr, id); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
		}
		logger.Error("Failed to find test report in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	tr.Uploaded = params.Uploaded

	verrs, err := database.UpdateTestReport(&tr)
	if err != nil {
		logger.Error("Failed to update test report in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(tr))
}
