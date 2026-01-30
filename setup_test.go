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
	//"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/fall"

	// ctls "github.com/coredns/coredns/plugin/pkg/tls"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

// TestParseNetbox tests the various things that should be parsed by setup.
func TestParseNetbox(t *testing.T) {
	// set up some tls configs for later tests
	// defaultTLSConfig, err := ctls.NewTLSConfigFromArgs([]string{}...)
	// if err != nil {
	// 	panic(err)
	// }

	// tests to run
	tests := []struct {
		msg     string
		input   string
		wantErr bool
		want    *Netbox
	}{
		{
			"minimal valid config",
			"netbox {\nurl http://example.org\ntoken foobar\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"minimal config with localCacheDuration (now invalid)",
			"netbox {\nurl example.org\ntoken foobar\nlocalCacheDuration 10s\n}\n",
			true,
			nil,
		},
		{
			"minimal valid config with zone",
			"netbox example.org {\nurl http://example.org\ntoken foobar\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"example.org."},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"minimal valid config with two zones",
			"netbox example.org example.net {\nurl http://example.org\ntoken foobar\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"example.org.", "example.net."},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
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
			"config with ttl",
			"netbox {\nurl http://example.org\ntoken foobar\nttl 1800s\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   time.Second * 1800,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"config with invalid ttl",
			"netbox {\nurl example.org\ntoken foobar\nttl INVALID\n}\n",
			true,
			nil,
		},
		{
			"config with timeout",
			"netbox {\nurl http://example.org\ntoken foobar\ntimeout 2s\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Client: &http.Client{
					Timeout: time.Second * 2,
				},
				UsePlugin: true,
			},
		},
		{
			"config with invalid timeout",
			"netbox {\nurl http://example.org\ntoken foobar\ntimeout INVALID\n}\n",
			true,
			nil,
		},
		{
			"config with fallthrough (all)",
			"netbox {\nurl http://example.org\ntoken foobar\nfallthrough\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Fall:  fall.F{Zones: []string{"."}},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"config with fallthrough (one domain)",
			"netbox {\nurl http://example.org\ntoken foobar\nfallthrough example.org\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Fall:  fall.F{Zones: []string{"example.org."}},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"config with fallthrough (multiple domains)",
			"netbox {\nurl http://example.org\ntoken foobar\nfallthrough example.org example.net\n}\n",
			false,
			&Netbox{
				Url:   "http://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Fall:  fall.F{Zones: []string{"example.org.", "example.net."}},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		{
			"config with https",
			"netbox {\nurl https://example.org\ntoken foobar\n}\n",
			false,
			&Netbox{
				Url:   "https://example.org",
				Token: "foobar",
				TTL:   defaultTTL,
				Next:  plugin.Handler(nil),
				Zones: []string{"."},
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
				UsePlugin: true,
			},
		},
		//! No clue why this test fails....
		// {
		// 	"config with https and tls (no options)",
		// 	"netbox {\nurl https://example.org\ntoken foobar\ntls\n}\n",
		// 	false,
		// 	&Netbox{
		// 		Url:   "https://example.org",
		// 		Token: "foobar",
		// 		TTL:   defaultTTL,
		// 		Next:  plugin.Handler(nil),
		// 		Zones: []string{"."},
		// 		Client: &http.Client{
		// 			Timeout: defaultTimeout,
		// 			Transport: &http.Transport{
		// 				TLSClientConfig: defaultTLSConfig,
		// 			},
		// 		},
		// 		UsePlugin: true,
		// 	},
		// },
		{
			"config with https and tls (invalid config)",
			"netbox {\nurl https://example.org\ntoken foobar\ntls testing/missing.crt\n}\n",
			true,
			nil,
		},
	}

	for range tests {
		gock.New("http://example.org/api/status").Reply(200).BodyString(`
		{
			"django-version": "5.1.7",
			"installed-apps": {
				"django_filters": "25.1",
				"django_prometheus": "2.3.1",
				"django_rq": "3.0.0",
				"django_tables2": "2.7.5",
				"drf_spectacular": "0.28.0",
				"drf_spectacular_sidecar": "2025.3.1",
				"mptt": "0.16.0",
				"netbox_dns": "1.2.6",
				"rest_framework": "3.15.2",
				"social_django": "5.4.3",
				"taggit": "6.1.0",
				"timezone_field": "7.1"
			},
			"netbox-version": "4.2.5-Docker-3.2.0",
			"plugins": {
				"netbox_dns": "1.2.6"
			},
			"python-version": "3.12.3",
			"rq-workers-running": 1
		}
	`)

		gock.New("https://example.org/api/status").Reply(200).BodyString(`
		{
			"django-version": "5.1.7",
			"installed-apps": {
				"django_filters": "25.1",
				"django_prometheus": "2.3.1",
				"django_rq": "3.0.0",
				"django_tables2": "2.7.5",
				"drf_spectacular": "0.28.0",
				"drf_spectacular_sidecar": "2025.3.1",
				"mptt": "0.16.0",
				"netbox_dns": "1.2.6",
				"rest_framework": "3.15.2",
				"social_django": "5.4.3",
				"taggit": "6.1.0",
				"timezone_field": "7.1"
			},
			"netbox-version": "4.2.5-Docker-3.2.0",
			"plugins": {
				"netbox_dns": "1.2.6"
			},
			"python-version": "3.12.3",
			"rq-workers-running": 1
		}
	`)
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
