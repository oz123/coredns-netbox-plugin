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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypePTR   DNSRecordType = "PTR"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeSOA   DNSRecordType = "SOA"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
)

var DNSRecordReverseMap map[DNSRecordType]uint16 = map[DNSRecordType]uint16{
	DNSRecordTypeA:     dns.TypeA,
	DNSRecordTypeAAAA:  dns.TypeAAAA,
	DNSRecordTypePTR:   dns.TypePTR,
	DNSRecordTypeCNAME: dns.TypeCNAME,
	DNSRecordTypeNS:    dns.TypeNS,
	DNSRecordTypeSOA:   dns.TypeSOA,
	DNSRecordTypeMX:    dns.TypeMX,
	DNSRecordTypeTXT:   dns.TypeTXT,
}

type DNSRecord struct {
	Type          DNSRecordType `json:"type"`
	TTL           uint32        `json:"ttl"`
	Value         string        `json:"value"`
	AbsoluteValue string        `json:"absolute_value"`
	FQDN          string        `json:"fqdn"`
}

func (r *DNSRecord) RR() dns.RR {
	var rr dns.RR
	header := dns.RR_Header{
		Name:   r.FQDN,
		Rrtype: DNSRecordReverseMap[r.Type],
		Class:  dns.ClassINET,
		Ttl:    r.TTL,
	}
	switch r.Type {
	case DNSRecordTypeA:
		rr = &dns.A{
			Hdr: header,
			A:   net.ParseIP(r.AbsoluteValue),
		}
	case DNSRecordTypeAAAA:
		rr = &dns.AAAA{
			Hdr:  header,
			AAAA: net.ParseIP(r.AbsoluteValue),
		}
	case DNSRecordTypeCNAME:
		rr = &dns.CNAME{
			Hdr:    header,
			Target: r.AbsoluteValue,
		}
	case DNSRecordTypePTR:
		rr = &dns.PTR{
			Hdr: header,
			Ptr: r.AbsoluteValue,
		}
	case DNSRecordTypeNS:
		rr = &dns.NS{
			Hdr: header,
			Ns:  r.AbsoluteValue,
		}
	case DNSRecordTypeMX:
		// we receive "[pref] [host]" from Netbox Plugin
		prefAndHost := strings.Split(r.AbsoluteValue, " ")
		if len(prefAndHost) != 2 {
			log.Error("received malformed MX record from Netbox. Abort.")
			return &dns.NULL{}
		}
		preference, err := strconv.ParseInt(prefAndHost[0], 10, 32)
		if err != nil {
			log.Errorf("can not parse int from Netbox MX record: %s", err.Error())
			return &dns.NULL{}
		}
		rr = &dns.MX{
			Hdr:        header,
			Preference: uint16(preference),
			Mx:         prefAndHost[1],
		}
	case DNSRecordTypeTXT:
		rr = &dns.TXT{
			Hdr: header,
			Txt: []string{
				r.AbsoluteValue,
			},
		}
	default:
		return &dns.NULL{}
	}
	return rr
}

type DNSRecordsList struct {
	Records []DNSRecord `json:"results"`
}

type DNSZone struct {
	Name  string `json:"name"`
	MName struct {
		Name string `json:"name"`
	} `json:"soa_mname"`
	RName   string `json:"soa_rname"`
	Serial  uint32 `json:"soa_serial"`
	Refresh uint32 `json:"soa_refresh"`
	Retry   uint32 `json:"soa_retry"`
	Expire  uint32 `json:"soa_expire"`
	Minimum uint32 `json:"soa_minimum"`
	TTL     uint32 `json:"soa_ttl"`
}

func (z *DNSZone) RR() dns.RR {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: z.Name + ".", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: z.TTL},
		Ns:      z.MName.Name + ".",
		Mbox:    z.RName + ".",
		Serial:  z.Serial,
		Expire:  z.Expire,
		Refresh: z.Refresh,
		Retry:   z.Retry,
		Minttl:  z.Minimum,
	}
}

type DNSZoneList struct {
	Zones []DNSZone `json:"results"`
}

type DNSQuerySet string

const (
	DNSQuerySetA     DNSQuerySet = "type=A&type=CNAME"
	DNSQuerySetAAAA  DNSQuerySet = "type=AAAA&type=CNAME"
	DNSQuerySetPTR   DNSQuerySet = "type=PTR"
	DNSQuerySetCNAME DNSQuerySet = "type=CNAME"
	DNSQuerySetNS    DNSQuerySet = "type=NS"
	DNSQuerySetMX    DNSQuerySet = "type=MX"
	DNSQuerySetTXT   DNSQuerySet = "type=TXT"
)

var DNSQueryReverseMap map[uint16]DNSQuerySet = map[uint16]DNSQuerySet{
	dns.TypeA:     DNSQuerySetA,
	dns.TypeAAAA:  DNSQuerySetAAAA,
	dns.TypePTR:   DNSQuerySetPTR,
	dns.TypeCNAME: DNSQuerySetCNAME,
	dns.TypeNS:    DNSQuerySetNS,
	dns.TypeMX:    DNSQuerySetMX,
	dns.TypeTXT:   DNSQuerySetTXT,
}

func (n *Netbox) queryRecord(zone string, fqdn string, querySet DNSQuerySet) ([]DNSRecord, error) {
	var (
		requrl  = fmt.Sprintf("%s/api/plugins/netbox-dns/records/?zone=%s&fqdn=%s&active=true&%s", n.Url, strings.TrimRight(zone, "."), fqdn, querySet)
		records DNSRecordsList
	)

	// do http request against NetBox instance
	resp, err := get(n.Client, requrl, n.Token)
	if err != nil {
		return records.Records, fmt.Errorf("problem performing request: %w", err)
	}
	// ensure body is closed once we are done
	defer resp.Body.Close()

	// status code must be http.StatusOK
	if resp.StatusCode != http.StatusOK {
		return records.Records, fmt.Errorf("bad HTTP response code: %d", resp.StatusCode)
	}

	// read and parse response body
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&records); err != nil {
		return records.Records, fmt.Errorf("could not unmarshal response: %w", err)
	}

	return records.Records, nil
}

func (n *Netbox) queryZone(zone string) ([]DNSZone, error) {
	var (
		requrl = fmt.Sprintf("%s/api/plugins/netbox-dns/zones/?name=%s&active=true", n.Url, strings.TrimSuffix(zone, "."))
		zones  DNSZoneList
	)

	// do http request against NetBox instance
	resp, err := get(n.Client, requrl, n.Token)
	if err != nil {
		return zones.Zones, fmt.Errorf("problem performing request: %w", err)
	}
	// ensure body is closed once we are done
	defer resp.Body.Close()

	// status code must be http.StatusOK
	if resp.StatusCode != http.StatusOK {
		return zones.Zones, fmt.Errorf("bad HTTP response code: %d", resp.StatusCode)
	}

	// read and parse response body
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&zones); err != nil {
		return zones.Zones, fmt.Errorf("could not unmarshal response: %w", err)
	}

	return zones.Zones, nil
}
