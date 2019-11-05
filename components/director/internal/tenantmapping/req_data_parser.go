package tenantmapping

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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
			return ReqData{}, errors.New("request body is empty")
		}

		return ReqData{}, errors.Wrap(err, "while decoding request body")
	}

	defer func() {
		err := req.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	return NewReqData(reqBody, req.Header), nil
}
