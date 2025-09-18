package schema

type Interface struct {
	Name      string
	LinkState string
	Tag       int    `json:"tag,omitempty" yaml:"tag,omitempty"`
	Trunks    string `json:"trunks,omitempty" yaml:"trunks,omitempty"`
	Mac       string
	ofport    int
}
