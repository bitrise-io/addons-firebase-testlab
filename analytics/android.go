package analytics

import (
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/testing/v1"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

const (
	eventTestingTestStarted          = "vdt_android_addon_test_started"
	eventTestingTestFinished         = "vdt_android_addon_test_finished"
	eventTestingTestStartedOnDevice  = "vdt_android_addon_test_started_on_device"
	eventTestingTestFinishedOnDevice = "vdt_android_addon_test_finished_on_device"
)

// SendAndroidTestStartedOnDeviceEvent ...
func (c *Client) SendAndroidTestStartedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendAndroidTestingEventDevices(eventTestingTestStartedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendAndroidTestFinishedOnDeviceEvent ...
func (c *Client) SendAndroidTestFinishedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	c.sendAndroidTestingEventDevices(eventTestingTestFinishedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendAndroidTestStartedEvent ...
func (c *Client) SendAndroidTestStartedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestStartedOnDevice, appSlug, buildSlug, testType, eventProperties)
}

// SendAndroidTestFinishedEvent ...
func (c *Client) SendAndroidTestFinishedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestFinishedOnDevice, appSlug, buildSlug, testType, eventProperties)
}

func (c *Client) sendAndroidTestingEventDevices(event, appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
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
