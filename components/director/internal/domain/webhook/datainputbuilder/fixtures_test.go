package datainputbuilder_test

const (
	ApplicationID               = "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	Application2ID              = "6f5389cf-4f9e-46b3-9870-624d792d94ad"
	ApplicationTemplateID       = "58963c6f-24f6-4128-a05c-51d5356e7e09"
	ApplicationTemplate2ID      = "b203a000-02af-449b-a8b5-fd0787a9fa4e"
	ApplicationTenantID         = "d456f5a3-9b1f-42d5-aa81-8fde48cddfbd"
	ApplicationTemplateTenantID = "946d2550-5d90-4727-93e7-87c48286a6e7"
	globalSubaccountIDLabelKey  = "global_subaccount_id"
)

func fixApplicationLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"app-label-key": "app-label-value",
	}
}

func fixApplicationLabelsMapWithUnquotableLabels() map[string]interface{} {
	return map[string]interface{}{
		"app-label-key": []string{"app-label-value"},
	}
}

func fixLabelsMapForApplicationWithLabels() map[string]string {
	return map[string]string{
		"app-label-key": "app-label-value",
	}
}

func fixLabelsMapForApplicationWithCompositeLabels() map[string]string {
	return map[string]string{
		"app-label-key": "[\"app-label-value\"]",
	}
}

func fixLabelsMapForApplicationTemplateWithLabels() map[string]string {
	return map[string]string{
		"apptemplate-label-key": "apptemplate-label-value",
	}
}

func fixLabelsMapForApplicationTemplateWithSubaccountLabels() map[string]string {
	return map[string]string{
		globalSubaccountIDLabelKey: ApplicationTemplateTenantID,
		"apptemplate-label-key":    "apptemplate-label-value",
	}
}
