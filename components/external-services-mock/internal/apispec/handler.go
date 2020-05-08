package apispec

import (
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randSpec(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func HandleFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write(randSpec(10))
	if err != nil {
		logrus.Fatal(err)
	}
}
