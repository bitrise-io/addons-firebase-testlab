package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/go-utils/log"
)

func main() {
	fmt.Println("Setup db connection for migrations ...")
	err := database.RunMigrations()
	if err != nil {
		log.Errorf("Migration failed: %s", err)
		os.Exit(1)
	}
	fmt.Println("Migration finished ...")
}
