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
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestQuery(t *testing.T) {
	// set up dummy Netbox
	n := newNetbox()
	n.Url = "https://example.org/api/ipam/ip-addresses"
	n.Token = "mytoken"

	tests := []struct {
		name    string
		body    string
		family  int
		wantErr bool
		want    []net.IP
	}{
		{
			"host1",
            `{"name": "example.com", "type": "A", "value": "10.0.0.2"}`,
			familyIP4,
			false,
			[]net.IP{net.ParseIP("10.0.0.2")},
		},
	}

	defer gock.Off() // Flush pending mocks after test execution

	// set up mock responses
	for _, tt := range tests {
		gock.New("https://example.org/api/ipam/ip-addresses/").Reply(
			200).BodyString(tt.body)
	}

	// run tests
	for _, tt := range tests {
		got, err := n.query(tt.name, tt.family)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.name)
		}
	}
}

