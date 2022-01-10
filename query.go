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
	"net/http"
	"strings"
	"time"

	"github.com/imkira/go-ttlmap"
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

var localCache = ttlmap.New(nil)

func query(url, token, dns_name string, duration time.Duration, family int) (string, error) {
	item, err := localCache.Get(dns_name)
	if err == nil {
		log.Debugf("Found in local cache %s", dns_name)
		return item.Value().(string), nil
	} else {
		records := RecordsList{}
		client := &http.Client{}
		var resp *http.Response
		log.Debugf("Querying %s/?dns_name=%s", url, dns_name)
		//logger.Printf("Querying %s ", fmt.Sprintf("%s/?dns_name=%s", url, dns_name))
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/?dns_name=%s", url, dns_name), nil)
		if err != nil {
			return "", err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

		for i := 1; i <= 10; i++ {
			resp, err = client.Do(req)

			if err != nil {
				return "", err
			}

			if resp.StatusCode == http.StatusOK {
				break
			}

			time.Sleep(1 * time.Second)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("invalid HTTP resoponse code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		jsonAns := string(body)
		err = json.Unmarshal([]byte(jsonAns), &records)
		if err != nil {
			return "", err
		}

		if len(records.Records) == 0 {
			log.Debugf("recored not found response: %s", jsonAns)
			return "", nil
		}

		var ip_address string
		switch family {
		case 4:
			for _, r := range records.Records {
				if r.Family.Version == 4 {
					ip_address = strings.Split(r.Address, "/")[0]
					localCache.Set(dns_name, ttlmap.NewItem(ip_address, ttlmap.WithTTL(duration)), nil)
					break
				}
			}
		case 6:
			for _, r := range records.Records {
				if r.Family.Version == 6 {
					ip_address = strings.Split(r.Address, "/")[0]
					localCache.Set(dns_name, ttlmap.NewItem(ip_address, ttlmap.WithTTL(duration)), nil)
					break
				}
			}

		}
		return ip_address, nil
	}
}
