package common

import (
	"io/ioutil"
	"net/http"
	"os"
)

type closer interface {
	Close()
}

type urler interface {
	URL() string
}

type FakeServer interface {
	closer
	urler
}

func writeError(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Write([]byte("{}"))
}

func getFileContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}