package worker

import (
	"github.com/gobuffalo/uuid"
	"github.com/gocraft/work"
	"github.com/pkg/errors"
)

// Service ...
type Service struct{}

// EnqueueCreateStepResult ...
func (*Service) EnqueueCreateStepResult(testReportID uuid.UUID, secondsFromNow int64) error {
	enqueuer := work.NewEnqueuer(namespace, redisPool)
	var err error
	jobParams := work.Q{"test_report_id": testReportID}
	if secondsFromNow == 0 {
		_, err = enqueuer.EnqueueUnique(createStepResult, jobParams)
	} else {
		_, err = enqueuer.EnqueueUniqueIn(createStepResult, secondsFromNow, jobParams)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
