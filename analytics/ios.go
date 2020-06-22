package analytics

import (
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/testing/v1"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

const (
	eventIOSTestingTestStarted          = "vdt_ios_addon_test_started"
	eventIOSTestingTestFinished         = "vdt_ios_addon_test_finished"
	eventIOSTestingTestStartedOnDevice  = "vdt_ios_addon_test_started_on_device"
	eventIOSTestingTestFinishedOnDevice = "vdt_ios_addon_test_finished_on_device"
)

// SendIOSTestStartedOnDeviceEvent ...
func (c *Client) SendIOSTestStartedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	c.sendIOSTestingEventDevices(eventIOSTestingTestStartedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendIOSTestFinishedOnDeviceEvent ...
func (c *Client) SendIOSTestFinishedOnDeviceEvent(appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	c.sendIOSTestingEventDevices(eventIOSTestingTestFinishedOnDevice, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendIOSTestStartedEvent ...
func (c *Client) SendIOSTestStartedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventIOSTestingTestStartedOnDevice, appSlug, buildSlug, testType, eventProperties)
}

// SendIOSTestFinishedEvent ...
func (c *Client) SendIOSTestFinishedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventIOSTestingTestFinishedOnDevice, appSlug, buildSlug, testType, eventProperties)
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
