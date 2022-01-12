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
	"net"
	"net/http"
	"strings"
)

type Record struct {
	Family   Family `json:"family"`
	Address  string `json:"address"`
	HostName string `json:"dns_name,omitempty"`
}

type Family struct {
	Version int    `json:"value"`
	Label   string `json:"label"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

func get(client *http.Client, url, token string) (*http.Response, error) {
	// handle if provided client was not set up
	if client == nil {
		return nil, fmt.Errorf("provided *http.Client was invalid")
	}

	// set up HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// set authorization header for request to NetBox
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	// do request
	return client.Do(req)
}

func (n *Netbox) query(host string, family int) ([]net.IP, error) {
	var (
		dns_name = strings.TrimSuffix(host, ".")
		requrl   = fmt.Sprintf("%s/?dns_name=%s", n.Url, dns_name)
		records  RecordsList
	)

	addresses := make([]net.IP, 0)

	qtype := "A"
	if family == familyIP6 {
		qtype = "AAAA"
	}
	log.Debugf("Querying %s for %s", requrl, qtype)

	// do http request against NetBox instance
	resp, err := get(n.Client, requrl, n.Token)
	if err != nil {
		return addresses, fmt.Errorf("Problem performing request: %w", err)
	}

	// ensure body is closed once we are done
	defer resp.Body.Close()

	// status code must be http.StatusOK
	if resp.StatusCode != http.StatusOK {
		return addresses, fmt.Errorf("Bad HTTP response code: %d", resp.StatusCode)
	}

	// read and parse response body
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&records); err != nil {
		return addresses, fmt.Errorf("Could not unmarshal response: %w", err)
	}

	if len(records.Records) == 0 {
		return addresses, nil
	}

	// grab returned address of specified address family
	for _, r := range records.Records {
		if r.Family.Version == family {
			if addr := net.ParseIP(strings.Split(r.Address, "/")[0]); addr != nil {
				addresses = append(addresses, addr)
			}
		}
	}

	log.Debugf("Got %d %s responses from %s", len(addresses), qtype, requrl)

	return addresses, nil
}
