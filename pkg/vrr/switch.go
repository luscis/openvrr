package vrr

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/luscis/openvrr/pkg/ovs"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
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
	CookieIn = 0x2021
)

const (
	DefaultVlanMac = "00:00:00:00:20:15"
)

type Composer struct {
	brname string
	client *ovs.Client
	ofctl  *ovs.OpenFlowService
	vsctl  *ovs.VSwitchService
	ns     netns.NsHandle
}

func (a *Composer) Start() {
	log.Printf("Composer.Start")
	if options, err := a.vsctl.Get.Bridge(a.brname); err == nil {
		for key, value := range options.OtherConfig {
			if source, found := strings.CutPrefix(key, "snat-"); found {
				a.addSNAT(source, value)
			} else if dest, found := strings.CutPrefix(key, "dnat-"); found {
				dest = strings.Replace(dest, "-", ":", 2)
				a.addDNAT(dest, value)
			}
		}
	} else {
		log.Printf("Composer.Start: bridge options: %+v", err)
	}
}

func (a *Composer) listPorts() ([]ovs.PortData, error) {
	ports, err := a.vsctl.ListPorts(a.brname)
	if err != nil {
		log.Printf("Composer.listPorts: %v", err)
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
		log.Printf("Composer.hasPort: %v", err)
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
			log.Printf("Composer.addVlanTag.add: %v", err)
			return err
		}
	}

	ps := ovs.PortOptions{Tag: tag}
	if err := a.vsctl.Set.Port(port, ps); err != nil {
		log.Printf("Composer.addVlanTag.set: %v", err)
		return err
	}
	return nil
}

func (a *Composer) delVlanTag(port string) error {
	if !a.hasPort(port) {
		return nil
	}

	if err := a.vsctl.ClearPort(port, "tag"); err != nil {
		log.Printf("Composer.delVlanTag: %v", err)
		return err
	}
	return nil
}

func (a *Composer) addVlanTrunks(port, trunks string) error {
	if !a.hasPort(port) {
		if err := a.vsctl.AddPort(a.brname, port); err != nil {
			log.Printf("Composer.addVlanTrunks.add: %v", err)
			return err
		}
	}
	ps := ovs.PortOptions{Trunks: trunks}
	if err := a.vsctl.Set.Port(port, ps); err != nil {
		log.Printf("Composer.addVlanTrunks.set: %v", err)
		return err
	}
	return nil
}

func (a *Composer) delVlanTrunks(port string) error {
	if !a.hasPort(port) {
		return nil
	}

	if err := a.vsctl.ClearPort(port, "trunks"); err != nil {
		log.Printf("Composer.delVlanTrunks: %v", err)
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
		log.Printf("Composer.addVlanPort: add: %v", err)
		return err
	}
	if err := a.vsctl.Set.Interface(vlan, is); err != nil {
		log.Printf("Composer.addVlanPort: set interface: %v", err)
		return err
	}
	tag := a.findVlanId(vlan)
	if tag > 0 {
		ps := ovs.PortOptions{Tag: tag}
		if err := a.vsctl.Set.Port(vlan, ps); err != nil {
			log.Printf("Composer.addVlanPort: set port: %v", err)
			return err
		}
		link, err := netlink.LinkByName(vlan)
		if err != nil {
			log.Printf("Composer.addVlanPort: find port: %v", err)
			return err
		}
		err = netlink.LinkSetNsFd(link, int(a.ns))
		if err != nil {
			log.Printf("Composer.addVlanPort: set netns failed: %v", err)
		}
	}
	return nil
}

func (a *Composer) delPort(vlan string) error {
	if err := a.vsctl.DeletePort(a.brname, vlan); err != nil {
		log.Printf("Composer.delPort: %v", err)
		return err
	}
	return nil
}

func (a *Composer) addBr(name string) error {
	if err := a.vsctl.AddBridge(name); err != nil {
		log.Fatalf("Composer.addBr: %v", err)
		return err
	}
	return nil
}

func (a *Composer) Init() {
	a.client = ovs.New()

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
		Priority: 200,
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
		Priority: 200,
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
		Cookie:   CookieIn,
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
		Cookie:   CookieIn,
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
		Cookie:   CookieIn,
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

func (a *Composer) AddHost(ipdst IPAddr, ethdst HWAddr, vlanif string) error {
	// table=20 FIB
	log.Printf("Compose.AddHost: %s -> %s on %s", ipdst, ethdst, vlanif)
	ethsrc := a.findPortAddr(vlanif)
	vlanid := fmt.Sprintf("0x%x", a.findVlanId(vlanif))
	portid := fmt.Sprintf("0x%x", a.findPortId(vlanif))

	return a.addFlow(&ovs.Flow{
		Priority: 100,
		Cookie:   CookieIn,
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
		Cookie:   CookieIn,
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
		log.Printf("Composer.addFlow: %v", err)
	}
	return err
}

func (a *Composer) delFlows(match *ovs.MatchFlow) error {
	err := a.ofctl.DelFlows(a.brname, match)
	if err != nil {
		log.Printf("Composer.delFlow: %v", err)
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
		Cookie:   CookieIn,
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
		Cookie:   CookieIn,
		Table:    TableRib,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(ipdst.Str()),
			ovs.DataLinkDestination(ethsrc),
		},
	})
}

func (a *Composer) addSNAT(source, sourceTo string) error {
	log.Printf("Compose.addSNAT: %s -> %s", source, sourceTo)
	return a.addFlow(&ovs.Flow{
		Priority: 50,
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

func (a *Composer) AddSNAT(source, sourceTo string) error {
	err := a.addSNAT(source, sourceTo)
	if err == nil {
		a.vsctl.Set.Bridge(a.brname, ovs.BridgeOptions{
			OtherConfig: map[string]string{toKey("snat", source): sourceTo},
		})
	}
	return err
}

func (a *Composer) DelSNAT(source string) error {
	log.Printf("Compose.DelSNAT: %s", source)

	err := a.delFlows(&ovs.MatchFlow{
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
	if err == nil {
		a.vsctl.RemoveBridge(a.brname, "other_config", toKey("snat", source))
	}
	return err
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

func (a *Composer) addDNAT(dest, destTo string) error {
	daddr, dport, err := parseDest(dest)
	if err != nil {
		return err
	}
	log.Printf("Compose.addDNAT: %s -> %s", dest, destTo)
	return a.addFlow(&ovs.Flow{
		Priority: 160,
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

func toKey(prefix, value string) string {
	key := fmt.Sprintf("%s-%s", prefix, value)
	key = strings.Replace(key, ":", "-", 2)
	return key
}

func (a *Composer) AddDNAT(dest, destTo string) error {
	err := a.addDNAT(dest, destTo)
	if err == nil {

		a.vsctl.Set.Bridge(a.brname, ovs.BridgeOptions{
			OtherConfig: map[string]string{toKey("dnat", dest): destTo},
		})
	}
	return err
}

func (a *Composer) DelDNAT(dest string) error {
	daddr, dport, err := parseDest(dest)
	if err != nil {
		return err
	}
	log.Printf("Compose.DelDNAT: %s", dest)

	err = a.delFlows(&ovs.MatchFlow{
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
	if err == nil {
		a.vsctl.RemoveBridge(a.brname, "other_config", toKey("dnat", dest))
	}
	return err
}

func (a *Composer) AddLocal(addr string) error {
	log.Printf("Compose.AddLocal: %s", addr)
	host := strings.SplitN(addr, "/", 2)[0]
	return a.addFlow(&ovs.Flow{
		Priority: 150,
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(host),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, TableRib),
		},
	})
}

func (a *Composer) DelLocal(addr string) error {
	log.Printf("Compose.DelLocal: %s", addr)
	host := strings.SplitN(addr, "/", 2)[0]
	return a.delFlows(&ovs.MatchFlow{
		Cookie:   CookieIn,
		Table:    TableNat,
		Protocol: ovs.ProtocolIPv4,
		Matches: []ovs.Match{
			ovs.NetworkDestination(host),
		},
	})
}

type HWAddr string

func (m HWAddr) Hex() string {
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
