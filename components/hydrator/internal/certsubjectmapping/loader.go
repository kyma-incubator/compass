package certsubjectmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"time"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// DirectorClient is a GraphQL client used to communicate with the director component and the DB
//
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
	InitialiseCertSubjectMappings(ctx context.Context, certSubjectMappingsFromEnv string) error
}

type certSubjectMappingLoader struct {
	certSubjectMappingCache *certSubjectMappingCache
	certSubjectMappingCfg   Config
	directorClient          DirectorClient
}

var (
	CertSubjectMappingLoaderCorrelationID = "cert-subject-mapping-loader-correlation-id"
	CertSubjectMappingRetryInterval       = 50 * time.Millisecond
)

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

	err := certSubjectLoader.InitialiseCertSubjectMappings(ctx, certSubjectMappingCfg.EnvironmentMappings)
	if err != nil {
		return nil, err
	}

	go certSubjectLoader.Run(ctx, certSubjectMappingCfg.EnvironmentMappings)

	return cache, nil
}

func (cl *certSubjectMappingLoader) InitialiseCertSubjectMappings(ctx context.Context, certSubjectMappingsFromEnv string) error {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, CertSubjectMappingLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	var mappings []SubjectConsumerTypeMapping
	var err error

	err = retry.Do(func() error {
		mappings, err = cl.loadCertSubjectMappings(ctx, certSubjectMappingsFromEnv)
		if err != nil {
			return err
		}
		return nil
	},
		retry.Attempts(0), // we want the default value here
		retry.Delay(CertSubjectMappingRetryInterval),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			log.C(ctx).Infof("Certificate subject mapping resync failed. Retrying request attempt (%d) after error %v", n, err)
		}))

	if err != nil {
		return errors.Errorf("Certificate subject mapping resync failed with error: %v", err)
	}

	log.C(ctx).Info("Update certificate subject mapping cache with the newly fetched data")
	cl.certSubjectMappingCache.Put(mappings)

	return nil
}

func (cl *certSubjectMappingLoader) Run(ctx context.Context, certSubjectMappingsFromEnv string) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, CertSubjectMappingLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	t := time.NewTicker(cl.certSubjectMappingCfg.ResyncInterval)
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Infof("Context cancelled, stopping certificate subject mapping resyncer...")
			t.Stop()
			return
		case <-t.C:
			mappings, err := cl.loadCertSubjectMappings(ctx, certSubjectMappingsFromEnv)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Certificate subject mapping resync failed with error: %v", err)
			} else {
				log.C(ctx).Info("Update certificate subject mapping cache with the newly fetched data")
				cl.certSubjectMappingCache.Put(mappings)
			}
		}
	}
}

func (cl *certSubjectMappingLoader) loadCertSubjectMappings(ctx context.Context, certSubjectMappingsFromEnv string) ([]SubjectConsumerTypeMapping, error) {
	after := ""
	mappings := make([]SubjectConsumerTypeMapping, 0)
	hasNextPage := true
	csmTotalCount := 0
	log.C(ctx).Infof("Getting certificate subject mapping(s) from environment.")
	mappingsFromEnv, err := unmarshalMappings(certSubjectMappingsFromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "while getting certificate subject mappings from environment")
	}
	log.C(ctx).Infof("Certificate subject mapping(s) count from environment: %d", len(mappingsFromEnv))

	mappings = append(mappings, mappingsFromEnv...)
	log.C(ctx).Info("Listing certificate subject mapping from DB...")
	for hasNextPage == true {
		csmGQLPage, err := cl.directorClient.ListCertificateSubjectMappings(ctx, after)
		if err != nil {
			return mappings, errors.Wrap(err, "while listing certificate subject mappings from DB")
		}
		csmTotalCount = csmGQLPage.TotalCount
		mappings = append(mappings, convertGQLCertSubjectMappings(csmGQLPage.Data)...)
		hasNextPage = csmGQLPage.PageInfo.HasNextPage
		after = string(csmGQLPage.PageInfo.EndCursor)
	}
	log.C(ctx).Infof("Certificate subject mapping(s) count from DB: %d", csmTotalCount)

	return mappings, nil
}

func convertGQLCertSubjectMappings(gqlMappings []*schema.CertificateSubjectMapping) []SubjectConsumerTypeMapping {
	m := make([]SubjectConsumerTypeMapping, 0, len(gqlMappings))
	var internalConsumerID string
	for _, e := range gqlMappings {
		if e.InternalConsumerID != nil {
			internalConsumerID = *e.InternalConsumerID
		}
		scm := SubjectConsumerTypeMapping{
			Subject:            e.Subject,
			ConsumerType:       e.ConsumerType,
			InternalConsumerID: internalConsumerID,
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
