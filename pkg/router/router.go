package router

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/luscis/openvrr/pkg/schema"
	"github.com/vishvananda/netlink"
)

type Router struct {
	ipneigh *IPNeighbor
	iproute *IPRoute
	compose *Composer
	http    *Http
}

func (v *Router) Init() {
	v.compose = &Composer{}
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

func (v *Router) Start() {
	v.ipneigh.Start()
	v.iproute.Start()
	v.compose.Start()
	v.http.Start()
}

func (v *Router) Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	<-x
}

func (v *Router) AddVlan(data schema.Vlan) error {
	if data.Tag > 0 {
		return v.compose.addVlanTag(data.Interface, data.Tag)
	}
	if data.Trunks != "" {
		return v.compose.addVlanTrunks(data.Interface, data.Trunks)
	}
	return nil
}

func (v *Router) DelVlan(data schema.Vlan) error {
	return nil
}

func (v *Router) AddInterface(data schema.Interface) error {
	return v.compose.addVlanPort(data.Name)
}

func (v *Router) DelInterface(data schema.Interface) error {
	return v.compose.delPort(data.Name)
}

func (v *Router) OnNeighbor(update uint16, nei netlink.Neigh) error {
	log.Printf("Router.OnNeighbor: Type=%d, Neigh=%+v", update, nei)
	link, err := netlink.LinkByIndex(nei.LinkIndex)
	if err != nil {
		log.Printf("Router.OnNeighbor: %v", err)
		return err
	}
	attr := link.Attrs()
	port := attr.Name
	if !strings.HasPrefix(port, "vlan") {
		return nil
	}

	ipdst := nei.IP.String()
	ethdst := nei.HardwareAddr.String()

	switch update {
	case 0, 28:
		if ethdst == "" {
			return nil
		}
		v.compose.AddRoute(ipdst, HwAddr(ethdst), port)
	case 29:
		v.compose.DelRoute(ipdst, port)
	}

	return nil
}

func (v *Router) OnRoute(update uint16, route netlink.Route) error {
	log.Printf("Router.OnRoute: Type=%d, Neigh=%+v", update, route)
	return nil
}
