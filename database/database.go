package database

import (
	"fmt"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/pkg/errors"
)

// DB ...
var DB *pop.Connection

// InitDB ...
func InitDB() error {
	var err error
	DB, err = pop.Connect(configs.GetENV())
	if err != nil {
		return err
	}
	pop.Debug = configs.GetENV() == "development"
	return nil
}

// DeleteApp ...
func DeleteApp(appSlug string) error {
	err := DB.RawQuery("DELETE FROM apps WHERE apps.app_slug = ?", appSlug).Exec()
	if err != nil {
		return fmt.Errorf("failed to delete provision from DB: %s", err)
	}
	return nil
}

// IsAppExists ...
func IsAppExists(appSlug string) (bool, error) {
	ct, err := DB.Q().Where("apps.app_slug = ?", appSlug).Count(&models.App{})
	if err != nil {
		return false, fmt.Errorf("failed to check if appSlug exists: %s", err)
	}
	return ct > 0, nil
}

// IsAppExistsWithToken ...
func IsAppExistsWithToken(appSlug, apiToken string) (bool, error) {
	ct, err := DB.Q().Where("apps.app_slug = ? AND apps.api_token = ?", appSlug, apiToken).Count(&models.App{})
	if err != nil {
		return false, fmt.Errorf("failed to check if appSlug exists: %s", err)
	}
	return ct > 0, nil
}

// UpdateApp ...
func UpdateApp(app *models.App) error {
	plan := app.Plan
	err := DB.Q().Where("apps.app_slug = ?", app.AppSlug).First(app)
	if err != nil {
		return fmt.Errorf("failed to get appSlug from DB, error: %s", err)
	}

	app.Plan = plan

	err = DB.Save(app)
	if err != nil {
		return fmt.Errorf("failed to save appSlug to DB, error: %s", err)
	}
	return nil
}

// GetApp ...
func GetApp(app *models.App) (*models.App, error) {
	err := DB.Q().Where("apps.app_slug = ?", app.AppSlug).First(app)
	if err != nil {
		return nil, fmt.Errorf("failed to get appSlug from DB, error: %s", err)
	}

	return app, nil
}

// AddApp ...
func AddApp(app *models.App) error {
	err := DB.Create(app)
	if err != nil {
		return fmt.Errorf("failed to create appSlug in DB, error: %s", err)
	}
	return nil
}

// GetBuild ...
func GetBuild(appSlug, buildSlug string) (*models.Build, error) {
	build := &models.Build{}
	err := DB.Q().Where("builds.app_slug = ? AND builds.build_slug = ?", appSlug, buildSlug).First(build)
	if err != nil {
		return nil, fmt.Errorf("failed to get appSlug from DB, error: %s", err)
	}
	return build, nil
}

// AddBuild ...
func AddBuild(build *models.Build) error {
	err := DB.Create(build)
	if err != nil {
		return fmt.Errorf("failed to create buildSlug in DB, error: %s", err)
	}
	return nil
}

// IsBuildExists ...
func IsBuildExists(appSlug, buildSlug string) (bool, error) {
	ct, err := DB.Q().Where("builds.app_slug = ? AND builds.build_slug = ?", appSlug, buildSlug).Count(&models.Build{})
	if err != nil {
		return false, fmt.Errorf("failed to check if buildSlug exists: %s", err)
	}
	return ct > 0, nil
}

// IsTestReportExistsForBuild ...
func IsTestReportExistsForBuild(buildSlug string, testReportID string) (bool, error) {
	ct, err := DB.Q().Where("test_reports.build_slug = ? AND test_reports.id = ?", buildSlug, testReportID).Count(&models.TestReport{})
	if err != nil {
		return false, fmt.Errorf("failed to check if test report exists: %s", err)
	}
	return ct > 0, nil
}

// CloseBuildSession ...
func CloseBuildSession(appSlug, buildSlug string) error {
	build := &models.Build{}
	err := DB.Q().Where("builds.build_slug = ?", buildSlug).First(build)
	if err != nil {
		return fmt.Errorf("Failed to get buildSlug from DB, error: %s", err)
	}

	build.BuildSessionEnabled = false

	err = DB.Update(build)
	if err != nil {
		return fmt.Errorf("failed to update buildSlug in DB, error: %s", err)
	}
	return nil
}

// UpdateBuild ...
func UpdateBuild(build *models.Build) error {
	err := DB.Save(build)
	if err != nil {
		return fmt.Errorf("failed to save appSlug to DB, error: %s", err)
	}
	return nil
}

// GetAllExpiredOpenBuilds ...
func GetAllExpiredOpenBuilds() (*models.Builds, error) {
	builds := &models.Builds{}

	timeoutBetweenRequests := 2 * time.Minute

	err := DB.Q().Where("builds.last_request <= ? AND builds.build_session_enabled = ?", time.Now().Add(-timeoutBetweenRequests), true).All(builds)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds from DB, error: %s", err)
	}
	return builds, nil
}

// GetAllBuilds ...
func GetAllBuilds() (*models.Builds, error) {
	builds := &models.Builds{}

	err := DB.All(builds)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds from DB, error: %s", err)
	}
	return builds, nil
}

// GetBuildsCount ...
func GetBuildsCount() (int, error) {
	builds := &models.Builds{}

	count, err := DB.Count(builds)
	if err != nil {
		return 0, fmt.Errorf("failed to get builds from DB, error: %s", err)
	}
	return count, nil
}

// CreateTestReport ...
func CreateTestReport(tr *models.TestReport) (*validate.Errors, error) {
	verrs, err := DB.ValidateAndCreate(tr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return verrs, nil
}

// FindTestReport ...
func FindTestReport(tr *models.TestReport, id string) error {
	if err := DB.Find(tr, id); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UpdateTestReport ...
func UpdateTestReport(tr *models.TestReport) (*validate.Errors, error) {
	verrs, err := DB.ValidateAndUpdate(tr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return verrs, nil
}

// GetTestReports ...
func GetTestReports(trs *[]models.TestReport, appSlug string, buildSlug string) error {
	if err := DB.Where("app_slug = ? AND build_slug = ? AND uploaded = ?", appSlug, buildSlug, true).All(trs); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// TestReportExistsForAppAndBuild ...
func TestReportExistsForAppAndBuild(testReportID string, appSlug string, buildSlug string) (bool, error) {
	trs := []models.TestReport{}
	if err := DB.Where("id = ? AND app_slug = ? AND build_slug = ? AND uploaded = ?", testReportID, appSlug, buildSlug, true).Limit(1).All(&trs); err != nil {
		return false, errors.WithStack(err)
	}
	return len(trs) > 0, nil
}

// CreateTestReportAsset ...
func CreateTestReportAsset(tra *models.TestReportAsset) (*validate.Errors, error) {
	verrs, err := DB.ValidateAndCreate(tra)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return verrs, nil
}
