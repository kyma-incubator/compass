package lms

import (
	"time"

	"sync"

	"errors"

	"crypto/x509/pkix"

	"github.com/google/uuid"
)

const (
	FakeCaCertificate     = "cert-ca-payload"
	FakeSignedCertificate = "signed-cert-payload"
	FakeLmsHost           = "lms.localhost"
	FakePrivateKey        = "private-key"
)

// FakeClient implements the lms client interface but does not call real external system
type FakeClient struct {
	mu   sync.Mutex
	data map[string]tenantInfo

	timeToReady    time.Duration
	requestedCerts map[string]struct{}
}

// NewFakeClient creates lms fake client which response tenant ready after timeToReady duration
func NewFakeClient(timeToReady time.Duration) *FakeClient {
	return &FakeClient{
		data:           make(map[string]tenantInfo, 0),
		timeToReady:    timeToReady,
		requestedCerts: make(map[string]struct{}, 0),
	}
}

type tenantInfo struct {
	createdAt time.Time
}

func (f *FakeClient) CreateTenant(input CreateTenantInput) (o CreateTenantOutput, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, _ := uuid.NewRandom()

	f.data[id.String()] = tenantInfo{
		createdAt: time.Now(),
	}

	return CreateTenantOutput{
		ID: id.String(),
	}, nil
}

func (f *FakeClient) GetTenantStatus(tenantID string) (status TenantStatus, err error) {
	ti, found := f.data[tenantID]
	if !found {
		return TenantStatus{}, errors.New("tenant not exists")
	}

	if time.Since(ti.createdAt) > f.timeToReady {
		return TenantStatus{KibanaDNSResolves: true, ElasticsearchDNSResolves: true}, nil
	} else {
		return TenantStatus{KibanaDNSResolves: false, ElasticsearchDNSResolves: false}, nil
	}
}

func (f *FakeClient) GetTenantInfo(tenantID string) (status TenantInfo, err error) {
	_, found := f.data[tenantID]
	if !found {
		return TenantInfo{}, errors.New("tenant not exists")
	}
	return TenantInfo{DNS: FakeLmsHost}, nil
}

func (f *FakeClient) GetCACertificate(tenantID string) (cert string, found bool, err error) {
	if !f.IsCertRequestedForTenant(tenantID) {
		return "", false, errors.New("certificate not requested")
	}
	return FakeCaCertificate, true, nil
}

func (f *FakeClient) GetSignedCertificate(tenantID string, certID string) (cert string, found bool, err error) {
	if !f.IsCertRequestedForTenant(tenantID) {
		return "", false, errors.New("certificate not requested")
	}
	return FakeSignedCertificate, true, nil
}

func (f *FakeClient) RequestCertificate(tenantID string, subj pkix.Name) (id string, privateKey []byte, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requestedCerts[tenantID] = struct{}{}
	return "id-001", []byte(FakePrivateKey), nil
}

// assert methods
func (f *FakeClient) IsCertRequestedForTenant(tenantID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, found := f.requestedCerts[tenantID]
	return found
}
