package service

import (
	"log"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Create(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Create")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Get(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Get")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Update(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Update")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Delete(rw http.ResponseWriter, rq *http.Request) {
	log.Println("Delete")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) List(rw http.ResponseWriter, rq *http.Request) {
	log.Println("List")
	rw.WriteHeader(http.StatusOK)
}
