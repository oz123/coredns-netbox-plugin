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
	"net/http"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

func (n Netbox) Ready() bool {
	client := &http.Client{}
	var resp *http.Response

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?limit=1", n.Url), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", n.Token))

	if err != nil {
		clog.Warning("Could not setup HTTP request, check your configuration\n")
		return false
	}

	resp, err = client.Do(req)

	if err != nil {
		clog.Warning("Could HTTP request failed, check your configuration\n")
		return false
	}

	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		clog.Warning(fmt.Sprintf("The server returned error code: %d\n", resp.StatusCode))
		return false
	}
}
