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
	"net/http"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	ctls "github.com/coredns/coredns/plugin/pkg/tls"

	"github.com/coredns/caddy"
)

var VERSION = "0.5.0"

const (
	defaultTTL     = time.Second * 3600 // 3600s
	defaultTimeout = time.Second * 5    // 5s
)

// init registers this plugin.
func init() { plugin.Register("netbox", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {

	// parse config block in Corefile
	n, err := parseNetbox(c)
	if err != nil {
		return plugin.Error("netbox", err)
	}

	log.Infof("Version %s", VERSION)

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

// newNetbox returns a basic *Netbox type with some defaults set
func newNetbox() *Netbox {
	return &Netbox{
		TTL:   defaultTTL,
		Zones: []string{"."},
		Client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// parseNetbox handles parsing of the plugins config
func parseNetbox(c *caddy.Controller) (*Netbox, error) {
	n := newNetbox()
	i := 0
	for c.Next() {
		// ensure plugin is only included once in each block
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++

		// handle netbox [zones...]
		zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
		if len(zones) > 0 {
			n.Zones = zones
		}
		log.Infof("Responsible zones: %v", n.Zones)

		// parse inside block
		for c.NextBlock() {
			switch c.Val() {
			case "fallthrough":
				n.Fall.SetZonesFromArgs(c.RemainingArgs())

			case "url":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				n.Url = c.Val()

			case "token":
				if !c.NextArg() {
					return n, c.ArgErr()
				}
				n.Token = c.Val()

			case "tls":
				args := c.RemainingArgs()
				tlsConfig, err := ctls.NewTLSConfigFromArgs(args...)
				if err != nil {
					return n, err
				}

				// add custom transport to client
				n.Client.Transport = &http.Transport{
					TLSClientConfig: tlsConfig,
				}

			case "ttl":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				duration, err := time.ParseDuration(c.Val())
				if err != nil {
					return n, c.Errf("could not parse 'ttl': %s", err)
				}
				n.TTL = duration

			case "timeout":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				duration, err := time.ParseDuration(c.Val())
				if err != nil {
					return n, c.Errf("could not parse 'timeout': %s", err)
				}
				n.Client.Timeout = duration

			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}

	// fail if url or token are not set
	if n.Url == "" || n.Token == "" {
		return nil, c.Err("Invalid config")
	}

	if !n.Ready() {
		return nil, c.Err("Netbox not reachable")
	}

	return n, nil
}
