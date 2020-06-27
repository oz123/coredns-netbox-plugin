package netbox

import (
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestQuery(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org").Get("/").Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	want := "10.0.0.2"
	got := query("https://example.org", "mytoken", "myhost")
	if got != want {
		t.Fatalf("Expected true")
	}
}
