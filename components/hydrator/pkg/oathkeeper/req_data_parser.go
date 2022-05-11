package oathkeeper

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// NewReqDataParser missing godoc
func NewReqDataParser() *reqDataParser {
	return &reqDataParser{}
}

type reqDataParser struct{}

// Parse returns parsed incomming request as ReqData with body as ReqBody struct and original headers collection
func (p *reqDataParser) Parse(req *http.Request) (ReqData, error) {
	var reqBody ReqBody
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		if err == io.EOF {
			return ReqData{}, apperrors.NewInternalError("request body is empty")
		}

		return ReqData{}, errors.Wrap(err, "while decoding request body")
	}

	defer func() {
		err := req.Body.Close()
		if err != nil {
			log.C(req.Context()).WithError(err).Errorf("An error has occurred while closing request body: %v", err)
		}
	}()

	return NewReqData(req.Context(), reqBody, req.Header), nil
}
