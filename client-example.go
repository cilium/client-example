// Copyright 2017 Authors of Cilium
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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/client"
	log "github.com/sirupsen/logrus"
)

var separator = strings.Repeat("-", 78)

func header(title string) {
	fmt.Printf("%s\n%s\n%s\n", separator, title, separator)
}

func main() {
	c, err := client.NewDefaultClient()
	if err != nil {
		log.WithError(err).Fatal("Cannot create client")
	}

	agentConfiguration, err := c.ConfigGet()
	if err != nil {
		log.WithError(err).Fatal("Cannot get agent configuration")
	}

	if result, err := json.MarshalIndent(agentConfiguration, "", "  "); err == nil {
		header("Agent configuration:")
		fmt.Println(string(result))
	}

	response, err := c.Daemon.GetHealthz(nil)
	if err != nil {
		log.WithError(err).Fatal("Cannot get agent status")
	}

	statusResponse := response.Payload
	if statusResponse.Cilium != nil && statusResponse.Cilium.State != models.StatusStateOk {
		log.Info("Cilium agent status not OK, aborting")
		os.Exit(1)
	}

	endpoints, err := c.EndpointList()
	if err != nil {
		log.WithError(err).Fatal("Cannot list endpoints")
	}

	header("List of running endpoints:")
	for _, ep := range endpoints {
		fmt.Printf("%8d %14s %16s %32s\n", ep.ID, ep.ContainerName, ep.Addressing.IPV4, ep.Addressing.IPV6)
	}
}
