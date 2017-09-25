# Cilium client API example

Simple example illustrating use of the Cilium API.

See [Cilium API reference](http://docs.cilium.io/en/stable/api/) for further
documentation on use of the API.

## Example Client Output

Note: As the cilium-agent is running as privileged process, the client must be
running as privileged process as well or the process must belong to group
`cilium`.


    $  sudo ./client-example
    ------------------------------------------------------------------------------
    Agent configuration:
    ------------------------------------------------------------------------------
    {
      "addressing": {
        "ipv4": {
          "alloc-range": "10.15.0.0/16",
          "enabled": true,
          "ip": "10.15.28.238"
        },
        "ipv6": {
          "alloc-range": "f00d::a00:20f:0:0/112",
          "enabled": true,
          "ip": "f00d::a00:20f:0:8ad6"
        }
      },
      "configuration": {
        "mutable": {
          "Conntrack": "Enabled",
          "ConntrackAccounting": "Enabled",
          "ConntrackLocal": "Disabled",
          "Debug": "Enabled",
          "DropNotification": "Enabled",
          "PolicyTracing": "Disabled",
          "TraceNotification": "Enabled"
        }
      },
      "nodeMonitor": {
        "cpus": 2,
        "npages": 64,
        "pagesize": 4096
      },
      "policy-enforcement": "default"
    }
    ------------------------------------------------------------------------------
    List of running endpoints:
    ------------------------------------------------------------------------------
       62006    musing_saha    10.15.116.202             f00d::a00:20f:0:f236
