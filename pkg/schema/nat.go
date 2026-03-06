package schema

type SNAT struct {
	Source   string `json:"source" yaml:"source"`
	SourceTo string `json:"sourceTo" yaml:"sourceTo"`
}

type DNAT struct {
	Protocol string `json:"protocol" yaml:"protocol"`
	Dest     string `json:"destination" yaml:"destination"`
	DestTo   string `json:"destinationTo" yaml:"destinationTo"`
}
