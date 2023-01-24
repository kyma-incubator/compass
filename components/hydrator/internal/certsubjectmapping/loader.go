package certsubjectmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"time"
)

// DirectorClient is a GraphQL client used to communicate with the director component and the DB
//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorClient interface {
	ListCertificateSubjectMappings(ctx context.Context, after string) (*schema.CertificateSubjectMappingPage, error)
}

type SubjectConsumerTypeMapping struct {
	Subject            string   `json:"subject"`
	ConsumerType       string   `json:"consumer_type"`
	InternalConsumerID string   `json:"internal_consumer_id"`
	TenantAccessLevels []string `json:"tenant_access_levels"`
}

// Loader provide mechanism to load certificate subject mappings' data into in-memory storage
type Loader interface {
	Run(ctx context.Context, certSubjectMappingsFromEnv string)
}

type certSubjectMappingLoader struct {
	certSubjectMappingCache *certSubjectMappingCache
	certSubjectMappingCfg   Config
	directorClient          DirectorClient
}

var certSubjectMappingLoaderCorrelationID = "cert-subject-mapping-loader-correlation-id"

func NewCertSubjectMappingLoader(certSubjectMappingCache *certSubjectMappingCache, certSubjectMappingCfg Config, directorClient DirectorClient) Loader {
	return &certSubjectMappingLoader{
		certSubjectMappingCache: certSubjectMappingCache,
		certSubjectMappingCfg:   certSubjectMappingCfg,
		directorClient:          directorClient,
	}
}

func (s *SubjectConsumerTypeMapping) Validate() error {
	if len(s.Subject) < 1 {
		return errors.New("subject is not provided")
	}

	if !inputvalidation.SupportedConsumerTypes[s.ConsumerType] {
		return fmt.Errorf("consumer type %s is not valid", s.ConsumerType)
	}

	for _, al := range s.TenantAccessLevels {
		if !inputvalidation.SupportedAccessLevels[al] {
			return fmt.Errorf("tenant access level %s is not valid", al)
		}
	}

	return nil
}

func StartCertSubjectMappingLoader(ctx context.Context, certSubjectMappingCfg Config, directorClient DirectorClient) (Cache, error) {
	cache := NewCertSubjectMappingCache()
	certSubjectLoader := NewCertSubjectMappingLoader(cache, certSubjectMappingCfg, directorClient)
	go certSubjectLoader.Run(ctx, certSubjectMappingCfg.environmentMappings)

	return cache, nil
}

func (cl *certSubjectMappingLoader) Run(ctx context.Context, certSubjectMappingsFromEnv string) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certSubjectMappingLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	t := time.NewTicker(cl.certSubjectMappingCfg.resyncInterval)
	for {
		select {
		case <-t.C:
			mappings, err := cl.loadCertSubjectMappings(ctx, certSubjectMappingsFromEnv)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Certificate subject mapping resync failed with error: %v", err)
				continue
			}
			log.C(ctx).Info("Update certificate subject mapping cache with the newly fetched data")
			cl.certSubjectMappingCache.Put(mappings)
			// todo::: remove
			spew.Dump(mappings)
		case <-ctx.Done():
			log.C(ctx).Infof("Context cancelled, stopping certificate subject mapping resyncer...")
			t.Stop()
			return
		}
	}
}

func (cl *certSubjectMappingLoader) loadCertSubjectMappings(ctx context.Context, certSubjectMappingsFromEnv string) ([]SubjectConsumerTypeMapping, error) {
	log.C(ctx).Info("Listing certificate subject mapping from DB...")
	after := ""
	certSubjectMappingsGQLPage, err := cl.directorClient.ListCertificateSubjectMappings(ctx, after)
	if err != nil {
		return nil, errors.Wrap(err, "while listing certificate subject mappings")
	}

	log.C(ctx).Infof("Total count of fetched certificate subject mappings from the DB: %d", certSubjectMappingsGQLPage.TotalCount)
	mappings := make([]SubjectConsumerTypeMapping, 0, certSubjectMappingsGQLPage.TotalCount)

	mappings = append(mappings, convertGQLCertSubjectMappings(certSubjectMappingsGQLPage.Data)...)

	hasNextPage := certSubjectMappingsGQLPage.PageInfo.HasNextPage
	after = string(certSubjectMappingsGQLPage.PageInfo.EndCursor)
	for hasNextPage == true {
		csmGQLPage, err := cl.directorClient.ListCertificateSubjectMappings(ctx, after)
		if err != nil {
			return nil, errors.Wrap(err, "while listing certificate subject mappings")
		}
		mappings = append(mappings, convertGQLCertSubjectMappings(csmGQLPage.Data)...)
		hasNextPage = csmGQLPage.PageInfo.HasNextPage
		after = string(csmGQLPage.PageInfo.EndCursor)
	}

	mappingsFromEnv, err := unmarshalMappings(certSubjectMappingsFromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "while getting certificate subject mappings from environment")
	}

	mappings = append(mappings, mappingsFromEnv...)

	return mappings, nil
}

func convertGQLCertSubjectMappings(gqlMappings []*schema.CertificateSubjectMapping) []SubjectConsumerTypeMapping {
	m := make([]SubjectConsumerTypeMapping, 0, len(gqlMappings))
	for _, e := range gqlMappings {
		scm := SubjectConsumerTypeMapping{
			Subject:            e.Subject,
			ConsumerType:       e.ConsumerType,
			InternalConsumerID: *e.InternalConsumerID,
			TenantAccessLevels: e.TenantAccessLevels,
		}
		m = append(m, scm)
	}
	return m
}

func unmarshalMappings(certSubjectMappingsFromEnv string) ([]SubjectConsumerTypeMapping, error) {
	var mappingsFromEnv []SubjectConsumerTypeMapping
	if err := json.Unmarshal([]byte(certSubjectMappingsFromEnv), &mappingsFromEnv); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling mappings")
	}

	return mappingsFromEnv, nil
}
