package webhook_test

import "github.com/kyma-incubator/compass/components/director/pkg/webhook/automock"

func unusedAppRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func unusedAppTemplateRepo() *automock.ApplicationTemplateRepository {
	return &automock.ApplicationTemplateRepository{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedRuntimeCtxRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}
