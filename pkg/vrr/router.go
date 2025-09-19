package vrr

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/luscis/openvrr/pkg/schema"
	"github.com/vishvananda/netlink"
)

type IPForwards map[string]schema.IPForward

func (f IPForwards) Add(value schema.IPForward) {
	f[value.Prefix] = value
}

func (f IPForwards) Remove(prefix string) {
	delete(f, prefix)
}

type Vrr struct {
	ipneigh   *IPNeighbor
	iproute   *IPRoute
	compose   *Composer
	http      *Http
	forward   IPForwards
	linkAttrs map[int]*netlink.LinkAttrs
	mutex     sync.RWMutex
}

func (v *Vrr) Init() {
	v.forward = make(map[string]schema.IPForward)
	v.linkAttrs = make(map[int]*netlink.LinkAttrs)

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

func (v *Vrr) AddVlan(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if data.Tag > 0 {
		return v.compose.addVlanTag(data.Name, data.Tag)
	}
	if data.Trunks != "" {
		return v.compose.addVlanTrunks(data.Name, data.Trunks)
	}
	return nil
}

func (v *Vrr) DelVlan(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return nil
}

func (v *Vrr) AddInterface(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.compose.addVlanPort(data.Name)
}

func (v *Vrr) DelInterface(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.compose.delPort(data.Name)
}

func (v *Vrr) ListInterface() ([]schema.Interface, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	ports, err := v.compose.listPorts()
	if err != nil {
		return nil, err
	}

	var items []schema.Interface
	for _, port := range ports {
		items = append(items, schema.Interface{
			Name:      port.Name,
			Tag:       port.Tag,
			Trunks:    port.Trunks,
			LinkState: port.LinkState,
			Mac:       port.Mac,
		})
	}
	return items, nil
}

func (v *Vrr) OnNeighbor(update uint16, host netlink.Neigh) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if host.Family == netlink.FAMILY_V6 {
		return nil
	}

	log.Printf("Vrr.OnNeighbor: Type=%d, Host=%+v", update, host)

	attr := v.findLinkAttr(host.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}

	port := attr.Name
	ipdst := host.IP.String()
	ethdst := host.HardwareAddr.String()
	switch update {
	case UpdateNeighNew, UpdateNeighAdd:
		if ethdst == "" || host.IP.IsMulticast() {
			return nil
		}

		v.compose.AddHost(IpAddr(ipdst), HwAddr(ethdst), port)
		v.forward.Add(schema.IPForward{
			Prefix:    ipdst,
			NextHop:   ipdst,
			LLAddr:    ethdst,
			Interface: port,
		})
	case UpdateNeighDel:
		v.compose.DelHost(IpAddr(ipdst), port)
		v.forward.Remove(ipdst)
	}

	return nil
}

func (v *Vrr) findLinkAttr(index int) *netlink.LinkAttrs {
	if link, err := netlink.LinkByIndex(index); err == nil {
		v.linkAttrs[index] = link.Attrs()
	}

	return v.linkAttrs[index]
}

func (v *Vrr) OnRoute(update uint16, rule netlink.Route) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if rule.Family == netlink.FAMILY_V6 {
		return nil
	}

	log.Printf("Vrr.OnRoute: Type=%d, Rule=%+v", update, rule)

	attr := v.findLinkAttr(rule.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}

	port := attr.Name
	ipdst := rule.Dst.String()
	ipgw := rule.Gw.String()
	switch update {
	case UpdateRouteAdd, UpdateRouteNew:
		if ipgw == "" || ipgw == "<nil>" {
			return nil
		}
		v.compose.AddRoute(IpPrefix(ipdst), IpAddr(ipgw), port)
		v.forward.Add(schema.IPForward{
			Prefix:    ipdst,
			NextHop:   ipgw,
			Interface: port,
		})
	case UpdateRouteDel:
		v.compose.DelRoute(IpPrefix(ipdst), port)
		v.forward.Remove(ipdst)
	}

	return nil
}

func (v *Vrr) ListForward() ([]schema.IPForward, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var items []schema.IPForward
	for _, value := range v.forward {
		items = append(items, value)
	}
	return items, nil
}
