package stepresult

import (
	"net/http"

	junitparser "github.com/joshdk/go-junit"

	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
)

// CreateTestStepResult ...
func CreateTestStepResult(id uuid.UUID) error {
	testReport := models.TestReport{}
	if err := database.FindTestReport(&testReport, id.String()); err != nil {
		return errors.WithStack(err)
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		return errors.WithStack(err)
	}

	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportWithTestSuite, err := testReportFiller.FillOne(testReport, fAPI, parser, &http.Client{}, "failed")
	if err != nil {
		return errors.WithStack(err)
	}

	failedTests := []junitparser.Test{}
	total := 0
	for _, suite := range testReportWithTestSuite.TestSuites {
		total += suite.Totals.Tests
		for _, test := range suite.Tests {
			failedTests = append(failedTests, test)
		}
	}

	status := "success"
	if len(failedTests) > 0 {
		status = "failed"
	}

	testStepResult := bitrise.TestStepResult{
		StepResult: bitrise.StepResult{
			Name:   testReport.Name,
			Status: status,
		},
		Total:       total,
		FailedTests: failedTests,
	}

	app := &models.App{AppSlug: testReport.AppSlug}
	app, err = database.GetApp(app)
	if err != nil {
		return errors.WithStack(err)
	}

	client := bitrise.NewClient(app.BitriseAPIToken)

	if err := client.CreateTestStepResult(testReport.AppSlug, testReport.BuildSlug, &testStepResult); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
