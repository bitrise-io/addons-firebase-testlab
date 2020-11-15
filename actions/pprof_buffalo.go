package actions

import (
	"net/http/pprof"

	"github.com/gobuffalo/buffalo"
)

// RegisterProfEndpoints Registers pprof endpoints
func RegisterProfEndpoints(app *buffalo.App) {
	profiles := []string{
		"allocs",
		"block",
		"heap",
		"mutex",
		"threadcreate",
		"trace",
		"goroutine",
	}

	for _, profile := range profiles {
		app.GET("/debug_pprof_prof/"+profile, profEndpoint(profile))
	}
}

func profEndpoint(name string) buffalo.Handler {
	return func(c buffalo.Context) error {
		pprof.Handler(name).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
