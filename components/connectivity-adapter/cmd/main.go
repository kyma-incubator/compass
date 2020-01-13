package main

import "net/http"

func main() {
	http.HandleFunc("/healthz", healthz)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}

func healthz(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
