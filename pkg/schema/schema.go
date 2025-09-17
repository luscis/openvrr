package schema

type Vlan struct {
	Interface string
	Tag       string
	Trunks    string
	VlanMode  string
}

type Interface struct {
	Name string
}
