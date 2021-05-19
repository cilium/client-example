// Copyright 2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build go1.16

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cilium/cilium/pkg/client"
)

func main() {
	// Connect to the default path /var/run/cilium/cilium.sock
	c, err := client.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	// List all endpoints
	eps, err := c.EndpointList()
	if err != nil {
		panic(err)
	}

	// Sort EPs per IDs
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].ID < eps[j].ID
	})

	// Print the IPs of the endpoints
	for _, ep := range eps {
		var v4s, v6s []string
		for _, ip := range ep.Status.Networking.Addressing {
			if ip.IPV4 != "" {
				v4s = append(v4s, ip.IPV4)
			}
			if ip.IPV4 != "" {
				v6s = append(v6s, ip.IPV6)
			}
		}
		ips := strings.Join(v4s, ", ")
		if ips != "" {
			fmt.Printf("EP ID %d has IP addresses: %s\n", ep.ID, ips)
		} else {
			fmt.Printf("EP ID %d does not have an IP address\n", ep.ID)
		}
	}
}
