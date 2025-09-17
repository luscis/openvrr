package router

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishvananda/netlink"
)

type IPNeighbor struct {
	On func(uint16, netlink.Neigh) error
}

func (r *IPNeighbor) Init() {
}

func (n *IPNeighbor) list() {
	neighbors, err := netlink.NeighList(0, syscall.AF_INET)
	if err != nil {
		log.Fatalf("IPNeighbor: list routes: %v", err)
	}

	for _, neigh := range neighbors {
		n.On(0, neigh)
	}
}

func (n *IPNeighbor) Start() {
	n.list()
	go n.watch()
}

func (n *IPNeighbor) watch() {
	neighCh := make(chan netlink.NeighUpdate)
	doneCh := make(chan struct{})

	err := netlink.NeighSubscribe(neighCh, doneCh)
	if err != nil {
		log.Fatalf("IPNeighbor.watch: subscribe to updates: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case update := <-neighCh:
			n.On(update.Type, update.Neigh)

		case <-sigCh:
			close(doneCh)
			return
		}
	}
}

type IPRoute struct {
	On func(update uint16, route netlink.Route) error
}

func (r *IPRoute) Init() {
}

func (r *IPRoute) list() {
	routes, err := netlink.RouteList(nil, syscall.AF_INET)
	if err != nil {
		log.Fatalf("IPRoute.watch: list routes: %v", err)
	}

	for _, route := range routes {
		r.On(0, route)
	}
}

func (r *IPRoute) Start() {
	r.list()
	go r.watch()
}

func (r *IPRoute) watch() {
	routeCh := make(chan netlink.RouteUpdate)
	doneCh := make(chan struct{})

	err := netlink.RouteSubscribe(routeCh, doneCh)
	if err != nil {
		log.Fatalf("IPRoute.watch: subscribe to updates: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case update := <-routeCh:
			r.On(update.Type, update.Route)

		case <-sigCh:
			close(doneCh)
			return
		}
	}
}
