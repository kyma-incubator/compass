package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func NewChunkedHandler() *ChunkedHandler {
	return &ChunkedHandler{}
}

type ChunkedHandler struct {
}

func (a *ChunkedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, req.Body)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Received body:%q\n", buf.Bytes())
	rw.Header().Add("Content-Length", strconv.FormatInt(n, 10))
	rw.WriteHeader(200)
	io.Copy(rw, buf)
}
