package operations

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

type ProcessingResult struct {
	Requeue bool
	Delay   time.Duration
}

type Stage interface {
	Name() model.OperationStage
	Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (StageResult, error)
	TimeLimit() time.Duration
}

type StageResult struct {
	Stage model.OperationStage
	Delay time.Duration
}

type NonRecoverableError struct {
	error error
}

func (r NonRecoverableError) Error() string {
	return r.error.Error()
}

func NewNonRecoverableError(err error) NonRecoverableError {
	return NonRecoverableError{error: err}
}

type FailureHandler interface {
	HandleFailure(operation model.Operation, cluster model.Cluster) error
}
