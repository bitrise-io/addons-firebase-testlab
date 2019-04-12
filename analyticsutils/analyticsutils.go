package analyticsutils

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	testing "google.golang.org/api/testing/v1"

	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/logging"
	"github.com/pkg/errors"
	amplitude "github.com/savaki/amplitude-go"
)

// cons ...
const (
	//
	// Addon Events
	EventAddonProvisioned   = "vdt_android_addon_provisioned"
	EventAddonDeprovisioned = "vdt_android_addon_deprovisioned"
	EventAddonPlanChanged   = "vdt_android_addon_plan_changed"
	EventAddonSSOLogin      = "vdt_android_addon_sso_login"

	//
	// Upload Events
	EventUploadFileUploadRequested = "vdt_android_addon_file_upload_requested"

	//
	// Testing Events
	EventTestingTestStarted          = "vdt_android_addon_test_started"
	EventTestingTestFinished         = "vdt_android_addon_test_finished"
	EventTestingTestStartedOnDevice  = "vdt_android_addon_test_started_on_device"
	EventTestingTestFinishedOnDevice = "vdt_android_addon_test_finished_on_device"

	EventIOSTestingTestStarted          = "vdt_ios_addon_test_started"
	EventIOSTestingTestFinished         = "vdt_ios_addon_test_finished"
	EventIOSTestingTestStartedOnDevice  = "vdt_ios_addon_test_started_on_device"
	EventIOSTestingTestFinishedOnDevice = "vdt_ios_addon_test_finished_on_device"
)

// Client ...
var Client *amplitude.Client

// Init ...
func Init() error {
	if configs.GetAmplitudeToken() == "" {
		Client = nil
		return fmt.Errorf("AMPLITUDE_TOKEN env is not set, not an issue but analytics won't work")
	}
	Client = amplitude.New(configs.GetAmplitudeToken())
	return nil
}

// SendTestingEvent ...
func SendTestingEvent(event, appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if Client == nil {
		return
	}

	go func(client *amplitude.Client, event, appSlug, buildSlug, testType string, eventProperties map[string]interface{}) {
		evt := amplitude.Event{UserId: appSlug, EventType: event, Time: time.Now(), EventProperties: map[string]interface{}{"build_slug": buildSlug}}

		if testType != "" {
			evt.EventProperties["test_type"] = testType
		}

		if eventProperties != nil {
			for key, val := range eventProperties {
				evt.EventProperties[key] = val
			}
		}

		if err := client.Publish(evt); err != nil {
			logger.Error("[!] Exception: failed to send analytics event", zap.Any("error", errors.WithStack(err)))
		}
	}(Client, event, appSlug, buildSlug, testType, eventProperties)
}

// SendTestingEventDevices ...
func SendTestingEventDevices(event, appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if Client == nil {
		return
	}

	go func(client *amplitude.Client, event, appSlug, buildSlug, testType string, devices []*testing.AndroidDevice, eventProperties map[string]interface{}) {
		for _, device := range devices {
			evt := amplitude.Event{UserId: appSlug, EventType: event, Time: time.Now(), EventProperties: map[string]interface{}{"build_slug": buildSlug}}

			if testType != "" {
				evt.EventProperties["test_type"] = testType
			}

			if eventProperties != nil {
				for key, val := range eventProperties {
					evt.EventProperties[key] = val
				}
			}

			evt.EventProperties["device_id"] = device.AndroidModelId
			evt.EventProperties["device_os_version"] = device.AndroidVersionId
			evt.EventProperties["device_language"] = device.Locale
			evt.EventProperties["device_orientation"] = device.Orientation

			if err := client.Publish(evt); err != nil {
				logger.Error("[!] Exception: failed to send analytics event", zap.Any("error", errors.WithStack(err)))
			}
		}
	}(Client, event, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendIOSTestingEventDevices ...
func SendIOSTestingEventDevices(event, appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if Client == nil {
		return
	}

	go func(client *amplitude.Client, event, appSlug, buildSlug, testType string, devices []*testing.IosDevice, eventProperties map[string]interface{}) {
		for _, device := range devices {
			evt := amplitude.Event{UserId: appSlug, EventType: event, Time: time.Now(), EventProperties: map[string]interface{}{"build_slug": buildSlug}}

			if testType != "" {
				evt.EventProperties["test_type"] = testType
			}

			if eventProperties != nil {
				for key, val := range eventProperties {
					evt.EventProperties[key] = val
				}
			}

			evt.EventProperties["device_id"] = device.IosModelId
			evt.EventProperties["device_os_version"] = device.IosVersionId
			evt.EventProperties["device_language"] = device.Locale
			evt.EventProperties["device_orientation"] = device.Orientation

			if err := client.Publish(evt); err != nil {
				logger.Error("[!] Exception: failed to send analytics event", zap.Any("error", errors.WithStack(err)))
			}
		}
	}(Client, event, appSlug, buildSlug, testType, devices, eventProperties)
}

// SendUploadEvent ...
func SendUploadEvent(event, appSlug, buildSlug string) {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if Client == nil {
		return
	}

	go func(client *amplitude.Client, event, appSlug, buildSlug string) {
		evt := amplitude.Event{UserId: appSlug, EventType: event, Time: time.Now(), EventProperties: map[string]interface{}{"build_slug": buildSlug}}

		if err := Client.Publish(evt); err != nil {
			logger.Error("[!] Exception: failed to send analytics event", zap.Any("error", errors.WithStack(err)))
		}
	}(Client, event, appSlug, buildSlug)
}

// SendAddonEvent ...
func SendAddonEvent(event, appSlug, currentPlan, newPlan string) {
	logger := logging.WithContext(nil)
	defer logging.Sync(logger)

	if Client == nil {
		return
	}
	go func(client *amplitude.Client, event, appSlug, currentPlan, newPlan string) {

		evt := amplitude.Event{UserId: appSlug, EventType: event, Time: time.Now()}

		var params map[string]interface{}

		if currentPlan != "" {
			if params == nil {
				params = map[string]interface{}{}
			}
			params["old_plan"] = currentPlan
		}

		if newPlan != "" {
			if params == nil {
				params = map[string]interface{}{}
			}
			params["plan"] = newPlan
		}

		if params != nil {
			evt.EventProperties = params
		}

		if err := Client.Publish(evt); err != nil {
			logger.Error("[!] Exception: failed to send analytics event", zap.Any("error", errors.WithStack(err)))
		}
	}(Client, event, appSlug, currentPlan, newPlan)
}
