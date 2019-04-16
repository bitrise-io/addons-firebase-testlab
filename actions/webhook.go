package actions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
)

const (
	abortedBuildStatus      int    = 3
	buildTriggeredEventType string = "build/triggered"
	buildFinishedEventType  string = "build/finished"
)

// GitData ...
type GitData struct {
	Provider      string `json:"provider"`
	SrcBranch     string `json:"src_branch"`
	DstBranch     string `json:"dst_branch"`
	PullRequestID int    `json:"pull_request_id"`
}

// AppData ...
type AppData struct {
	AppSlug                string  `json:"app_slug"`
	BuildSlug              string  `json:"build_slug"`
	BuildNumber            int     `json:"build_number"`
	BuildStatus            int     `json:"build_status"`
	BuildTriggeredWorkflow string  `json:"build_triggered_workflow"`
	Git                    GitData `json:"git"`
}

// WebhookHandler ...
func WebhookHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildType := c.Request().Header.Get("Bitrise-Event-Type")

	if buildType != buildTriggeredEventType && buildType != buildFinishedEventType {
		logger.Error("Invalid Bitrise event type")
		return c.Render(http.StatusInternalServerError, r.String("Invalid Bitrise event type"))
	}

	appData := &AppData{}
	if err := json.NewDecoder(c.Request().Body).Decode(appData); err != nil {
		return c.Render(http.StatusBadRequest, r.String("Request body has invalid format"))
	}

	app := &models.App{AppSlug: appData.AppSlug}
	app, err := database.GetApp(app)
	if err != nil {
		logger.Error("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal error"))
	}

	switch buildType {
	case buildFinishedEventType:
		if appData.BuildStatus == abortedBuildStatus {
			build, err := database.GetBuild(app.AppSlug, appData.BuildSlug)
			if err != nil {
				logger.Error("Failed to get build from database",
					zap.String("app_slug", app.AppSlug),
					zap.String("app_data_app_slug", appData.AppSlug),
					zap.String("build_slug", appData.BuildSlug),
					zap.Any("error", errors.WithStack(err)),
				)
				return c.Render(http.StatusInternalServerError, r.String("Internal error"))
			}
			if build.TestExecutionID != "" {
				_, err := firebaseutils.CancelTestMatrix(build.TestMatrixID)
				if err != nil {
					return fmt.Errorf("Failed to cancel test matrix(id: %s), error: %+v", build.TestMatrixID, err)
				}
			}
		}
	case buildTriggeredEventType:
		// Don't care
	default:
		logger.Error("Invalid build type", zap.String("build_event_type", buildType))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(200, r.JSON(app))
}
