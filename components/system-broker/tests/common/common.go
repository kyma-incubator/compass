package common

import (
	"encoding/json"
	"fmt"
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

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_, werr := w.Write([]byte(fmt.Sprintf(`{"description": %q`, err.Error())))
	if werr != nil {
		panic(werr)
	}
}

func writeGQLError(w http.ResponseWriter, errMsg string) {
	errJson := struct {
		Errors []struct {
			Message string
		}
	}{
		Errors: make([]struct{ Message string }, 1),
	}

	errJson.Errors[0] = struct{ Message string }{
		Message: errMsg,
	}
	if err := json.NewEncoder(w).Encode(errJson); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
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
