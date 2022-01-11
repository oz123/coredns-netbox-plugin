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
	"net/http"
	"net/url"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"

	"github.com/coredns/caddy"
)

var VERSION = "0.1.1-dev"

// init registers this plugin.
func init() { plugin.Register("netbox", setup) }

// setup is the function that gets called when the config parser see the token "netbox". Setup is responsible
// for parsing any extra options the netbox plugin may have. The first token this function sees is "netbox".
func setup(c *caddy.Controller) error {

	n, err := parseNetbox(c)
	if err != nil {
		return plugin.Error("netbox", err)
	}

	// Add a startup function that will -- after all plugins have been loaded -- check if the
	// prometheus plugin has been used - if so we will export metrics. We can only register
	// this metric once, hence the "once.Do".
	c.OnStartup(func() error {
		once.Do(func() {
			m := dnsserver.GetConfig(c).Handler("prometheus")
			if m == nil {
				return
			}
			if x, ok := m.(*metrics.Metrics); ok {
				x.MustRegister(requestCount)
			}
		})
		return nil
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		n.Next = next
		return n
	})

	// All OK, return a nil error.
	return nil
}

func newNetbox() *Netbox {
	return &Netbox{
		TTL:   3600,
		Zones: []string{"."},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func parseNetbox(c *caddy.Controller) (*Netbox, error) {
	var client *http.Client

	n := newNetbox()
	i := 0
	for c.Next() {
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++

		// netbox [zones...]
		zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
		if len(zones) == 0 {
			zones = []string{"."}
		}
		n.Zones = zones

		// parse inside block
		for c.NextBlock() {
			switch c.Val() {
			case "fallthrough":
				n.Fall.SetZonesFromArgs(c.RemainingArgs())

			case "url":
				if !c.NextArg() {
					return n, c.ArgErr()
				}
				apiurl, err := url.Parse(c.Val())
				if err != nil {
					return n, c.Errf("could not parse 'url': %s", err)
				}
				if apiurl.Scheme != "http" && apiurl.Scheme != "https" {
					return n, c.Errf("unsupported scheme for 'url': %s", apiurl.String())
				}
				n.Url = apiurl.String()

			case "timeout":
				if !c.NextArg() {
					return n, c.ArgErr()
				}
				duration, err := time.ParseDuration(c.Val())
				if err != nil {
					return n, c.Errf("could not parse 'timeout': %s", err)
				}

				// set timeout for http client
				client.Timeout = duration

			case "token":
				if !c.NextArg() {
					return n, c.ArgErr()
				}
				n.Token = c.Val()

			case "ttl":
				if !c.NextArg() {
					return n, c.ArgErr()
				}
				duration, err := time.ParseDuration(c.Val())
				if err != nil {
					return n, c.Errf("could not parse 'ttl': %s", err)
				}
				n.TTL = uint32(duration)

			default:
				return n, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}

	// fail if either url or token are not set
	if n.Url == "" || n.Token == "" {
		return n, c.Err("Invalid config")
	}

	// set http client for plugin
	n.client = client

	log.Infof("Version %s", VERSION)

	return n, nil
}
