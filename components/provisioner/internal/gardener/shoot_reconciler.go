package gardener

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"time"

	proviRuntime "github.com/kyma-incubator/compass/components/provisioner/internal/runtime"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"

	"k8s.io/client-go/util/retry"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewReconciler(
	mgr ctrl.Manager,
	dbsFactory dbsession.Factory,
	secretsClient v1core.SecretInterface,
	shootClient v1beta1.ShootInterface,
	installationService installation.Service,
	installationTimeout time.Duration,
	directorClient director.DirectorClient,
	runtimeConfigurator proviRuntime.Configurator) *Reconciler {
	return &Reconciler{
		client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		log:        logrus.WithField("Component", "ShootReconciler"),
		dbsFactory: dbsFactory,

		provisioningOperator: &ProvisioningOperator{
			dbsFactory:              dbsFactory,
			secretsClient:           secretsClient,
			shootClient:             shootClient,
			installationService:     installationService,
			kymaInstallationTimeout: installationTimeout,
			directorClient:          directorClient,
			runtimeConfigurator:     runtimeConfigurator,
		},
	}
}

type Reconciler struct {
	client     client.Client
	scheme     *runtime.Scheme
	dbsFactory dbsession.Factory

	log                  *logrus.Entry
	provisioningOperator *ProvisioningOperator
}

type ProvisioningOperator struct {
	installationErrorsThreshold int
	secretsClient               v1core.SecretInterface
	shootClient                 v1beta1.ShootInterface
	installationService         installation.Service
	dbsFactory                  dbsession.Factory
	kymaInstallationTimeout     time.Duration
	directorClient              director.DirectorClient
	runtimeConfigurator         proviRuntime.Configurator
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithField("Shoot", req.NamespacedName)
	log.Infof("reconciling Shoot")

	var shoot gardener_types.Shoot
	if err := r.client.Get(context.Background(), req.NamespacedName, &shoot); err != nil {
		if errors.IsNotFound(err) {
			return r.provisioningOperator.HandleShootDeletion(log, req.NamespacedName)
		}

		log.Error(err, "unable to get shoot")
		return ctrl.Result{}, err
	}

	shouldReconcile, err := r.shouldReconcileShoot(shoot)
	if err != nil {
		log.Errorf("Failed to verify if shoot should be reconciled")
		return ctrl.Result{}, err
	}
	if !shouldReconcile {
		log.Debugf("Gardener cluster not found in database, shoot will be ignored")
		return ctrl.Result{}, nil
	}
	runtimeId := getRuntimeId(shoot)
	log = log.WithField("RuntimeId", runtimeId)

	// Get operationId from annotation
	operationId := getOperationId(shoot)
	if operationId == "" {
		err := r.handleShootWithoutOperationId(log, shoot)
		if err != nil {
			log.Errorf("Error handling shoot without operation Id: %s", err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Proceed this path only if there is OperationId on Shoot
	log = log.WithField("OperationId", operationId)

	if isBeingDeleted(shoot) {
		// TODO - here we can ensure if last operation for cluster is Deprovision
		log.Infof("Shoot is being deleted: %s", getProvisioningState(shoot))
		return ctrl.Result{}, nil
	}

	if shoot.Status.LastOperation == nil {
		log.Warnf("Shoot does not have last operation status")
		return ctrl.Result{}, nil
	}

	provisioningStep := getProvisioningStep(shoot)

	switch provisioningStep {
	case ProvisioningInProgressStep:
		return r.provisioningOperator.ProvisioningInProgress(log, shoot, operationId)
	case InstallationInProgressStep:
		return r.provisioningOperator.InstallationInProgress(log, shoot, operationId)
	case ProvisioningFinishedStep:
		return r.provisioningOperator.ProvisioningFinished(log, shoot)
	case DeprovisioningInProgressStep:
		return r.provisioningOperator.DeprovisioningInProgress(log, shoot, operationId)
	default:
		log.Warnf("Unknown state of Shoot")
		return ctrl.Result{}, nil
	}
}

func (r *Reconciler) handleShootWithoutOperationId(log *logrus.Entry, shoot gardener_types.Shoot) error {
	// TODO: We can verify shoot status here - ensure it is ok
	log.Debug("Shoot without operation ID is ignored for now")
	return nil
}

func (r *Reconciler) shouldReconcileShoot(shoot gardener_types.Shoot) (bool, error) {
	session := r.dbsFactory.NewReadSession()

	_, err := session.GetGardenerClusterByName(shoot.Name)
	if err != nil {
		if err.Code() == dberrors.CodeNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func isShootFailed(shoot gardener_types.Shoot) bool {
	return shoot.Status.LastOperation.State == gardencorev1beta1.LastOperationStateFailed
}

func isBeingDeleted(shoot gardener_types.Shoot) bool {
	return shoot.DeletionTimestamp != nil
}

func (r *ProvisioningOperator) HandleShootDeletion(log *logrus.Entry, name types.NamespacedName) (ctrl.Result, error) {
	log.Info("Shoot deleted")

	session := r.dbsFactory.NewReadSession()

	cluster, dberr := session.GetGardenerClusterByName(name.Name)
	if dberr != nil {
		log.Errorf("error getting Gardener cluster by name: %s", dberr.Error())
		return ctrl.Result{}, dberr
	}

	lastOperation, dberr := session.GetLastOperation(cluster.ID)
	if dberr != nil {
		log.Errorf("error getting last operation for %s Runtime: %s", cluster.ID, dberr.Error())
		return ctrl.Result{}, dberr
	}

	if lastOperation.Type == model.Deprovision {
		if lastOperation.State == model.InProgress || lastOperation.State == model.Succeeded {
			err := r.setDeprovisioningFinished(cluster, lastOperation)
			if err != nil {
				log.Errorf("error setting deprovisioning finished: %s", err.Error())
				return ctrl.Result{}, err
			}
		} else {
			// TODO: we can implement some more cases here
			err := fmt.Errorf("Error: Invalid state. Shoot deleted, last operation: %s, state: %s", lastOperation.Type, lastOperation.State)
			log.Errorf(err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// TODO: here we can implement some logic to recover from such state
	err := fmt.Errorf("Error: Invalid state. Shoot deleted, last operation: %s, state: %s", lastOperation.Type, lastOperation.State)
	log.Errorf(err.Error())
	return ctrl.Result{}, err
}

func (r *ProvisioningOperator) updateShoot(shoot gardener_types.Shoot, modifyShootFn func(s *gardener_types.Shoot)) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		refetchedShoot, err := r.shootClient.Get(shoot.Name, v1.GetOptions{})
		if err != nil {
			return err
		}

		modifyShootFn(refetchedShoot)

		refetchedShoot, err = r.shootClient.Update(refetchedShoot)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *ProvisioningOperator) setDeprovisioningFinished(cluster model.Cluster, lastOp model.Operation) error {
	session, dberr := r.dbsFactory.NewSessionWithinTransaction()
	if dberr != nil {
		return fmt.Errorf("error starting db session with transaction: %s", dberr.Error())
	}
	defer session.RollbackUnlessCommitted()

	dberr = session.MarkClusterAsDeleted(cluster.ID)
	if dberr != nil {
		return fmt.Errorf("error marking cluster for deletion: %s", dberr.Error())
	}

	dberr = session.UpdateOperationState(lastOp.ID, "Operation succeeded.", model.Succeeded)
	if dberr != nil {
		return fmt.Errorf("error setting deprovisioning operation %s as succeeded: %s", lastOp.ID, dberr.Error())
	}

	err := r.directorClient.DeleteRuntime(cluster.ID, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error deleting Runtime form Director: %s", err.Error())
	}

	dberr = session.Commit()
	if dberr != nil {
		return fmt.Errorf("error commiting transaction: %s", dberr.Error())
	}

	return nil
}
