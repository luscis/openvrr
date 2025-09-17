package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openvrr/pkg/schema"
)

type Vlan struct {
	call Caller
}

func (l Vlan) Router(r *mux.Router) {
	r.HandleFunc("/api/vlan", l.List).Methods("GET")
	r.HandleFunc("/api/vlan", l.Add).Methods("POST")
	r.HandleFunc("/api/vlan", l.Remove).Methods("DELETE")
}

func (l Vlan) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, nil)
}

func (l Vlan) Add(w http.ResponseWriter, r *http.Request) {
	data := schema.Vlan{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.AddVlan(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (l Vlan) Remove(w http.ResponseWriter, r *http.Request) {
	data := schema.Vlan{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.DelVlan(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}
