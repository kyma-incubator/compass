package apispec

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func randSpec() []byte {
	return []byte(fmt.Sprintf(specTemplate, randString(10)))
}

func HandleFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write(randSpec())
	if err != nil {
		rw.WriteHeader(503)
		logrus.Info(err)
	}
}
