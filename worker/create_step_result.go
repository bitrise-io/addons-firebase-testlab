package worker

import (
	"github.com/bitrise-io/addons-firebase-testlab/stepresult"
	"github.com/gobuffalo/uuid"
	"github.com/gocraft/work"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var createStepResult = "create_step_result"

// CreateStepResult ...
func (c *Context) CreateStepResult(job *work.Job) error {
	c.logger.Info("[i] Job CreateStepResult started")
	trID := job.ArgString("test_report_id")
	if trID == "" {
		c.logger.Error("Failed to get test_report_id")
		return errors.New("Failed to get test_report_id")
	}
	testReportID, err := uuid.FromString(trID)
	if err != nil {
		c.logger.Error("Failed to parse test_report_id to UUID", zap.Error(err))
		return errors.New("Failed to parse test_report_id to UUID")
	}
	err = stepresult.CreateTestStepResult(testReportID)
	if err != nil {
		c.logger.Error("Failed to create test step result", zap.Error(err))
		return errors.New("Failed to create test step result")
	}

	c.logger.Info("[i] Job CreateStepResult finished")
	return nil
}
