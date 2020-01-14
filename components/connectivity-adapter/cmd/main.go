package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", dummy)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}

func healthz(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func dummy(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("%+v", req.Header)

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}
