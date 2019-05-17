package main

import (
	"log"
	"net/http"
)

const serverPort = "3000"

func EchoHandler(writer http.ResponseWriter, request *http.Request) {

	log.Println("Echoing back request made to " + request.URL.Path + " to client (" + request.RemoteAddr + ")")

	writer.Header().Set("Access-Control-Allow-Origin", "*")

	writer.Header().Set("Access-Control-Allow-Headers", "Content-Range, Content-Disposition, Content-Type, ETag")

	err := request.Write(writer)
	if err != nil {
		log.Println("error: ", err)
	}
}

func main() {

	log.Println("starting server, listening on port " + serverPort)

	http.HandleFunc("/", EchoHandler)
	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil {
		log.Println("error: ", err)
	}
}
