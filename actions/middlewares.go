package actions

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
)

func validateUserLoginStatus(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		log.Printf("ValidateUserLoginStatus")

		if configs.GetShouldSkipSessionAuthentication() {
			c.Session().Set("app_slug", os.Getenv("BITRISE_APP_SLUG"))
		}

		sessionAppSlug, ok := c.Session().Get("app_slug").(string)

		if ok {
			fmt.Printf("stored appSlug: %s", sessionAppSlug)

			exists, err := database.IsAppExists(sessionAppSlug)
			if err != nil {
				return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
			}
			if exists {
				fmt.Printf("app exists, next...")
				return next(c)
			}
		}
		fmt.Printf("app not exists: %s", sessionAppSlug)
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
		exists, err := database.IsAppExistsWithToken(c.Param("app_slug"), c.Param("token"))
		if err != nil {
			log.Errorf("Failed to check if token valid under app, error: %s", err)
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
		buildExists, err := database.IsBuildExists(c.Param("app_slug"), c.Param("build_slug"))
		if err != nil {
			log.Errorf(" [!] Exception: Failed to check if build exists: %+v", err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		if !buildExists {
			log.Errorf("Build doesn't exist")
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}

		return next(c)
	}
}

func authorizeForTestReport(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		testReportExists, err := database.IsTestReportExistsForBuild(c.Param("build_slug"), c.Param("test_report_id"))
		if err != nil {
			log.Errorf(" [!] Exception: Failed to check if test report exists: %+v", err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		if !testReportExists {
			log.Errorf("Test report doesn't exist")
			return c.Render(http.StatusForbidden, r.JSON(map[string]string{"error": "Unauthorized request"}))
		}

		return next(c)
	}
}

func authorizeForRunningBuildViaBitriseAPI(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if configs.GetShouldSkipBuildAuthorizationWithBitriseAPI() {
			return next(c)
		}

		app := &models.App{AppSlug: c.Param("app_slug")}
		app, err := database.GetApp(app)
		if err != nil {
			log.Errorf("Failed to get app from DB, error: %+v", errors.WithStack(err))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}

		client := bitrise.NewClient(app.BitriseAPIToken)
		resp, build, err := client.GetBuildOfApp(c.Param("build_slug"), c.Param("app_slug"))
		if err != nil {
			log.Errorf("Failed to get build from Bitrise API, error: %+v", err)
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
		// read and serve svg contents as html
		svgAssetsBaseDir := "./frontend/assets/compiled/images"
		files, err := ioutil.ReadDir(svgAssetsBaseDir)
		if err != nil {
			return fmt.Errorf("Failed to read dir, error: %+v", err)
		}
		svgContents := map[string]template.HTML{}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".svg") {
				content, err := fileutil.ReadStringFromFile(filepath.Join(svgAssetsBaseDir, file.Name()))
				if err != nil {
					return fmt.Errorf("Failed to get svg data, error: %s", err)
				}
				svgContents[file.Name()] = template.HTML(content)
			}
		}
		c.Set("Svg", svgContents)
		return next(c)
	}
}
