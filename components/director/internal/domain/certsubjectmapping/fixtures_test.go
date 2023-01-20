package certsubjectmapping_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
)

const (
	TestID                 = "455e47ea-5eab-49c5-ba35-a67e1d9125f6"
	TestSubject            = "C=DE, L=test, O=SAP SE, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass"
	TestConsumerType       = "TestConsumerType"
)

var (
	TestInternalConsumerID = "c889e656-2623-4828-916c-c9d46fafac7c"
	TestTenantAccessLevelsAsString        = "[\"global\",\"account\"]"
	TestInvalidTenantAccessLevelsAsString = "[invalid"
	TestTenantAccessLevels                = []string{"global", "account"}
	nilModelEntity                        *model.CertSubjectMapping

	CertSubjectMappingModel = &model.CertSubjectMapping{
		ID:                 TestID,
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingGQLModel = &graphql.CertificateSubjectMapping{
		ID:                 TestID,
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingGQLModelInput = graphql.CertificateSubjectMappingInput{
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingGQLInvalidInput = graphql.CertificateSubjectMappingInput{
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevels,
	}

	CertSubjectMappingEntity = &certsubjectmapping.Entity{
		ID:                 TestID,
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestTenantAccessLevelsAsString,
	}

	CertSubjectMappingEntityInvalidTntAccessLevels = &certsubjectmapping.Entity{
		ID:                 TestID,
		Subject:            TestSubject,
		ConsumerType:       TestConsumerType,
		InternalConsumerID: &TestInternalConsumerID,
		TenantAccessLevels: TestInvalidTenantAccessLevelsAsString,
	}

	CertificateSubjectMappingModelPage = &model.CertSubjectMappingPage{
		Data:       []*model.CertSubjectMapping{CertSubjectMappingModel},
		PageInfo:   &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 0,
	}

	CertificateSubjectMappingsGQL = []*graphql.CertificateSubjectMapping{CertSubjectMappingGQLModel}

	CertificateSubjectMappingGQLPage = &graphql.CertificateSubjectMappingPage{
		Data:       CertificateSubjectMappingsGQL,
		PageInfo:   &graphql.PageInfo{
			StartCursor: graphql.PageCursor("start"),
			EndCursor:   graphql.PageCursor("end"),
			HasNextPage: false,
		},
		TotalCount: 0,
	}
)

func fixColumns() []string {
	return []string{"id", "subject", "consumer_type", "internal_consumer_id", "tenant_access_levels"}
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
