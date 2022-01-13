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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var (
	anotherHostWithIPv4 = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                         "address": "10.0.0.2/25", "dns_name": "my_host"}]}`

	hostWithIPv6 = `{"results": [{"family": {"value": 6, "label": "IPv6"},
                                  "address": "fe80::250:56ff:fe3d:83af/64", "dns_name": "mail.foo.com"}]}`
	hostWithMultipleAddresses = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                              "address": "10.0.0.1/25", "dns_name": "mail.foo.com"},
                                              {"family": {"value": 6, "label": "IPv6"},
                                               "address": "fe80::250:56ff:fe3d:83af/64",
                                               "dns_name": "mail.foo.com"}]}`
)

func TestQuery(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(anotherHostWithIPv4)

	want := "10.0.0.2"
	got, _ := query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100, 4)
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}

}

// dig google.com AAAA +short
func TestQueryIPv6(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "mail.foo.com"}).Reply(
		200).BodyString(hostWithIPv6)

	want := "fe80::250:56ff:fe3d:83af"
	got, _ := query("https://example.org/api/ipam/ip-addresses", "mytoken", "mail.foo.com", time.Millisecond*100, 6)
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}

}

func TestQueryMultipleAddresses(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "mail.foo.com"}).Reply(
		200).BodyString(hostWithMultipleAddresses)
	want := "fe80::250:56ff:fe3d:83af"
	got, _ := query("https://example.org/api/ipam/ip-addresses", "mytoken", "mail.foo.com", time.Millisecond*100, 6)
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}

}
func TestNoSuchHost(t *testing.T) {

	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(
		200).BodyString(`{"count":0,"next":null,"previous":null,"results":[]}`)

	want := ""
	got, _ := query("https://example.org/api/ipam/ip-addresses", "mytoken", "NoSuchHost", time.Millisecond*100, 4)
	if got != want {
		t.Fatalf("Expected empty string but got %s", got)
	}

}

func TestLocalCache(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(anotherHostWithIPv4)

	ip_address := ""

	got, _ := query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100, 4)

	item, err := localCache.Get("my_host")
	if err == nil {
		ip_address = item.Value().(string)
	}

	assert.Equal(t, got, ip_address, "local cache item didn't match")

}

func TestLocalCacheExpiration(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(anotherHostWithIPv4)

	_, _ = query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100, 4)
	<-time.After(101 * time.Millisecond)
	item, err := localCache.Get("my_host")
	if err != nil {
		t.Fatalf("Expected errors, but got: %v", item)
	}
}
