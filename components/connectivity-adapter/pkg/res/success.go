package res

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func WriteJSONResponse(writer http.ResponseWriter, res interface{}) error {
	log.Infoln("returning response...")
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)
	return enc.Encode(&res)
}
