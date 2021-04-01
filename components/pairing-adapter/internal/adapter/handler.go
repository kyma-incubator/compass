package adapter

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

//go:generate mockery --name=Client --output=automock --outpkg=automock
type Client interface {
	Do(ctx context.Context, req RequestData) (*ExternalToken, error)
}

func NewHandler(cli Client) *Handler {
	return &Handler{cli: cli}
}

type Handler struct {
	cli Client
}

// swagger:route POST /adapter adapter
// Request token from external solution
// 		Consumes:
//		- application/json
//   	Produces:
//		- application/json
//		Responses:
// 		200: externalToken
//		400:
// 		500:
func (a *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := req.Body.Close(); err != nil {
			logrus.Error("Got error on closing request body", err)
		}
	}()
	decoder := json.NewDecoder(req.Body)
	var reqData RequestData
	err := decoder.Decode(&reqData)
	if err != nil {
		logrus.Warnf("Got error on decoding Application Data: %v\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	logrus.Infof("Got ApplicationData %v", reqData)
	token, err := a.cli.Do(req.Context(), reqData)
	if err != nil {
		logrus.Warnf("Got error on calling external pairing server: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(rw).Encode(token); err != nil {
		logrus.Warnf("Got error on encoding response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
