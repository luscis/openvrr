#!/bin/bash

set -ex

# wait ovs-vswitchd service
while ! find /run/openvswitch -name "ovs-vswitchd.*.ctl"; do
  sleep 5
done

/usr/bin/openvrr