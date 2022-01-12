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

// Ready returns true if the NetBox API can be reached and returns a HTTP 200
// response using the configured url and token
func (n Netbox) Ready() bool {
	resp, err := get(n.Client, fmt.Sprintf("%s/?dns_name=%s", n.Url, "test-dns-name.example.org"), n.Token)
	if err != nil {
		return false
	}

	return resp.StatusCode == http.StatusOK
}
