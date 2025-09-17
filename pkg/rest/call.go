package rest

import "github.com/luscis/openvrr/pkg/schema"

type Caller interface {
	AddVlan(data schema.Vlan) error
	DelVlan(data schema.Vlan) error
	AddInterface(data schema.Interface) error
	DelInterface(data schema.Interface) error
}
