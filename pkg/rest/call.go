package rest

import "github.com/luscis/openvrr/pkg/schema"

type Caller interface {
	AddVlan(data schema.Interface) error
	DelVlan(data schema.Interface) error
	AddInterface(data schema.Interface) error
	DelInterface(data schema.Interface) error
	ListInterface() ([]schema.Interface, error)
}
