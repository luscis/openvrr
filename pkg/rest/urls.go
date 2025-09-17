package rest

import "github.com/gorilla/mux"

func Add(r *mux.Router, call Caller) {
	Interface{call: call}.Router(r)
	Vlan{call: call}.Router(r)
}
