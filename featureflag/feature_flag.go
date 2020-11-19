package featureflag

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	ld "gopkg.in/launchdarkly/go-server-sdk.v5"
)

var Client *ld.LDClient

func InitClient() error {
	key, ok := os.LookupEnv("LAUNCHDARKLY_SDK_KEY")
	if !ok {
		return errors.New("No value set for env LAUNCHDARKLY_SDK_KEY")
	}

	var err error
	Client, err = ld.MakeClient(key, 5*time.Second)
	if err != nil {
		return err
	}

	return nil
}

func BoolVariationForApp(flagKey string, appSlug string, fallback bool) bool {
	return BoolVariation(flagKey, fmt.Sprintf("app-%s", appSlug), fallback)
}

func BoolVariation(flagKey string, userID string, fallback bool) bool {
	user := lduser.NewUser(userID)

	flagValue, err := Client.BoolVariation(flagKey, user, fallback)
	if err != nil {
		return fallback
	}

	return flagValue
}

// TODO: find out where to add graceful teardown to Buffalo
func Close() error {
	if Client == nil {
		return nil
	}

	return Client.Close()
}
