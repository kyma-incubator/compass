package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers/automock"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Health Handler", func() {
	When("Storage connection is not ok", func() {
		It("Should return 500 and report storage status Down", func() {
			service := &automock.HealthService{}
			service.On("CheckHealth", mock.Anything).
				Return(types.HealthStatus{Storage: types.StatusDown}, errors.New("ping failed"))
			handler := HealthHandler{
				Service: service,
			}

			w, ctx := createTestRequest(nil)

			handler.Health(ctx)
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(`{"storageStatus":"Down"}`))
		})
	})
	When("Storage connection is ok", func() {
		It("Should return 200 and report storage status Up", func() {
			service := &automock.HealthService{}
			service.On("CheckHealth", mock.Anything).
				Return(types.HealthStatus{Storage: types.StatusUp}, nil)
			handler := HealthHandler{
				Service: service,
			}

			w, ctx := createTestRequest(nil)

			handler.Health(ctx)
			Expect(w.Code).To(Equal(http.StatusOK))
			responseBody, err := io.ReadAll(w.Body)
			Expect(err).Error().ToNot(HaveOccurred())
			Expect(responseBody).To(ContainSubstring(`{"storageStatus":"Up"}`))
		})
	})
})
