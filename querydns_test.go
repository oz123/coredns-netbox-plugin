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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestQueryRecord(t *testing.T) {
	n := newNetbox()
	n.Url = "https://example.org"
	n.Token = "123456789"

	tests := []struct {
		name    string
		zone    string
		fqdn    string
		rType   DNSRecordType
		body    string
		wantErr bool
		want    []string
	}{
		{
			"Query A Record No TTL",
			"example.org.",
			"mail1.example.org.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "A",
					"ttl": null,
					"value": "192.168.0.1",
					"absolute_value": "192.168.0.1",
					"fqdn": "mail1.example.org."
				}]
			}`,
			false,
			[]string{
				"mail1.example.org.\t0\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query A Record With TTL",
			"example.org.",
			"mail1.example.org.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "A",
					"ttl": 8600,
					"value": "192.168.0.1",
					"absolute_value": "192.168.0.1",
					"fqdn": "mail1.example.org."
				}]
			}`,
			false,
			[]string{
				"mail1.example.org.\t8600\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query AAAA Record",
			"example.org.",
			"mail1.example.org.",
			DNSRecordTypeAAAA,
			`{
				"results": [
				{
					"type": "AAAA",
					"ttl": 8600,
					"value": "2001:db8::1",
					"absolute_value": "2001:db8::1",
					"fqdn": "mail1.example.org."
				}]
			}`,
			false,
			[]string{
				"mail1.example.org.\t8600\tIN\tAAAA\t2001:db8::1",
			},
		},
		{
			"Query CNAME Record",
			"example.org.",
			"test.example.org.",
			DNSRecordTypeCNAME,
			`{
				"results": [
				{
					"type": "CNAME",
					"ttl": 8600,
					"value": "mail1",
					"absolute_value": "mail1.example.org.",
					"fqdn": "test.example.org."
				}]
			}`,
			false,
			[]string{
				"test.example.org.\t8600\tIN\tCNAME\tmail1.example.org.",
			},
		},
		{
			"Query A Record but receive CNAME Record with resolved A record",
			"example.org.",
			"test.example.org.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "CNAME",
					"ttl": 8600,
					"value": "mail1",
					"absolute_value": "mail1.example.org.",
					"fqdn": "test.example.org."
				}]
			}`,
			false,
			[]string{
				"test.example.org.\t8600\tIN\tCNAME\tmail1.example.org.",
				"mail1.example.org.\t8600\tIN\tA\t192.168.0.1",
			},
		},
		{
			"Query MX Record",
			"example.org.",
			"example.org.",
			DNSRecordTypeMX,
			`{
				"results": [
				{
					"type": "MX",
					"ttl": 8600,
					"value": "10 mail1",
					"absolute_value": "10 mail1.example.org.",
					"fqdn": "example.org."
				}]
			}`,
			false,
			[]string{
				"example.org.\t8600\tIN\tMX\t10 mail1.example.org.",
			},
		},
		{
			"Query PTR v4 Record",
			"0.168.192.in-addr.arpa.",
			"1.0.168.192.in-addr.arpa.",
			DNSRecordTypePTR,
			`{
				"results": [
				{
					"type": "PTR",
					"ttl": 8600,
					"value": "mail1.example.org.",
					"absolute_value": "mail1.example.org.",
					"fqdn": "1.0.168.192.in-addr.arpa."
				}]
			}`,
			false,
			[]string{
				"1.0.168.192.in-addr.arpa.\t8600\tIN\tPTR\tmail1.example.org.",
			},
		},
		{
			"Query unsupported query type",
			"example.org.",
			"test.example.org.",
			DNSRecordType("BBBB"),
			`{
				"type": [
					"Select a valid choice. BBBB is not one of the available choices."
				]
			}`,
			false,
			[]string{},
		},
		{
			"Query TXT record",
			"example.org.",
			"example.org.",
			DNSRecordTypeTXT,
			`{
				"results": [
				{
					"type": "TXT",
					"ttl": 8600,
					"value": "TEST ENTRY",
					"absolute_value": "TEST ENTRY",
					"fqdn": "example.org."
				}]
			}`,
			false,
			[]string{
				"example.org.\t8600\tIN\tTXT\t\"TEST ENTRY\"",
			},
		},
		{
			"Query NS record",
			"example.org.",
			"example.org.",
			DNSRecordTypeNS,
			`{
				"results": [
				{
					"type": "NS",
					"ttl": 8600,
					"value": "ns1.example.org.",
					"absolute_value": "ns1.example.org.",
					"fqdn": "example.org."
				}]
			}`,
			false,
			[]string{
				"example.org.\t8600\tIN\tNS\tns1.example.org.",
			},
		},
		{
			"Query not existing record",
			"example.org.",
			"test2.example.org.",
			DNSRecordTypeA,
			`{
				"results": [
				{
					"type": "PTR",
					"ttl": 8600,
					"value": "mail1.example.org.",
					"absolute_value": "mail1.example.org.",
					"fqdn": "1.0.168.192.in-addr.arpa."
				}]
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
				"type":   string(tt.rType),
			}).Reply(
			200).BodyString(tt.body)
	}

	for _, tt := range tests {
		responses, err := n.queryRecord(tt.zone, tt.fqdn, DNSQuerySet(fmt.Sprintf("type=%s", tt.rType)))
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			if len(tt.want) != 0 && assert.NotEmpty(t, responses, tt.name) {
				for i, response := range responses {
					assert.Equal(t, tt.want[i], response.RR().String(), tt.name)
				}
			}
		}
	}
}

func TestQueryZone(t *testing.T) {
	n := newNetbox()
	n.Url = "https://example.org"
	n.Token = "123456789"

	tests := []struct {
		name    string
		zone    string
		body    string
		wantErr bool
		want    []string
	}{
		{
			"Query A Record No TTL",
			"example.org.",
			`{
				"results": [
				{
					"name": "example.org",
					"soa_ttl": 86400,
					"soa_mname": {
						"name": "ns1.example.org"
					},
					"soa_rname": "admin.example.org",
					"soa_serial": 1742857987,
					"soa_refresh": 43200,
					"soa_retry": 7200,
					"soa_expire": 2419200,
					"soa_minimum": 3600
				}]
			}`,
			false,
			[]string{
				"example.org.\t86400\tIN\tSOA\tns1.example.org. admin.example.org. 1742857987 43200 7200 2419200 3600",
			},
		},
	}

	// set up mock responses
	for _, tt := range tests {
		gock.New("https://example.org/api/plugins/netbox-dns/zones/").MatchParams(
			map[string]string{
				"name":   strings.TrimRight(tt.zone, "."),
				"active": "true",
			}).Reply(
			200).BodyString(tt.body)
	}

	for _, tt := range tests {
		responses, err := n.queryZone(tt.zone)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			if len(tt.want) != 0 && assert.NotEmpty(t, responses, tt.name) {
				for i, response := range responses {
					assert.Equal(t, tt.want[i], response.RR().String(), tt.name)
				}
			}
		}
	}
}
