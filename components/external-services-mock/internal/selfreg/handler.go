package selfreg

import "net/http"

type SelfRegisterHandler struct {
}

func NewSelfRegisterHandler() *SelfRegisterHandler {
	return &SelfRegisterHandler{}
}

func (H *SelfRegisterHandler) HandleSelfRegPrep(rw http.ResponseWriter, req *http.Request) {

}

func (H *SelfRegisterHandler) HandleSelfRegCleanup(rw http.ResponseWriter, req *http.Request) {

}
