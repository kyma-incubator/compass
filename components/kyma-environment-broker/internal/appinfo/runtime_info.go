package appinfo

import (
	"net/http"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/httputil"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/predicate"

	"github.com/pkg/errors"
)

//go:generate mockery -name=InstanceFinder -output=automock -outpkg=automock -case=underscore

type (
	InstanceFinder interface {
		FindAllJoinedWithOperations(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, error)
	}

	ResponseWriter interface {
		InternalServerError(rw http.ResponseWriter, r *http.Request, err error, context string)
	}
)

type RuntimeInfoHandler struct {
	instanceFinder          InstanceFinder
	respWriter              ResponseWriter
	defaultSubaccountRegion string
}

func NewRuntimeInfoHandler(instanceFinder InstanceFinder, region string, respWriter ResponseWriter) *RuntimeInfoHandler {
	return &RuntimeInfoHandler{
		instanceFinder:          instanceFinder,
		respWriter:              respWriter,
		defaultSubaccountRegion: region,
	}
}

func (h *RuntimeInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allInstances, err := h.instanceFinder.FindAllJoinedWithOperations(predicate.SortAscByCreatedAt())
	if err != nil {
		h.respWriter.InternalServerError(w, r, err, "while fetching all instances")
		return
	}

	dto, err := h.mapToDTO(allInstances)
	if err != nil {
		h.respWriter.InternalServerError(w, r, err, "while mapping instance model to dto")
	}

	if err := httputil.JSONEncode(w, dto); err != nil {
		h.respWriter.InternalServerError(w, r, err, "while encoding response to JSON")
		return
	}
}

func (h *RuntimeInfoHandler) mapToDTO(instances []internal.InstanceWithOperation) ([]*RuntimeDTO, error) {
	items := make([]*RuntimeDTO, 0, len(instances))
	indexer := map[string]int{}

	for _, inst := range instances {
		region, err := h.getRegionOrDefault(inst)
		if err != nil {
			return nil, errors.Wrap(err, "while getting region")
		}

		idx, found := indexer[inst.InstanceID]
		if !found {
			items = append(items, &RuntimeDTO{
				RuntimeID:         inst.RuntimeID,
				SubAccountID:      inst.SubAccountID,
				SubAccountRegion:  region,
				ServiceInstanceID: inst.InstanceID,
				GlobalAccountID:   inst.GlobalAccountID,
				ServiceClassID:    inst.ServiceID,
				ServiceClassName:  svcNameOrDefault(inst),
				ServicePlanID:     inst.ServicePlanID,
				ServicePlanName:   planNameOrDefault(inst),
				Status: StatusDTO{
					CreatedAt: getIfNotZero(inst.CreatedAt),
					UpdatedAt: getIfNotZero(inst.UpdatedAt),
					DeletedAt: getIfNotZero(inst.DeletedAt),
				},
			})
			idx = len(items) - 1
			indexer[inst.InstanceID] = idx
		}

		// TODO: consider to merge the rows in sql query
		opStatus := &OperationStatusDTO{
			State:       inst.State.String,
			Description: inst.Description.String,
		}
		switch dbmodel.OperationType(inst.Type.String) {
		case dbmodel.OperationTypeProvision:
			items[idx].Status.Provisioning = opStatus
		case dbmodel.OperationTypeDeprovision:
			items[idx].Status.Deprovisioning = opStatus
		}
	}

	return items, nil
}

func (h *RuntimeInfoHandler) getRegionOrDefault(inst internal.InstanceWithOperation) (string, error) {
	pp, err := inst.GetProvisioningParameters()
	if err != nil {
		return "", errors.Wrap(err, "while getting provisioning parameters")
	}

	if pp.PlatformRegion == "" {
		return h.defaultSubaccountRegion, nil
	}
	return pp.PlatformRegion, nil
}

func svcNameOrDefault(inst internal.InstanceWithOperation) string {
	if inst.ServiceName != "" {
		return inst.ServiceName
	}
	return broker.KymaServiceName
}

func planNameOrDefault(inst internal.InstanceWithOperation) string {
	if inst.ServicePlanName != "" {
		return inst.ServicePlanName
	}
	return broker.Plans[inst.ServicePlanID].PlanDefinition.Name
}

func getIfNotZero(in time.Time) *time.Time {
	if in.IsZero() {
		return nil
	}
	return ptr.Time(in)
}
