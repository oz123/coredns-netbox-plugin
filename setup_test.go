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
	"net/http"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/stretchr/testify/assert"
)

// TestParseNetbox tests the various things that should be parsed by setup.
func TestParseNetbox(t *testing.T) {
	// tests to run
	tests := []struct {
		msg     string
		input   string
		wantErr bool
		want    *Netbox
	}{
		{
			"minimal valid config",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"."},
				Client:        &http.Client{},
			},
		},
		{
			"minimal valid config with zone",
			"netbox example.org {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"example.org."},
				Client:        &http.Client{},
			},
		},
		{
			"minimal valid config with two zones",
			"netbox example.org example.net {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"example.org.", "example.net."},
				Client:        &http.Client{},
			},
		},
		{
			"empty config",
			"netbox {}\n",
			true,
			nil,
		},
		{
			"empty config with zone",
			"netbox example.org {}\n",
			true,
			nil,
		},
		{
			"invalid localCacheDuration",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration Wrong\n}\n",
			true,
			nil,
		},
		{
			"config with ttl",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nttl 1800s\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           time.Second * 1800,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"."},
				Client:        &http.Client{},
			},
		},
		{
			"config with invalid ttl",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nttl INVALID\n}\n",
			true,
			nil,
		},
		{
			"config with fallthrough (all)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"."},
				Fall:          fall.F{Zones: []string{"."}},
				Client:        &http.Client{},
			},
		},
		{
			"config with fallthrough (one domain)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough example.org\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"."},
				Fall:          fall.F{Zones: []string{"example.org."}},
				Client:        &http.Client{},
			},
		},
		{
			"config with fallthrough (multiple domains)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\nfallthrough example.org example.net\n}\n",
			false,
			&Netbox{
				Url:           "example.org",
				Token:         "foobar",
				TTL:           defaultTTL,
				CacheDuration: time.Second * 10,
				Next:          plugin.Handler(nil),
				Zones:         []string{"."},
				Fall:          fall.F{Zones: []string{"example.org.", "example.net."}},
				Client:        &http.Client{},
			},
		},
	}

	// run tests
	for _, tt := range tests {
		c := caddy.NewTestController("dns", tt.input)
		got, err := parseNetbox(c)
		if tt.wantErr {
			assert.Error(t, err, tt.msg)
		} else {
			assert.Nil(t, err, tt.msg)
			assert.Equal(t, tt.want, got, tt.msg)
		}
	}
}
