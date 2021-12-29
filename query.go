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
	"encoding/json"
	"fmt"
	"io/ioutil"
	logger "log"
	"net/http"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/imkira/go-ttlmap"
)

type Record struct {
	Family   Family `json:"faily"`
	Address  string `json:"address"`
	HostName string `json:"dns_name,omitempty"`
}

type Family struct {
	Protocol int    `json:"value"`
	Label    string `json:"label"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

// TODO: deal with multiple IP addresses for the same dns name:
//{
//  "count": 2,
//  "next": null,
//  "previous": null,
//  "results": [
//    {
//      "id": 7,
//      "url": "http://localhost:8000/api/ipam/ip-addresses/7/",
//      "display": "10.0.0.1/25",
//      "family": {
//        "value": 4,
//        "label": "IPv4"
//      },
//      "address": "10.0.0.1/25",
//      "vrf": null,
//      "tenant": null,
//      "status": {
//        "value": "active",
//        "label": "Active"
//      },
//      "role": null,
//      "assigned_object_type": null,
//      "assigned_object_id": null,
//      "assigned_object": null,
//      "nat_inside": null,
//      "nat_outside": null,
//      "dns_name": "mail.foo.com",
//      "description": "",
//      "tags": [],
//      "custom_fields": {},
//      "created": "2021-12-28",
//      "last_updated": "2021-12-28T21:36:39.883703Z"
//    },
//    {
//      "id": 6,
//      "url": "http://localhost:8000/api/ipam/ip-addresses/6/",
//      "display": "fe80::250:56ff:fe3d:83af/64",
//      "family": {
//        "value": 6,
//        "label": "IPv6"
//      },
//      "address": "fe80::250:56ff:fe3d:83af/64",
//      "vrf": null,
//      "tenant": null,
//      "status": {
//        "value": "active",
//        "label": "Active"
//      },
//      "role": null,
//      "assigned_object_type": null,
//      "assigned_object_id": null,
//      "assigned_object": null,
//      "nat_inside": null,
//      "nat_outside": null,
//      "dns_name": "mail.foo.com",
//      "description": "",
//      "tags": [],
//      "custom_fields": {},
//      "created": "2021-12-28",
//      "last_updated": "2021-12-28T14:42:05.160608Z"
//    }
//  ]
//}
//
var localCache = ttlmap.New(nil)

func query(url, token, dns_name string, duration time.Duration, protocol int) string {
	item, err := localCache.Get(dns_name)
	if err == nil {
		clog.Debug(fmt.Sprintf("Found in local cache %s", dns_name))
		logger.Printf("Found in local cache %s", dns_name)
		return item.Value().(string)
	} else {
		records := RecordsList{}
		client := &http.Client{}
		var resp *http.Response
		clog.Debug("Querying ", fmt.Sprintf("%s/?dns_name=%s", url, dns_name))
		logger.Printf("Querying %s ", fmt.Sprintf("%s/?dns_name=%s", url, dns_name))
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/?dns_name=%s", url, dns_name), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

		for i := 1; i <= 10; i++ {
			resp, err = client.Do(req)

			if err != nil {
				clog.Fatalf("HTTP Error %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				break
			}

			time.Sleep(1 * time.Second)
		}

		if resp.StatusCode != http.StatusOK {
			return ""
		}

		body, err := ioutil.ReadAll(resp.Body)
		logger.Printf("%s", body)
		if err != nil {
			clog.Fatalf("Error reading body %v", err)
		}

		jsonAns := string(body)
		err = json.Unmarshal([]byte(jsonAns), &records)
		if err != nil {
			clog.Fatalf("could not unmarshal response %v", err)
		}

		if len(records.Records) == 0 {
			clog.Info("Recored not found in", jsonAns)
			return ""
		}

		ip_address := strings.Split(records.Records[0].Address, "/")[0]
		localCache.Set(dns_name, ttlmap.NewItem(ip_address, ttlmap.WithTTL(duration)), nil)
		return ip_address
	}
}
