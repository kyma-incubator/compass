package httpcommon

import (
	"io"
	"log"
)

func CloseBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Printf("while closing body %+v\n", err)
	}
}
