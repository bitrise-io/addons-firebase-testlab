package main

import (
	"log"
	"os"

	"github.com/bitrise-io/addons-firebase-testlab/actions"
	"github.com/bitrise-io/addons-firebase-testlab/worker"
	"github.com/bitrise-io/api-utils/logging"
	"github.com/gobuffalo/envy"
	"github.com/pkg/errors"
)

func main() {
	port := envy.Get("PORT", "3000")
	app := actions.App()
	log.Fatal(app.Start(port))

	workerMode := os.Getenv("WORKER") == "true"
	if workerMode {
		logger := logging.WithContext(nil)
		defer logging.Sync(logger)
		log.Println("Starting worker mode...")
		log.Fatal(errors.WithStack(worker.Start(logger)))
	}
}
