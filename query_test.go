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
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var (
	noCacheHost = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                         "address": "10.0.0.2/24", "dns_name": "no_cache"}]}`
	expireCacheHost = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                         "address": "10.0.0.2/24", "dns_name": "expire_cache"}]}`
	anotherHostWithIPv4 = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                         "address": "10.0.0.2/24", "dns_name": "my_host"}]}`

	hostWithIPv6 = `{"results": [{"family": {"value": 6, "label": "IPv6"},
                                  "address": "fe80::250:56ff:fe3d:83af/64", "dns_name": "mail.foo.com"}]}`
	hostWithMultipleAddressFamilies = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                              "address": "10.0.0.1/24", "dns_name": "mail.foo.com"},
                                              {"family": {"value": 6, "label": "IPv6"},
                                               "address": "fe80::250:56ff:fe3d:83af/64",
                                               "dns_name": "mail.foo.com"}]}`
	hostWithMultipleAddresses = `{"results": [{"family": {"value": 4, "label": "IPv4"},
											   "address": "10.0.0.1/24", "dns_name": "mail.foo.com"},
											   {"family": {"value": 4, "label": "IPv4"},
											   "address": "10.0.0.2/24", "dns_name": "mail.foo.com"}]}`
	n = Netbox{
		Url:   "https://example.org/api/ipam/ip-addresses",
		Token: "mytoken",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
)

func TestQuery(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(anotherHostWithIPv4)

	want := []net.IP{net.ParseIP("10.0.0.2")}
	got, _ := n.query("my_host", familyIP4)
	assert.Equal(t, want, got, "basic IPv4 query did not match")
}

func TestQueryIPv6(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "mail.foo.com"}).Reply(
		200).BodyString(hostWithIPv6)

	want := []net.IP{net.ParseIP("fe80::250:56ff:fe3d:83af")}
	got, _ := n.query("mail.foo.com", familyIP6)
	assert.Equal(t, want, got, "basic IPv6 query did not match")
}

func TestQueryMultipleAddresses(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "mail.foo.com"}).Reply(
		200).BodyString(hostWithMultipleAddresses)
	want := []net.IP{
		net.ParseIP("10.0.0.1"),
		net.ParseIP("10.0.0.2"),
	}
	got, _ := n.query("mail.foo.com", familyIP4)
	assert.Equal(t, want, got, "host with multiple IPv4 addresses did not match")
}

func TestQueryMultipleAddressFamilies(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "mail.foo.com"}).Reply(
		200).BodyString(hostWithMultipleAddressFamilies)
	want := []net.IP{net.ParseIP("fe80::250:56ff:fe3d:83af")}
	got, _ := n.query("mail.foo.com", familyIP6)
	assert.Equal(t, want, got, "host with both IPv4 and IPv6 addresses did not match")
}

func TestNoSuchHost(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(
		200).BodyString(`{"count":0,"next":null,"previous":null,"results":[]}`)

	want := make([]net.IP, 0)
	got, _ := n.query("NoSuchHost", familyIP4)
	assert.Equal(t, want, got, "missing host response was not empty")
}
