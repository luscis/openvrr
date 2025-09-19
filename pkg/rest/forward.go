package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Forward struct {
	call Caller
}

func (l Forward) Router(r *mux.Router) {
	r.HandleFunc("/api/forward", l.List).Methods("GET")
}

func (l Forward) List(w http.ResponseWriter, r *http.Request) {
	items, _ := l.call.ListForward()
	ResponseJson(w, items)
}
