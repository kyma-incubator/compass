package auth_test

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	directorjwt "github.com/kyma-incubator/compass/components/director/pkg/jwt"
	"github.com/stretchr/testify/suite"
)

var directorJwtConfig = directorjwt.Config{
	ExpireAfter: 1000,
}

func TestSelfSignedJwtTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(SelfSignedJwtTokenAuthorizationProviderTestSuite))
}

type SelfSignedJwtTokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_New() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorjwt.Config{})
	suite.Require().NotNil(provider)
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_Name() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorjwt.Config{})

	name := provider.Name()

	suite.Require().Equal(name, "SelfSignedTokenAuthorizationProvider")
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_Matches() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorjwt.Config{})

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.SelfSignedTokenCredentials{}))
	suite.Require().Equal(matches, true)
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_DoesNotMatchWhenBasicCredentialsInContext() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorjwt.Config{})

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))
	suite.Require().Equal(matches, false)
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_DoesNotMatchNoCredentialsInContext() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorjwt.Config{})

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, false)
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_GetAuthorization() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorJwtConfig)

	secretName := "system-fetcher-external-keys"

	encodedPrivateKey := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQ1h4TmJrY1lkYkNxUFMKOHRVeUxpYlhTVUx1M0ZPNTY4VWZGbjlYVGNIbktDbVJnUzhiMERoZWgzbDc2emhleXZZWURENS9LTnJUeWRKcgpCSzdVVWVzRFowaXdMeDdvQTR1bEppSmdVaSsrMENGOWZlc2U5dmtCZzRyMkxrYkpMTTVIV1dpWXg1dDJGYlBWClJETFV6TXBEazJOejZYWGZZTHFBVlNockdZdjBmUVZhdTN6em5CQlpVOVYvSzhIc0RTc3g3cDJQZmhYR1E2VXIKaHFNWWt6UEhrSDBUTjNXQ3NDR2dYN2t0TjEzcmFiNitzWWdXNHpWRW9DZnNBUGpGQkx2M0k3MWJJQkxhU1ZLRQorK0wvb2xQdTdYV3MyYkRXeURYVlo0LzRtaURjTHYzaXp2SjZ1R2hyNTFjK2ZiQjYxcVRLYXZSTVFRVUdHek9jCmF0SG1EaGpuQWdNQkFBRUNnZ0VBTmNuYWkyNDlGYVFvdWF1OHFhTTN1dGRKTkpTN3k4bm12QVRpTHRQdEkvclUKK0srN1BYVkhkU0U0aWhXc2pkUUs4aXpzdlc2Q1Y4dFFtd00yM3lNRlV0aDVKNUFidVFrQXBoQms0SlJnUFpWUQpPVVMyWHV3VEJsbFRsN3FBOWUyK1VnVTdEK0syazF4UHR3Y0xxT1hIemJsZjV3WFg4OG81YnlBL1NlM3M3MElRCk5iN0lnNjI5Y1puKy9oQWl4R29RRXlreGJGWjZ0UHZWUUlBMFZ2UW53QjQ0QWtYWnZlRnN5bUJ1U1JiUWVHZ1IKb2MzLzZiV2tQaUJyYlQ2aytTYThjYit0VTNQRUVZZC9pbzA4dFZTMDFwc3VGVi8xYUkxczE4OTRrUHJXNktsMgpUTGVOem1EcW1ZblRuMDd3MnR5V2JyL21kNGhnL0M4dk9ZT3pYS3R3Z1FLQmdRREljRDFqZWJlaVA5aC80UVJkCkRoekZRUlVDby83OExLcHBWRmpDVUY1MytvWC9OQ3hXVWlUNzRsSEtWTVNkSXZtRWxhd2Jsdm1YdHMrU0cxUHoKTmJIN0MyMys2TUtBQ2QwUzAxaFBWWHlKSjYza0xkcG5rQVZyUmRCbmtYeEh4d3NWc0pjcWEyNTB2R2ZwTHJUVQo4d2d1UjFoMFpsaTB6WDhQbnVkVi9WSHZKd0tCZ1FEQjF0b2dvcGVzaGI0ZWYzSTEvRnhXd3ZQd3d3RUYxVU9NCmZrdXBDelpTOEJubzJtSUJZUWFtRWRtQnR1QU1kSWtkR1BkNU9NdHpINFpaSjlocFhQckMydThDK2x4WUoxVlkKVTZZVGErRzRjaU9PQWRMa3RZMXJoNU1DY0hLTTlpU3R6LzZZYWZIdTVhb2RjcUd5SDlMRTU4TXZkUTcxNmh5ZwpaU0VIWi9pZ1FRS0JnRU5IV1hQQWNXRW1xUmNUZ3BGeG9UcWN3OTZsQ1h1L3lsdWNra1ozRDU2YUdzdzB5UVVZCmdZMkN4QTEwTXFMRUVKanVYRnpPYW0wQVVlQXJDQnpFMHo3KzhTYjFIZ1E0UzFwOFVsSWUwYlIvK3lCeU83TXoKWm41QmF0aTR2c3loQlJsOHN1RHNPcUU0ZEhDUzJ1UDN4N0V4QllIY3NMM1BsR3k3Mjg3RFB2TlZBb0dCQUtHVAptNjJhZXIzbm1ndklCb2J6dm5EZi93R1JPemdHaGxFRk1jSk9RMUV0TFJ2SmRlcGFXM1Z3NlpMVHdyei9JeEFyCk1KWk9mbUNQUmFqcHF0NWhEL0gvRno2dlBPeUtsUHlVZFpvNDBpV3lsdXFYb0pqZXNXeVJ6VHc2U1hJdzkzQWYKTWNVUWd3ZTFUM0ZPalhSeVRIbXdDeVp5K0M2S29LUWV5RUpwSzNsQkFvR0JBSjZFZ2xPTFY0ZklRKzIrYmtvaApnblNPVVloYUR5aDFFY08wT0hDa01vSndOaHBSNldrQWhKTTY1OTBkY3pFMFpoUU54Ym9LT0ZUbVkxODFyS2JsCmUxU1RiRXloZGRycWJJb25kMkR2V0FvNndPVTcrZDZOcGNCVjRTcDBKWndqMStRUnM5U2ZWR2t5b0R5NlMxbnAKaXNxQVFOcUhCWHdNV0JrNHpXSytMSEV1Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0="
	decodedPrivateKey, _ := base64.StdEncoding.DecodeString(encodedPrivateKey)

	rsaKey, _ := rsaConfigSetup(decodedPrivateKey)
	keys := map[string]*credloader.KeyStore{
		secretName: {
			PrivateKey: rsaKey,
		},
	}

	ctx := auth.SaveToContext(context.Background(), &auth.SelfSignedTokenCredentials{
		KeysCache:                 credloader.NewKeyCacheWithKeys(keys),
		JwtSelfSignCertSecretName: secretName,
		Claims:                    map[string]interface{}{auth.CustomerIDClaimKey: tenant},
	})

	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)

	suite.Require().Contains(authorization, "Bearer")
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_GetAuthorizationFailsWhenNoCredentialsInContext() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorJwtConfig)

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().True(apperrors.IsNotFoundError(err))
	suite.Require().Empty(authorization)
}

func (suite *SelfSignedJwtTokenAuthorizationProviderTestSuite) TestSelfSignedJwtTokenAuthorizationProvider_GetAuthorizationFailsWhenBasicCredentialsAreInContext() {
	provider := auth.NewSelfSignedJWTTokenAuthorizationProvider(directorJwtConfig)

	authorization, err := provider.GetAuthorization(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to cast credentials to self-signed token credentials type")
	suite.Require().Empty(authorization)
}

func rsaConfigSetup(pKey []byte) (*rsa.PrivateKey, error) {
	privPem, _ := pem.Decode(pKey)
	privPemBytes := privPem.Bytes

	var parsedKey interface{}
	var err error
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			return nil, err
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, err
	}

	return privateKey, nil
}
