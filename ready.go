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
)

func (n *Netbox) Ready() bool {
	resp, err := get(n.Client, fmt.Sprintf("%s/?limit=1", n.Url), n.Token)
	if err != nil {
		log.Warning("HTTP request failed, check your configuration")
		return false
	}

	if resp.StatusCode != http.StatusOK {
		log.Warning(fmt.Sprintf("The server returned error code: %d", resp.StatusCode))
		return false
	}

	return true
}
