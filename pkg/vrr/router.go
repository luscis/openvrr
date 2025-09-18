package vrr

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/luscis/openvrr/pkg/schema"
	"github.com/vishvananda/netlink"
)

type Vrr struct {
	ipneigh *IPNeighbor
	iproute *IPRoute
	compose *Composer
	http    *Http
}

func (v *Vrr) Init() {
	v.compose = &Composer{
		brname: "br-vrr",
	}
	v.compose.Init()

	v.iproute = &IPRoute{
		On: v.OnRoute,
	}
	v.iproute.Init()

	v.ipneigh = &IPNeighbor{
		On: v.OnNeighbor,
	}
	v.ipneigh.Init()

	v.http = &Http{
		listen:    "127.0.0.1:10001",
		adminFile: "/etc/openvrr/token",
		caller:    v,
	}
	v.http.Init()
}

func (v *Vrr) Start() {
	v.ipneigh.Start()
	v.iproute.Start()
	v.compose.Start()
	v.http.Start()
}

func (v *Vrr) Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	<-x
}

func (v *Vrr) AddVlan(data schema.Vlan) error {
	if data.Tag > 0 {
		return v.compose.addVlanTag(data.Interface, data.Tag)
	}
	if data.Trunks != "" {
		return v.compose.addVlanTrunks(data.Interface, data.Trunks)
	}
	return nil
}

func (v *Vrr) DelVlan(data schema.Vlan) error {
	return nil
}

func (v *Vrr) AddInterface(data schema.Interface) error {
	return v.compose.addVlanPort(data.Name)
}

func (v *Vrr) DelInterface(data schema.Interface) error {
	return v.compose.delPort(data.Name)
}

func (v *Vrr) OnNeighbor(update uint16, host netlink.Neigh) error {
	attr := FindLinkAttr(host.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}

	log.Printf("Vrr.OnNeighbor: Type=%d, Host=%+v", update, host)

	port := attr.Name
	ipdst := host.IP.String()
	ethdst := host.HardwareAddr.String()
	switch update {
	case UpdateNeighNew, UpdateNeighAdd:
		if ethdst == "" || host.IP.IsMulticast() {
			return nil
		}
		v.compose.AddHost(IpAddr(ipdst), HwAddr(ethdst), port)
	case UpdateNeighDel:
		v.compose.DelHost(IpAddr(ipdst), port)
	}

	return nil
}

func FindLinkAttr(index int) *netlink.LinkAttrs {
	link, err := netlink.LinkByIndex(index)
	if err != nil {
		return nil
	}
	return link.Attrs()
}

func (v *Vrr) OnRoute(update uint16, rule netlink.Route) error {
	attr := FindLinkAttr(rule.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}
	log.Printf("Vrr.OnRoute: Type=%d, Rule=%+v", update, rule)

	port := attr.Name
	ipdst := rule.Dst.String()
	ipgw := rule.Gw.String()
	switch update {
	case UpdateRouteAdd, UpdateRouteNew:
		if ipgw == "" || ipgw == "<nil>" {
			return nil
		}
		v.compose.AddRoute(IpPrefix(ipdst), IpAddr(ipgw), port)
	case UpdateRouteDel:
		v.compose.DelRoute(IpPrefix(ipdst), port)
	}

	return nil
}
