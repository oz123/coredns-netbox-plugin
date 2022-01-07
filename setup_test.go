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
	//"fmt"

	"testing"

	"github.com/coredns/caddy"
	"github.com/stretchr/testify/assert"
)

var ()

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"minimal valid config", "netbox {\nurl http://example.org\ntoken foobar\n}\n", false},
		{"minimal valid config with zones", "netbox example.org {\nurl http://example.org\ntoken foobar\n}\n", false},
		{"minimal config with fallthrough", "netbox {\nurl http://example.org\ntoken foobar\nfallthrough\n}\n", false},
		{"empty config a", "netbox { }\n", true},
		{"empty config b", "netbox\n", true},
		{"missing url", "netbox {\nurl\ntoken foobar\n}\n", true},
		{"invalid url", "netbox {\nurl ftp://blah/something\ntoken foobar\n}\n", true},
		{"valid duration for ttl", "netbox {\nurl http://example.org\ntoken foobar\nttl 300s\n}\n", false},
		{"invalid duration for ttl", "netbox {\nurl http://example.org\ntoken foobar\nttl INVALID\n}\n", true},
	}
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		err := setup(c)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestParseNetbox(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantUrl   string
		wantToken string
		wantZones []string
	}{
		{"all zones", "netbox {\nurl http://example.org\ntoken foobar\n}\n", false, "http://example.org", "foobar", []string{"."}},
		{"specific zone", "netbox example.org {\nurl http://example.org\ntoken foobar\n}\n", false, "http://example.org", "foobar", []string{"example.org."}},
		{"multiple zones", "netbox example.org example.net {\nurl http://example.org\ntoken foobar\n}\n", false, "http://example.org", "foobar", []string{"example.org.", "example.net."}},
	}
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		n, err := parseNetbox(c)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.wantToken, n.Token, tt.name)
			assert.Equal(t, tt.wantUrl, n.Url, tt.name)
			assert.Equal(t, tt.wantZones, n.Zones, tt.name)
		}
	}

}
