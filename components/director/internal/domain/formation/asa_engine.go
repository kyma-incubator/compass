package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ASAEngine processes Automatic Scenario Assignments
type ASAEngine struct {
	asaRepository               automaticFormationAssignmentRepository
	runtimeRepository           runtimeRepository
	runtimeContextRepository    runtimeContextRepository
	formationRepository         FormationRepository
	formationTemplateRepository FormationTemplateRepository
	runtimeTypeLabelKey         string
	applicationTypeLabelKey     string
}

// NewASAEngine returns ASAEngine
func NewASAEngine(asaRepository automaticFormationAssignmentRepository, runtimeRepository runtimeRepository, runtimeContextRepository runtimeContextRepository, formationRepository FormationRepository, formationTemplateRepository FormationTemplateRepository, runtimeTypeLabelKey, applicationTypeLabelKey string) *ASAEngine {
	return &ASAEngine{
		asaRepository:               asaRepository,
		runtimeRepository:           runtimeRepository,
		runtimeContextRepository:    runtimeContextRepository,
		formationRepository:         formationRepository,
		formationTemplateRepository: formationTemplateRepository,
		runtimeTypeLabelKey:         runtimeTypeLabelKey,
		applicationTypeLabelKey:     applicationTypeLabelKey,
	}
}

// EnsureScenarioAssigned ensures that the scenario is assigned to all the runtimes and runtimeContexts that are in the ASAs target_tenant_id
func (s *ASAEngine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error {
	return s.processScenario(ctx, in, processScenarioFunc, model.AssignFormation)
}

// RemoveAssignedScenario removes all the scenarios that are coming from the provided ASA
func (s *ASAEngine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error {
	return s.processScenario(ctx, in, processScenarioFunc, model.UnassignFormation)
}

func (s *ASAEngine) processScenario(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc, processingType model.FormationOperation) error {
	runtimeTypes, err := s.getFormationTemplateRuntimeTypes(ctx, in.ScenarioName, in.Tenant)
	if err != nil {
		return err
	}

	lblFilters := make([]*labelfilter.LabelFilter, 0, len(runtimeTypes))
	for _, runtimeType := range runtimeTypes {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, runtimeType)
		lblFilters = append(lblFilters, labelfilter.NewForKeyWithQuery(s.runtimeTypeLabelKey, query))
	}

	ownedRuntimes, err := s.runtimeRepository.ListOwnedRuntimes(ctx, in.TargetTenantID, lblFilters)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range ownedRuntimes {
		hasRuntimeContext, err := s.runtimeContextRepository.ExistsByRuntimeID(ctx, in.TargetTenantID, r.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting runtime contexts for runtime with id %q", r.ID)
		}

		// If the runtime has runtime context that is so called "multi-tenant" runtime, and we should not assign the runtime to formation.
		// In such cases only the runtime context should be assigned to formation. That happens with the "for" cycle below.
		if hasRuntimeContext {
			continue
		}

		// If the runtime has not runtime context, then it's a "single tenant" runtime, and we have to assign it to formation.
		if _, err = processScenarioFunc(ctx, in.Tenant, r.ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime with id %s from formation %s coming from ASA", processingType, r.ID, in.ScenarioName)
		}
	}

	// The part below covers the "multi-tenant" runtime case that we skipped above and
	// gets all runtimes(with and without owner access) and assign every runtime context(if there is any) for each of the runtimes to formation.
	runtimes, err := s.runtimeRepository.ListAllWithUnionSetCombination(ctx, in.TargetTenantID, lblFilters)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range runtimes {
		runtimeContext, err := s.runtimeContextRepository.GetByRuntimeID(ctx, in.TargetTenantID, r.ID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				continue
			}
			return errors.Wrapf(err, "while getting runtime context for runtime with ID: %q", r.ID)
		}

		if _, err = processScenarioFunc(ctx, in.Tenant, runtimeContext.ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime context with id %s from formation %s coming from ASA", processingType, runtimeContext.ID, in.ScenarioName)
		}
	}

	return nil
}

func (s *ASAEngine) isASAMatchingRuntime(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error) {
	runtimeTypes, err := s.getFormationTemplateRuntimeTypes(ctx, asa.ScenarioName, asa.Tenant)
	if err != nil {
		return false, err
	}

	lblFilters := make([]*labelfilter.LabelFilter, 0, len(runtimeTypes))
	for _, runtimeType := range runtimeTypes {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, runtimeType)
		lblFilters = append(lblFilters, labelfilter.NewForKeyWithQuery(s.runtimeTypeLabelKey, query))
	}

	runtimeExists, err := s.runtimeRepository.OwnerExistsByFiltersAndID(ctx, asa.TargetTenantID, runtimeID, lblFilters)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if runtime with id %q have owner=true", runtimeID)
	}

	if !runtimeExists {
		return false, nil
	}

	// If the runtime has runtime contexts then it's a "multi-tenant" runtime, and it should NOT be matched by the ASA and should NOT be added to formation.
	hasRuntimeContext, err := s.runtimeContextRepository.ExistsByRuntimeID(ctx, asa.TargetTenantID, runtimeID)
	if err != nil {
		return false, errors.Wrapf(err, "while cheking runtime context existence for runtime with ID: %q", runtimeID)
	}

	return !hasRuntimeContext, nil
}

func (s *ASAEngine) isASAMatchingRuntimeContext(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeContextID string) (bool, error) {
	runtimeTypes, err := s.getFormationTemplateRuntimeTypes(ctx, asa.ScenarioName, asa.Tenant)
	if err != nil {
		return false, err
	}

	lblFilters := make([]*labelfilter.LabelFilter, 0, len(runtimeTypes))
	for _, runtimeType := range runtimeTypes {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, runtimeType)
		lblFilters = append(lblFilters, labelfilter.NewForKeyWithQuery(s.runtimeTypeLabelKey, query))
	}

	rtmCtx, err := s.runtimeContextRepository.GetByID(ctx, asa.TargetTenantID, runtimeContextID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "while getting runtime contexts with ID: %q", runtimeContextID)
	}

	_, err = s.runtimeRepository.GetByFiltersAndIDUsingUnion(ctx, asa.TargetTenantID, rtmCtx.RuntimeID, lblFilters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "while getting runtime with ID: %q and label with key: %q and value: %q", rtmCtx.RuntimeID, s.runtimeTypeLabelKey, runtimeTypes)
	}

	return true, nil
}

func (s *ASAEngine) IsFormationComingFromASA(ctx context.Context, objectID, formation string, objectType graphql.FormationObjectType) (bool, error) {
	formationsFromASA, err := s.GetScenariosFromMatchingASAs(ctx, objectID, objectType)
	if err != nil {
		return false, errors.Wrapf(err, "while getting formations from ASAs for %s with id: %q", objectType, objectID)
	}

	for _, formationFromASA := range formationsFromASA {
		if formation == formationFromASA {
			return true, nil
		}
	}

	return false, nil
}

func (s *ASAEngine) getFormationTemplateRuntimeTypes(ctx context.Context, scenarioName, tenant string) ([]string, error) {
	log.C(ctx).Debugf("Getting formation with name: %q in tenant: %q", scenarioName, tenant)
	formation, err := s.formationRepository.GetByName(ctx, scenarioName, tenant)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation by name %q", scenarioName)
	}

	log.C(ctx).Debugf("Getting formation template with ID: %q", formation.FormationTemplateID)
	formationTemplate, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation template by id %q", formation.FormationTemplateID)
	}

	return formationTemplate.RuntimeTypes, nil
}

func (s *ASAEngine) GetMatchingFuncByFormationObjectType(objType graphql.FormationObjectType) (MatchingFunc, error) {
	switch objType {
	case graphql.FormationObjectTypeRuntime:
		return s.isASAMatchingRuntime, nil
	case graphql.FormationObjectTypeRuntimeContext:
		return s.isASAMatchingRuntimeContext, nil
	}
	return nil, errors.Errorf("unexpected formation object type %q", objType)
}

// GetScenariosFromMatchingASAs gets all the scenarios that should be added to the runtime based on the matching Automatic Scenario Assignments
// In order to do that, the ASAs should be searched in the caller tenant as this is the tenant that modifies the runtime and this is the tenant that the ASA
// produced labels should be added to.
func (s *ASAEngine) GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error) {
	log.C(ctx).Infof("Getting scenarios matching from ASA for object with ID: %q and type: %q", objectID, objType)
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	matchFunc, err := s.GetMatchingFuncByFormationObjectType(objType)
	if err != nil {
		return nil, err
	}

	scenarioAssignments, err := s.asaRepository.ListAll(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Automatic Scenario Assignments in tenant: %s", tenantID)
	}
	log.C(ctx).Infof("Found %d ASA(s) in tenant with ID: %q", len(scenarioAssignments), tenantID)

	matchingASAs := make([]*model.AutomaticScenarioAssignment, 0, len(scenarioAssignments))
	for _, scenarioAssignment := range scenarioAssignments {
		matches, err := matchFunc(ctx, scenarioAssignment, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while checking if asa matches runtime with ID %s", objectID)
		}
		if matches {
			matchingASAs = append(matchingASAs, scenarioAssignment)
		}
	}

	scenarios := make([]string, 0)
	for _, sa := range matchingASAs {
		scenarios = append(scenarios, sa.ScenarioName)
	}
	log.C(ctx).Infof("Matched scenarios from ASA are: %v", scenarios)

	return scenarios, nil
}
