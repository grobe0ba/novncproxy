// Copyright (c) 2020 Byron Grobe
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package vm

import (
	"encoding/json"
	"log"
)

type Disk struct {
	Boot  bool   `json:"boot"`
	Size  uint   `json:"size"`
	Model string `json:"model,omitempty"`
}

type NIC struct {
	Gateway string `json:"gateway"`
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Model   string `json:"model,omitempty"`
	MTU     uint   `json:"mtu"`
	Tag     string `json:"nic_tag"`
	VLAN    uint   `json:"vlan_id,omitempty"`
	Primary bool   `json:"primary"`
}

type VNC struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Display int    `json:"display"`
}

type VM struct {
	Alias     string   `json:"alias,omitempty"`
	Autoboot  bool     `json:"autoboot"`
	Boot      string   `json:"boot"`
	Brand     string   `json:"brand"`
	Disks     []Disk   `json:"disks"`
	DNSDomain string   `json:"dns_domain,omitempty"`
	Hostname  string   `json:"hostname,omitempty"`
	NICs      []NIC    `json:"nics,omitempty"`
	RAM       uint     `json:"ram"`
	Resolvers []string `json:"resolvers,omitempty"`
	VCPUs     uint     `json:"vcpus"`
	CPUType   string   `json:"cpu_type,omitempty"`
	UUID      string   `json:"uuid"`
	Type      string   `json:"type"`  // read-only: joyent*, lx, kvm, bhyve
	State     string   `json:"state"` // read-only: failed, stopped, running
	VNC       VNC      `json:"vnc"`   // read-only
}

func FromJSON(data []byte) VM {
	var (
		vm VM
		e  error
	)
	e = json.Unmarshal(data, &vm)
	if e != nil {
		log.Fatal(e)
	}

	return vm
}
