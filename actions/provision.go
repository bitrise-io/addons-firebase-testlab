package actions

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/analyticsutils"

	"github.com/bitrise-io/go-utils/log"
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
	provData := &ProvisionData{}

	err := json.NewDecoder(c.Request().Body).Decode(provData)
	if err != nil {
		log.Errorf("Failed to decode request body, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Failed to decode provisioning data"}))
	}

	exists, err := database.IsAppExists(provData.AppSlug)
	if err != nil {
		log.Errorf("Failed to check if App exists in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if exists {
		log.Warnf("  [!] App already exists")
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
			log.Errorf("Failed to add app to DB, error: %+v", errors.WithStack(err))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		analyticsutils.SendAddonEvent(analyticsutils.EventAddonProvisioned, app.AppSlug, "", app.Plan)

		envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
		return c.Render(200, r.JSON(envs))
	}

	app, err = database.GetApp(app)
	if err != nil {
		log.Errorf("Failed to get app from DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
	return c.Render(200, r.JSON(envs))
}

// ProvisionPutHandler ...
func ProvisionPutHandler(c buffalo.Context) error {
	appSlug := c.Param("app_slug")

	provData := &ProvisionData{}

	err := json.NewDecoder(c.Request().Body).Decode(provData)
	if err != nil {
		log.Errorf("Failed to decode request body, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	app := &models.App{AppSlug: appSlug, Plan: provData.Plan}

	exists, err := database.IsAppExists(app.AppSlug)
	if err != nil {
		log.Errorf("Failed to check if App exists in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}
	if !exists {
		log.Errorf("App doesn't exists")
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "App doesn't exists exists"}))
	}

	err = database.UpdateApp(app)
	if err != nil {
		log.Errorf("Failed to update app in DB, error: %+v", errors.WithStack(err))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	return c.Render(200, nil)
}

// ProvisionDeleteHandler ...
func ProvisionDeleteHandler(c buffalo.Context) error {
	appSlug := c.Param("app_slug")

	err := database.DeleteApp(appSlug)
	if err != nil {
		log.Errorf("Failed to delete App from DB, error: %+v", errors.WithStack(err))
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
