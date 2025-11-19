package vrr

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/luscis/openvrr/pkg/ovs"
)

const (
	TableIn  = 0
	TableRib = 19
	TableFib = 20
	TableFdb = 30
)

const (
	CookieIn  = 0x2021
	CookieRib = 0x2022
	CookieFib = 0x2023
	CookieFdb = 0x2024
)

const (
	DefaultVlanMac = "00:00:00:00:20:15"
)

type Composer struct {
	brname string
	client *ovs.Client
	ofctl  *ovs.OpenFlowService
	vsctl  *ovs.VSwitchService
}

func (a *Composer) Start() {
	log.Printf("Composer.Start")
}

func (a *Composer) listPorts() ([]ovs.PortData, error) {
	ports, err := a.vsctl.ListPorts(a.brname)
	if err != nil {
		log.Printf("Composer.listPorts: %v\n", err)
		return nil, err
	}

	var items []ovs.PortData
	for _, port := range ports {
		data, err := a.vsctl.Get.Port(port)
		if err != nil {
			continue
		}
		items = append(items, data)
	}

	return items, nil
}

func (a *Composer) hasPort(name string) bool {
	ports, err := a.vsctl.ListPorts(a.brname)
	if err != nil {
		log.Printf("Composer.hasPort: %v\n", err)
		return false
	}

	for _, port := range ports {
		if port == name {
			return true
		}
	}
	return false
}

func (a *Composer) addVlanTag(port string, tag int) error {
	if !a.hasPort(port) {
		if err := a.vsctl.AddPort(a.brname, port); err != nil {
			log.Printf("Composer.addVlanTag.add: %v\n", err)
			return err
		}
	}

	ps := ovs.PortOptions{Tag: tag}
	if err := a.vsctl.Set.Port(port, ps); err != nil {
		log.Printf("Composer.addVlanTag.set: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) addVlanTrunks(port, trunks string) error {
	if !a.hasPort(port) {
		if err := a.vsctl.AddPort(a.brname, port); err != nil {
			log.Printf("Composer.addVlanTrunks.add: %v\n", err)
			return err
		}
	}
	ps := ovs.PortOptions{Trunks: trunks}
	if err := a.vsctl.Set.Port(port, ps); err != nil {
		log.Printf("Composer.addVlanTrunks.set: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) addVlanPort(vlan string) error {
	is := ovs.InterfaceOptions{
		OfportRequest: a.findPortId(vlan),
		Mac:           a.findPortAddr(vlan),
		Type:          ovs.InterfaceTypeInternal,
	}
	if err := a.vsctl.AddPortWith(a.brname, vlan, is); err != nil {
		log.Printf("Composer.addVlanPort.add: %v\n", err)
		return err
	}
	if err := a.vsctl.Set.Interface(vlan, is); err != nil {
		log.Printf("Composer.addVlanPort.set.interface: %v\n", err)
		return err
	}

	ps := ovs.PortOptions{Tag: a.findVlanId(vlan)}
	if err := a.vsctl.Set.Port(vlan, ps); err != nil {
		log.Printf("Composer.addVlanPort.set.port: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) delPort(vlan string) error {
	if err := a.vsctl.DeletePort(a.brname, vlan); err != nil {
		log.Printf("Composer.delPort: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) addBr(name string) error {
	if err := a.vsctl.AddBridge(name); err != nil {
		log.Fatalf("Composer.addBr: %v\n", err)
		return err
	}
	return nil
}

func (a *Composer) Init() {
	a.client = ovs.New(
		ovs.Sudo(),
	)

	a.vsctl = a.client.VSwitch
	a.ofctl = a.client.OpenFlow

	a.addBr(a.brname)
	a.delFlows(nil)

	// table=0 IN
	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieIn,
		Protocol: ovs.ProtocolIPv4,
		Table:    TableIn,
		Actions: []ovs.Action{
			ovs.Resubmit(0, 19),
		},
	})
	// table=0 IN
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   CookieIn,
		Table:    TableIn,
		Actions: []ovs.Action{
			ovs.Normal(),
		},
	})
	// table=19 RIB
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   CookieRib,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Actions: []ovs.Action{
			ovs.Push("OXM_OF_IPV4_DST"),
			ovs.Pop("reg0"),
			ovs.Resubmit(0, 20),
		},
	})
	// table=20 FIB
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   CookieFib,
		Table:    TableFib,
		Actions: []ovs.Action{
			ovs.Load("0x0", "reg0"),
			ovs.Resubmit(0, 30),
		},
	})
	// table=30 FDB
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Table:    TableFdb,
		Cookie:   CookieFdb,
		Actions: []ovs.Action{
			ovs.Normal(),
		},
	})
}

func (a *Composer) findPortId(name string) int {
	vlanid := 0
	fmt.Sscanf(name, "vlan%d", &vlanid)
	return vlanid + 32768
}

func (a *Composer) findPortAddr(name string) string {
	if strings.HasPrefix(name, "vlan") {
		return DefaultVlanMac
	}
	return "unkonwn"
}

func (a *Composer) findVlanId(name string) int {
	vlanid := 0
	fmt.Sscanf(name, "vlan%d", &vlanid)
	return vlanid
}

func (a *Composer) AddHost(ipdst IpAddr, ethdst HwAddr, vlanif string) error {
	// table=20 FIB
	log.Printf("Compose.AddHost: %s -> %s on %s", ipdst, ethdst, vlanif)
	ethsrc := a.findPortAddr(vlanif)
	vlanid := fmt.Sprintf("0x%x", a.findVlanId(vlanif))
	portid := fmt.Sprintf("0x%x", a.findPortId(vlanif))

	return a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieFib,
		Table:    TableFib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.FieldMatch("reg0", ipdst.Hex()),
			ovs.DataLinkDestination(ethsrc),
		},
		Actions: []ovs.Action{
			ovs.Push("NXM_OF_ETH_DST"),
			ovs.Pop("NXM_OF_ETH_SRC"),
			ovs.Load(ethdst.Hex(), "NXM_OF_ETH_DST"),
			ovs.Load(vlanid, "NXM_OF_VLAN_TCI"),
			ovs.Load(portid, "NXM_OF_IN_PORT"),
			ovs.DecTTL(),
			ovs.Resubmit(0, 30),
		},
	})
}

func (a *Composer) DelHost(ipdst IpAddr, vlanif string) error {
	log.Printf("Compose.DelHost: %s on %s", ipdst, vlanif)
	ethsrc := a.findPortAddr(vlanif)

	return a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieFib,
		Table:    TableFib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.FieldMatch("reg0", ipdst.Hex()),
			ovs.DataLinkDestination(ethsrc),
		},
	})
}

func (a *Composer) addFlow(flow *ovs.Flow) error {
	err := a.ofctl.AddFlow(a.brname, flow)
	if err != nil {
		log.Printf("Composer.addFlow: %v\n", err)
	}
	return err
}

func (a *Composer) delFlows(match *ovs.MatchFlow) error {
	err := a.ofctl.DelFlows(a.brname, match)
	if err != nil {
		log.Printf("Composer.delFlow: %v\n", err)
	}
	return err
}

func (a *Composer) AddRoute(ipdst IpPrefix, ipgw IpAddr, vlanif string) error {
	// table=19 RIB
	log.Printf("Compose.AddRoute: %s -> %s on %s", ipdst, ipgw, vlanif)
	ethsrc := a.findPortAddr(vlanif)

	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieRib,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst.Str()),
			ovs.DataLinkDestination(ethsrc),
		},
		Actions: []ovs.Action{
			ovs.Load(ipgw.Hex(), "reg0"),
			ovs.Resubmit(0, 20),
		},
	})
	return nil
}

func (a *Composer) DelRoute(ipdst IpPrefix, vlanif string) error {
	// table=19 RIB
	log.Printf("Compose.DelRoute: %s on %s", ipdst, vlanif)
	ethsrc := a.findPortAddr(vlanif)

	a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieRib,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst.Str()),
			ovs.DataLinkDestination(ethsrc),
		},
	})
	return nil
}

type HwAddr string

func (m HwAddr) Hex() string {
	return fmt.Sprintf("0x%s", strings.Replace(string(m), ":", "", 5))
}

type IpAddr string

func (i IpAddr) Hex() string {
	addr := net.ParseIP(string(i))
	if addr != nil {
		bytes := addr.To4()
		return fmt.Sprintf("0x%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3])
	}
	return ""
}

func (i IpAddr) Str() string {
	return string(i)
}

type IpPrefix string

func (i IpPrefix) Str() string {
	return string(i)
}
