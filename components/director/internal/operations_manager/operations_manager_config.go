package operationsmanager

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
)

// OperationsManagerConfig destination service configuration
type OperationsManagerConfig struct {
	// RescheduleOpsElectionConfig is a leader election configuration for reschedule operations job
	RescheduleOpsElectionConfig cronjob.ElectionConfig
	// RescheduleHangedOpsElectionConfig is a leader election configuration for reschedule hanged operations job
	RescheduleHangedOpsElectionConfig cronjob.ElectionConfig
	// ElectionConfig ia s congiguratin for leader election
	ElectionConfig cronjob.ElectionConfig
	// PriorityQueueLimit is the number of operations returned from priority queue. Should be larger (+1) than the number of pods.
	PriorityQueueLimit int `envconfig:"APP_OPERATIONS_MANAGER_PRIORITY_QUEUE_LIMIT,default=10"`
	// RescheduleOperationsJobInterval how frequently the reschedule job for (refresh data) will be executed
	RescheduleOperationsJobInterval time.Duration `envconfig:"APP_OPERATIONS_MANAGER_RESCHEDULE_JOB_INTERVAL,default=24h"`
	// OperationReschedulePeriod the period when data harvested from this operation is considered obsolete and need to be refetched
	OperationReschedulePeriod time.Duration `envconfig:"APP_OPERATIONS_MANAGER_RESCHEDULE_PERIOD,default=72h"`
	// RescheduleHangedOperationsJobInterval how frequently the reschedule job for hanged operations will be executed
	RescheduleHangedOperationsJobInterval time.Duration `envconfig:"APP_OPERATIONS_MANAGER_RESCHEDULE_HANGED_JOB_INTERVAL,default=1h"`
	// OperationHangPeriod the max time period when operation is considered as hanged if not completed
	OperationHangPeriod time.Duration `envconfig:"APP_OPERATIONS_MANAGER_RESCHEDULE_HANGED_PERIOD,default=2h"`
}
