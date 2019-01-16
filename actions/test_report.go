package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/pkg/errors"
)

type testReportsPostParams struct {
	Filename string `json:"filename"`
	Filesize int    `json:"filesize"`
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
	buildSlug := c.Param("build_slug")

	params := testReportsPostParams{}
	if err := json.NewDecoder(c.Request().Body).Decode(&params); err != nil {
		log.Errorf("Failed to decode request body, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode test report data"}))
	}

	testReport := &models.TestReport{
		Filename:  params.Filename,
		Filesize:  params.Filesize,
		Uploaded:  false,
		BuildSlug: buildSlug,
	}

	verrs, err := database.CreateTestReport(testReport)
	if err != nil {
		log.Errorf("Failed to create test report in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		log.Errorf("Failed to create Firebase API model, error: %s", err)
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	preSignedURL, err := fAPI.UploadURLforPath(testReport.PathInBucket())
	if err != nil {
		log.Errorf("Failed to create upload url, error: %s", err)
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
	id := c.Param("test_report_id")

	params := testReportPatchParams{}
	if err := json.NewDecoder(c.Request().Body).Decode(&params); err != nil {
		log.Errorf("Failed to decode request body, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode test report data"}))
	}

	tr := models.TestReport{}
	if err := database.FindTestReport(&tr, id); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Not found"}))
		}
		log.Errorf("Failed to find test report in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	tr.Uploaded = params.Uploaded

	verrs, err := database.UpdateTestReport(&tr)
	if err != nil {
		log.Errorf("Failed to update test report in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(tr))
}
