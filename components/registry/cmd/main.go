package main

import (
	"net/http"
	"log"
)

const serverPort = "3000"

func EchoHandler(writer http.ResponseWriter, request *http.Request) {

	log.Println("Echoing back request made to " + request.URL.Path + " to client (" + request.RemoteAddr + ")")

	writer.Header().Set("Access-Control-Allow-Origin", "*")

	writer.Header().Set("Access-Control-Allow-Headers", "Content-Range, Content-Disposition, Content-Type, ETag")

	request.Write(writer)
}

func main() {

	log.Println("starting server, listening on port " + serverPort)

	http.HandleFunc("/", EchoHandler)
	http.ListenAndServe(":" + serverPort, nil)
}
