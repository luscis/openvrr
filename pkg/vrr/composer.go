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
	TableCt  = 10
	TableNat = 12
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

func (a *Composer) delVlanTag(port string) error {
	if !a.hasPort(port) {
		return nil
	}

	if err := a.vsctl.ClearPort(port, "tag"); err != nil {
		log.Printf("Composer.delVlanTag: %v\n", err)
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

func (a *Composer) delVlanTrunks(port string) error {
	if !a.hasPort(port) {
		return nil
	}

	if err := a.vsctl.ClearPort(port, "trunks"); err != nil {
		log.Printf("Composer.delVlanTrunks: %v\n", err)
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
	tag := a.findVlanId(vlan)
	if tag > 0 {
		ps := ovs.PortOptions{Tag: tag}
		if err := a.vsctl.Set.Port(vlan, ps); err != nil {
			log.Printf("Composer.addVlanPort.set.port: %v\n", err)
			return err
		}
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
			ovs.Resubmit(0, TableCt),
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
	// table=10 CT
	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieIn,
		Table:    TableCt,
		Protocol: ovs.ProtocolIPv4,
		Actions: []ovs.Action{
			ovs.ConnectionTracking(fmt.Sprintf("nat,zone=10,table=%d", TableNat)),
		},
	})
	// table=12 NAT
	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateReply),
			),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, TableRib),
		},
	})
	a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateEstablished),
			),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, TableRib),
		},
	})
	a.addFlow(&ovs.Flow{
		Priority: 10,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Actions: []ovs.Action{
			ovs.Resubmit(0, TableRib),
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
			ovs.Resubmit(0, TableFib),
		},
	})
	// table=20 FIB
	a.addFlow(&ovs.Flow{
		Priority: 0,
		Cookie:   CookieFib,
		Table:    TableFib,
		Actions: []ovs.Action{
			ovs.Load("0x0", "reg0"),
			ovs.Resubmit(0, TableFdb),
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
	if strings.HasPrefix(name, "vlan") {
		vlanid := 0
		fmt.Sscanf(name, "vlan%d", &vlanid)
		return vlanid + 32768
	}
	return 0
}

func (a *Composer) findPortAddr(name string) string {
	if strings.HasPrefix(name, "vlan") {
		return DefaultVlanMac
	}
	return ""
}

func (a *Composer) findVlanId(name string) int {
	vlanid := 0
	fmt.Sscanf(name, "vlan%d", &vlanid)
	return vlanid
}

func (a *Composer) AddHost(ipdst IPAddr, ethdst HwAddr, vlanif string) error {
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
			ovs.Resubmit(0, TableFdb),
		},
	})
}

func (a *Composer) DelHost(ipdst IPAddr, vlanif string) error {
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

func (a *Composer) AddRoute(ipdst IPPrefix, ipgw IPAddr, vlanif string) error {
	// table=19 RIB
	log.Printf("Compose.AddRoute: %s -> %s on %s", ipdst, ipgw, vlanif)
	ethsrc := a.findPortAddr(vlanif)

	var actions []ovs.Action
	if ipgw == "<nil>" {
		actions = []ovs.Action{
			ovs.Push("OXM_OF_IPV4_DST"),
			ovs.Pop("reg0"),
			ovs.Resubmit(0, TableFib),
		}
	} else {
		actions = []ovs.Action{
			ovs.Load(ipgw.Hex(), "reg0"),
			ovs.Resubmit(0, TableFib),
		}
	}
	a.addFlow(&ovs.Flow{
		Priority: 100 + ipdst.Prefixlen(),
		Cookie:   CookieRib,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst.Str()),
			ovs.DataLinkDestination(ethsrc),
		},
		Actions: actions,
	})
	return nil
}

func (a *Composer) DelRoute(ipdst IPPrefix, vlanif string) error {
	// table=19 RIB
	log.Printf("Compose.DelRoute: %s on %s", ipdst, vlanif)
	ethsrc := a.findPortAddr(vlanif)

	return a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieRib,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst.Str()),
			ovs.DataLinkDestination(ethsrc),
		},
	})
}

func (a *Composer) AddSNAT(source, sourceTo string) error {
	log.Printf("Compose.AddSNAT: %s -> %s", source, sourceTo)

	return a.addFlow(&ovs.Flow{
		Priority: 60,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateNew),
			),
			ovs.NetworkSource(source),
		},
		Actions: []ovs.Action{
			ovs.ConnectionTracking(fmt.Sprintf("commit,nat(src=%s),zone=10,table=%d", sourceTo, TableRib)),
		},
	})
}

func (a *Composer) DelSNAT(source string) error {
	log.Printf("Compose.DelSNAT: %s", source)

	return a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateNew),
			),
			ovs.NetworkSource(source),
		},
	})
}

func parseDest(data string) (string, uint16, error) {
	var dAddr string
	var dport uint16

	valus := strings.SplitN(data, ":", 2)
	if len(valus) != 2 {
		return "", 0, fmt.Errorf("invalid destination: %s", data)
	}
	dAddr = valus[0]
	if _, err := fmt.Sscanf(valus[1], "%d", &dport); err != nil {
		return "", 0, fmt.Errorf("invalid destination port: %v", err)
	}
	return dAddr, dport, nil
}

func (a *Composer) AddDNAT(dest, destTo string) error {
	daddr, dport, err := parseDest(dest)
	if err != nil {
		return err
	}
	log.Printf("Compose.AddDNAT: %s -> %s", dest, destTo)

	return a.addFlow(&ovs.Flow{
		Priority: 80,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolTCPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateNew),
			),
			ovs.NetworkDestination(daddr),
			ovs.TransportDestinationPort(dport),
		},
		Actions: []ovs.Action{
			ovs.ConnectionTracking(fmt.Sprintf("commit,nat(dst=%s),zone=10,table=%d", destTo, TableRib)),
		},
	})
}

func (a *Composer) DelDNAT(dest string) error {
	daddr, dport, err := parseDest(dest)
	if err != nil {
		return err
	}
	log.Printf("Compose.DelDNAT: %s", dest)

	return a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolTCPv4,
		Matches: []ovs.Match{
			ovs.ConnectionTrackingState(
				ovs.SetState(ovs.CTStateTracked),
				ovs.SetState(ovs.CTStateNew),
			),
			ovs.NetworkDestination(daddr),
			ovs.TransportDestinationPort(dport),
		},
	})
}

type HwAddr string

func (m HwAddr) Hex() string {
	return fmt.Sprintf("0x%s", strings.Replace(string(m), ":", "", 5))
}

type IPAddr string

func (i IPAddr) Hex() string {
	addr := net.ParseIP(string(i))
	if addr != nil {
		bytes := addr.To4()
		return fmt.Sprintf("0x%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3])
	}
	return ""
}

func (i IPAddr) Str() string {
	return string(i)
}

type IPPrefix string

func (i IPPrefix) Str() string {
	return string(i)
}

func (i IPPrefix) Prefixlen() int {
	_, ipnet, err := net.ParseCIDR(string(i))
	if err != nil {
		return 0
	}
	ones, _ := ipnet.Mask.Size()
	return ones
}
