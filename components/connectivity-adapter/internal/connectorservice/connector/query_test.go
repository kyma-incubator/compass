package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryProvider(t *testing.T) {
	t.Run("Should remove white chars in Sign CSR query", func(t *testing.T) {
		// given
		csr := `
		aa	

		bb

        cc 
   `
		expectedQuery := `mutation {
	result: signCertificateSigningRequest(csr: "aabbcc")
  	{
	 	certificateChain
		caCertificate
		clientCertificate
	}
    }`
		qp := queryProvider{}

		// when
		query := qp.signCSR(csr)

		// then
		assert.Equal(t, expectedQuery, query)
	})
}
