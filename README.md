# OpenVRR

OpenVRR is a solution to improve kernel routing and performance by openvswitch.

# Architecture

```
                               [VLANs] [Interfaces] [LACP]
                                            |
    +----------+  +-----------+      +-------------+
    |    FRR   |  |  iproute  |      |   OpenVRR   |-------------------+
    +----------+  +-----------+      +-------------+                   |
         |             |                  ^                            |
         |             |                  | Wath route table changes   | Translate FIB to OpenFlow
         |             |                  |                            |
         +------[netlink]-----------------+          +--------------+  |
                    |                                | openvswith   |<-+
                    |                                +--------------+  |
                    |                                                  |
    +---------------+--------------+     +---------------------+       |
    |         kernel               |<----|        dpdk         |<------+ Fast forwarding by Megaflows
    |        datapath              |     |      datapath       |
    +------------------------------+     +---------------------+
```

# As a Gateway

```
                        +---------+
    [PCa] -- vlan10 --> | OpenVRR | <-- vlan20 --- [PCb]
                        +---------+
```

The PCa access to OpenVRR via vlan 10, and as a vlan 10 subnet gateway.
```
vrrcli interface vlan add --tag 10 --interface eth1
vrrcli interface create --name vlan10

ifconfig vlan10 192.168.1.1/24
```
The PCb access to OpenVRR via vlan 20, and as a vlan 20 subnet gateway..
```
vrrcli interface vlan add --tag 20 --interface eth2
vrrcli interface create --name vlan20

ifconfig vlan20 192.168.2.1/24
```