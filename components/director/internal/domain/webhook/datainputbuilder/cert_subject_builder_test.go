package datainputbuilder_test

import (
	"testing"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	certSubjectMappingsForApplication = []*model.CertSubjectMapping{{
		ID:                 "id-1",
		Subject:            "subject-1",
		InternalConsumerID: str.Ptr(ApplicationID),
	}, {
		ID:                 "id-2",
		Subject:            "subject-2",
		InternalConsumerID: str.Ptr(ApplicationID),
	}}

	certSubjectMappingsForApplication2 = []*model.CertSubjectMapping{{
		ID:                 "id-3",
		Subject:            "subject-3",
		InternalConsumerID: str.Ptr(Application2ID),
	}}

	expectedTrustDetailsForApplications = map[string]*webhook.TrustDetails{
		ApplicationID: {
			Subjects: []string{certSubjectMappingsForApplication[0].Subject, certSubjectMappingsForApplication[1].Subject},
		},
		Application2ID: {
			Subjects: []string{certSubjectMappingsForApplication2[0].Subject},
		},
	}
)

func TestWebhookLabelBuilder_GetTrustDetailsForObjects(t *testing.T) {
	testCases := []struct {
		name                 string
		certSubjectRepo      func() *automock.CertSubjectRepository
		objectIDs            []string
		expectedTrustDetails map[string]*webhook.TrustDetails
		expectedErrMsg       string
	}{
		{
			name: "success",
			certSubjectRepo: func() *automock.CertSubjectRepository {
				repo := &automock.CertSubjectRepository{}
				repo.On("ListByConsumerID", emptyCtx, ApplicationID).Return(certSubjectMappingsForApplication, nil).Once()
				repo.On("ListByConsumerID", emptyCtx, Application2ID).Return(certSubjectMappingsForApplication2, nil).Once()
				return repo
			},
			objectIDs:            []string{ApplicationID, Application2ID},
			expectedTrustDetails: expectedTrustDetailsForApplications,
		},
		{
			name:      "success when there are no object ids",
			objectIDs: []string{},
		},
		{
			name: "returns error when can't list cert subject mappings by consumer id",
			certSubjectRepo: func() *automock.CertSubjectRepository {
				repo := &automock.CertSubjectRepository{}
				repo.On("ListByConsumerID", emptyCtx, ApplicationID).Return(nil, testErr).Once()
				return repo
			},
			objectIDs:      []string{ApplicationID},
			expectedErrMsg: "while listing cert subject mappings",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			certSubjectRepo := &automock.CertSubjectRepository{}
			if tCase.certSubjectRepo != nil {
				certSubjectRepo = tCase.certSubjectRepo()
			}
			defer mock.AssertExpectationsForObjects(t, certSubjectRepo)

			webhookDataInputBuilder := databuilder.NewWebhookCertSubjectBuilder(certSubjectRepo)

			// WHEN
			resultTrustDetails, err := webhookDataInputBuilder.GetTrustDetailsForObjects(emptyCtx, tCase.objectIDs)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, resultTrustDetails)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTrustDetails, resultTrustDetails)
			}
		})
	}
}

func TestWebhookLabelBuilder_GetTrustDetailsForObject(t *testing.T) {
	testCases := []struct {
		name                 string
		certSubjectRepo      func() *automock.CertSubjectRepository
		objectID             string
		expectedTrustDetails *webhook.TrustDetails
		expectedErrMsg       string
	}{
		{
			name: "success",
			certSubjectRepo: func() *automock.CertSubjectRepository {
				repo := &automock.CertSubjectRepository{}
				repo.On("ListByConsumerID", emptyCtx, ApplicationID).Return(certSubjectMappingsForApplication, nil).Once()
				return repo
			},
			objectID:             ApplicationID,
			expectedTrustDetails: expectedTrustDetailsForApplications[ApplicationID],
		},
		{
			name: "returns error when can't list cert subject mappings by consumer id",
			certSubjectRepo: func() *automock.CertSubjectRepository {
				repo := &automock.CertSubjectRepository{}
				repo.On("ListByConsumerID", emptyCtx, ApplicationID).Return(nil, testErr).Once()
				return repo
			},
			objectID:       ApplicationID,
			expectedErrMsg: "while listing cert subject mappings",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			certSubjectRepo := &automock.CertSubjectRepository{}
			if tCase.certSubjectRepo != nil {
				certSubjectRepo = tCase.certSubjectRepo()
			}
			defer mock.AssertExpectationsForObjects(t, certSubjectRepo)

			webhookDataInputBuilder := databuilder.NewWebhookCertSubjectBuilder(certSubjectRepo)

			// WHEN
			resultTrustDetails, err := webhookDataInputBuilder.GetTrustDetailsForObject(emptyCtx, tCase.objectID)

			// THEN
			if tCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tCase.expectedErrMsg)
				require.Nil(t, resultTrustDetails)
			} else {
				require.NoError(t, err)
				require.Equal(t, tCase.expectedTrustDetails, resultTrustDetails)
			}
		})
	}
}
