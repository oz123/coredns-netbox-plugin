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
	"context"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"gopkg.in/h2non/gock.v1"
)

var hostWithIPv4 = `{"results": [{"family": {"value": 4, "label": "IPv4"},
                                 "address": "10.0.0.2/25", "dns_name": "my_host"}]}`

func TestNetbox(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(hostWithIPv4)
	nb := newNetbox()
	nb.Url = "https://example.org/api/ipam/ip-addresses"
	nb.Token = "s3kr3tt0ken"
	nb.CacheDuration = time.Second * 10

	if nb.Name() != "netbox" {
		t.Errorf("expected plugin name: %s, got %s", "netbox", nb.Name())
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("my_host.", dns.TypeA)

	rcode, err := nb.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
	}
	IP := rec.Msg.Answer[0].(*dns.A).A.String()

	if IP != "10.0.0.2" {
		t.Errorf("Expected %v, got %v", "10.0.0.2", IP)
	}

}
