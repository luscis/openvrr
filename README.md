# OpenVRR

OpenVRR is a solution to improve kernel routing and performance by openvswitch.

## Architecture

```
                             [VLANs] [Interfaces] [LACP]
                                          |
     +----------+  +-----------+        +-------------+
     |   FRR    |  |  iproute  |        |   OpenVRR   |-------------------+
     +----------+  +-----------+        +-------------+                   |
            |             |                  ^                            |
            |             |                  | Wath route table changes   | Translate FIB to OpenFlow
            |             |                  |                            |
            +---------[netlink]--------------+          +--------------+  |
                          |                             | openvswith   |<-+
                          |                             +--------------+  |
                          |                                               |
     +---------------+--------------+      +---------------------+        |
     |            kernel            |<-----|          dpdk       |<-------+ Fast forwarding by Megaflows
     |           datapath           |      |         datapath    |
     +------------------------------+      +---------------------+
```

## As a Gateway

```
                          [internet]
                              |
                             SNAT
                              |
                         +---------+
     [PCa] -- vlan10 --> | OpenVRR | <-- vlan20 --- [PCb]
                         +---------+
                              |
                              |
                             [PCc]
```

The PCa access to OpenVRR via vlan 10, and as a vlan 10 subnet gateway.

```
vrrcli vlan add --tag 10 --interface eth1
vrrcli interface add --name vlan10

ip netns exec vrr ifconfig vlan10 192.168.1.1/24
```

The PCb access to OpenVRR via vlan 20, and as a vlan 20 subnet gateway..

```
vrrcli vlan add --tag 20 --interface eth2
vrrcli interface add --name vlan20

ip netns exec vrr ifconfig vlan20 192.168.2.1/24
```

The OpenVRR connect to internet via vlan 11, and use 10.10.10.0/24 as external network.

```
vrrcli vlan add --tag 11 --interface eth3
vrrcli interface add --name vlan11

ip netns exec vrr ip r add default via 10.10.1.254
ip netns exec vrr ifconfig vlan11 10.10.10.1/24
```

The SNAT rule translates the source IP address of outbound traffic originating from the internal subnet 192.168.1.0/24 to the external IP 10.10.10.1.
```
vrrcli snat add --source 192.168.1.0/24 --source-to 10.10.10.1
```
The DNAT rule redirects all incoming TCP traffic destined for the external IP 10.10.10.1 on port 80 to the internal server 192.168.1.2 on port 8000.
```
vrrcli dnat add --dest 10.10.10.1:80 --dest-to 192.168.1.2:8000 --protocol tcp
```