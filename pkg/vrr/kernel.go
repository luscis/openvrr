package vrr

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	UpdateNeighNew = 0
	UpdateNeighAdd = 28
	UpdateNeighDel = 29
	UpdateRouteNew = 0
	UpdateRouteAdd = 24
	UpdateRouteDel = 25
)

func NeighListAt(ns netns.NsHandle) ([]netlink.Neigh, error) {
	if ns != netns.None() {
		if h, err := netlink.NewHandleAt(ns); err != nil {
			return nil, err
		} else {
			defer h.Close()
			return h.NeighList(0, syscall.AF_INET)
		}
	}
	return netlink.NeighList(0, syscall.AF_INET)
}

type KernelNeighbor struct {
	ns netns.NsHandle
	On func(uint16, netlink.Neigh) error
}

func (n *KernelNeighbor) Init() {
}

func (n *KernelNeighbor) list() {
	neighbors, err := NeighListAt(n.ns)
	if err != nil {
		log.Fatalf("KernelNeighbor.list: %v", err)
	}

	for _, neigh := range neighbors {
		n.On(0, neigh)
	}
}

func (n *KernelNeighbor) Start() {
	n.list()
	go n.watch()
}

func (n *KernelNeighbor) watch() {
	neighCh := make(chan netlink.NeighUpdate)
	doneCh := make(chan struct{})

	err := netlink.NeighSubscribeAt(n.ns, neighCh, doneCh)
	if err != nil {
		log.Fatalf("KernelNeighbor.watch: subscribe %v", err)
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

func RouteListAt(ns netns.NsHandle) ([]netlink.Route, error) {
	if ns != netns.None() {
		if h, err := netlink.NewHandleAt(ns); err != nil {
			return nil, err
		} else {
			defer h.Close()
			return h.RouteList(nil, syscall.AF_INET)
		}
	}
	return netlink.RouteList(nil, syscall.AF_INET)
}

type KernelRoute struct {
	ns netns.NsHandle
	On func(uint16, netlink.Route) error
}

func (r *KernelRoute) Init() {
}

func (r *KernelRoute) list() {
	routes, err := RouteListAt(r.ns)
	if err != nil {
		log.Fatalf("KernelRoute.list: %v", err)
	}

	for _, route := range routes {
		r.On(0, route)
	}
}

func (r *KernelRoute) Start() {
	r.list()
	go r.watch()
}

func (r *KernelRoute) watch() {
	routeCh := make(chan netlink.RouteUpdate)
	doneCh := make(chan struct{})

	err := netlink.RouteSubscribeAt(r.ns, routeCh, doneCh)
	if err != nil {
		log.Fatalf("KernelRoute.watch: subscribe %v", err)
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

func AddrListAt(ns netns.NsHandle) ([]netlink.Addr, error) {
	if ns != netns.None() {
		if h, err := netlink.NewHandleAt(ns); err != nil {
			return nil, err
		} else {
			defer h.Close()
			return h.AddrList(nil, syscall.AF_INET)
		}
	}
	return netlink.AddrList(nil, syscall.AF_INET)
}

type KernelAddr struct {
	ns netns.NsHandle
	On func(netlink.AddrUpdate) error
}

func (r *KernelAddr) Init() {
}

func (r *KernelAddr) list() {
	addrs, err := AddrListAt(r.ns)
	if err != nil {
		log.Fatalf("KernelAddr.list: %v", err)
	}

	for _, addr := range addrs {
		r.On(netlink.AddrUpdate{
			NewAddr:     true,
			LinkIndex:   addr.LinkIndex,
			LinkAddress: *addr.IPNet,
		})
	}
}

func (r *KernelAddr) Start() {
	r.list()
	go r.watch()
}

func (r *KernelAddr) watch() {
	routeCh := make(chan netlink.AddrUpdate)
	doneCh := make(chan struct{})

	err := netlink.AddrSubscribeAt(r.ns, routeCh, doneCh)
	if err != nil {
		log.Fatalf("KernelAddr.watch: subscribe %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case update := <-routeCh:
			r.On(update)

		case <-sigCh:
			close(doneCh)
			return
		}
	}
}

type KernelRegister struct {
	ns         netns.NsHandle
	neighbor   *KernelNeighbor
	route      *KernelRoute
	addr       *KernelAddr
	OnAddress  func(netlink.AddrUpdate) error
	OnRoute    func(uint16, netlink.Route) error
	OnNeighbor func(uint16, netlink.Neigh) error
}

func (r *KernelRegister) Init() {
	r.addr = &KernelAddr{
		ns: r.ns,
		On: r.OnAddress,
	}
	r.neighbor = &KernelNeighbor{
		ns: r.ns,
		On: r.OnNeighbor,
	}
	r.route = &KernelRoute{
		ns: r.ns,
		On: r.OnRoute,
	}
}

func (r *KernelRegister) Start() {
	r.neighbor.Start()
	r.route.Start()
	r.addr.Start()
}

func (r *KernelRegister) Stop() {
	// TODO
}
