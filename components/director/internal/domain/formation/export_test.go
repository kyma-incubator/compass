package formation

import "github.com/kyma-incubator/compass/components/director/pkg/persistence"

// NewServiceWithAsaEngine creates formation service with the provided ASAEngine
func NewServiceWithAsaEngine(transact persistence.Transactioner, applicationRepository applicationRepository, labelDefRepository labelDefRepository, labelRepository labelRepository, formationRepository FormationRepository, formationTemplateRepository FormationTemplateRepository, labelService labelService, uuidService uuidService, labelDefService labelDefService, asaRepo automaticFormationAssignmentRepository, asaService automaticFormationAssignmentService, tenantSvc tenantService, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, formationAssignmentService formationAssignmentService, formationAssignmentNotificationService FormationAssignmentNotificationsService, notificationsService NotificationsService, constraintEngine constraintEngine, runtimeTypeLabelKey, applicationTypeLabelKey string, engine asaEngine) *service {
	return &service{
		applicationRepository:       applicationRepository,
		labelDefRepository:                     labelDefRepository,
		labelRepository:                        labelRepository,
		formationRepository:                    formationRepository,
		formationTemplateRepository:            formationTemplateRepository,
		labelService:                           labelService,
		labelDefService:                        labelDefService,
		asaService:                             asaService,
		uuidService:                            uuidService,
		tenantSvc:                              tenantSvc,
		repo:                                   asaRepo,
		runtimeRepo:                            runtimeRepo,
		runtimeContextRepo:                     runtimeContextRepo,
		formationAssignmentService:             formationAssignmentService,
		formationAssignmentNotificationService: formationAssignmentNotificationService,
		notificationsService:                   notificationsService,
		constraintEngine:                       constraintEngine,
		transact:                               transact,
		runtimeTypeLabelKey:                    runtimeTypeLabelKey,
		applicationTypeLabelKey:                applicationTypeLabelKey,
		asaEngine:                              engine,
	}
}
