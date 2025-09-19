package schema

type Interface struct {
	Name      string `json:"name" yaml:"name"`
	LinkState string `json:"linkstate,omitempty" yaml:"linkstate,omitempty"`
	Tag       int    `json:"tag,omitempty" yaml:"tag,omitempty"`
	Trunks    string `json:"trunks,omitempty" yaml:"trunks,omitempty"`
	Mac       string `json:"mac,omitempty" yaml:"mac,omitempty"`
	Ofport    int    `json:"ofport,omitempty" yaml:"ofport,omitempty"`
}
