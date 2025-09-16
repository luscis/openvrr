# OpenVRR

OpenVRR is a solution to improve kernel routing and performance by openvswitch.

# Architecture

```
                                  [VLANs] [Interfaces] [LACP]
                                               |
       +----------+  +-----------+      +-------------+
       |    FRR   |  |  iproute  |      |   openvrr   |-------------------+
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