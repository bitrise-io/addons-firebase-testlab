package analytics

import (
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/testing/v1"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

const (
	eventTestingTestStartedOnDevice  = "vdt_android_addon_test_started_on_device"
	eventTestingTestFinishedOnDevice = "vdt_android_addon_test_finished_on_device"
)

// SendTestStartedOnDeviceEvent ...
func (c *Client) SendTestStartedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendTestingEventDevices(eventTestingTestStartedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendTestFinishedOnDeviceEvent ...
func (c *Client) SendTestFinishedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendTestingEventDevices(eventTestingTestFinishedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
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
