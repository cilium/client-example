module github.com/cilium/client-example/v1.8

go 1.14

require github.com/cilium/cilium v1.8.10

replace (
	github.com/miekg/dns => github.com/cilium/dns v1.1.4-0.20190417235132-8e25ec9a0ff3
	github.com/optiopay/kafka => github.com/cilium/kafka v0.0.0-20180809090225-01ce283b732b
	// Using cilium/netlink until XFRM patches merged upstream
	github.com/vishvananda/netlink => github.com/cilium/netlink v0.0.0-20210223023818-d826f2a4c934
	k8s.io/client-go => github.com/cilium/client-go v0.0.0-20210417023617-aeb4c6f1b557
)
