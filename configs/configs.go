package configs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/gobuffalo/envy"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	iam "google.golang.org/api/iam/v1"
)

var (
	port = ""
	//
	shouldSkipBuildAuthorizationWithBitriseAPI          = false
	shouldSkipSessionAuthentication                     = false
	shouldSkipProvisioningBitriseAPIWebhookRegistration = false
	// GCS
	gcBucket           = ""
	gcKeyJSON          = ""
	gcProjectID        = ""
	gcJWTModel         = &JWTModel{}
	addonConfiguration = AddonConfiguration{}
	env                = ""
	accessToken        = ""
	ssoToken           = ""
	host               = ""
	amplitudeToken     = ""
)

// AddonConfiguration ...
type AddonConfiguration struct {
	ID           string        `json:"id" yaml:"id"`
	Details      *Details      `json:"details" yaml:"details"`
	Subscription *Subscription `json:"subscription" yaml:"subscription"`
	Provisioned  bool          `json:"enabled" yaml:"-"`
}

// Details ...
type Details struct {
	Title       string `json:"title" yaml:"title"`
	Summary     string `json:"summary" yaml:"summary"`
	Description string `json:"description" yaml:"description"`
}

// ServerConfig ...
type ServerConfig struct {
	Host      string `json:"host" yaml:"host"`
	Token     string `json:"token" yaml:"token"`
	SSOSecret string `json:"sso_secret" yaml:"sso_secret"`
}

// Subscription ...
type Subscription struct {
	Unit  string         `json:"unit" yaml:"unit"`
	Plans map[string]int `json:"plans" yaml:"plans"`
}

// JWTModel ...
type JWTModel struct {
	Client *http.Client
	Config *jwt.Config
}

func newAddonConfig() (AddonConfiguration, error) {
	conf := AddonConfiguration{}

	configFilePath := "./addon-config.yml"

	dat, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return AddonConfiguration{}, err
	}

	err = yaml.Unmarshal(dat, &conf)
	if err != nil {
		return AddonConfiguration{}, err
	}

	return conf, nil
}

func setRequiredEnv(envKey string, storeIn *string) error {
	envVar := os.Getenv(envKey)
	if envVar == "" {
		return fmt.Errorf("Environment Variable missing: %s", envKey)
	}
	*storeIn = envVar
	return nil
}

func newJWTModel() (*JWTModel, error) {
	config, err := google.JWTConfigFromJSON([]byte(GetGCSKeyJSON()), iam.CloudPlatformScope, "https://www.googleapis.com/auth/firebase")
	if err != nil {
		return nil, err
	}

	client := config.Client(oauth2.NoContext)

	return &JWTModel{Config: config, Client: client}, nil
}

// Setup ...
func Setup() error {
	env = envy.Get("GO_ENV", "development")

	amplitudeToken = os.Getenv("AMPLITUDE_TOKEN")

	if env == "development" {
		shouldSkipBuildAuthorizationWithBitriseAPI = os.Getenv("SKIP_AUTH_WITH_BITRISE_API") == "yes"
		shouldSkipSessionAuthentication = os.Getenv("SKIP_SESSION_AUTH") == "yes"
		shouldSkipProvisioningBitriseAPIWebhookRegistration = os.Getenv("SKIP_PROVISIONING_BITRISE_API_WEBHOOK_REGISTRATION") == "yes"
	}

	if err := setRequiredEnv("PORT", &port); err != nil {
		return err
	}

	if err := setRequiredEnv("BUCKET", &gcBucket); err != nil {
		return err
	}

	if err := setRequiredEnv("PROJECT_ID", &gcProjectID); err != nil {
		return err
	}
	if err := setRequiredEnv("SERVICE_ACCOUNT_KEY_JSON", &gcKeyJSON); err != nil {
		return err
	}
	if err := setRequiredEnv("ADDON_ACCESS_TOKEN", &accessToken); err != nil {
		return err
	}
	if err := setRequiredEnv("ADDON_SSO_TOKEN", &ssoToken); err != nil {
		return err
	}
	if err := setRequiredEnv("ADDON_HOST", &host); err != nil {
		return err
	}

	JWTModel, err := newJWTModel()
	if err != nil {
		return err
	}
	gcJWTModel = JWTModel

	addonConfiguration, err = newAddonConfig()
	if err != nil {
		return err
	}

	return nil
}

// GetAmplitudeToken ...
func GetAmplitudeToken() string {
	return amplitudeToken
}

// GetENV ...
func GetENV() string {
	return env
}

// GetPort ...
func GetPort() string {
	return port
}

// GetShouldSkipBuildAuthorizationWithBitriseAPI ...
func GetShouldSkipBuildAuthorizationWithBitriseAPI() bool {
	return shouldSkipBuildAuthorizationWithBitriseAPI
}

// GetShouldSkipSessionAuthentication ...
func GetShouldSkipSessionAuthentication() bool {
	return shouldSkipSessionAuthentication
}

// GetShouldSkipProvisioningBitriseAPIWebhookRegistration ...
func GetShouldSkipProvisioningBitriseAPIWebhookRegistration() bool {
	return shouldSkipProvisioningBitriseAPIWebhookRegistration
}

// GetProjectID ...
func GetProjectID() string {
	return gcProjectID
}

// GetGCSBucket ...
func GetGCSBucket() string {
	return gcBucket
}

// GetGCSKeyJSON ...
func GetGCSKeyJSON() string {
	return gcKeyJSON
}

// GetJWTModel ...
func GetJWTModel() *JWTModel {
	return gcJWTModel
}

// GetAddonConfig ...
func GetAddonConfig() AddonConfiguration {
	return addonConfiguration
}

// GetAddonHost ...
func GetAddonHost() string {
	return host
}

// GetAddonAccessToken ...
func GetAddonAccessToken() string {
	return accessToken
}

// GetAddonSSOToken ...
func GetAddonSSOToken() string {
	return ssoToken
}

// GetPlanLimit ...
func GetPlanLimit(name string) int64 {
	for planName, limit := range GetAddonConfig().Subscription.Plans {
		if planName == name {
			return int64(limit) * 60
		}
	}
	return 0
}
