package schema

type Vlan struct {
	Interface string
	Tag       int
	Trunks    string
	VlanMode  string
}

type Interface struct {
	Name string
}
