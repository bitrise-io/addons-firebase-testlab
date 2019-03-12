package actions

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func addLogger(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		ctx := logging.NewContext(c, zap.String("request_id", uuid.NewV4().String()))
		return next(ctx)
	}
}

func validateUserLoginStatus(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if configs.GetShouldSkipSessionAuthentication() {
			c.Session().Set("app_slug", os.Getenv("BITRISE_APP_SLUG"))
		}

		sessionAppSlug, ok := c.Session().Get("app_slug").(string)

		if ok {
			exists, err := database.IsAppExists(sessionAppSlug)
			if err != nil {
				return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
			}
			if exists {
				return next(c)
			}
		}
		return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized"}))
	}
}

func authenticateWithAccessToken(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if c.Request().Header.Get("Authentication") != configs.GetAddonAccessToken() {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}
		return next(c)
	}
}

func authenticateRequestWithToken(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		logger := logging.WithContext(c)
		defer logging.Sync(logger)

		exists, err := database.IsAppExistsWithToken(c.Param("app_slug"), c.Param("token"))
		if err != nil {
			logger.Error("Failed to check if token valid under app", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Invalid request"}))
		}
		if !exists {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}
		return next(c)
	}
}

func authorizeForBuild(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		logger := logging.WithContext(c)
		defer logging.Sync(logger)

		buildExists, err := database.IsBuildExists(c.Param("app_slug"), c.Param("build_slug"))
		if err != nil {
			logger.Error(" [!] Exception: Failed to check if build exists", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		if !buildExists {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}

		return next(c)
	}
}

func authorizeForTestReport(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		logger := logging.WithContext(c)
		defer logging.Sync(logger)

		testReportExists, err := database.IsTestReportExistsForBuild(c.Param("build_slug"), c.Param("test_report_id"))
		if err != nil {
			logger.Error(" [!] Exception: Failed to check if test report exists", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		if !testReportExists {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}

		return next(c)
	}
}

func authorizeForRunningBuildViaBitriseAPI(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		logger := logging.WithContext(c)
		defer logging.Sync(logger)
		if configs.GetShouldSkipBuildAuthorizationWithBitriseAPI() {
			return next(c)
		}

		app := &models.App{AppSlug: c.Param("app_slug")}
		app, err := database.GetApp(app)
		if err != nil {
			logger.Error("Failed to get app from DB", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		client := bitrise.NewClient(app.BitriseAPIToken)
		resp, build, err := client.GetBuildOfApp(c.Param("build_slug"), c.Param("app_slug"))
		if err != nil {
			logger.Error("Failed to get build from Bitrise API", zap.Any("error", errors.WithStack(err)))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		if resp.StatusCode != http.StatusOK {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}

		if build.Status != 0 {
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Build has already finished"}))
		}

		return next(c)
	}
}

func serveSVGs(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		logger := logging.WithContext(c)
		defer logging.Sync(logger)

		// read and serve svg contents as html
		svgAssetsBaseDir := "./frontend/assets/compiled/images"
		files, err := ioutil.ReadDir(svgAssetsBaseDir)
		if err != nil {
			return errors.Wrap(err, "Failed to read dir")
		}
		svgContents := map[string]template.HTML{}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".svg") {
				content, err := fileutil.ReadStringFromFile(filepath.Join(svgAssetsBaseDir, file.Name()))
				if err != nil {
					return errors.Wrap(err, "Failed to get svg data")
				}
				svgContents[file.Name()] = template.HTML(content)
			}
		}
		c.Set("Svg", svgContents)
		return next(c)
	}
}
