package router

import (
	"fmt"
	"log"
	"strings"

	"github.com/luscis/openvrr/pkg/ovs"
)

type Composer struct {
	brname string
	client *ovs.Client
	ofctl  *ovs.OpenFlowService
	vsctl  *ovs.VSwitchService
}

func (a *Composer) Start() {
	a.AddRoute("192.168.1.2", HwAddr("00:01"), "vlan10")
	a.AddRoute("192.168.2.2", HwAddr("00:02"), "vlan20")
}

func (a *Composer) addVlanPort(vlan string) error {
	ifs := ovs.InterfaceOptions{
		OfportRequest: a.findPortId(vlan),
		Mac:           a.findPortAddr(vlan),
		Type:          ovs.InterfaceTypeInternal,
	}
	if err := a.vsctl.AddPortWith(a.brname, vlan, ifs); err != nil {
		log.Printf("failed to add port: %v\n", err)
		return err
	}
	if err := a.vsctl.Set.Interface(vlan, ifs); err != nil {
		log.Printf("failed to set interface: %v\n", err)
		return err
	}

	ips := ovs.PortOptions{
		Tag: a.findVlanId(vlan),
	}
	if err := a.vsctl.Set.Port(vlan, ips); err != nil {
		log.Printf("failed to set port: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) delVlanPort(vlan string) error {
	if err := a.vsctl.DeletePort(a.brname, vlan); err != nil {
		log.Printf("failed to delete port: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) addBr(name string) error {
	if err := a.vsctl.AddBridge(name); err != nil {
		log.Fatalf("failed to add bridge: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) Init() {
	a.brname = "br-vrr"
	a.client = ovs.New(
		ovs.Sudo(),
	)

	a.vsctl = a.client.VSwitch
	a.ofctl = a.client.OpenFlow

	a.addBr(a.brname)
	a.addVlanPort("vlan10")
	a.addVlanPort("vlan20")
	a.addVlanPort("vlan30")

	a.delFlows(nil)

	// table=0
	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   0x2021,
		Table:    0,
		Actions: []ovs.Action{
			ovs.Resubmit(0, 20),
		},
	})
	// table=0
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   0x2021,
		Table:    0,
		Actions: []ovs.Action{
			ovs.Drop(),
		},
	})
	// table=20
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   0x2041,
		Table:    20,
		Actions: []ovs.Action{
			ovs.Resubmit(0, 30),
		},
	})
	// table=30
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Table:    30,
		Cookie:   0x2051,
		Actions: []ovs.Action{
			ovs.Normal(),
		},
	})
}

func (a *Composer) findPortId(name string) int {
	vlanid := 0
	fmt.Sscanf(name, "vlan%d", &vlanid)
	return vlanid + 1000
}

type HwAddr string

func (m HwAddr) hex() string {
	return fmt.Sprintf("0x%s", strings.Replace(string(m), ":", "", 5))
}

func (a *Composer) findPortAddr(name string) string {
	if strings.HasPrefix(name, "vlan") {
		return "00:00:00:00:10:00"
	}
	return "unkonwn"
}

func (a *Composer) findVlanId(name string) int {
	vlanid := 0
	fmt.Sscanf(name, "vlan%d", &vlanid)
	return vlanid
}

func (a *Composer) AddRoute(ipdst string, ethdst HwAddr, vlanif string) {
	// table=20
	ethsrc := a.findPortAddr(vlanif)
	vlanid := fmt.Sprintf("0x%x", a.findVlanId(vlanif))
	portid := fmt.Sprintf("0x%x", a.findPortId(vlanif))

	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   0x2041,
		Table:    20,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst),
			ovs.DataLinkDestination(ethsrc),
		},
		Actions: []ovs.Action{
			ovs.Push("NXM_OF_ETH_DST"),
			ovs.Pop("NXM_OF_ETH_SRC"),
			ovs.Load(ethdst.hex(), "NXM_OF_ETH_DST"),
			ovs.Load(vlanid, "NXM_OF_VLAN_TCI"),
			ovs.Load(portid, "NXM_OF_IN_PORT"),
			ovs.Resubmit(0, 30),
		},
	})
}

func (a *Composer) addFlow(flow *ovs.Flow) error {
	err := a.ofctl.AddFlow(a.brname, flow)
	if err != nil {
		log.Printf("failed to add flow: %v\n", err)
	}
	return err
}

func (a *Composer) delFlows(match *ovs.MatchFlow) error {
	err := a.ofctl.DelFlows(a.brname, match)
	if err != nil {
		log.Printf("failed to add flow: %v\n", err)
	}
	return err
}
