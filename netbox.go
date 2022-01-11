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
	"net"
	"net/http"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/fall"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("netbox")

type Netbox struct {
	Url    string
	Token  string
	TTL    uint32
	Next   plugin.Handler
	Fall   fall.F
	Zones  []string
	client *http.Client
}

// constants to match IP address family used by NetBox
const (
	familyIP4 = 4
	familyIP6 = 6
)

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		ips []net.IP
		err error
	)

	// TODO: check DNS request type here
	// https://en.wikipedia.org/wiki/List_of_DNS_record_types
	// if this is not A or AAAA return early
	state := request.Request{W: w, Req: r}

	// only handle zones we are configured to respond for
	zone := plugin.Zones(n.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	qname := state.Name()
	answers := []dns.RR{}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// handle A and AAAA records only
	switch state.QType() {
	case dns.TypeA:
		ips, err = n.query(qname, familyIP4)
		answers = a(qname, n.TTL, ips)
	case dns.TypeAAAA:
		ips, err = n.query(qname, familyIP6)
		answers = aaaa(qname, n.TTL, ips)
	}

	if err != nil {
		// fallthrough on error
		if n.Fall.Through(qname) {
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}

		// return failure here without fallthrough
		return dnserror(dns.RcodeServerFailure, state, err)
	}

	if len(answers) == 0 {
		// fallthrough on NXDOMAIN too
		if n.Fall.Through(qname) {
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}

		// return NXDOMAIN without fallthrough
		// TODO: this is not cached for some reason so subsequent requests trigger HTTP calls
		return dnserror(dns.RcodeNameError, state, nil)
	}

	// create DNS response
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	w.WriteMsg(m)

	// signal response sent back to client
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (n Netbox) Name() string { return "netbox" }

func a(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.A = ip
		answers[i] = r
	}
	return answers
}

// aaaa takes a slice of net.IPs and returns a slice of AAAA RRs.
func aaaa(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = ip
		answers[i] = r
	}
	return answers
}

// writes a DNS error response back to the client. Based on plugin.BackendError
func dnserror(rcode int, state request.Request, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative = true

	// send response
	state.W.WriteMsg(m)

	// return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
