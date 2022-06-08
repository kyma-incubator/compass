package formationtemplate_test

import "github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"

const (
	testID = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
)

func UnusedFormationTemplateRepository() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func UnusedFormationTemplateConverter() *automock.FormationTemplateConverter {
	return &automock.FormationTemplateConverter{}
}
