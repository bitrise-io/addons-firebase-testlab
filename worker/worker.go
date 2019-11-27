package worker

import (
	"os"
	"os/signal"

	"go.uber.org/zap"

	redispkg "github.com/bitrise-io/addons-firebase-testlab/redis"
	"github.com/bitrise-io/api-utils/utils"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

var namespace = "test_addon_workers"
var redisPool *redis.Pool

// Context ...
type Context struct {
	logger *zap.Logger
}

func init() {
	if redisPool == nil {
		redisPool = redispkg.NewPool(
			os.Getenv("REDIS_URL"),
			int(utils.GetInt64EnvWithDefault("WORKER_MAX_IDLE_CONNECTION", 50)),
			int(utils.GetInt64EnvWithDefault("WORKER_MAX_ACTIVE_CONNECTION", 1000)),
		)
	}
}

// Start ...
func Start(logger *zap.Logger) error {
	context := Context{logger: logger}
	pool := work.NewWorkerPool(context, 10, namespace, redisPool)

	pool.Job(createStepResult, (&context).CreateStepResult)

	pool.Start()
	defer pool.Stop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	return nil
}
