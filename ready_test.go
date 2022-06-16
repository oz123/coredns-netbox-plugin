package netbox

import (
	"net/http"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestNetboxReady(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"limit": "1"}).Reply(200)

	nb := Netbox{Url: "https://example.org/api/ipam/ip-addresses", Token: "s3kr3tt0ken", Client: &http.Client{}}
	ready := nb.Ready()
	if !ready {
		t.Errorf("Expected ready %v, got %v", true, ready)
	}
}

func TestNetboxNotReady(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"limit": "1"}).Reply(403)

	nb := Netbox{Url: "https://example.org/api/ipam/ip-addresses", Token: "s3kr3tt0ken", Client: &http.Client{}}
	not_ready := nb.Ready()
	if not_ready {
		t.Errorf("Expected ready %v, got %v", false, not_ready)
	}
}
