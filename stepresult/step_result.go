package stepresult

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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

const lintStepIDListenvKey = "BITRISE_LINT_STEPS"

// CreateStepResult ...
func CreateStepResult(id uuid.UUID) error {
	testReport := models.TestReport{}
	if err := database.FindTestReport(&testReport, id.String()); err != nil {
		return errors.WithStack(err)
	}

	stepInfo := models.StepInfo{}
	if err := json.Unmarshal(testReport.Step, &stepInfo); err != nil {
		return errors.WithStack(err)
	}

	lintSteps := make(map[string]bool)
	for _, stepID := range strings.Split(getEnv(lintStepIDListenvKey, ""), ",") {
		lintSteps[stepID] = true
	}

	if lintSteps[stepInfo.ID] {
		return CreateLintStepResult(testReport, stepInfo)
	}

	return CreateTestStepResult(testReport, stepInfo)
}

// CreateTestStepResult ...
func CreateTestStepResult(testReport models.TestReport, stepInfo models.StepInfo) error {
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

	name := stepInfo.Title
	if len(testReport.Name) > 0 && testReport.Name != stepInfo.Title {
		name = fmt.Sprintf("%s (%s)", stepInfo.Title, testReport.Name)
	}

	status := "success"
	if len(failedTests) > 0 {
		status = "failed"
	}

	testStepResult := bitrise.TestStepResult{
		StepResult: bitrise.StepResult{
			Name:   name,
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

// CreateLintStepResult ...
func CreateLintStepResult(testReport models.TestReport, stepInfo models.StepInfo) error {
	fAPI, err := firebaseutils.New()
	if err != nil {
		return errors.WithStack(err)
	}

	testReportFiller := testreportfiller.Filler{}

	annotations, err := testReportFiller.Annotate(testReport, fAPI, &http.Client{})
	if err != nil {
		return errors.WithStack(err)
	}

	name := stepInfo.Title
	if len(testReport.Name) > 0 && testReport.Name != stepInfo.Title {
		name = fmt.Sprintf("%s (%s)", stepInfo.Title, testReport.Name)
	}

	status := "success"
	if len(annotations) > 0 {
		status = "failed"
	}

	lintStepResult := bitrise.LintStepResult{
		StepResult: bitrise.StepResult{
			Name:   name,
			Status: status,
		},
		Annotations: annotations,
	}

	app := &models.App{AppSlug: testReport.AppSlug}
	app, err = database.GetApp(app)
	if err != nil {
		return errors.WithStack(err)
	}

	client := bitrise.NewClient(app.BitriseAPIToken)

	if err := client.CreateLintStepResult(testReport.AppSlug, testReport.BuildSlug, &lintStepResult); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}
