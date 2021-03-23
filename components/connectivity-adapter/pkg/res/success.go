package res

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

func WriteJSONResponse(writer http.ResponseWriter, ctx context.Context, res interface{}) error {
	log.C(ctx).Infoln("returning response...")
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)
	return enc.Encode(&res)
}
