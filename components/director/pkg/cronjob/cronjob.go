package cronjob

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const hostnameEnvVar = "HOSTNAME"

// ElectionConfig configuration for k8s leader election with lease
type ElectionConfig struct {
	LeaseLockName      string        `envconfig:"APP_ELECTION_LEASE_LOCK_NAME"`
	LeaseLockNamespace string        `envconfig:"APP_ELECTION_LEASE_LOCK_NAMESPACE"`
	LeaseDuration      time.Duration `envconfig:"optional,default=60s,APP_ELECTION_LEASE_DURATION"`
	RenewDeadline      time.Duration `envconfig:"optional,default=15s,APP_ELECTION_RENEW_DEADLINE"`
	RetryPeriod        time.Duration `envconfig:"optional,default=5s,APP_ELECTION_RETRY_PERIOD"`
	ElectionEnabled    bool          `envconfig:"optional,default=true,APP_ELECTION_ENABLED"`
	ClientConfig       kubernetes.Config
}

// CronJob represents a job that executes Fn on every SchedulePeriod.
type CronJob struct {
	Name           string
	Fn             func(ctx context.Context)
	SchedulePeriod time.Duration
}

type cronJobRunner struct {
	CronJob CronJob
	stop    context.CancelFunc
}

func (r *cronJobRunner) Start(ctx context.Context) {
	newCtx, stop := context.WithCancel(ctx)
	r.stop = stop
	defer r.Stop()

	for {
		start := time.Now()
		log.C(ctx).Infof("Starting CronJob %s execution", r.CronJob.Name)
		r.CronJob.Fn(newCtx)
		if newCtx.Err() != nil {
			log.C(ctx).Infof("CronJob %s stopped due to context done", r.CronJob.Name)
			return
		}
		jobTime := time.Since(start)
		log.C(ctx).Infof("CronJob %s executed for %s", r.CronJob.Name, jobTime.String())
		if jobTime < r.CronJob.SchedulePeriod {
			waitPeriod := r.CronJob.SchedulePeriod - jobTime
			log.C(ctx).Infof("Scheduling CronJob %s to run after %s", r.CronJob.Name, waitPeriod.String())

			select {
			case <-newCtx.Done():
				log.C(ctx).Infof("Context of CronJob %s is done. Exiting CronJob loop...", r.CronJob.Name)
				return
			case <-time.After(waitPeriod):
				log.C(ctx).Infof("Waited %s to run next iteration of CronJob %s",
					waitPeriod.String(), r.CronJob.Name)
			}
		}
	}
}

func (r *cronJobRunner) Stop() {
	if r.stop != nil {
		r.stop()
	}
}

func runLeaderLeaseLoop(ctx context.Context, electionConfig ElectionConfig, job CronJob) error {
	k8sConfig := electionConfig.ClientConfig
	client, err := kubernetes.NewKubernetesClientSet(
		ctx, k8sConfig.PollInterval, k8sConfig.PollTimeout, k8sConfig.Timeout)
	if err != nil {
		return err
	}
	electionID := os.Getenv(hostnameEnvVar)
	if electionID == "" {
		return fmt.Errorf("not running in k8s pod. Env variable %s not set", hostnameEnvVar)
	}

	runner := cronJobRunner{
		CronJob: job,
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      electionConfig.LeaseLockName,
			Namespace: electionConfig.LeaseLockNamespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: electionID,
		},
	}

	leaderElectionConfig := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   electionConfig.LeaseDuration,
		RenewDeadline:   electionConfig.RenewDeadline,
		RetryPeriod:     electionConfig.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.C(ctx).Infof("Starting CronJob executor on %s", electionID)
				runner.Start(ctx)
				log.C(ctx).Infof("CronJob executor on %s exited", electionID)
			},
			OnStoppedLeading: func() {
				log.C(ctx).Errorf("Instance %s is no longer leader. Stopping CronJob executor", electionID)
				runner.Stop()
			},
			OnNewLeader: func(identity string) {
				log.C(ctx).Debugf("Instance %s elected as leader", identity)
			},
		},
	}

	leaderElection, err := leaderelection.NewLeaderElector(leaderElectionConfig)
	if err != nil {
		return err
	}

	leaderElection.Run(ctx)
	return nil
}

func runCronJobWithElection(ctx context.Context, cfg ElectionConfig, job CronJob) error {
	for {
		if err := runLeaderLeaseLoop(ctx, cfg, job); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Leader lease loop context is done, exiting leader lease loop...")
			return nil
		default:
			log.C(ctx).Error("Leader lease loop ended, re-running...")
		}
	}
}

// RunCronJob runs a CronJob and blocks.
// If cfg.LeaseEnabled is true then only one pod (if application is scaled) will run the cron job.
// This is done using leader election from k8s with leases.
// Returns error in case of bad configuration or bad connection to k8s cluster
func RunCronJob(ctx context.Context, cfg ElectionConfig, job CronJob) error {
	if cfg.ElectionEnabled {
		return runCronJobWithElection(ctx, cfg, job)
	}

	runner := cronJobRunner{
		CronJob: job,
	}
	runner.Start(ctx)
	return nil
}
