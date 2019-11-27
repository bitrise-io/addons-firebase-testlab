package bitrise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/pkg/errors"

	junitparser "github.com/joshdk/go-junit"
)

const (
	baseURLenvKey  = "BITRISE_API_URL"
	defaultBaseURL = "https://api.bitrise.io"
	version        = "v0.1"
)

// Client manages communication with the Bitrise API.
type Client struct {
	client   *http.Client
	BaseURL  string
	apiToken string
}

// StepResult ...
type StepResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// TestStepResult ...
type TestStepResult struct {
	StepResult
	Total       int                `json:"total"`
	FailedTests []junitparser.Test `json:"failed_test"`
}

// NewClient returns a new instance of *Client.
func NewClient(apiToken string) *Client {
	return &Client{
		client:   &http.Client{Timeout: 10 * time.Second},
		apiToken: apiToken,
		BaseURL:  fmt.Sprintf("%s/%s", getEnv(baseURLenvKey, defaultBaseURL), version),
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func closeResponseBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Errorf("Failed to close response body, error: %+v", errors.WithStack(err))
	}
}

// newRequest creates an authenticated API request that is ready to send.
func (c *Client) newRequest(method string, action string, payload []byte) (*http.Request, error) {
	method = strings.ToUpper(method)
	endpoint := fmt.Sprintf("%s/%s", c.BaseURL, action)

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Bitrise-Addon-Auth-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, bp *Build) (*http.Response, error) {
	req.Close = true
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return resp, nil
	}

	var successResp struct {
		Data Build
	}

	if err = json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return resp, errors.WithStack(err)
	}

	*bp = successResp.Data
	return resp, nil
}

// Build represents a build
type Build struct {
	Status int    `json:"status"`
	Slug   string `json:"slug"`
}

// GetBuildOfApp returns information about a single build.
func (c *Client) GetBuildOfApp(buildSlug string, appSlug string) (*http.Response, *Build, error) {
	action := fmt.Sprintf("apps/%s/builds/%s", appSlug, buildSlug)
	req, err := c.newRequest("GET", action, nil)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	var build Build
	resp, err := c.do(req, &build)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		return resp, nil, errors.WithStack(err)
	}

	return resp, &build, nil
}

// GetLatestBuildOfApp returns information about the latest build of app.
func (c *Client) GetLatestBuildOfApp(appSlug string) (*Build, error) {
	action := fmt.Sprintf("apps/%s/builds?limit=1", appSlug)
	req, err := c.newRequest("GET", action, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Close = true
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to fetch latest build info: status: %d", resp.StatusCode)
	}

	var successResp struct {
		Data []Build
	}
	defer closeResponseBody(resp)

	if err = json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(successResp.Data) == 0 {
		return nil, errors.New("No builds found for app")
	}

	return &successResp.Data[0], nil
}

// RegisterWebhook ...
func (c *Client) RegisterWebhook(app *models.App) (*http.Response, error) {
	appSecret, err := app.Secret()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	testingAddonHost, ok := os.LookupEnv("ADDON_HOST")
	if !ok {
		return nil, errors.New("No ADDON_HOST env var is set")
	}
	payloadStruct := map[string]interface{}{
		"url":    fmt.Sprintf("%s/webhook", testingAddonHost),
		"events": []string{"build"},
		"secret": appSecret,
	}

	payload, err := json.Marshal(payloadStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/apps/%s/outgoing-webhooks", c.BaseURL, app.AppSlug), bytes.NewBuffer(payload))
	req.Header.Set("Bitrise-Addon-Auth-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			fmt.Println(errors.WithStack(err))
		}
	}()

	if response.StatusCode != http.StatusCreated {
		return nil, errors.New("Internal error: Failed to register webhook")
	}

	return response, nil
}

// CreateTestStepResult creates a new test step result
func (c *Client) CreateTestStepResult(appSlug string, buildSlug string, tsr *TestStepResult) error {
	payload, err := json.Marshal(tsr)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/apps/%s/builds/%s/test_step_results", c.BaseURL, appSlug, buildSlug), bytes.NewBuffer(payload))
	req.Header.Set("Bitrise-Addon-Auth-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return errors.WithStack(err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			fmt.Println(errors.WithStack(err))
		}
	}()

	if response.StatusCode != http.StatusCreated {
		return errors.New("Internal error: Failed to create test step result")
	}

	return nil
}
