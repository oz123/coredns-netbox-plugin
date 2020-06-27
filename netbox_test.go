package netbox

import (
	"context"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestNetbox(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org").Get("/").Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)
	nb := Netbox{Url: "https://example.org", Token: "s3kr3tt0ken"}

	if nb.Name() != "netbox" {
		t.Errorf("expected plugin name: %s, got %s", "netbox", nb.Name())
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("my_host.", dns.TypeA)

	rcode, err := nb.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
	}
	IP := rec.Msg.Answer[0].(*dns.A).A.String()

	if IP != "10.0.0.2" {
		t.Errorf("Expected %v, got %v", "10.0.0.2", IP)
	}

}
