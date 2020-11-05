package osb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"

	"github.com/kyma-incubator/compass/components/system-broker/internal/osb/osbfakes"
)

func TestServicesReturnsErrorForInvalidApplications(t *testing.T) {
	app := schema.ApplicationExt{
		Application: schema.Application{
			Name: "test-app",
		},
	}
	output := director.ApplicationsOutput{
		Result: &schema.ApplicationPageExt{
			ApplicationPage: schema.ApplicationPage{},
			Data:            []*schema.ApplicationExt{&app},
		},
	}

	lister := &osbfakes.FakeApplicationsLister{}
	lister.FetchApplicationsReturns(&output, nil)
	converter := &osbfakes.FakeConverter{}
	converter.ConvertReturns(nil, errors.New("test-error"))

	endpoint := osb.NewCatalogEndpoint(lister, converter)
	_, err := endpoint.Services(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 1, lister.FetchApplicationsCallCount())
	assert.Equal(t, 1, converter.ConvertCallCount())
}
