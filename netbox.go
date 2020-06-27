package netbox

import (
	"context"
	"io"
	"net"
	"os"

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
	Url   string
	Token string
	Next  plugin.Handler
}

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	answers := []dns.RR{}
	state := request.Request{W: w, Req: r}

	ip_address := query(n.Url, n.Token, state.QName())
	// no IP is found in netbox pass processing to the next plugin
	if len(ip_address) == 0 {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}
	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
	rec.A = net.ParseIP(ip_address)
	answers = append(answers, rec)
	m := new(dns.Msg)
	m.Answer = answers
	m.SetReply(r)
	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (n Netbox) Name() string { return "netbox" }

// Make out a reference to os.Stdout so we can easily overwrite it for testing.
var out io.Writer = os.Stdout
