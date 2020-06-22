package analytics

import (
	"errors"
	"os"
	"time"

	"github.com/gobuffalo/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/testing/v1"

	segment "gopkg.in/segmentio/analytics-go.v3"
)

var client segment.Client

const (
	eventAddonProvisioned   = "vdt_android_addon_provisioned"
	eventAddonDeprovisioned = "vdt_android_addon_deprovisioned"
	eventAddonPlanChanged   = "vdt_android_addon_plan_changed"
	eventAddonSSOLogin      = "vdt_android_addon_sso_login"

	eventUploadFileUploadRequested = "vdt_android_addon_file_upload_requested"

	eventTestingTestStarted          = "vdt_android_addon_test_started"
	eventTestingTestFinished         = "vdt_android_addon_test_finished"
	eventTestingTestStartedOnDevice  = "vdt_android_addon_test_started_on_device"
	eventTestingTestFinishedOnDevice = "vdt_android_addon_test_finished_on_device"

	eventIOSTestingTestStarted          = "vdt_ios_addon_test_started"
	eventIOSTestingTestFinished         = "vdt_ios_addon_test_finished"
	eventIOSTestingTestStartedOnDevice  = "vdt_ios_addon_test_started_on_device"
	eventIOSTestingTestFinishedOnDevice = "vdt_ios_addon_test_finished_on_device"
)

// Client ...
type Client struct {
	client segment.Client
	logger *zap.Logger
}

// Initialize ...
func Initialize() error {
	writeKey, ok := os.LookupEnv("SEGMENT_WRITE_KEY")
	if !ok {
		return errors.New("No value set for env SEGMENT_WRITEKEY")
	}
	client = segment.New(writeKey)
	return nil
}

// GetClient ...
func GetClient(logger *zap.Logger) *Client {
	return &Client{
		client: client,
		logger: logger,
	}
}

// TestReportSummaryGenerated ...
func (c *Client) TestReportSummaryGenerated(appSlug, buildSlug, result string, numberOfTests int, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Test report summary generated",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("result", result).
			Set("number_of_tests", numberOfTests).
			Set("datetime", time),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (TestReportSummaryGenerated)", zap.Error(err))
	}
}

// TestReportResult ...
func (c *Client) TestReportResult(appSlug, buildSlug, result, testType string, testResultID uuid.UUID, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Test report result",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("result", result).
			Set("test_type", testType).
			Set("datetime", time).
			Set("test_report_id", testResultID.String()),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (TestReportResult)", zap.Error(err))
	}
}

// NumberOfTestReports ...
func (c *Client) NumberOfTestReports(appSlug, buildSlug string, count int, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Number of test reports",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("count", count).
			Set("datetime", time),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (NumberOfTestReports)", zap.Error(err))
	}
}

// SendTestStartedEvent ...
func (c *Client) SendTestStartedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestStarted, appSlug, buildSlug, testType, eventProperties)
}

// SendTestFinishedEvent ...
func (c *Client) SendTestFinishedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestFinished, appSlug, buildSlug, testType, eventProperties)
}

// SendTestStartedOnDeviceEvent ...
func (c *Client) SendTestStartedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendTestingEventDevices(eventTestingTestStartedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendTestFinishedOnDeviceEvent ...
func (c *Client) SendTestFinishedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendTestingEventDevices(eventTestingTestFinishedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendIOSTestStartedOnDeviceEvent ...
func (c *Client) SendIOSTestStartedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	c.sendIOSTestingEventDevices(eventIOSTestingTestStartedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendIOSTestFinishedOnDeviceEvent ...
func (c *Client) SendIOSTestFinishedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	c.sendIOSTestingEventDevices(eventIOSTestingTestFinishedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendAddonProvisionedEvent ...
func (c *Client) SendAddonProvisionedEvent(appSlug, currentPlan, newPlan string) {
	c.sendAddonEvent(eventAddonProvisioned, appSlug, currentPlan, newPlan)
}

// SendAddonDeprovisionedEvent ...
func (c *Client) SendAddonDeprovisionedEvent(appSlug, currentPlan, newPlan string) {
	c.sendAddonEvent(eventAddonDeprovisioned, appSlug, currentPlan, newPlan)
}

// SendAddonPlanChangedEvent ...
func (c *Client) SendAddonPlanChangedEvent(appSlug, currentPlan, newPlan string) {
	c.sendAddonEvent(eventAddonPlanChanged, appSlug, currentPlan, newPlan)
}

// SendAddonSSOLoginEvent ...
func (c *Client) SendAddonSSOLoginEvent(appSlug, currentPlan, newPlan string) {
	c.sendAddonEvent(eventAddonSSOLogin, appSlug, currentPlan, newPlan)
}

// SendUploadRequestedEvent ...
func (c *Client) SendUploadRequestedEvent(event, appSlug, buildSlug string) {
	if c.client == nil {
		return
	}
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  eventUploadFileUploadRequested,
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug),
		Timestamp: time.Now(),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (SendUploadRequestedEvent)", zap.Error(err))
	}
}

func (c *Client) sendTestingEvent(event, appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	if c.client == nil {
		return
	}

	trackProps := segment.NewProperties().
		Set("app_slug", appSlug).
		Set("build_slug", buildSlug)

	if testType != "" {
		trackProps = trackProps.Set("test_type", testType)
	}
	if eventProperties != nil {
		for key, value := range eventProperties {
			trackProps = trackProps.Set(key, value)
		}
	}
	err := c.client.Enqueue(segment.Track{
		UserId:     appSlug,
		Event:      event,
		Properties: trackProps,
		Timestamp:  time.Now(),
	})

	if err != nil {
		c.logger.Warn("Failed to track analytics (sendTestingEvent)", zap.String("event", event), zap.Error(err))
	}
}

func (c *Client) sendTestingEventDevices(event, appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	if c.client == nil {
		return
	}

	for _, device := range devices {
		trackProps := segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug)

		if testType != "" {
			trackProps = trackProps.Set("test_type", testType)
		}
		if eventProperties != nil {
			for key, value := range eventProperties {
				trackProps = trackProps.Set(key, value)
			}
		}
		trackProps = trackProps.Set("device_id", device.AndroidModelId)
		trackProps = trackProps.Set("device_os_version", device.AndroidVersionId)
		trackProps = trackProps.Set("device_language", device.Locale)
		trackProps = trackProps.Set("device_orientation", device.Orientation)

		err := c.client.Enqueue(segment.Track{
			UserId:     appSlug,
			Event:      event,
			Properties: trackProps,
			Timestamp:  time.Now(),
		})

		if err != nil {
			c.logger.Warn("Failed to track analytics (sendTestingEventDevices)", zap.String("event", event), zap.Error(err))
		}
	}
}

func (c *Client) sendIOSTestingEventDevices(event, appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	if c.client == nil {
		return
	}

	for _, device := range devices {
		trackProps := segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug)

		if testType != "" {
			trackProps = trackProps.Set("test_type", testType)
		}
		if eventProperties != nil {
			for key, value := range eventProperties {
				trackProps = trackProps.Set(key, value)
			}
		}
		trackProps = trackProps.Set("device_id", device.IosModelId)
		trackProps = trackProps.Set("device_os_version", device.IosVersionId)
		trackProps = trackProps.Set("device_language", device.Locale)
		trackProps = trackProps.Set("device_orientation", device.Orientation)

		err := c.client.Enqueue(segment.Track{
			UserId:     appSlug,
			Event:      event,
			Properties: trackProps,
			Timestamp:  time.Now(),
		})

		if err != nil {
			c.logger.Warn("Failed to track analytics (sendIOSTestingEventDevices)", zap.String("event", event), zap.Error(err))
		}
	}
}

func (c *Client) sendAddonEvent(event, appSlug, currentPlan, newPlan string) {
	if c.client == nil {
		return
	}

	trackProps := segment.NewProperties().
		Set("app_slug", appSlug)
	if currentPlan != "" {
		trackProps = trackProps.Set("old_plan", currentPlan)
	}
	if newPlan != "" {
		trackProps = trackProps.Set("plan", newPlan)
	}
	err := c.client.Enqueue(segment.Track{
		UserId:     appSlug,
		Event:      event,
		Properties: trackProps,
		Timestamp:  time.Now(),
	})

	if err != nil {
		c.logger.Warn("Failed to track analytics (sendTestingEvent)", zap.String("event", event), zap.Error(err))
	}
}

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}
