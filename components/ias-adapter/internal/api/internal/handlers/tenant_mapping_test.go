package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Tenant Mapping Handler", func() {
	var tenantMappingOkReverse = types.TenantMapping{
		FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
		ReceiverTenant: types.ReceiverTenant{
			ApplicationURL: "localhost",
		},
		AssignedTenants: []types.AssignedTenant{
			{
				UCLApplicationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				LocalTenantID:    "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				Operation:        types.OperationAssign,
				Parameters: types.AssignedTenantParameters{
					ClientID: "clientID",
				},
				ReverseAssignmentState: "READY",
			},
		},
	}

	var tenantMappingFailedReverse = types.TenantMapping{
		FormationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
		ReceiverTenant: types.ReceiverTenant{
			ApplicationURL: "localhost",
		},
		AssignedTenants: []types.AssignedTenant{
			{
				UCLApplicationID: "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				LocalTenantID:    "2d933ae2-10c4-4d6f-b4d4-5e1553e4ff05",
				Operation:        types.OperationAssign,
				Parameters: types.AssignedTenantParameters{
					ClientID: "clientID",
				},
				ReverseAssignmentState: "CREATE_ERROR",
			},
		},
	}
	When("Tenant mapping cannot be decoded", func() {
		It("Should fail with 400", func() {
			handler := TenantMappingsHandler{
				Service: &automock.TenantMappingsService{},
			}

			body := strings.NewReader("unprocessable body")
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(url.QueryEscape("failed to decode tenant mapping body")))
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	When("Tenant mapping is invalid", func() {
		It("Should fail with 400", func() {
			handler := TenantMappingsHandler{
				Service: &automock.TenantMappingsService{},
			}

			body := strings.NewReader(`{"assignedTenants":[{"configuration": ""}]}`)
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(url.QueryEscape("tenant mapping body is invalid")))
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	When("Reverse assignment state is not ready", func() {
		It("Should fail with 422", func() {
			service := &automock.TenantMappingsService{}
			handler := TenantMappingsHandler{
				Service: service,
			}

			data, err := json.Marshal(tenantMappingFailedReverse)
			Expect(err).Error().ToNot(HaveOccurred())
			body := bytes.NewReader(data)
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
		})
	})
	When("Consumed APIs cannot be updated", func() {
		It("Should fail with 500", func() {
			service := &automock.TenantMappingsService{}
			service.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(errors.New("error"))
			handler := TenantMappingsHandler{
				Service: service,
			}

			data, err := json.Marshal(tenantMappingOkReverse)
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
			service.On("ProcessTenantMapping", mock.Anything, mock.Anything).Return(nil)
			handler := TenantMappingsHandler{
				Service: service,
			}

			data, err := json.Marshal(tenantMappingOkReverse)
			Expect(err).Error().ToNot(HaveOccurred())
			body := bytes.NewReader(data)
			w, ctx := createTestRequest(body)

			handler.Patch(ctx)
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
})
