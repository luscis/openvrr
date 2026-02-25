package schema

type SNAT struct {
	Source   string `json:"source" yaml:"source"`
	SourceTo string `json:"source_to" yaml:"source_to"`
}

type DNAT struct {
	Dest   string `json:"dest" yaml:"dest"`
	DestTo string `json:"dest_to" yaml:"dest_to"`
}
