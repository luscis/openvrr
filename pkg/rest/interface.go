package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openvrr/pkg/schema"
)

type Interface struct {
	call Caller
}

func (l Interface) Router(r *mux.Router) {
	r.HandleFunc("/api/interface", l.List).Methods("GET")
	r.HandleFunc("/api/interface", l.Add).Methods("POST")
	r.HandleFunc("/api/interface", l.Remove).Methods("DELETE")
}

func (l Interface) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, nil)
}

func (l Interface) Add(w http.ResponseWriter, r *http.Request) {
	data := schema.Interface{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.AddInterface(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (l Interface) Remove(w http.ResponseWriter, r *http.Request) {
	data := schema.Interface{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.DelInterface(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}
