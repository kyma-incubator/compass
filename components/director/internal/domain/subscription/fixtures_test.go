package subscription_test

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	testDescription    = "{{display-name}}"
	testProviderName   = "provider-display-name"
	testURL            = "http://valid.url"
	appInputJSONString = `{"Name":"foo","ProviderName":"compass","Description":"Lorem ipsum","Labels":{"test":["val","val2"]},"HealthCheckURL":"https://foo.bar","Webhooks":[{"Type":"","URL":"webhook1.foo.bar","Auth":null},{"Type":"","URL":"webhook2.foo.bar","Auth":null}],"IntegrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
)

func fixJSONApplicationCreateInput(name string) string {
	return fmt.Sprintf(`{"name": "%s", "providerName": "%s", "description": "%s", "healthCheckURL": "%s"}`, name, testProviderName, testDescription, testURL)
}

func fixModelAppTemplateWithAppInputJSON(id, name, appInputJSON string) *model.ApplicationTemplate {
	out := fixModelApplicationTemplate(id, name)
	out.ApplicationInputJSON = appInputJSON

	return out
}

func fixModelApplicationTemplate(id, name string) *model.ApplicationTemplate {
	desc := testDescription
	out := model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholders(),
		Webhooks:             []model.Webhook{},
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}

	return &out
}

func fixModelApplication(id, name, appTemplateID string) *model.Application {
	return &model.Application{
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:                  name,
		BaseEntity:            &model.BaseEntity{ID: id},
		ApplicationTemplateID: &appTemplateID,
	}
}

func fixModelRuntime(name string) *model.Runtime {
	desc := testDescription
	out := model.Runtime{
		Name:        name,
		Description: &desc,
	}

	return &out
}

func fixModelPlaceholders() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := testDescription
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "name",
			Description: &placeholderDesc,
		},
		{
			Name:        "display-name",
			Description: &placeholderDesc,
		},
	}
}

func fixModelApplicationFromTemplateInput(name, subscribedAppName string) model.ApplicationFromTemplateInput {
	return model.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*model.ApplicationTemplateValueInput{
			{Placeholder: "name", Value: subscribedAppName},
			{Placeholder: "display-name", Value: subscribedAppName},
		},
	}
}

func fixGQLApplicationCreateInput(name string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:           name,
		ProviderName:   &testProviderName,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixModelApplicationCreateInput(name string) model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixModelApplicationCreateInputWithLabels(name, subscribedSubaccountID string) model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
		Labels: map[string]interface{}{
			"managed":                          "false",
			scenarioassignment.SubaccountIDKey: subscribedSubaccountID,
		},
	}
}
