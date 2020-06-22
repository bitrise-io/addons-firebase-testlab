package analytics

import (
	"time"

	"go.uber.org/zap"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

const (
	eventTestingTestStarted  = "vdt_android_addon_test_started"
	eventTestingTestFinished = "vdt_android_addon_test_finished"
)

// SendTestStartedEvent ...
func (c *Client) SendTestStartedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestStarted, appSlug, buildSlug, testType, eventProperties)
}

// SendTestFinishedEvent ...
func (c *Client) SendTestFinishedEvent(appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	c.sendTestingEvent(eventTestingTestFinished, appSlug, buildSlug, testType, eventProperties)
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
