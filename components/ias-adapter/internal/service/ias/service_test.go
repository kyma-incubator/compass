package ias

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testConsumedAPI = types.ApplicationConsumedAPI{
	Name:    "name1",
	APIName: "apiname1",
	AppID:   "appId1",
}

type testTransport struct {
	err    error
	status int
	body   string
}

func (tt *testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if tt.err != nil {
		return nil, tt.err
	}

	return &http.Response{
		StatusCode: tt.status,
		Body:       io.NopCloser(strings.NewReader(tt.body)),
	}, nil
}

var _ = Describe("Adding consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Should add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			newConsumedAPI := types.ApplicationConsumedAPI{
				Name:    "name2",
				APIName: "apiname2",
				AppID:   "appId2",
			}
			addConsumedAPI(&consumedAPIs, newConsumedAPI)
			Expect(len(consumedAPIs)).To(Equal(2))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Shouldn't add it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			addConsumedAPI(&consumedAPIs, testConsumedAPI)
			Expect(len(consumedAPIs)).To(Equal(1))
		})
	})
})

var _ = Describe("Removing consumed API", func() {
	When("It doesn't exist in the current consumed APIs", func() {
		It("Shouldn't do anything", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			removeConsumedAPI(&consumedAPIs, "non-existent-api-name")
			Expect(len(consumedAPIs)).To(Equal(1))
		})
	})
	When("It already exists in the current consumed APIs", func() {
		It("Should remove it", func() {
			consumedAPIs := []types.ApplicationConsumedAPI{testConsumedAPI}
			removeConsumedAPI(&consumedAPIs, "apiname1")
			Expect(len(consumedAPIs)).To(Equal(0))
		})
	})
})

var _ = Describe("Getting application", func() {
	config := config.IAS{}
	ctx := context.Background()
	iasHost := "ias-host"
	clientID := "client-id"
	appTenantId := "app-tenant-id"

	When("IAS returns an error", func() {
		It("Returns an error", func() {
			err := errors.New("connection reset")
			service := NewService(config, &http.Client{Transport: &testTransport{err: err}})
			_, err = service.GetApplication(ctx, iasHost, clientID, appTenantId)
			Expect(err).To(Equal(err))
		})
	})

	When("There are no applications with the specified client id", func() {
		It("Returns IAS App not found error", func() {
			apps := types.Applications{}
			b, err := json.Marshal(apps)
			Expect(err).ToNot(HaveOccurred())

			service := NewService(config, &http.Client{Transport: &testTransport{status: http.StatusOK, body: string(b)}})
			_, err = service.GetApplication(ctx, iasHost, clientID, appTenantId)
			Expect(err).To(MatchError(errors.IASApplicationNotFound))
		})
	})

	When("There are no applications with the specified app tenant id", func() {
		It("Returns IAS App not found error", func() {
			apps := types.Applications{
				Applications: []types.Application{
					{
						Authentication: types.ApplicationAuthentication{
							SAPManagedAttributes: types.SAPManagedAttributes{
								AppTenantID: "some-other-app-tenant-id",
							},
						},
					},
				},
			}
			b, err := json.Marshal(apps)
			Expect(err).ToNot(HaveOccurred())

			service := NewService(config, &http.Client{Transport: &testTransport{status: http.StatusOK, body: string(b)}})
			_, err = service.GetApplication(ctx, iasHost, clientID, appTenantId)
			Expect(err).To(MatchError(errors.IASApplicationNotFound))
		})
	})
})
