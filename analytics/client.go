package analytics

import (
	"errors"
	"os"
	"time"

	"go.uber.org/zap"

	segment "gopkg.in/segmentio/analytics-go.v3"
)

// Client ...
type Client struct {
	client segment.Client
	logger *zap.Logger
}

// NewClient ...
func NewClient(logger *zap.Logger) (Client, error) {
	writeKey, ok := os.LookupEnv("SEGMENT_WRITE_KEY")
	if !ok {
		return Client{}, errors.New("No value set for env SEGMENT_WRITEKEY")
	}

	return Client{
		client: segment.New(writeKey),
		logger: logger,
	}, nil
}

// TestReportSummaryGenerated ...
func (c *Client) TestReportSummaryGenerated(appSlug, buildSlug, result string, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Test report summary generated",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("result", result).
			Set("datetime", time),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (TestReportSummaryGenerated)", zap.Error(err))
	}
}

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}
