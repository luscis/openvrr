package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openvrr/pkg/schema"
)

type SNAT struct {
	call Caller
}

func (l SNAT) Router(r *mux.Router) {
	r.HandleFunc("/api/snat", l.List).Methods("GET")
	r.HandleFunc("/api/snat", l.Add).Methods("POST")
	r.HandleFunc("/api/snat", l.Remove).Methods("DELETE")
}

func (l SNAT) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, nil)
}

func (l SNAT) Add(w http.ResponseWriter, r *http.Request) {
	data := schema.SNAT{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.AddSNAT(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (l SNAT) Remove(w http.ResponseWriter, r *http.Request) {
	data := schema.SNAT{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.DelSNAT(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

type DNAT struct {
	call Caller
}

func (l DNAT) Router(r *mux.Router) {
	r.HandleFunc("/api/dnat", l.List).Methods("GET")
	r.HandleFunc("/api/dnat", l.Add).Methods("POST")
	r.HandleFunc("/api/dnat", l.Remove).Methods("DELETE")
}

func (l DNAT) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, nil)
}

func (l DNAT) Add(w http.ResponseWriter, r *http.Request) {
	data := schema.DNAT{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.AddDNAT(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (l DNAT) Remove(w http.ResponseWriter, r *http.Request) {
	data := schema.DNAT{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := l.call.DelDNAT(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}
