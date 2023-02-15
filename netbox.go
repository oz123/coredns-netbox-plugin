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
	"strings"
	"time"

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
	Next   plugin.Handler
	TTL    time.Duration
	Fall   fall.F
	Zones  []string
	Client *http.Client
}

// constants to match IP address family used by NetBox
const (
	familyIP4 = 4
	familyIP6 = 6
)

// ServeDNS implements the plugin.Handler interface
func (n *Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		ips     []net.IP
		domains []string
		err     error
	)

	state := request.Request{W: w, Req: r}

	// only handle zones we are configured to respond for
	zone := plugin.Zones(n.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	qname := state.Name()

	// check record type here and bail out if not A, AAAA or PTR
	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA && state.QType() != dns.TypePTR {
		// always fallthrough if configured
		if n.Fall.Through(qname) {
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}

		// otherwise return SERVFAIL here without fallthrough
		return dnserror(dns.RcodeServerFailure, state, err)
	}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	answers := []dns.RR{}

	// handle A and AAAA records only
	switch state.QType() {
	case dns.TypeA:
		ips, err = n.query(strings.TrimRight(qname, "."), familyIP4)
		answers = a(qname, uint32(n.TTL), ips)
	case dns.TypeAAAA:
		ips, err = n.query(strings.TrimRight(qname, "."), familyIP6)
		answers = aaaa(qname, uint32(n.TTL), ips)
	case dns.TypePTR:
		domains, err = n.queryreverse(qname)
		answers = ptr(qname, uint32(n.TTL), domains)
	}

	if len(answers) == 0 {
		// always fallthrough if configured
		if n.Fall.Through(qname) {
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}

		if err != nil {
			// return SERVFAIL here without fallthrough
			return dnserror(dns.RcodeServerFailure, state, err)
		}

		// otherwise return NXDOMAIN
		return dnserror(dns.RcodeNameError, state, nil)
	}

	// create DNS response
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	// send response back to client
	_ = w.WriteMsg(m)

	// signal response sent back to client
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (n *Netbox) Name() string { return "netbox" }

// a takes a slice of net.IPs and returns a slice of A RRs.
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

// ptr takes a slice of strings and returns a slice of PTR RRs.
func ptr(zone string, ttl uint32, domains []string) []dns.RR {

	answers := make([]dns.RR, len(domains))
	for i, domain := range domains {
		r := new(dns.PTR)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: ttl}
		r.Ptr = domain
		answers[i] = r
	}

	return answers
}

// dnserror writes a DNS error response back to the client. Based on plugin.BackendError
func dnserror(rcode int, state request.Request, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative = true

	// send response
	_ = state.W.WriteMsg(m)

	// return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
