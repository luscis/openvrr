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
	"github.com/vishvananda/netns"
)

type IPForwards map[string]schema.IPForward

func (f IPForwards) Add(value schema.IPForward) {
	f[value.Prefix] = value
}

func (f IPForwards) Remove(prefix string) {
	delete(f, prefix)
}

type Gateway struct {
	kernel    *KernelRegister
	scomo     *Composer
	http      *Http
	forward   IPForwards
	linkAttrs map[int]*netlink.LinkAttrs
	mutex     sync.RWMutex
	ns        netns.NsHandle
}

const (
	vrname     = "vrr"
	tokenFile  = "/etc/openvrr/token"
	httpListen = "127.0.0.1:10001"
)

func (v *Gateway) Init() {
	v.forward = make(map[string]schema.IPForward)
	v.linkAttrs = make(map[int]*netlink.LinkAttrs)

	ns, err := netns.GetFromName(vrname)
	if err != nil {
		log.Fatalf("Gateway.Init: Get netns %v", err)
	} else {
		v.ns = ns
	}

	v.scomo = &Composer{
		brname: vrname,
		ns:     v.ns,
	}
	v.scomo.Init()

	v.kernel = &KernelRegister{
		ns:         v.ns,
		OnAddress:  v.OnAddress,
		OnRoute:    v.OnRoute,
		OnNeighbor: v.OnNeighbor,
	}
	v.kernel.Init()

	v.http = &Http{
		listen:    httpListen,
		adminFile: tokenFile,
		caller:    v,
	}
	v.http.Init()
}

func (v *Gateway) Start() {
	v.kernel.Start()
	v.scomo.Start()
	v.http.Start()
}

func (v *Gateway) Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	<-x
}

func (v *Gateway) AddVlan(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if data.Tag > 0 {
		if err := v.scomo.addVlanTag(data.Name, data.Tag); err != nil {
			return err
		}
	}
	if data.Trunks != "" {
		if err := v.scomo.addVlanTrunks(data.Name, data.Trunks); err != nil {
			return err
		}
	}
	return nil
}

func (v *Gateway) DelVlan(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if data.Tag == 4095 {
		if err := v.scomo.delVlanTag(data.Name); err != nil {
			return err
		}
	}
	if data.Trunks == "all" {
		if err := v.scomo.delVlanTrunks(data.Name); err != nil {
			return err
		}
	}

	return nil
}

func (v *Gateway) AddInterface(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.addVlanPort(data.Name)
}

func (v *Gateway) DelInterface(data schema.Interface) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.delPort(data.Name)
}

func (v *Gateway) ListInterface() ([]schema.Interface, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	ports, err := v.scomo.listPorts()
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

func (v *Gateway) OnNeighbor(update uint16, host netlink.Neigh) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if host.Family == netlink.FAMILY_V6 {
		return nil
	}

	log.Printf("Gateway.OnNeighbor: Type=%d, Host=%+v", update, host)

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

		v.scomo.AddHost(IPAddr(ipdst), HWAddr(ethdst), port)
		v.forward.Add(schema.IPForward{
			Prefix:    ipdst,
			NextHop:   ipdst,
			LLAddr:    ethdst,
			Interface: port,
		})
	case UpdateNeighDel:
		v.scomo.DelHost(IPAddr(ipdst), port)
		v.forward.Remove(ipdst)
	}

	return nil
}

func (v *Gateway) findLinkAttr(index int) *netlink.LinkAttrs {
	if v.ns != netns.None() {
		if h, err := netlink.NewHandleAt(v.ns); err != nil {
			return nil
		} else {
			if link, err := h.LinkByIndex(index); err == nil {
				v.linkAttrs[index] = link.Attrs()
			}
			return v.linkAttrs[index]
		}
	}
	if link, err := netlink.LinkByIndex(index); err == nil {
		v.linkAttrs[index] = link.Attrs()
	}

	return v.linkAttrs[index]
}

func (v *Gateway) OnRoute(update uint16, rule netlink.Route) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if rule.Family == netlink.FAMILY_V6 {
		return nil
	}

	log.Printf("Gateway.OnRoute: Type=%d, Rule=%+v", update, rule)

	attr := v.findLinkAttr(rule.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}

	port := attr.Name
	ipdst := rule.Dst.String()
	ipgw := rule.Gw.String()
	switch update {
	case UpdateRouteAdd, UpdateRouteNew:
		v.scomo.AddRoute(IPPrefix(ipdst), IPAddr(ipgw), port)
		v.forward.Add(schema.IPForward{
			Prefix:    ipdst,
			NextHop:   ipgw,
			Interface: port,
		})
	case UpdateRouteDel:
		v.scomo.DelRoute(IPPrefix(ipdst), port)
		v.forward.Remove(ipdst)
	}

	return nil
}

func (v *Gateway) ListForward() ([]schema.IPForward, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var items []schema.IPForward
	for _, value := range v.forward {
		items = append(items, value)
	}
	return items, nil
}

func (v *Gateway) AddSNAT(data schema.SNAT) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.AddSNAT(data.Source, data.SourceTo)
}

func (v *Gateway) DelSNAT(data schema.SNAT) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.DelSNAT(data.Source)
}

func (v *Gateway) ListSNAT() ([]schema.SNAT, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var results []schema.SNAT
	for key, value := range v.scomo.ListSNAT() {
		results = append(results, schema.SNAT{
			Source:   SplitSNAT(key),
			SourceTo: value,
		})
	}
	return results, nil
}

func (v *Gateway) AddDNAT(data schema.DNAT) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.AddDNAT(data.Protocol, data.Dest, data.DestTo)
}

func (v *Gateway) DelDNAT(data schema.DNAT) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.scomo.DelDNAT(data.Protocol, data.Dest)
}

func (v *Gateway) ListDNAT() ([]schema.DNAT, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var results []schema.DNAT
	for key, value := range v.scomo.ListDNAT() {
		protocol, dest := SplitDNAT(key)
		results = append(results, schema.DNAT{
			Dest:     dest,
			DestTo:   value,
			Protocol: protocol,
		})
	}
	return results, nil
}

func (v *Gateway) OnAddress(data netlink.AddrUpdate) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	log.Printf("Gateway.OnAddress: new=%v, Data=%+v", data.NewAddr, data)

	attr := v.findLinkAttr(data.LinkIndex)
	if attr == nil || !strings.HasPrefix(attr.Name, "vlan") {
		return nil
	}

	switch data.NewAddr {
	case true:
		v.scomo.AddLocal(data.LinkAddress.String())
	case false:
		v.scomo.DelLocal(data.LinkAddress.String())
	}

	return nil
}
