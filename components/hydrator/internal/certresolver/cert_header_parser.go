package certresolver

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

//go:generate mockery --name=CertificateHeaderParser --output=automock --outpkg=automock --case=underscore
type CertificateHeaderParser interface {
	GetCertificateData(*http.Request) *CertificateData
	GetIssuer() string
}

type CertificateData struct {
	ClientID         string
	CertificateHash  string
	AuthSessionExtra map[string]interface{}
}

type certificateInfo struct {
	Hash    string
	Subject string
}

type headerParser struct {
	certHeaderName                 string
	issuer                         string
	isSubjectMatching              func(subject string) bool
	getClientIDFromSubject         func(subject string) string
	getAuthSessionExtraFromSubject func(ctx context.Context, subject string) map[string]interface{}
}

func NewHeaderParser(certHeaderName, issuer string, isSubjectMatching func(subject string) bool, getClientIDFromSubject func(subject string) string, getAuthSessionExtraFromSubject func(ctx context.Context, subject string) map[string]interface{}) *headerParser {
	return &headerParser{
		certHeaderName:                 certHeaderName,
		issuer:                         issuer,
		isSubjectMatching:              isSubjectMatching,
		getClientIDFromSubject:         getClientIDFromSubject,
		getAuthSessionExtraFromSubject: getAuthSessionExtraFromSubject,
	}
}

func (hp *headerParser) GetCertificateData(r *http.Request) *CertificateData {
	certHeader := r.Header.Get(hp.certHeaderName)
	if certHeader == "" {
		return nil
	}

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)
	hashRegex := regexp.MustCompile(`Hash=([0-9a-f]*)`)

	subjects := extractFromHeader(certHeader, subjectRegex)
	hashes := extractFromHeader(certHeader, hashRegex)

	certificateInfos := createCertInfos(subjects, hashes)

	log.C(r.Context()).Debugf("Trying to match certificate subjects [%s] for issuer %s", strings.Join(subjects, ","), hp.GetIssuer())

	certificateInfo, found := hp.getCertificateInfoWithMatchingSubject(certificateInfos)
	if !found {
		return nil
	}

	certData := &CertificateData{
		ClientID:         hp.getClientIDFromSubject(certificateInfo.Subject),
		CertificateHash:  certificateInfo.Hash,
		AuthSessionExtra: hp.getAuthSessionExtraFromSubject(r.Context(), certificateInfo.Subject),
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
		if hp.isSubjectMatching(info.Subject) {
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
