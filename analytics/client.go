package analytics

import (
	"errors"
	"os"
	"time"

	"github.com/gobuffalo/uuid"
	"go.uber.org/zap"

	segment "gopkg.in/segmentio/analytics-go.v3"
)

var client segment.Client

const (
	eventUploadFileUploadRequested = "vdt_android_addon_file_upload_requested"
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

// SendUploadRequestedEvent ...
func (c *Client) SendUploadRequestedEvent(appSlug, buildSlug string) {
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

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}
