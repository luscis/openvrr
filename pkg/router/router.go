package router

import (
	"os"
	"os/signal"
	"syscall"
)

type Router struct {
	ipneigh *IPNeighbor
	iproute *IPRoute
}

func (v *Router) Init() {
	v.iproute = &IPRoute{}
	v.ipneigh = &IPNeighbor{}
}

func (v *Router) Start() {
	v.ipneigh.Start()
	v.iproute.Start()
}

func (v *Router) Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	<-x
}
