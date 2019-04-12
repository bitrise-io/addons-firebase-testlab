package actions

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/analyticsutils"
	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"go.uber.org/zap"

	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
)

// ProvisionData ...
type ProvisionData struct {
	Plan            string `json:"plan"`
	AppSlug         string `json:"app_slug"`
	BitriseAPIToken string `json:"api_token"`
}

// Env ...
type Env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ProvisionPostHandler ...
func ProvisionPostHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	provData := &ProvisionData{}

	err := json.NewDecoder(c.Request().Body).Decode(provData)
	if err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode provisioning data"}))
	}

	exists, err := database.IsAppExists(provData.AppSlug)
	if err != nil {
		logger.Error("Failed to check if App exists in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if exists {
		logger.Warn("  [!] App already exists")
		//return c.Render(http.StatusConflict, r.JSON(map[string]string{"error": "App already exists"}))
	}

	envs := map[string][]Env{}
	hostURL := configs.GetAddonHost()
	envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_URL", Value: fmt.Sprintf("%s/test", hostURL)})

	app := &models.App{
		AppSlug:         provData.AppSlug,
		Plan:            provData.Plan,
		BitriseAPIToken: provData.BitriseAPIToken,
	}

	if !exists {
		app.APIToken = generateRandomHash(50)

		err = database.AddApp(app)
		if err != nil {
			logger.Error("Failed to add app to DB", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		analyticsutils.SendAddonEvent(analyticsutils.EventAddonProvisioned, app.AppSlug, "", app.Plan)

		client := bitrise.NewClient(app.BitriseAPIToken)
		_, err = client.RegisterWebhook(app)
		if err != nil {
			logger.Error("Failed to register webhook for app", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
		return c.Render(200, r.JSON(envs))
	}

	app, err = database.GetApp(app)
	if err != nil {
		logger.Error("Failed to get app from DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
	return c.Render(200, r.JSON(envs))
}

// ProvisionPutHandler ...
func ProvisionPutHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug := c.Param("app_slug")
	provData := &ProvisionData{}
	err := json.NewDecoder(c.Request().Body).Decode(provData)
	if err != nil {
		logger.Error("Failed to decode request body", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	app := &models.App{AppSlug: appSlug, Plan: provData.Plan}

	exists, err := database.IsAppExists(app.AppSlug)
	if err != nil {
		logger.Error("Failed to check if App exists in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if !exists {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "App doesn't exists exists"}))
	}

	err = database.UpdateApp(app)
	if err != nil {
		logger.Error("Failed to update app in DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(200, nil)
}

// ProvisionDeleteHandler ...
func ProvisionDeleteHandler(c buffalo.Context) error {
	logger := logging.WithContext(c)
	defer logging.Sync(logger)

	appSlug := c.Param("app_slug")
	err := database.DeleteApp(appSlug)
	if err != nil {
		logger.Error("Failed to delete App from DB", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	analyticsutils.SendAddonEvent(analyticsutils.EventAddonProvisioned, appSlug, "", "")

	return c.Render(200, nil)
}

func generateRandomHash(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
