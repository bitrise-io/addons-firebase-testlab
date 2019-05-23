package models

import (
	"fmt"

	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/gobuffalo/pop"
)

// DB ...
var DB *pop.Connection

func init() {
	var err error
	DB, err = pop.Connect(configs.GetENV())
	if err != nil {
		fmt.Printf("Failed to init DB, error: %+v", err)
	}
	pop.Debug = configs.GetENV() == "development"
}
