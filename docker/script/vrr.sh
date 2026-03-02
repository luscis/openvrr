#!/bin/bash

set -ex

ip netns exec vrr ip l || ip netns add vrr

# wait ovs-vswitchd service
while ! find /run/openvswitch -name "ovs-vswitchd.*.ctl"; do
  sleep 5
done

/usr/bin/openvrr