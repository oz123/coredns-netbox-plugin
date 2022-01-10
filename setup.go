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
	"errors"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"

	"github.com/coredns/caddy"
)

var VERSION = "0.1.1-dev"

// init registers this plugin.
func init() { plugin.Register("netbox", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {

	n, err := newNetBox(c)
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

func newNetBox(c *caddy.Controller) (Netbox, error) {
	var n = Netbox{
		TTL: 3600 * time.Second,
	}

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "fallthrough":
					n.Fall.SetZonesFromArgs(c.RemainingArgs())

				case "url":
					if !c.NextArg() {
						return Netbox{}, c.ArgErr()
					}
					n.Url = c.Val()

				case "token":
					if !c.NextArg() {
						return Netbox{}, c.ArgErr()
					}
					n.Token = c.Val()

				case "localCacheDuration":
					if !c.NextArg() {
						return Netbox{}, c.ArgErr()
					}
					duration, err := time.ParseDuration(c.Val())
					if err != nil {
						return Netbox{}, c.Errf("invalid 'localCacheDuration': %s", err)
					}
					n.CacheDuration = duration
				case "ttl":
					if !c.NextArg() {
						return Netbox{}, c.ArgErr()
					}
					duration, err := time.ParseDuration(c.Val())
					if err != nil {
						return Netbox{}, c.Errf("invalid 'ttl': %s", err)
					}
					n.TTL = duration
				}

				if !c.Next() {
					break
				}
			}
		}

	}

	if n.Url == "" || n.Token == "" || n.CacheDuration == 0 {
		return Netbox{}, errors.New("Could not parse config")
	}

	log.Infof("Started plugin version %s", VERSION)
	return n, nil

}
