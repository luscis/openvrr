package router

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/luscis/openvrr/pkg/schema"
)

type Router struct {
	ipneigh *IPNeighbor
	iproute *IPRoute
	compose *Composer
	http    *Http
}

func (v *Router) Init() {
	v.iproute = &IPRoute{}
	v.iproute.Init()

	v.ipneigh = &IPNeighbor{}
	v.ipneigh.Init()

	v.compose = &Composer{}
	v.compose.Init()

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
	return nil
}

func (v *Router) DelVlan(data schema.Vlan) error {
	return nil
}

func (v *Router) AddInterface(data schema.Interface) error {
	return v.compose.addVlanPort(data.Name)
}

func (v *Router) DelInterface(data schema.Interface) error {
	return v.compose.delVlanPort(data.Name)
}
