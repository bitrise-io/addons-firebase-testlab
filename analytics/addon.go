package analytics

import (
	"time"

	"go.uber.org/zap"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

const (
	eventAddonProvisioned   = "vdt_android_addon_provisioned"
	eventAddonDeprovisioned = "vdt_android_addon_deprovisioned"
	eventAddonPlanChanged   = "vdt_android_addon_plan_changed"
	eventAddonSSOLogin      = "vdt_android_addon_sso_login"
)

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
