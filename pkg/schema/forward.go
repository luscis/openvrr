package schema

type IPForward struct {
	Prefix    string `json:"prefix" yaml:"prefix"`
	NextHop   string `json:"nexthop,omitempty" yaml:"nexthop,omitempty"`
	Interface string `json:"interface,omitempty" yaml:"interface,omitempty"`
	LLAddr    string `json:"lladdr,omitempty" yaml:"lladdr,omitempty"`
}
