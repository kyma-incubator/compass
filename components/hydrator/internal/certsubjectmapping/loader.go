package certsubjectmapping

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/certsubjmapping"

	"github.com/avast/retry-go/v4"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// DirectorClient is a GraphQL client used to communicate with the director component and the DB
//
//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorClient interface {
	ListCertificateSubjectMappings(ctx context.Context, after string) (*schema.CertificateSubjectMappingPage, error)
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
	CertSubjectMappingLoaderCorrelationID        = "cert-subject-mapping-loader-correlation-id"
	CertSubjectMappingInitialLoaderCorrelationID = "cert-subject-mapping-initial-loader-correlation-id"
	CertSubjectMappingRetryInterval              = 50 * time.Millisecond
)

func NewCertSubjectMappingLoader(certSubjectMappingCache *certSubjectMappingCache, certSubjectMappingCfg Config, directorClient DirectorClient) Loader {
	return &certSubjectMappingLoader{
		certSubjectMappingCache: certSubjectMappingCache,
		certSubjectMappingCfg:   certSubjectMappingCfg,
		directorClient:          directorClient,
	}
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
	entry = entry.WithField(log.FieldRequestID, CertSubjectMappingInitialLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	var mappings []certsubjmapping.SubjectConsumerTypeMapping
	var err error

	err = retry.Do(func() error {
		mappings, err = cl.loadCertSubjectMappings(ctx, certSubjectMappingsFromEnv)
		if err != nil {
			return err
		}
		return nil
	},
		retry.Attempts(0), // we want to try until the call succeeds; if it keeps failing and failing, the pod will be stuck, and we leave the decision when to terminate it to kubernetes.
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

func (cl *certSubjectMappingLoader) loadCertSubjectMappings(ctx context.Context, certSubjectMappingsFromEnv string) ([]certsubjmapping.SubjectConsumerTypeMapping, error) {
	after := ""
	mappings := make([]certsubjmapping.SubjectConsumerTypeMapping, 0)
	hasNextPage := true
	csmTotalCount := 0
	log.C(ctx).Infof("Getting certificate subject mapping(s) from environment.")
	mappingsFromEnv, err := certsubjmapping.UnmarshalMappings(certSubjectMappingsFromEnv)
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

		// The graphql response could be nil, and there could be NO graphql error in case something else failed.
		// E.g., an error from the oathkeeper or other components
		if csmGQLPage == nil {
			return mappings, errors.Errorf("the certificate subject mappings page cannot be nil")
		}
		csmTotalCount = csmGQLPage.TotalCount
		mappings = append(mappings, convertGQLCertSubjectMappings(csmGQLPage.Data)...)

		if csmGQLPage.PageInfo == nil {
			return mappings, errors.Errorf("the certificate subject mappings page info cannot be nil")
		}
		hasNextPage = csmGQLPage.PageInfo.HasNextPage
		after = string(csmGQLPage.PageInfo.EndCursor)
	}
	log.C(ctx).Infof("Certificate subject mapping(s) count from DB: %d", csmTotalCount)

	return mappings, nil
}

func convertGQLCertSubjectMappings(gqlMappings []*schema.CertificateSubjectMapping) []certsubjmapping.SubjectConsumerTypeMapping {
	m := make([]certsubjmapping.SubjectConsumerTypeMapping, 0, len(gqlMappings))
	for _, e := range gqlMappings {
		var internalConsumerID string
		if e.InternalConsumerID != nil {
			internalConsumerID = *e.InternalConsumerID
		}
		scm := certsubjmapping.SubjectConsumerTypeMapping{
			Subject:            e.Subject,
			ConsumerType:       e.ConsumerType,
			InternalConsumerID: internalConsumerID,
			TenantAccessLevels: e.TenantAccessLevels,
		}
		m = append(m, scm)
	}
	return m
}
