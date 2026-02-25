package rest

import "github.com/luscis/openvrr/pkg/schema"

type Caller interface {
	AddVlan(data schema.Interface) error
	DelVlan(data schema.Interface) error
	AddInterface(data schema.Interface) error
	DelInterface(data schema.Interface) error
	ListInterface() ([]schema.Interface, error)
	ListForward() ([]schema.IPForward, error)
	AddSNAT(data schema.SNAT) error
	DelSNAT(data schema.SNAT) error
	AddDNAT(data schema.DNAT) error
	DelDNAT(data schema.DNAT) error
}
