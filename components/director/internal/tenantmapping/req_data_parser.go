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

func (p *reqDataParser) Parse(req *http.Request) (ReqData, error) {
	var data ReqData
	err := json.NewDecoder(req.Body).Decode(&data)
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

	if data.Extra == nil {
		data.Extra = make(map[string]interface{})
	}

	return data, nil
}
