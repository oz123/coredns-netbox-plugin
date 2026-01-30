// Copyright 2025 Lucas Kirsche <kontakt@lucas-kirsche.de>
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
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestNetboxReady(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	resp := gock.New("https://example.org/api/status").Reply(http.StatusOK)
	resp.JSON(status{
		Apps: statusApps{
			DNSPlugin: "1.2.6",
		},
		Version: "4.2.5-Docker-3.2.0",
	})

	nb := Netbox{Url: "https://example.org/api/status", Token: "s3kr3tt0ken", Client: &http.Client{}}
	ready := nb.Ready()
	if !ready {
		t.Errorf("Expected ready be %v, got %v", true, ready)
	}
}

func TestNetboxNotReady(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/status").Reply(403)

	nb := Netbox{Url: "https://example.org/api/ipam/ip-addresses", Token: "s3kr3tt0ken", Client: &http.Client{}}
	not_ready := nb.Ready()
	if not_ready {
		t.Errorf("Expected ready to be %v, got %v", false, not_ready)
	}
}
