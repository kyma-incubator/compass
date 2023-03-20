package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Tenant Mapping Handler", func() {
	var tenantMapping = types.TenantMapping{
		FormationID: "formationID",
		ReceiverTenant: types.ReceiverTenant{
			ApplicationURL: "localhost",
		},
		AssignedTenants: []types.AssignedTenant{
			{
				Operation: types.OperationAssign,
				Parameters: types.AssignedTenantParameters{
					ClientID: "clientID",
				},
			},
		},
	}
	When("Tenant mapping cannot be decoded", func() {
		It("Should fail with 422", func() {
			handler := TenantMappingsHandler{
				Service: &automock.TenantMappingsService{},
			}

			body := strings.NewReader("unprocessable body")
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring("Failed to decode tenant mapping body"))
			Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
		})
	})
	When("Tenant mapping is invalid", func() {
		It("Should fail with 422", func() {
			handler := TenantMappingsHandler{
				Service: &automock.TenantMappingsService{},
			}

			body := strings.NewReader("{}")
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring("Tenant mapping body is invalid"))
			Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
		})
	})
	When("Consumed APIs cannot be updated", func() {
		It("Should fail with 500", func() {
			service := &automock.TenantMappingsService{}
			service.On("UpdateApplicationsConsumedAPIs", mock.Anything, mock.Anything).Return(errors.New("error"))
			handler := TenantMappingsHandler{
				Service: service,
			}

			data, err := json.Marshal(tenantMapping)
			Expect(err).Error().ToNot(HaveOccurred())
			body := bytes.NewReader(data)
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})
	When("Consumed APIs are successfully updated", func() {
		It("Should return 200", func() {
			service := &automock.TenantMappingsService{}
			service.On("UpdateApplicationsConsumedAPIs", mock.Anything, mock.Anything).Return(nil)
			handler := TenantMappingsHandler{
				Service: service,
			}

			data, err := json.Marshal(tenantMapping)
			Expect(err).Error().ToNot(HaveOccurred())
			body := bytes.NewReader(data)
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
})
