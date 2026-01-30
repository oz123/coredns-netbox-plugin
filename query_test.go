// Lucas Kirsche
// Copyright 2020 Oz Tiram <oz.tiram@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netbox

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestQuery(t *testing.T) {
	// set up dummy Netbox
	n := newNetbox()
	n.Url = "https://example.org"
	n.Token = "mytoken"

	tests := []struct {
		name    string
		host    string
		body    string
		family  int
		wantErr bool
		want    []net.IP
	}{
		{
			"IPv4 Only Host",
			"host1",
			`{
				"results": [
					{"family": {"value": 4, "label": "IPv4"}, "address": "10.0.0.1/24", "dns_name": "host1"}
				]
			}`,
			familyIP4,
			false,
			[]net.IP{
				net.ParseIP("10.0.0.1"),
			},
		},
		{
			"IPv6 Only Host",
			"host2",
			`{
				"results": [
					{"family": {"value": 6, "label": "IPv6"}, "address": "fe80::250:56ff:fe3d:83af/64", "dns_name": "host2"}
				]
			}`,
			familyIP6,
			false,
			[]net.IP{
				net.ParseIP("fe80::250:56ff:fe3d:83af"),
			},
		},
		{
			"IPv4 Query with Dual Stack Host",
			"host3",
			`{
				"results": [
					{"family": {"value": 4, "label": "IPv4"}, "address": "10.0.0.1/24", "dns_name": "host3"},
					{"family": {"value": 6, "label": "IPv6"}, "address": "fe80::250:56ff:fe3d:83af/64", "dns_name": "host3"}
				]
			}`,
			familyIP4,
			false,
			[]net.IP{
				net.ParseIP("10.0.0.1"),
			},
		},
		{
			"Multiple IPv4 addresses",
			"host4",
			`{
				"results": [
					{"family": {"value": 4, "label": "IPv4"}, "address": "10.0.0.1/24", "dns_name": "host4"},
					{"family": {"value": 4, "label": "IPv4"}, "address": "10.0.0.2/24", "dns_name": "host4"}
				]
			}`,
			familyIP4,
			false,
			[]net.IP{
				net.ParseIP("10.0.0.1"),
				net.ParseIP("10.0.0.2"),
			},
		},
		{
			"Not found",
			"host5",
			`{"results": []}`,
			familyIP4,
			false,
			[]net.IP{},
		},
	}

	defer gock.Off() // Flush pending mocks after test execution

	// set up mock responses
	for _, tt := range tests {
		gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
			map[string]string{"dns_name": tt.host}).Reply(
			200).BodyString(tt.body)
	}

	// run tests
	for _, tt := range tests {
		got, err := n.query(tt.host, tt.family)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.name)
		}
	}
}

func TestReverseQuery(t *testing.T) {
	// set up dummy Netbox
	n := newNetbox()
	n.Url = "https://example.org"
	n.Token = "mytoken"

	tests := []struct {
		name    string
		reverse string
		body    string
		wantErr bool
		want    []string
	}{
		{
			"Reverse Query",
			"2.0.0.10.in-addr.arpa.",
			`{
				"results": [
					{"address": "10.0.0.2", "dns_name": "domain.com"}
				]
			}`,
			false,
			[]string{
				"domain.com.",
			},
		},
	}

	defer gock.Off() // Flush pending mocks after test execution

	// set up mock responses
	for _, tt := range tests {
		gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
			map[string]string{"address": "10.0.0.2"}).Reply(
			200).BodyString(tt.body)
	}

	// run tests
	for _, tt := range tests {
		got, err := n.queryreverse(tt.reverse)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			fmt.Printf("%#v\n", got)
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.reverse)
		}
	}
}
