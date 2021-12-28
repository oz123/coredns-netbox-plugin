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
	"context"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("netbox")

type Netbox struct {
	Url           string
	Token         string
	CacheDuration time.Duration
	Next          plugin.Handler
}

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	answers := []dns.RR{}
	// TODO: check DNS request type here
	// https://en.wikipedia.org/wiki/List_of_DNS_record_types
	// if this is not A or AAAA return early
	state := request.Request{W: w, Req: r}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	switch state.QType() {
	case dns.TypeA:
		ip_address := query(n.Url, n.Token, strings.TrimRight(state.QName(), "."), n.CacheDuration, 4)
		// no IP is found in netbox pass processing to the next plugin
		if len(ip_address) == 0 {
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}
		rec := a(state, ip_address, 3600)
		answers = append(answers, rec)
		m := new(dns.Msg)
		m.Answer = answers
		m.SetReply(r)
		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	// TODO: add querying for AAAA records
	case dns.TypeAAAA:
		ip_address := query(n.Url, n.Token, strings.TrimRight(state.QName(), "."), n.CacheDuration, 6)
		rec := a6(state, ip_address, 3600)
		answers = append(answers, rec)
		m := new(dns.Msg)
		m.Answer = answers
		m.SetReply(r)
		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	default:
		// TODO: is dns.RcodeServerFailure the correct answer?
		// maybe dns.RcodeNotImplemented is more appropriate?
		return dns.RcodeServerFailure, nil
	}

}

// Name implements the Handler interface.
func (n Netbox) Name() string { return "netbox" }

func a(state request.Request, ip_addr string, ttl uint32) *dns.A {
	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
	rec.A = net.ParseIP(ip_addr)
	return rec
}

func a6(state request.Request, ip_addr string, ttl uint32) *dns.A {
	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
	rec.A = net.ParseIP(ip_addr)
	return rec
}

// Make out a reference to os.Stdout so we can easily overwrite it for testing.
var out io.Writer = os.Stdout
