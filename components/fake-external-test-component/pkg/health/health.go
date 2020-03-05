package health

import "net/http"

func HandleFunc(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
