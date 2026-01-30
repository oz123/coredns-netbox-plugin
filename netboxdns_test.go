// Copyright 2025 Lucas Kirsche <kontakt@lucas-kirsche.de>
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
	"strings"
	"testing"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestQueryDNSPlugin(t *testing.T) {
	// set up dummy Netbox
	n := newNetbox()
	n.Url = "https://example.org"
	n.Token = "mytoken"

	tests := []struct {
		name    string
		zone    string
		fqdn    string
		dnsType DNSRecordType
		body    string
		wantErr bool
		want    []string
	}{
		{
			"Query A Record No TTL",
			"example.com.",
			"mail1.example.com.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "A",
					"ttl": null,
					"value": "192.168.0.1",
					"absolute_value": "192.168.0.1",
					"fqdn": "mail1.example.com."
				}]
			}`,
			false,
			[]string{
				"mail1.example.com.\t0\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query A Record With TTL",
			"example.com.",
			"mail1.example.com.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "A",
					"ttl": 8600,
					"value": "192.168.0.1",
					"absolute_value": "192.168.0.1",
					"fqdn": "mail1.example.com."
				}]
			}`,
			false,
			[]string{
				"mail1.example.com.\t8600\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query AAAA Record",
			"example.com.",
			"mail1.example.com.",
			DNSRecordTypeAAAA,
			`{
				"results": [
				{
					"type": "AAAA",
					"ttl": 8600,
					"value": "2001:db8::1",
					"absolute_value": "2001:db8::1",
					"fqdn": "mail1.example.com."
				}]
			}`,
			false,
			[]string{
				"mail1.example.com.\t8600\tIN\tAAAA\t2001:db8::1",
			},
		},
		{
			"Query CNAME Record",
			"example.com.",
			"test.example.com.",
			DNSRecordTypeCNAME,
			`{
				"results": [
				{
					"type": "CNAME",
					"ttl": 8600,
					"value": "mail1",
					"absolute_value": "mail1.example.com.",
					"fqdn": "test.example.com."
				}]
			}`,
			false,
			[]string{
				"test.example.com.\t8600\tIN\tCNAME\tmail1.example.com.",
			},
		},
		{
			"Query A Record but receive CNAME Record with resolved A record",
			"example.com.",
			"test.example.com.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "CNAME",
					"ttl": 8600,
					"value": "mail1",
					"absolute_value": "mail1.example.com.",
					"fqdn": "test.example.com."
				}]
			}`,
			false,
			[]string{
				"test.example.com.\t8600\tIN\tCNAME\tmail1.example.com.",
				"mail1.example.com.\t8600\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query unsupported record",
			"example.com.",
			"test.example.com.",
			DNSRecordType("BBBB"),
			`{
				"type": [
					"Select a valid choice. BBBB is not one of the available choices."
				]
			}`,
			true,
			[]string{},
		},
		{
			"Query not existing record",
			"example.com.",
			"test2.example.com.",
			DNSRecordType("A"),
			`{
				"results": []
			}`,
			false,
			[]string{},
		},
	}

	// set up mock responses
	for _, tt := range tests {
		gock.New("https://example.org/api/plugins/netbox-dns/records/").MatchParams(
			map[string]string{
				"zone":   strings.TrimRight(tt.zone, "."),
				"active": "true",
				"fqdn":   tt.fqdn,
				"type":   string(tt.dnsType),
			}).Reply(
			200).BodyString(tt.body)
	}

	// run tests
	for _, tt := range tests {
		r := new(dns.Msg)
		r.SetQuestion(tt.fqdn, DNSRecordReverseMap[tt.dnsType])
		responses, _, err := n.queryDNSPlugin(tt.zone, request.Request{Req: r})

		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			if len(tt.want) != 0 && assert.NotEmpty(t, responses) {
				for i, response := range responses {
					assert.Equal(t, tt.want[i], response.String(), tt.name)
				}
			}
		}
	}
}

// {
// 	"Query SOA Record",
// 	"example.com.",
// 	"example.com.",
// 	DNSRecordTypeSOA,
// 	`{
// 	"results": [
// 	{
// 	"name": "example.com",
// 	"default_ttl": 86400,
// 	"soa_ttl": 86400,
// 	"soa_mname": {
// 		"name": "ns1.example.com",
// 	},
// 	"soa_rname": "admin.example.com",
// 	"soa_serial": 1742759410,
// 	"soa_serial_auto": true,
// 	"soa_refresh": 43200,
// 	"soa_retry": 7200,
// 	"soa_expire": 2419200,
// 	"soa_minimum": 3600,
// 	}]}`,
// 	false,
// 	[]string{
// 		"ns1.example.com. admin.example.com. 1742759410 43200 7200 2419200 3600",
// 	},
// },
