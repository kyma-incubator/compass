package operations

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
	"time"
)

type Stage interface {
	Name() model.OperationStage
	Run(cluster model.Cluster, logger logrus.FieldLogger) (StageResult, error)
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
