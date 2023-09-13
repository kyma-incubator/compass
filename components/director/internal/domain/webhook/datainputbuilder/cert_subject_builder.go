package datainputbuilder

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=certSubjectRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type certSubjectRepository interface {
	ListByConsumerID(ctx context.Context, consumerID string) ([]*model.CertSubjectMapping, error)
}

// WebhookCertSubjectBuilder takes care to get and build trust details for objects
type WebhookCertSubjectBuilder struct {
	certSubjectRepository certSubjectRepository
}

// NewWebhookCertSubjectBuilder creates a WebhookCertSubjectBuilder
func NewWebhookCertSubjectBuilder(certSubjectRepository certSubjectRepository) *WebhookCertSubjectBuilder {
	return &WebhookCertSubjectBuilder{
		certSubjectRepository: certSubjectRepository,
	}
}

// GetTrustDetailsForObjects builds trust details for objects with IDs objectIDs
func (b *WebhookCertSubjectBuilder) GetTrustDetailsForObjects(ctx context.Context, objectIDs []string) (map[string]*webhook.TrustDetails, error) {
	if len(objectIDs) == 0 {
		return nil, nil
	}

	result := make(map[string]*webhook.TrustDetails, len(objectIDs))

	for _, objectID := range objectIDs {
		certSubjectMappingsForObject, err := b.certSubjectRepository.ListByConsumerID(ctx, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while listing cert subject mappings for consumer with ID %q", objectID)
		}

		result[objectID] = &webhook.TrustDetails{Subjects: certSubjectMappingsToCertificateDetails(certSubjectMappingsForObject)}
	}

	return result, nil
}

// GetTrustDetailsForObject builds trust details for objects with ID objectID
func (b *WebhookCertSubjectBuilder) GetTrustDetailsForObject(ctx context.Context, objectID string) (*webhook.TrustDetails, error) {
	trustDetails, err := b.GetTrustDetailsForObjects(ctx, []string{objectID})
	if err != nil {
		return nil, err
	}

	result, ok := trustDetails[objectID]
	if !ok {
		return nil, errors.Errorf("There are no trust details for object with ID %q", objectID)
	}

	return result, nil
}

func certSubjectMappingsToCertificateDetails(certSubjectMappings []*model.CertSubjectMapping) []string {
	result := make([]string, 0, len(certSubjectMappings))

	for _, certSubjectMapping := range certSubjectMappings {
		result = append(result, certSubjectMapping.Subject)
	}

	return result
}
