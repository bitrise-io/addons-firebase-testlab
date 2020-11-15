package main

import (
	"log"

	"github.com/bitrise-io/addons-firebase-testlab/actions"
	"github.com/gobuffalo/envy"

	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := envy.Get("PORT", "3000")
	app := actions.App()
	log.Fatal(app.Start(port))
}
