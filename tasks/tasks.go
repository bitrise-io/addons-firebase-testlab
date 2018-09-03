package tasks

import (
	"fmt"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/firebaseutils"
	"github.com/bitrise-io/addons-firebase-testlab/models"
)

// CleanBuilds ...
func CleanBuilds() error {
	fmt.Println("Closing builds....")

	builds, err := database.GetAllExpiredOpenBuilds()
	if err != nil {
		return fmt.Errorf("Failed to get builds to clean-up, error: %+v", err)
	}
	fmt.Println("Builds to clean-up:")

	for _, build := range *builds {
		fmt.Printf("AppSlug: %s; BuildSlug: %s; BuildSessionEnabled: %t; LastRequest: %s\n", build.AppSlug, build.BuildSlug, build.BuildSessionEnabled, build.LastRequest.Time.String())

		build.BuildSessionEnabled = false
		err := database.UpdateBuild(&build)
		if err != nil {
			return fmt.Errorf("Failed to update build, error: %+v", err)
		}

		if build.TestMatrixID != "" {
			response, err := firebaseutils.CancelTestMatrix(build.TestMatrixID)
			if err != nil {
				return fmt.Errorf("Failed to cancel test matrix(id: %s), error: %+v", build.TestMatrixID, err)
			}
			fmt.Printf("Cancel request sent, response: %s", response)
		}
	}
	return nil
}

// AddDefaultUser ...
func AddDefaultUser(appSlug string) error {
	fmt.Println("Adding default user")
	return database.AddApp(&models.App{Plan: "free", AppSlug: appSlug, APIToken: "test-api-token"})
}

// GetBuildsCount ...
func GetBuildsCount() error {
	fmt.Println("Getting builds count...")

	buildsCount, err := database.GetBuildsCount()
	if err != nil {
		return fmt.Errorf("Failed to get builds count, error: %+v", err)
	}

	fmt.Printf("Builds: %d", buildsCount)

	return nil
}
