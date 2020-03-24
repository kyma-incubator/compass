package health

import (
	"net/http"
)

func LivenessHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	}
}
