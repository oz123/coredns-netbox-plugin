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
	Url           string
	Token         string
	CacheDuration time.Duration
	Next          plugin.Handler
	TTL           time.Duration
	Fall          fall.F
	Zones         []string
}

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	// TODO: check DNS request type here
	// https://en.wikipedia.org/wiki/List_of_DNS_record_types
	// if this is not A or AAAA return early
	state := request.Request{W: w, Req: r}

	// only handle zones we are configured to respond for
	zone := plugin.Zones(n.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	var (
		ip_address string
		record4    *dns.A
		record6    *dns.AAAA
		err        error
	)

	qname := state.Name()

	switch state.QType() {

	case dns.TypeA:
		ip_address, err = query(n.Url, n.Token, strings.TrimRight(qname, "."), n.CacheDuration, 4)
		// no IP is found in netbox pass processing to the next plugin
		record4 = a(state, ip_address, uint32(n.TTL))
	case dns.TypeAAAA:
		ip_address, err = query(n.Url, n.Token, strings.TrimRight(qname, "."), n.CacheDuration, 6)
		record6 = a6(state, ip_address, uint32(n.TTL))
	}

	if len(ip_address) == 0 {
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

	writeDNSAnswer(record4, record6, w, r)
	return dns.RcodeSuccess, nil

}

// Name implements the Handler interface.
func (n Netbox) Name() string { return "netbox" }

// this is probably a bad way of doing things if we want to support other types too
func writeDNSAnswer(record4 *dns.A, record6 *dns.AAAA, w dns.ResponseWriter, r *dns.Msg) {

	answers := []dns.RR{}

	if record4 != nil {
		answers = append(answers, record4)
	}

	if record6 != nil {
		answers = append(answers, record6)
	}

	m := new(dns.Msg)
	m.Answer = answers
	m.SetReply(r)
	_ = w.WriteMsg(m)
}

func a(state request.Request, ip_addr string, ttl uint32) *dns.A {
	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
	rec.A = net.ParseIP(ip_addr)
	return rec
}

func a6(state request.Request, ip_addr string, ttl uint32) *dns.AAAA {
	rec := new(dns.AAAA)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
	rec.AAAA = net.ParseIP(ip_addr)
	return rec
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
