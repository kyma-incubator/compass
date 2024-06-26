package certsubjectmapping_test

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

const (
	TestID            = "455e47ea-5eab-49c5-ba35-a67e1d9125f6"
	TestSubject       = "C=DE, L=test, O=SAP SE, OU=TestRegion, OU=Test Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass"
	TestSubjectSorted = "CN=test-compass, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, OU=Test Cloud Platform Clients, OU=TestRegion, O=SAP SE, L=test, C=DE"
	TestConsumerType  = string(consumer.Runtime)
)

var (
	TestInternalConsumerID                = "c889e656-2623-4828-916c-c9d46fafac7c"
	TestTenantAccessLevelsAsString        = "[\"account\",\"subaccount\"]"
	TestInvalidTenantAccessLevelsAsString = "[invalid"
	TestTenantAccessLevels                = []string{string(tenantEntity.Account), string(tenantEntity.Subaccount)}
	nilModelEntity                        *model.CertSubjectMapping
	testTime                              = time.Date(2024, 04, 24, 9, 9, 9, 9, time.Local)

	CertSubjectMappingEntity                       = fixCertSubjectMappingEntity(TestID, TestSubject, TestConsumerType, TestInternalConsumerID, TestTenantAccessLevelsAsString, testTime)
	CertSubjectMappingEntityInvalidTntAccessLevels = fixCertSubjectMappingEntity(TestID, TestSubject, TestConsumerType, TestInternalConsumerID, TestInvalidTenantAccessLevelsAsString, testTime)
	CertSubjectMappingModel                        = fixCertSubjectMappingModel(TestID, TestSubject, TestConsumerType, TestInternalConsumerID, TestTenantAccessLevels, testTime)

	CertSubjectMappingGQLModel = &graphql.CertificateSubjectMapping{
		ID:                 TestID,
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
		CreatedAt:          graphql.Timestamp(testTime),
	}

	CertSubjectMappingGQLModelInput = graphql.CertificateSubjectMappingInput{
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingGQLModelInputWithSortedSubject = graphql.CertificateSubjectMappingInput{
		Subject:            TestSubjectSorted,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingGQLInvalidInput = graphql.CertificateSubjectMappingInput{
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertificateSubjectMappingModelPage = &model.CertSubjectMappingPage{
		Data: []*model.CertSubjectMapping{CertSubjectMappingModel},
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 0,
	}

	CertificateSubjectMappingsGQL = []*graphql.CertificateSubjectMapping{CertSubjectMappingGQLModel}

	CertificateSubjectMappingGQLPage = &graphql.CertificateSubjectMappingPage{
		Data: CertificateSubjectMappingsGQL,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor("start"),
			EndCursor:   graphql.PageCursor("end"),
			HasNextPage: false,
		},
		TotalCount: 0,
	}
)

func fixCertSubjectMappingEntity(ID, subject, consumerType, internalConsumerID, tenantAccessLevels string, createdAt time.Time) *certsubjectmapping.Entity {
	return &certsubjectmapping.Entity{
		ID:                 ID,
		Subject:            subject,
		ConsumerType:       consumerType,
		InternalConsumerID: &internalConsumerID,
		TenantAccessLevels: tenantAccessLevels,
		CreatedAt:          createdAt,
	}
}

func fixCertSubjectMappingModel(ID, subject, consumerType, internalConsumerID string, tenantAccessLevels []string, createdAt time.Time) *model.CertSubjectMapping {
	return &model.CertSubjectMapping{
		ID:                 ID,
		Subject:            subject,
		ConsumerType:       consumerType,
		InternalConsumerID: &internalConsumerID,
		TenantAccessLevels: tenantAccessLevels,
		CreatedAt:          createdAt,
	}
}

func fixColumns() []string {
	return []string{"id", "subject", "consumer_type", "internal_consumer_id", "tenant_access_levels", "created_at", "updated_at"}
}

func fixUnusedCertSubjectMappingRepository() *automock.CertMappingRepository {
	return &automock.CertMappingRepository{}
}

func fixUnusedUIDService() *automock.UIDService {
	return &automock.UIDService{}
}

func fixUnusedTransactioner() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
	return &persistenceautomock.PersistenceTx{}, &persistenceautomock.Transactioner{}
}

func fixUnusedCertSubjectMappingSvc() *automock.CertSubjectMappingService {
	return &automock.CertSubjectMappingService{}
}

func fixUnusedConverter() *automock.Converter {
	return &automock.Converter{}
}
