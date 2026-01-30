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
	"encoding/json"
	"fmt"
	"net/http"
)

type status struct {
	Apps    statusApps `json:"installed-apps"`
	Version string     `json:"netbox-version"`
}

type statusApps struct {
	DNSPlugin string `json:"netbox_dns"`
}

// Ready tests the connection to netbox and gathers version and capabilities
func (n *Netbox) Ready() bool {
	resp, err := get(n.Client, fmt.Sprintf("%s/api/status", n.Url), n.Token)
	if err != nil {
		log.Warning("HTTP request failed, check your configuration")
		return false
	}

	if resp.StatusCode != http.StatusOK {
		log.Warning(fmt.Sprintf("The server returned error code: %d", resp.StatusCode))
		return false
	}

	var s status
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&s); err != nil {
		log.Warning(fmt.Errorf("could not parse netbox status: %w", err))
		return false
	}

	n.UsePlugin = s.Apps.DNSPlugin != ""
	log.Infof("Netbox Version: %s, Netbox DNS Plugin Version: %s, Use Plugin: %t", s.Version, s.Apps.DNSPlugin, n.UsePlugin)

	return true
}
