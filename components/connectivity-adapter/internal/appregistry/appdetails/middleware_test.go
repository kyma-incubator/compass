package appdetails_test

import (
	"github.com/gorilla/mux"
	"net/http"
	"testing"
)

func TestMiddleware(t *testing.T) {
	//GIVEN
	router := mux.NewRouter()
	router.HandleFunc("/", dummy)

}

func dummy(writer http.ResponseWriter, r *http.Request) {

}
