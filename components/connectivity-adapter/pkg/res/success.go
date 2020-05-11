package res

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func WriteJSONResponse(writer http.ResponseWriter, res interface{}) error {
	log.Infoln("returning response...")
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	return json.NewEncoder(writer).Encode(&res)
}
