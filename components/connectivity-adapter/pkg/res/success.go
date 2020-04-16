package res

import (
	"encoding/json"
	"net/http"
)

func WriteJSONResponse(writer http.ResponseWriter, res interface{}) error {
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	return json.NewEncoder(writer).Encode(&res)
}
