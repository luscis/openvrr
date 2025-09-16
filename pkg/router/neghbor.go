package router

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishvananda/netlink"
)

type IPNeighbor struct {
}

func (r *IPNeighbor) Init() {
}

func (n *IPNeighbor) list() {
	neighbors, err := netlink.NeighList(0, syscall.AF_INET)
	if err != nil {
		log.Fatalf("failed to list routes: %v\n", err)
	}

	for _, neigh := range neighbors {
		log.Printf("Neighbor update received: Neigh=%+v\n", neigh)

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
		log.Fatalf("failed to subscribe to neighbor updates: %v\n", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case update := <-neighCh:
			log.Printf("Neighbor update received: Type=%v, Neigh=%+v\n", update.Type, update.Neigh)

		case <-sigCh:
			close(doneCh)
			return
		}
	}
}

type IPRoute struct {
}

func (r *IPRoute) Init() {
}

func (r *IPRoute) list() {
	routes, err := netlink.RouteList(nil, syscall.AF_INET)
	if err != nil {
		log.Fatalf("failed to list routes: %v\n", err)
	}

	for _, route := range routes {
		log.Printf("Route update received: Route=%+v\n", route)
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
		log.Fatalf("failed to subscribe to route updates: %v\n", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case update := <-routeCh:
			log.Printf("Route update received: Type=%v, Route=%+v\n", update.Type, update.Route)

		case <-sigCh:
			close(doneCh)
			return
		}
	}
}
