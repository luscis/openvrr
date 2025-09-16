package router

import (
	"os"
	"os/signal"
	"syscall"
)

type Router struct {
	ipneigh *IPNeighbor
	iproute *IPRoute
	compose *Composer
}

func (v *Router) Init() {
	v.iproute = &IPRoute{}
	v.iproute.Init()

	v.ipneigh = &IPNeighbor{}
	v.ipneigh.Init()

	v.compose = &Composer{}
	v.compose.Init()
}

func (v *Router) Start() {
	v.ipneigh.Start()
	v.iproute.Start()
	v.compose.Start()
}

func (v *Router) Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	<-x
}
