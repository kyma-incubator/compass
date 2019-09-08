package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const serverPort = "3000"

type Data struct {
	Subject string      `json:"subject"`
	Extra   interface{} `json:"extra"`
	Header  interface{} `json:"header"`
}

func EchoHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("=== HEADERS ===")
	for key, val := range request.Header {
		log.Printf("%s: %+v\n", key, val)
	}
	log.Println("=== ==== ===")

	writer.Header().Set("Content-Type", "application/json")

	var data Data
	err := json.NewDecoder(request.Body).Decode(&data)

	log.Printf("Body: \n %+v\n", data)

	defer func() {
		err := request.Body.Close()
		if err != nil {
			log.Println("error: ", err)
		}
	}()

	if data.Extra == nil {
		data.Extra = make(map[string]interface{})
	}

	extraMap, ok := data.Extra.(map[string]interface{})
	if !ok {
		log.Printf("error: Incorrect type %T\n", data.Extra)
	}
	extraMap["tenant"] = "9ac609e1-7487-4aa6-b600-0904b272b11f"

	data.Extra = extraMap

	log.Println("=== ==== ===")
	log.Printf("Response: \n %+v\n\n", data)

	err = json.NewEncoder(writer).Encode(data)

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
