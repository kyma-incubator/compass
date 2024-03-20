package certresolver

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

//go:generate mockery --name=CertificateHeaderParser --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertificateHeaderParser interface {
	GetCertificateData(*http.Request) *CertificateData
	GetIssuer() string
}

type CertificateData struct {
	ClientID         string
	CertificateHash  string
	AuthSessionExtra map[string]interface{}
	Subject          string
}

type certificateInfo struct {
	Hash    string
	Subject string
}

type subjectMatcherFunc func(subject string) bool
type clientIDRetrieverFunc func(subject string) string
type authSessionExtraRetrieverFunc func(ctx context.Context, subject string) map[string]interface{}

type headerParser struct {
	certHeaderName              string
	issuer                      string
	subjectMatcherFn            subjectMatcherFunc
	clientIDRetrieverFn         clientIDRetrieverFunc
	authSessionExtraRetrieverFn authSessionExtraRetrieverFunc
}

func NewHeaderParser(certHeaderName, issuer string, subjectMatcherFn subjectMatcherFunc, clientIDRetrieverFn clientIDRetrieverFunc, authSessionExtraRetrieverFn authSessionExtraRetrieverFunc) *headerParser {
	return &headerParser{
		certHeaderName:              certHeaderName,
		issuer:                      issuer,
		subjectMatcherFn:            subjectMatcherFn,
		clientIDRetrieverFn:         clientIDRetrieverFn,
		authSessionExtraRetrieverFn: authSessionExtraRetrieverFn,
	}
}

func (hp *headerParser) GetCertificateData(r *http.Request) *CertificateData {
	ctx := r.Context()

	certHeader := r.Header.Get(hp.certHeaderName)
	if certHeader == "" {
		return nil
	}

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)
	hashRegex := regexp.MustCompile(`Hash=([0-9a-f]*)`)

	subjects := extractFromHeader(certHeader, subjectRegex)
	hashes := extractFromHeader(certHeader, hashRegex)
	fmt.Println("ALEX GetCertificateData subjects", subjects)
	certificateInfos := createCertInfos(subjects, hashes)

	log.C(ctx).Debugf("Trying to match certificate subjects [%s] for issuer %s", strings.Join(subjects, ","), hp.GetIssuer())

	certificateInfo, found := hp.getCertificateInfoWithMatchingSubject(certificateInfos)
	if !found {
		return nil
	}

	fmt.Printf("ALEX GetCertificateData subject - %s", certificateInfo.Subject)
	fmt.Printf("ALEX GetCertificateData clientID - %s", hp.clientIDRetrieverFn(certificateInfo.Subject))
	fmt.Printf("ALEX GetCertificateData AuthSessionExtra - %+v", hp.authSessionExtraRetrieverFn(ctx, certificateInfo.Subject))

	certData := &CertificateData{
		ClientID:         hp.clientIDRetrieverFn(certificateInfo.Subject),
		CertificateHash:  certificateInfo.Hash,
		AuthSessionExtra: hp.authSessionExtraRetrieverFn(ctx, certificateInfo.Subject),
		Subject:          certificateInfo.Subject,
	}

	return certData
}

func (hp *headerParser) GetIssuer() string {
	return hp.issuer
}

func createCertInfos(subjects, hashes []string) []certificateInfo {
	if len(subjects) != len(hashes) {
		return []certificateInfo{}
	}

	certInfos := make([]certificateInfo, len(subjects))
	for i := 0; i < len(subjects); i++ {
		certInfo := newCertificateInfo(subjects[i], hashes[i])
		certInfos[i] = certInfo
	}
	return certInfos
}

func (hp *headerParser) getCertificateInfoWithMatchingSubject(infos []certificateInfo) (certificateInfo, bool) {
	for _, info := range infos {
		fmt.Println("ALEX cert_header_parser - subject", info.Subject, hp.subjectMatcherFn(info.Subject))
		if hp.subjectMatcherFn(info.Subject) {
			return info, true
		}
	}
	return certificateInfo{}, false
}

func newCertificateInfo(subject, hash string) certificateInfo {
	certInfo := certificateInfo{
		Hash:    hash,
		Subject: subject,
	}
	return certInfo
}

func extractFromHeader(certHeader string, regex *regexp.Regexp) []string {
	var matchedStrings []string

	matches := regex.FindAllStringSubmatch(certHeader, -1)

	for _, match := range matches {
		value := get(match, 1)
		matchedStrings = append(matchedStrings, value)
	}

	return matchedStrings
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}
