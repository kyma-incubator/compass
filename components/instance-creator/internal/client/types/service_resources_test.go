package types

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ServiceBinding_MatchMultiple(t *testing.T) {
	testCases := []struct {
		Name                          string
		ServiceBindingMatchParameters ServiceBindingMatchParameters
		InputResources                resources.Resources
		ExpectedMatchedResources      []string
		ExpectedErrMsg                string
	}{
		{
			Name:                          "Success - single service instance ID in parameters",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{"instance-id-1"}},
			InputResources: &ServiceBindings{
				NumItems: 3,
				Items: []*ServiceBinding{
					{
						ID:                "binding-id-1",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-2",
						ServiceInstanceID: "instance-id-2",
					},
					{
						ID:                "binding-id-3",
						ServiceInstanceID: "instance-id-1",
					},
				},
			},
			ExpectedMatchedResources: []string{"binding-id-1", "binding-id-3"},
		},
		{
			Name:                          "Success - multiple service instance ID in parameters",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{"instance-id-1", "instance-id-2"}},
			InputResources: &ServiceBindings{
				NumItems: 4,
				Items: []*ServiceBinding{
					{
						ID:                "binding-id-1",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-2",
						ServiceInstanceID: "instance-id-2",
					},
					{
						ID:                "binding-id-3",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-4",
						ServiceInstanceID: "instance-id-3",
					},
				},
			},
			ExpectedMatchedResources: []string{"binding-id-1", "binding-id-2", "binding-id-3"},
		},
		{
			Name:                          "Success - match zero bindings",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{"instance-id-10"}},
			InputResources: &ServiceBindings{
				NumItems: 4,
				Items: []*ServiceBinding{
					{
						ID:                "binding-id-1",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-2",
						ServiceInstanceID: "instance-id-2",
					},
					{
						ID:                "binding-id-3",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-4",
						ServiceInstanceID: "instance-id-3",
					},
				},
			},
			ExpectedMatchedResources: []string{},
		},
		{
			Name:                          "Success - empty parameters - match zero bindings",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{}},
			InputResources: &ServiceBindings{
				NumItems: 4,
				Items: []*ServiceBinding{
					{
						ID:                "binding-id-1",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-2",
						ServiceInstanceID: "instance-id-2",
					},
					{
						ID:                "binding-id-3",
						ServiceInstanceID: "instance-id-1",
					},
					{
						ID:                "binding-id-4",
						ServiceInstanceID: "instance-id-3",
					},
				},
			},
			ExpectedMatchedResources: []string{},
		},
		{
			Name:                          "Success - empty resources - match zero bindings",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{"instance-id-1"}},
			InputResources:                &ServiceBindings{},
			ExpectedMatchedResources:      []string{},
		},
		{
			Name:                          "Error - input resources are not service bindings",
			ServiceBindingMatchParameters: ServiceBindingMatchParameters{ServiceInstancesIDs: []string{"instance-id-10"}},
			InputResources:                &ServiceInstances{},
			ExpectedErrMsg:                "while type asserting Resources to ServiceBindings",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualMatchedResources, err := testCase.ServiceBindingMatchParameters.MatchMultiple(testCase.InputResources)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				assert.ElementsMatch(t, testCase.ExpectedMatchedResources, actualMatchedResources)
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ServiceInstance_Match(t *testing.T) {
	testCases := []struct {
		Name                           string
		ServiceInstanceMatchParameters ServiceInstanceMatchParameters
		InputResources                 resources.Resources
		ExpectedMatchedResourceID      string
		ExpectedErrMsg                 string
	}{
		{
			Name:                           "Success",
			ServiceInstanceMatchParameters: ServiceInstanceMatchParameters{ServiceInstanceName: "instance-name-1"},
			InputResources: &ServiceInstances{
				NumItems: 2,
				Items: []*ServiceInstance{
					{
						ID:   "instance-id-1",
						Name: "instance-name-1",
					},
					{
						ID:   "instance-id-2",
						Name: "instance-name-2",
					},
				},
			},
			ExpectedMatchedResourceID: "instance-id-1",
		},
		{
			Name:                           "Success - can't find service instance with the given name",
			ServiceInstanceMatchParameters: ServiceInstanceMatchParameters{ServiceInstanceName: "instance-name-10"},
			InputResources: &ServiceInstances{
				NumItems: 2,
				Items: []*ServiceInstance{
					{
						ID:   "instance-id-1",
						Name: "instance-name-1",
					},
					{
						ID:   "instance-id-2",
						Name: "instance-name-2",
					},
				},
			},
			ExpectedMatchedResourceID: "",
		},
		{
			Name:                           "Error - input resources are not service bindings",
			ServiceInstanceMatchParameters: ServiceInstanceMatchParameters{ServiceInstanceName: "instance-name-1"},
			InputResources:                 &ServiceBindings{},
			ExpectedErrMsg:                 "while type asserting Resources to ServiceInstances",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualMatchedResourceID, err := testCase.ServiceInstanceMatchParameters.Match(testCase.InputResources)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				assert.Equal(t, testCase.ExpectedMatchedResourceID, actualMatchedResourceID)
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ServiceOffering_Match(t *testing.T) {
	testCases := []struct {
		Name                           string
		ServiceOfferingMatchParameters ServiceOfferingMatchParameters
		InputResources                 resources.Resources
		ExpectedMatchedResourceID      string
		ExpectedErrMsg                 string
	}{
		{
			Name:                           "Success",
			ServiceOfferingMatchParameters: ServiceOfferingMatchParameters{CatalogName: "catalog-name-1"},
			InputResources: &ServiceOfferings{
				NumItems: 2,
				Items: []*ServiceOffering{
					{
						ID:          "offering-id-1",
						CatalogName: "catalog-name-1",
					},
					{
						ID:          "offering-id-2",
						CatalogName: "catalog-name-2",
					},
				},
			},
			ExpectedMatchedResourceID: "offering-id-1",
		},
		{
			Name:                           "Error - can't find service instance with the given name",
			ServiceOfferingMatchParameters: ServiceOfferingMatchParameters{CatalogName: "catalog-name-10"},
			InputResources: &ServiceOfferings{
				NumItems: 2,
				Items: []*ServiceOffering{
					{
						ID:          "offering-id-1",
						CatalogName: "catalog-name-1",
					},
					{
						ID:          "offering-id-2",
						CatalogName: "catalog-name-2",
					},
				},
			},
			ExpectedErrMsg: fmt.Sprintf("couldn't find service offering for catalog name: %s", "catalog-name-10"),
		},
		{
			Name:                           "Error - input resources are not service bindings",
			ServiceOfferingMatchParameters: ServiceOfferingMatchParameters{},
			InputResources:                 &ServiceBindings{},
			ExpectedErrMsg:                 "while type asserting Resources to ServiceOfferings",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualMatchedResourceID, err := testCase.ServiceOfferingMatchParameters.Match(testCase.InputResources)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				assert.Equal(t, testCase.ExpectedMatchedResourceID, actualMatchedResourceID)
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ServicePlan_Match(t *testing.T) {
	testCases := []struct {
		Name                       string
		ServicePlanMatchParameters ServicePlanMatchParameters
		InputResources             resources.Resources
		ExpectedMatchedResourceID  string
		ExpectedErrMsg             string
	}{
		{
			Name:                       "Success",
			ServicePlanMatchParameters: ServicePlanMatchParameters{PlanName: "catalog-name-1", OfferingID: "offering-id-1"},
			InputResources: &ServicePlans{
				NumItems: 2,
				Items: []*ServicePlan{
					{
						ID:                "plan-id-1",
						CatalogName:       "catalog-name-1",
						ServiceOfferingId: "offering-id-1",
					},
					{
						ID:                "plan-id-1",
						CatalogName:       "catalog-name-2",
						ServiceOfferingId: "offering-id-2",
					},
				},
			},
			ExpectedMatchedResourceID: "plan-id-1",
		},
		{
			Name:                       "Error - can't find service instance with the given name",
			ServicePlanMatchParameters: ServicePlanMatchParameters{PlanName: "catalog-name-10", OfferingID: "catalog-name-10"},
			InputResources: &ServicePlans{
				NumItems: 2,
				Items: []*ServicePlan{
					{
						ID:                "plan-id-1",
						CatalogName:       "catalog-name-1",
						ServiceOfferingId: "offering-id-1",
					},
					{
						ID:                "plan-id-1",
						CatalogName:       "catalog-name-2",
						ServiceOfferingId: "offering-id-2",
					},
				},
			},
			ExpectedErrMsg: fmt.Sprintf("couldn't find service plan for catalog name: %s and offering ID: %s", "catalog-name-10", "catalog-name-10"),
		},
		{
			Name:                       "Error - input resources are not service bindings",
			ServicePlanMatchParameters: ServicePlanMatchParameters{},
			InputResources:             &ServiceBindings{},
			ExpectedErrMsg:             "while type asserting Resources to ServicePlans",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualMatchedResourceID, err := testCase.ServicePlanMatchParameters.Match(testCase.InputResources)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				assert.Equal(t, testCase.ExpectedMatchedResourceID, actualMatchedResourceID)
				assert.NoError(t, err)
			}
		})
	}
}
