package actions

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
)

const (
	abortedBuildStatus int = 3
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
	BuildTriggeredWorkflow string  `json:"build_triggered_workflow"`
	Git                    GitData `json:"git"`
}

// WebhookHandler ...
func WebhookHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	buildType := c.Request().Header.Get("Bitrise-Event-Type")

	if buildType != "build/triggered" && buildType != "build/finished" {
		logger.Error("Invalid Bitrise event type")
		return c.Render(http.StatusInternalServerError, r.String("Invalid Bitrise event type"))
	}

	appData := &AppData{}
	if err := json.NewDecoder(c.Request().Body).Decode(appData); err != nil {
		return c.Render(http.StatusBadRequest, r.String("Request body has invalid format"))
	}

	app := &models.App{AppSlug: appData.AppSlug}
	app, err = database.GetApp(app)
	if err != nil {
		logger.Errorf("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal Server Error"))
	}

	client := bitrise.NewClient(app.BitriseAPIToken)
	_, build, err := client.GetBuildOfApp(appData.BuildSlug, app.AppSlug)
	if err != nil {
		logger.Errorf("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.String("Internal Server Error"))
	}

	if build.Status == abortedBuildStatus {
		// will do something
	}

	return c.Render(200, nil)
}
